package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
)

type OnFunc func(json.RawMessage) error

type Node struct {
	UUID string
	Type string

	Client Websocket
	//DisconnectClient context.CancelFunc
	wg        *sync.WaitGroup
	Config    *models.Config
	X509      *x509.Certificate
	TLS       *tls.Certificate
	CA        *x509.CertPool
	callbacks map[string][]OnFunc
	devices   models.Devices
}

func NewNode(client Websocket) *Node {
	return &Node{
		Client:    client,
		wg:        &sync.WaitGroup{},
		callbacks: make(map[string][]OnFunc),
		devices:   make(models.Devices),
	}
}
func (n *Node) Wait() {
	n.Client.Wait()
	n.wg.Wait()
}

func (n *Node) setup() {
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{TimestampFormat: time.RFC3339Nano, FullTimestamp: true})

	//Make sure we have a config
	n.Config = &models.Config{}
	n.Config.MustLoad()
	n.Config.Save("config.json")
}

func (n *Node) WriteMessage(msgType string, data interface{}) error {
	msg, err := models.NewMessage(msgType, data)
	if err != nil {
		return err
	}
	return n.Client.WriteJSON(msg)
}

func (n *Node) Connect() error {
	n.setup()

	// Load our signed certificate and get our UUID
	err := n.LoadCertificateKeyPair("crt")

	if err != nil {
		logrus.Error("Error trying to load certificate: ", err)

		// Start with creating a CSR and assign a UUID
		csr, err := n.GenerateCSR()
		if err != nil {
			return err
		}

		u := fmt.Sprintf("ws://%s:%s/ws", n.Config.Host, n.Config.Port)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err = n.connect(ctx, u)
		if err != nil {
			return err
		}

		// wait for server info so we can update our config
		serverInfo := &models.ServerInfo{}
		err = n.Client.WaitForMessage("server-info", serverInfo)
		if err != nil {
			return err
		}
		n.Config.Port = serverInfo.Port
		n.Config.TLSPort = serverInfo.TLSPort
		n.Config.Save("config.json")

		n.WriteMessage("certificate-signing-request", string(csr))
		if err != nil {
			return err
		}

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

		logrus.Info("Disconnect inseure connection")
		cancel()
		n.Wait()

		// We should have a certificate now. Try to load it
		err = n.LoadCertificateKeyPair("crt")
		if err != nil {
			return err
		}
	}

	//If we have certificate we can connect to TLS immedietly
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*n.TLS},
		RootCAs:      n.CA,
		ServerName:   "localhost",
	}

	n.Client.SetTLSConfig(tlsConfig)

	u := fmt.Sprintf("wss://%s:%s/ws", n.Config.Host, n.Config.TLSPort)
	ctx, cancel := context.WithCancel(context.Background())

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		<-interrupt
		cancel()
	}()

	n.Client.OnConnect(func() {
		n.sendNodeUpdate()
	})
	err = n.connect(ctx, u)
	if err != nil {
		logrus.Error(err)
	}
	n.wg.Add(1)
	go n.reader(ctx)
	return nil
}

func (n *Node) reader(ctx context.Context) {
	defer n.wg.Done()
	for {
		select {
		case <-ctx.Done():
			logrus.Info("Stopping node reader because:", ctx.Err())
			return
		case msg := <-n.Client.Message():
			for _, cb := range n.callbacks[msg.Type] {
				err := cb(msg.Body)
				if err != nil {
					logrus.Error(err)
					return
				}
			}
		}
	}
}

func (n *Node) connect(ctx context.Context, addr string) error {
	//ctx, cancel := context.WithCancel(ctx)
	//if n.DisconnectClient != nil {
	//n.DisconnectClient()
	//}
	//n.DisconnectClient = cancel
	headers := http.Header{}
	headers.Add("X-UUID", n.UUID)
	headers.Add("X-TYPE", n.Type)
	headers.Add("Sec-WebSocket-Protocol", "node")

	n.Client.ConnectWithRetry(ctx, addr, headers)

	return nil
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
	n.UUID = certX509.Subject.CommonName

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
		CommonName: uuid.New().String(),
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

	n.UUID = template.Subject.CommonName

	return d, nil
}

func (n *Node) on(what string, cb OnFunc) {
	n.callbacks[what] = append(n.callbacks[what], cb)
}

//OnConfig is run when node recieves updated configuration from the server
func (n *Node) OnConfig(cb OnFunc) {

	n.on("setup", func(data json.RawMessage) error {
		conf := &models.Node{}
		err := json.Unmarshal(data, conf)
		if err != nil {
			return err
		}
		return cb(conf.Config)
	})
}

func (n *Node) OnRequestStateChange(cb func(state models.DeviceState, device *models.Device) error) {
	n.on("state-change", func(data json.RawMessage) error {
		conf := &models.Node{}
		err := json.Unmarshal(data, conf)
		if err != nil {
			return err
		}

		if conf.Devices == nil {
			logrus.Error("No devices found in incoming update-node, skipping.")
			return nil
		}

		// loop over all devices and compare state

		stateChange := make(models.DeviceState)

		foundAnyChange := false
		for k, dev := range n.devices {
			foundChange := false
			for s, oldState := range dev.State {
				newState := conf.Devices[k].State[s]
				if newState != oldState {
					stateChange[s] = newState
					foundChange = true
					foundAnyChange = true
				}
			}
			if foundChange {
				err := cb(stateChange, dev)
				if err != nil {
					// set state back to before. we could not change it as requested
					// continue to next device
					logrus.Error(err)
					continue
				}
				// set the new state and send it to the server
				n.devices[k].State = conf.Devices[k].State

			}
		}
		if foundAnyChange {
			return n.sendNodeUpdate()
		}

		return nil
	})
}

func (n *Node) AddOrUpdate(d *models.Device) error {
	key := fmt.Sprintf("%s.%s", n.UUID, d.ID)
	d.Node = n.UUID
	n.devices[key] = d

	return n.sendNodeUpdate()
}

func (n *Node) sendNodeUpdate() error {
	node := models.Node{
		UUID: n.UUID,
		//Connected: true,
		//Type:      n.Type,
		Devices: n.devices,
		//Config
	}
	return n.WriteMessage("update-node", node)
}
