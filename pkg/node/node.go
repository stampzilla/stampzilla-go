package node

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/build"
	"github.com/stampzilla/stampzilla-go/pkg/websocket"
)

// OnFunc is used in all the callbacks
type OnFunc func(body json.RawMessage) error
type OnCloudRequestFunc func(req *http.Request) (*http.Response, error)
type callbackFunc func(body json.RawMessage, requestID json.RawMessage) error

// Node is the main struct.
type Node struct {
	UUID     string
	Type     string
	Version  string
	Protocol string

	Client websocket.Websocket
	// DisconnectClient context.CancelFunc
	wg         *sync.WaitGroup
	Config     *models.Config
	X509       *x509.Certificate
	TLS        *tls.Certificate
	CA         *x509.CertPool
	callbacks  map[string][]callbackFunc
	requestID  int
	requests   map[int]chan models.Message
	Devices    *devices.List
	shutdown   []func()
	stop       chan struct{}
	sendUpdate chan devices.ID

	mutex sync.Mutex
}

// New returns a new Node.
func New(t string) *Node {
	client := websocket.New()
	node := NewWithClient(client)
	node.Type = t

	node.setup()

	return node
}

// NewWithClient returns a new Node with a custom websocket client.
func NewWithClient(client websocket.Websocket) *Node {
	return &Node{
		Client:     client,
		wg:         &sync.WaitGroup{},
		callbacks:  make(map[string][]callbackFunc),
		requests:   make(map[int]chan models.Message),
		Devices:    devices.NewList(),
		stop:       make(chan struct{}),
		sendUpdate: make(chan devices.ID),
	}
}

// Stop will shutdown the node similar to a SIGTERM.
func (n *Node) Stop() {
	close(n.stop)
}

// Stopped is closed when the node is stopped by n.Stop or os signal.
func (n *Node) Stopped() <-chan struct{} {
	return n.stop
}

// Wait for node to be done after shutdown.
func (n *Node) Wait() {
	n.Client.Wait()
	n.wg.Wait()
}

func (n *Node) setup() {
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{TimestampFormat: time.RFC3339Nano, FullTimestamp: true})

	// Make sure we have a config
	n.Config = &models.Config{}
	n.Config.MustLoad()

	if n.Config.Version {
		fmt.Println(build.String())
		os.Exit(1)
	}
	n.Version = build.String()

	if n.Config.LogLevel != "" {
		lvl, err := logrus.ParseLevel(n.Config.LogLevel)
		if err != nil {
			logrus.Fatal(err)
			return
		}
		logrus.SetLevel(lvl)
	}

	// n.Config.Save("config.json")
}

// WriteMessage writes a message to the server over websocket client.
func (n *Node) WriteMessage(msgType string, data interface{}) error {
	msg, err := models.NewMessage(msgType, data)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"type": msgType,
		"body": data,
	}).Tracef("Send to server")

	return n.Client.WriteJSON(msg)
}
func (n *Node) WriteRequest(msgType string, data interface{}) (*models.Message, error) {
	msg, err := models.NewMessage(msgType, data)
	if err != nil {
		return nil, err
	}

	ch := make(chan models.Message)

	n.mutex.Lock()
	msg.Request = json.RawMessage(strconv.Itoa(n.requestID))
	n.requests[n.requestID] = ch
	n.requestID++
	n.mutex.Unlock()
	defer func() {
		n.mutex.Lock()
		if n.requests[n.requestID] != nil {
			close(n.requests[n.requestID])
			delete(n.requests, n.requestID)
		}
		n.mutex.Unlock()
	}()

	logrus.WithFields(logrus.Fields{
		"type":      msgType,
		"body":      data,
		"requestID": msg.Request,
	}).Debugf("Send request to server")

	err = n.Client.WriteJSON(msg)
	if err != nil {
		return nil, err
	}

	select {
	case resp := <-ch:
		switch resp.Type {
		case "failure":
			return nil, fmt.Errorf(string(resp.Body))
		case "success":
			return &resp, nil
		}
	case <-time.After(time.Second * 10):
		return nil, fmt.Errorf("timeout")
	}

	return nil, fmt.Errorf("should never happen")
}
func (n *Node) WriteResponse(msgType string, data interface{}, request json.RawMessage) error {
	msg, err := models.NewMessage(msgType, data)
	logrus.WithFields(logrus.Fields{
		"type": msgType,
		"body": data,
	}).Tracef("Send to server")
	if err != nil {
		return err
	}

	msg.Request = request

	logrus.WithFields(logrus.Fields{
		"type":    msg.Type,
		"request": msg.Request,
	}).Debug("Wrote response")

	return n.Client.WriteJSON(msg)
}

