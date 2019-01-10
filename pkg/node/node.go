package node

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
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/websocket"
)

type OnFunc func(json.RawMessage) error

type Node struct {
	UUID     string
	Type     string
	Version  string
	Protocol string

	Client websocket.Websocket
	//DisconnectClient context.CancelFunc
	wg        *sync.WaitGroup
	Config    *models.Config
	X509      *x509.Certificate
	TLS       *tls.Certificate
	CA        *x509.CertPool
	callbacks map[string][]OnFunc
	Devices   *devices.List
	shutdown  []func()
	stop      chan struct{}
}

// New returns a new Node
func New(t string) *Node {
	client := websocket.New()
	node := NewWithClient(client)
	node.Type = t

	return node
}

// NewWithClient returns a new Node with a custom websocket client
func NewWithClient(client websocket.Websocket) *Node {
	return &Node{
		Client:    client,
		wg:        &sync.WaitGroup{},
		callbacks: make(map[string][]OnFunc),
		Devices:   devices.NewList(),
		stop:      make(chan struct{}),
	}
}

// Stop will shutdown the node similar to a SIGTERM
func (n *Node) Stop() {
	close(n.stop)
}

// Wait for node to be done after shutdown
func (n *Node) Wait() {
	n.Client.Wait()
	n.wg.Wait()
}

func (n *Node) setup() {
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{TimestampFormat: time.RFC3339Nano, FullTimestamp: true})

	//Make sure we have a config
	n.Config = &models.Config{}
	n.Config.Load()
	//n.Config.Save("config.json")
}

// WriteMessage writes a message to the server over websocket client
func (n *Node) WriteMessage(msgType string, data interface{}) error {
	msg, err := models.NewMessage(msgType, data)
	logrus.WithFields(logrus.Fields{
		"type": msgType,
		"body": data,
	}).Tracef("Send to server")
	if err != nil {
		return err
	}
	return n.Client.WriteJSON(msg)
}

// WaitForMessage is a helper method to wait for a specific message type
func (n *Node) WaitForMessage(msgType string, dst interface{}) error {

	for data := range n.Client.Read() {
		msg, err := models.ParseMessage(data)
		if err != nil {
			return err
		}
		if msg.Type == msgType {
			return json.Unmarshal(msg.Body, dst)
		}
	}
	return nil
}

func (n *Node) fetchCertificate() error {
	// Start with creating a CSR and assign a UUID
	csr, err := n.generateCSR()
	if err != nil {
		return err
	}

	u := fmt.Sprintf("ws://%s:%s/ws", n.Config.Host, n.Config.Port)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	n.connect(ctx, u)

	// wait for server info so we can update our config
	serverInfo := &models.ServerInfo{}
	err = n.WaitForMessage("server-info", serverInfo)
	if err != nil {
		return err
	}
	n.Config.Port = serverInfo.Port
	n.Config.TLSPort = serverInfo.TLSPort
	n.Config.Save("config.json")

	n.WriteMessage("certificate-signing-request", models.Request{
		Type:    n.Type,
		Version: n.Version,
		CSR:     string(csr),
	})
	if err != nil {
		return err
	}

	// wait for our new certificate

	var rawCert string
	err = n.WaitForMessage("approved-certificate-signing-request", &rawCert)

	err = ioutil.WriteFile("crt.crt", []byte(rawCert), 0644)
	if err != nil {
		return err
	}

	var caCert string
	err = n.WaitForMessage("certificate-authority", &caCert)

	err = ioutil.WriteFile("ca.crt", []byte(caCert), 0644)
	if err != nil {
		return err
	}

	logrus.Info("Disconnect inseure connection")
	cancel()
	n.Wait()

	// We should have a certificate now. Try to load it
	return n.loadCertificateKeyPair("crt")
}

