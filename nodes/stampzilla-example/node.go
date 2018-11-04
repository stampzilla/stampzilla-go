package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
)

type Node struct {
	Client    *WebsocketClient
	Cancel    context.CancelFunc
	wg        *sync.WaitGroup
	stopRetry chan struct{}
	Config    *models.Config
	X509      *x509.Certificate
	TLS       *tls.Certificate
	CA        *x509.CertPool
}

func NewNode(client *WebsocketClient) *Node {
	return &Node{
		Client:    client,
		wg:        &sync.WaitGroup{},
		stopRetry: make(chan struct{}),
	}
}
func (n *Node) Wait() {
	n.wg.Wait()
}

func (n *Node) Connect() error {
	logrus.SetReportCaller(true)
	//Make sure we have a config
	n.Config = &models.Config{}
	n.Config.MustLoad()

	n.Config.Save("config.json")

	err := n.LoadCertificateKeyPair("crt")

	if err != nil {
		u := fmt.Sprintf("ws://%s:%s/ws", n.Config.Host, n.Config.Port)
		err = n.ConnectWithRetry(u)
		if err != nil {
			return err
		}

		// {"fromUUID":"","type":"server-info","body":{"name":"stampzilla server","uuid":"2beec593-758a-4d48-b39a-13b36a760389","tlsPort":"6443","port":"8080"}}
		// wait for server info so we can update our config

		serverInfo := &models.ServerInfo{}
		err = n.Client.WaitForMessage("server-info", serverInfo)
		if err != nil {
			return err
		}
		n.Config.Port = serverInfo.Port
		n.Config.TLSPort = serverInfo.TLSPort
		n.Config.Save("config.json")

		csr, err := n.GenerateCSR()
		if err != nil {
			return err
		}

		msg, err := models.NewMessage("certificate-signing-request", string(csr))
		if err != nil {
			return err
		}
		n.Client.WriteJSON(msg)

		// wait for our new certificate

		var rawCert string
		err = n.Client.WaitForMessage("approved-certificate-signing-request", &rawCert)

		err = ioutil.WriteFile("crt.crt", []byte(rawCert), 0644)
		if err != nil {
			return err
		}

		var caCert string
		err = n.Client.WaitForMessage("certificate-authority", &caCert)

		err = ioutil.WriteFile("ca.crt", []byte(caCert), 0644)
		if err != nil {
			return err
		}

		return err
	}

	//If we have certificate we can connect to TLS immedietly
	// connect tls
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*n.TLS},
		RootCAs:      n.CA,
	}

	n.Client.TLSClientConfig = tlsConfig

	u := fmt.Sprintf("wss://%s:%s/ws", n.Config.Host, n.Config.TLSPort)
	err = n.ConnectWithRetry(u)
	if err != nil {
		logrus.Error(err)
	}
	// if we dont have tls we must request certificate over http

	n.Wait()
	n.Client.Wait()

	return nil
}

func (n *Node) connect(addr string) error {
	ctx, cancel := context.WithCancel(context.Background())
	n.Cancel = cancel
	logrus.Info("Connecting to ", addr)
	err := n.Client.ConnectContext(ctx, addr)

	if err != nil {
		cancel()
		return err
	}
	return nil
}

func (n *Node) Disconnect() {
	n.Cancel()
	n.stopRetry <- struct{}{}
}

func (n *Node) ConnectWithRetry(addr string) error {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)

	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		for {
			select {
			case <-n.stopRetry:
				logrus.Info("Stopping ws retry worker")
				return
			case err := <-n.Client.Disconnected():
				logrus.Error("disconnected")
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					return
				}
				logrus.Info("connection retry because: ", err)
				//TODO this makes shutdown using interrupt delayed for maximum 5 secs. Other solutions?
				time.Sleep(5 * time.Second)
				n.connect(addr)
			case <-interrupt:
				n.Cancel()
				return
			}
		}

	}()

	return n.connect(addr)

}

func (n *Node) LoadCertificateKeyPair(name string) error {
	certTLS, err := tls.LoadX509KeyPair(name+".crt", name+".key")
	if err != nil {
		return err
	}
	certX509, err := x509.ParseCertificate(certTLS.Certificate[0])
	if err != nil {
		return err
	}

	n.TLS = &certTLS
	n.X509 = certX509

	// Load CA cert
	caCert, err := ioutil.ReadFile("ca.crt")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	n.CA = caCertPool

	return nil
}

func (n *Node) loadOrGenerateKey() (*rsa.PrivateKey, error) {
	data, err := ioutil.ReadFile("crt.key")
	if err != nil {
		if os.IsNotExist(err) {
			return n.generateKey()
		}
		return nil, err
	}
	block, _ := pem.Decode(data)
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func (n *Node) generateKey() (*rsa.PrivateKey, error) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	keyOut, err := os.OpenFile("crt.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	err = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()
	return priv, err
}

func (n *Node) GenerateCSR() ([]byte, error) {

	subj := pkix.Name{
		CommonName: "example",
		Country:    []string{"SE"},
		//Province:           []string{"Some-State"},
		//Locality:           []string{"MyCity"},
		//Organization:       []string{"Company Ltd"},
		//OrganizationalUnit: []string{"IT"},
	}

	template := x509.CertificateRequest{
		Subject:            subj,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	priv, err := n.loadOrGenerateKey()
	if err != nil {
		return nil, err
	}

	csrBytes, _ := x509.CreateCertificateRequest(rand.Reader, &template, priv)
	d := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})

	return d, nil
}