// WaitForMessage is a helper method to wait for a specific message type.
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

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		select {
		case <-interrupt:
			close(n.stop)
		case <-n.stop:
		}
		cancel()
		go func() {
			<-time.After(time.Second * 10)
			log.Fatal("force shutdown after 10 seconds")
		}()
		for _, f := range n.shutdown {
			f()
		}
	}()

	n.connect(ctx, u)

	// wait for server info so we can update our config
	serverInfo := &models.ServerInfo{}
	err = n.WaitForMessage("server-info", serverInfo)
	if err != nil {
		return err
	}
	n.Config.Port = serverInfo.Port
	n.Config.TLSPort = serverInfo.TLSPort

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
	if n.Config.Host == "" {
		ip, port := queryMDNS()
		n.Config.Host = ip
		n.Config.Port = port
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

	// If we have certificate we can connect to TLS immediately
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
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		select {
		case <-interrupt:
			close(n.stop)
		case <-n.stop:
		}
		cancel()
		go func() {
			<-time.After(time.Second * 10)
			log.Fatal("force shutdown after 10 seconds")
		}()
		for _, f := range n.shutdown {
			f()
		}
	}()

	n.Client.OnConnect(func() {
		for what := range n.getCallbacks() {
			n.Subscribe(what)
		}
		n.SyncDevices()
	})
	n.connect(ctx, u)
	n.wg.Add(1)
	go n.reader(ctx)
	go n.syncWorker()
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

			// Dont wait for callbacks because they could send requests and then we need to be able to receive the result
			cbs := n.getCallbacks()
			go func() {
				for _, cb := range cbs[msg.Type] {
					err := cb(msg.Body, msg.Request)
					if err != nil {
						logrus.Error(err)
						continue
					}
				}
			}()

			var ok bool
			if id, err := strconv.Atoi(string(msg.Request)); err == nil {
				logrus.WithFields(logrus.Fields{
					"type":      msg.Type,
					"requestID": msg.Request,
				}).Debug("Received request response")

				n.mutex.Lock()
				var ch chan models.Message
				ch, ok = n.requests[id]
				n.mutex.Unlock()

				if ok {
					go func() {
						ch <- *msg
					}()
				}
			}

			if (cbs[msg.Type] == nil || len(cbs[msg.Type]) == 0) && !ok {
				logrus.WithFields(logrus.Fields{
					"type":      msg.Type,
					"requestID": msg.Request,
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
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	subj := pkix.Name{
		CommonName:         uuid.New().String(),
		Organization:       []string{"stampzilla-go"},
		OrganizationalUnit: []string{hostname, n.Type},
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

// On sets up a callback that is run when a message received with type what.
func (n *Node) On(what string, cb OnFunc) {
	n.mutex.Lock()
	n.callbacks[what] = append(n.callbacks[what], func(raw json.RawMessage, req json.RawMessage) error {
		return cb(raw)
	})
	n.mutex.Unlock()
	n.Subscribe(what)
}

func (n *Node) OnCloudRequest(cb OnCloudRequestFunc) {
	n.mutex.Lock()
	n.callbacks["cloud-request"] = append(n.callbacks["cloud-request"], func(raw json.RawMessage, reqID json.RawMessage) error {
		msg, err := models.ParseForwardedRequest(raw)
		if err != nil {
			return err
		}

		req, err := msg.ParseRequest()
		if err != nil {
			return err
		}

		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%s", r)
				}

				if len(reqID) == 0 {
					logrus.Error(err)
					return
				}

				err = n.WriteResponse("failure", err.Error(), reqID)
				if err != nil {
					logrus.Error(err)
				}
			}
		}()

		resp, err := cb(req)

		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return err
		}

		// The message contains a request ID, so respond with the result
		if len(reqID) > 0 {
			if err != nil {
				err := n.WriteResponse("failure", err.Error(), reqID)
				if err != nil {
					logrus.Error(err)
				}
			} else {
				err := n.WriteResponse("success", dump, reqID)
				if err != nil {
					logrus.Error(err)
				}
			}
		}

		if err != nil {
			logrus.Error(err)
		}
		return err
	})
	n.mutex.Unlock()
	n.Subscribe("cloud-request")
}

func (n *Node) getCallbacks() map[string][]callbackFunc {
	cbs := make(map[string][]callbackFunc)
	n.mutex.Lock()
	for k, v := range n.callbacks {
		cbs[k] = v
	}
	n.mutex.Unlock()
	return cbs
}

// OnConfig is run when node receives updated configuration from the server.
func (n *Node) OnConfig(cb OnFunc) {
	n.On("setup", func(data json.RawMessage) error {
		conf := &models.Node{}
		err := json.Unmarshal(data, conf)
		if err != nil {
			return err
		}

		if len(conf.Config) == 0 {
			return nil
		}

		return cb(conf.Config)
	})
}

