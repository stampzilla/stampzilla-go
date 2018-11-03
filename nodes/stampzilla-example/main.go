package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
)

func main() {

	client := NewWebsocketClient()
	node := NewNode(client)
	csr, err := node.GenerateCSR()
	if err != nil {
		logrus.Error(err)
		return
	}

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	err = node.ConnectWithRetry(u.String())
	if err != nil {
		logrus.Error(err)
	}

	msg, err := models.NewMessage("certificate-signing-request", string(csr))
	if err != nil {
		logrus.Error(err)
		return
	}

	node.Client.WriteJSON(msg)

	node.Wait()

	logrus.Info("node done...")
	logrus.Info("waiting for client to be done")
	node.Client.Wait()
	logrus.Info("client done")
}

type Node struct {
	Client *WebsocketClient
	Cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func NewNode(client *WebsocketClient) *Node {
	return &Node{
		Client: client,
		wg:     &sync.WaitGroup{},
	}
}
func (n *Node) Wait() {
	n.wg.Wait()
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
func (n *Node) ConnectWithRetry(addr string) error {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)

	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		for {
			select {
			case err := <-n.Client.Disconnected():
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

	err := n.connect(addr)

	if err != nil {
		return err
	}

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