// Connect starts the node and makes connection to the server. Normally discovered using mdns but can be configured aswell.
func (n *Node) Connect() error {
	n.setup()

	if n.Config.Host == "" {
		ip, port, err := queryMDNS()
		if err != nil {
			return err
		}

		n.Config.Host = ip
		n.Config.Port = strconv.Itoa(port)
	}

	// Load our signed certificate and get our UUID
	err := n.loadCertificateKeyPair("crt")

	if err != nil {
		logrus.Error("Error trying to load certificate: ", err)
		err = n.fetchCertificate()
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
	signal.Notify(interrupt, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		select {
		case <-interrupt:
		case <-n.stop:
		}
		cancel()
		for _, f := range n.shutdown {
			f()
		}
	}()

	n.Client.OnConnect(func() {
		for what := range n.callbacks {
			n.Subscribe(what)
		}
		n.SyncDevices()
	})
	n.connect(ctx, u)
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
		case data := <-n.Client.Read():
			msg, err := models.ParseMessage(data)
			if err != nil {
				logrus.Error("node:", err)
				continue
			}
			for _, cb := range n.callbacks[msg.Type] {
				err := cb(msg.Body)
				if err != nil {
					logrus.Error(err)
					continue
				}
			}
			if n.callbacks[msg.Type] == nil || len(n.callbacks[msg.Type]) == 0 {
				logrus.WithFields(logrus.Fields{
					"type": msg.Type,
				}).Warn("Received message but no one cared")
			}
		}
	}
}

func (n *Node) connect(ctx context.Context, addr string) {
	headers := http.Header{}
	headers.Add("X-UUID", n.UUID)
	headers.Add("X-TYPE", n.Type)
	headers.Set("Sec-WebSocket-Protocol", n.Protocol)
	if n.Protocol == "" {
		headers.Set("Sec-WebSocket-Protocol", "node")
	}
	n.Client.ConnectWithRetry(ctx, addr, headers)
}

func (n *Node) loadCertificateKeyPair(name string) error {
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

func (n *Node) generateCSR() ([]byte, error) {

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

// On sets up a callback that is run when a message recieved with type what
func (n *Node) On(what string, cb OnFunc) {
	n.Subscribe(what)
	n.callbacks[what] = append(n.callbacks[what], cb)
}

//OnConfig is run when node recieves updated configuration from the server
func (n *Node) OnConfig(cb OnFunc) {
	n.On("setup", func(data json.RawMessage) error {
		conf := &models.Node{}
		err := json.Unmarshal(data, conf)
		if err != nil {
			return err
		}

		return cb(conf.Config)
	})
}

// OnShutdown registeres a callback that is run before the server shuts down
func (n *Node) OnShutdown(cb func()) {
	n.shutdown = append(n.shutdown, cb)
}

// OnRequestStateChange is run if we get a state-change request from the server to update our devices (for example we are requested to turn on a light)
func (n *Node) OnRequestStateChange(cb func(state devices.State, device *devices.Device) error) {
	n.On("state-change", func(data json.RawMessage) error {
		//devs := devices.NewList()
		devs := make(map[devices.ID]devices.State)
		err := json.Unmarshal(data, &devs)
		if err != nil {
			return err
		}

		// loop over all devices and compare state
		stateChange := make(devices.State)

		for devID, state := range devs {
			foundChange := false
			oldDev := n.Devices.Get(devID)
			for s, newState := range state {
				oldState := oldDev.State[s]
				if newState != oldState {
					stateChange[s] = newState
					foundChange = true
				}
			}
			if foundChange {
				err := cb(stateChange, oldDev)
				if err != nil {
					// set state back to before. we could not change it as requested
					// continue to next device
					logrus.Error(err)
					continue
				}

				// set the new state and send it to the server
				//TODO create new type for device.ID that contains both node uuid and device id and remove all splits currently used
				err = n.Devices.SetState(devID, state)
				if err != nil {
					logrus.Error(err)
					continue
				}

				err = n.WriteMessage("update-device", n.Devices.Get(devID))
				if err != nil {
					logrus.Error(err)
					continue
				}
			}
		}

		return nil
	})
}

// AddOrUpdate adds or updates a device in our local device store and notifies the server about the new state of the device.
func (n *Node) AddOrUpdate(d *devices.Device) error {
	d.ID.Node = n.UUID
	n.Devices.Add(d)
	return n.WriteMessage("update-device", d)
}

//SyncDevices notifies the server about the state of all our known devices.
func (n *Node) SyncDevices() error {
	return n.WriteMessage("update-devices", n.Devices)
}

//Subscribe subscribes to a topic in the server
func (n *Node) Subscribe(what ...string) error {
	return n.WriteMessage("subscribe", what)
}