// WaitForFirstConfig blocks until we receive the first config from server.
func (n *Node) WaitForFirstConfig() func() error {
	var once sync.Once
	waitForConfig := make(chan struct{})
	n.OnConfig(func(data json.RawMessage) error {
		var err error
		once.Do(func() {
			close(waitForConfig)
		})
		return err
	})

	return func() error {
		select {
		case <-waitForConfig:
			return nil
		case <-n.stop:
			return fmt.Errorf("node: stopped before first config")
		}
	}
}

// OnShutdown registers a callback that is run before the server shuts down.
func (n *Node) OnShutdown(cb func()) {
	n.shutdown = append(n.shutdown, cb)
}

// OnRequestStateChange is run if we get a state-change request from the server to update our devices (for example we are requested to turn on a light).
func (n *Node) OnRequestStateChange(cb func(state devices.State, device *devices.Device) error) {
	n.On("state-change", func(data json.RawMessage) error {
		// devs := devices.NewList()
		devs := make(map[devices.ID]devices.State)
		err := json.Unmarshal(data, &devs)
		if err != nil {
			return err
		}

		for devID, state := range devs {
			// loop over all devices and compare state
			stateChange := make(devices.State)
			foundChange := false
			oldDev := n.Devices.Get(devID)
			for s, newState := range state {
				oldState := oldDev.State[s]
				if newState != oldState {
					// fmt.Printf("oldstate %T %#v\n", oldState, newState)
					// fmt.Printf("newState %T %#v\n", newState, newState)
					stateChange[s] = newState
					foundChange = true
				}
			}
			if foundChange {
				err := cb(stateChange, oldDev)
				if err != nil {
					// set state back to before. we could not change it as requested
					// continue to next device
					if err == ErrSkipSync { // skip sync without logging error if needed
						continue
					}
					logrus.Error(err)
					continue
				}

				// set the new state and send it to the server
				err = n.Devices.SetState(devID, state.Merge(stateChange))
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

var ErrSkipSync = fmt.Errorf("skipping device sync after RequestStateChange")

// AddOrUpdate adds or updates a device in our local device store and notifies the server about the new state of the device.
func (n *Node) AddOrUpdate(d *devices.Device) {
	d.Lock()
	d.ID.Node = n.UUID
	d.Unlock()
	n.Devices.Add(d)
	n.sendUpdate <- d.ID
}

// syncWorker is a debouncer to send multiple devices to the server if we change many rapidly.
func (n *Node) syncWorker() {
	for {
		que := make([]devices.ID, 0)
		id := <-n.sendUpdate
		que = append(que, id)

		max := time.NewTimer(10 * time.Millisecond)
	outer:
		for {
			select {
			case id := <-n.sendUpdate:
				que = append(que, id)
			case <-time.After(1 * time.Millisecond):
				break outer
			case <-max.C:
				break outer
			}
		}

		// send message to server
		devs := make(devices.DeviceMap)
		for _, id := range que {
			d := n.GetDevice(id.ID)
			devs[d.ID] = d.Copy()
		}
		err := n.WriteMessage("update-devices", devs)
		if err != nil {
			logrus.Error(err)
		}
	}
}

func (n *Node) GetDevice(id string) *devices.Device {
	return n.Devices.Get(devices.ID{Node: n.UUID, ID: id})
}

// UpdateState updates the new state on the node if if differs and sends update to server if there was a diff.
func (n *Node) UpdateState(id string, newState devices.State) {
	device := n.GetDevice(id)

	if device == nil {
		return
	}

	if len(newState) == 0 {
		return
	}

	if diff := device.State.Diff(newState); len(diff) != 0 {
		device.Lock()
		device.State.MergeWith(diff)
		device.Unlock()
		n.SyncDevice(id)
	}
}

// SyncDevices notifies the server about the state of all our known devices.
func (n *Node) SyncDevices() error {
	return n.WriteMessage("update-devices", n.Devices)
}

// SyncDevice sync single device.
func (n *Node) SyncDevice(id string) {
	n.sendUpdate <- devices.ID{ID: id}
}

// Subscribe subscribes to a topic in the server.
func (n *Node) Subscribe(what ...string) error {
	return n.WriteMessage("subscribe", what)
}

func (n *Node) SendThruCloud(service string, req *http.Request) (*http.Response, error) {
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}

	msg, err := n.WriteRequest("cloud-request", models.ForwardedRequest{
		Dump:    dump,
		Service: service,
	})
	if err != nil {
		return nil, err
	}

	var body string
	err = json.Unmarshal(msg.Body, &body)
	if err != nil {
		return nil, err
	}

	raw, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(raw)
	rd := bufio.NewReader(b)
	resp, err := http.ReadResponse(rd, req)
	if err != nil {
		return nil, err
	}

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}

	resp.Body = reader

	return resp, nil
}
