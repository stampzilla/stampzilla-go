package cloud

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/ca"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/store"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket"
)

type Connection struct {
	store     *store.Store
	config    *models.Config
	ca        *ca.CA
	tlsConfig *tls.Config
	sender    websocket.Sender

	conn          net.Conn
	_connected    bool
	_reconnect    bool
	disconnected  *chan struct{}
	reconnectLoop *chan struct{}
	requestID     int
	requests      map[int]chan models.Message
	subscriptions map[string]struct{}

	sync.RWMutex
}

func New(store *store.Store, config *models.Config, ca *ca.CA, sender websocket.Sender) *Connection {
	caCert, err := ioutil.ReadFile(path.Join("certificates", "ca.crt"))
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	c := &Connection{
		store:  store,
		config: config,
		ca:     ca,
		tlsConfig: &tls.Config{
			GetCertificate: ca.GetServerCertificate,

			// Needed to verify client certificates
			ClientCAs:  caCertPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		},
		sender:        sender,
		requests:      make(map[int]chan models.Message),
		subscriptions: make(map[string]struct{}),
	}

	cl := store.GetCloud()
	c._reconnect = cl.Config.Enable

	go c.Reconnect()

	return c
}

func (c *Connection) Reconnect() {
	reconnectLoop := make(chan struct{})
	defer func() {
		close(reconnectLoop)
		c.Lock()
		c.reconnectLoop = nil
		c.Unlock()
	}()

	c.Lock()
	c.reconnectLoop = &reconnectLoop
	c.Unlock()

	cl := c.store.GetCloud()

	for cl.Config.Enable && c.reconnect() {
		select {
		case <-time.After(time.Second):
			host, port, err := net.SplitHostPort(cl.Config.Server)
			if err != nil {
				c.store.UpdateCloudState(models.CloudState{
					Connected: false,
					Error:     err.Error(),
				})
				continue
			}

			conn, err := net.Dial("tcp", net.JoinHostPort(host, port))
			if err != nil {
				c.store.UpdateCloudState(models.CloudState{
					Connected: false,
					Error:     err.Error(),
				})
				continue
			}

			go c.worker(conn, false)

			return
		}
	}
}

func (c *Connection) Connect(config models.CloudConfig) error {
	host, port, err := net.SplitHostPort(config.Server)
	if err != nil {
		return err
	}

	// Try to connect to test if the information is correct
	conn, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return err
	}

	directUpgrade, err := c.setup(conn)
	if err != nil {
		return err
	}

	// Stop the reconnect loop
	c.Lock()
	c._reconnect = false
	reconnectLoop := c.reconnectLoop
	c.Unlock()
	if reconnectLoop != nil {
		<-*reconnectLoop
	}

	// Disconnect the previous connection if any
	if c.connected() {
		c.conn.Close()

		// Wait until the connection is closed
		c.RLock()
		disconnected := c.disconnected
		c.RUnlock()
		<-*disconnected
	}

	// Start using the new connection
	c.store.UpdateCloudConfig(config)
	go c.worker(conn, directUpgrade)

	return nil
}

func (c *Connection) Disconnect() error {
	cl := c.store.GetCloud()
	cl.Config.Enable = false
	c.store.UpdateCloudConfig(cl.Config)

	// Stop the reconnect loop
	c.Lock()
	c._reconnect = false
	reconnectLoop := c.reconnectLoop
	c.Unlock()
	if reconnectLoop != nil {
		<-*reconnectLoop
	}

	// Disconnect the previous connection if any
	if c.connected() {
		c.conn.Close()

		// Wait until the connection is closed
		c.RLock()
		disconnected := c.disconnected
		c.RUnlock()
		<-*disconnected
	}

	return nil
}

// setup is only done when connecting with a new config
func (c *Connection) setup(conn net.Conn) (directUpgrade bool, err error) {
	logrus.Trace("cloud setup started")
	defer logrus.Trace("cloud setup done", err)

	cl := c.store.GetCloud()
	err = c.SendTo(conn, "instance", models.ServerInfo{
		UUID: c.config.UUID,
		Name: c.config.Name,

		Instance: cl.Config.Instance,
		Phrase:   cl.Config.Phrase,
	})
	if err != nil {
		return false, err
	}

	for {
		d := json.NewDecoder(conn)

		var msg models.Message

		err := d.Decode(&msg)
		if err != nil {
			return false, err
		}

		switch msg.Type {
		case "upgrade":
			logrus.Infof("cloud setup - server already has certificates")
			return true, nil
		case "certificate-signing-request":
			logrus.Trace("cloud setup - got request")
			var body models.Request
			err := json.Unmarshal(msg.Body, &body)
			if err != nil {
				return false, err
			}

			logrus.Trace("cloud setup - build cert")
			cert := &strings.Builder{}
			err = c.ca.CreateCertificateFromCloudRequest(cert, body)
			if err != nil {
				return false, err
			}

			logrus.Trace("cloud setup - send cert")
			err = c.SendTo(conn, "approved-certificate-signing-request", cert.String())
			if err != nil {
				return false, err
			}

			logrus.Trace("cloud setup - send ca")
			ca := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: c.ca.CAX509.Raw})
			err = c.SendTo(conn, "certificate-authority", string(ca))
			if err != nil {
				return false, err
			}
			logrus.Trace("cloud setup - done sending certs")
			return false, nil
		default:
			logrus.WithFields(logrus.Fields{
				"server": conn.RemoteAddr().String(),
				"type":   msg.Type,
			}).Warn("Received unknown message type from cloud during setup")
		}
	}
}

func (c *Connection) Send(msgType string, data interface{}) error {
	if !c.connected() {
		return fmt.Errorf("not connected")
	}

	return c.SendTo(c.conn, msgType, data)
}

func (c *Connection) SendTo(conn net.Conn, msgType string, data interface{}) error {
	message, err := models.NewMessage(msgType, data)
	if err != nil {
		return err
	}

	_, err = message.WriteToWriter(conn)
	return err
}

func (c *Connection) Request(req *models.Message) (json.RawMessage, error) {
	if !c.connected() {
		return nil, fmt.Errorf("not connected")
	}

	msg, err := models.NewMessage("request", req.Body)
	if err != nil {
		return nil, err
	}

	ch := make(chan models.Message)

	c.Lock()
	msg.Request = json.RawMessage(strconv.Itoa(c.requestID))
	c.requests[c.requestID] = ch
	c.requestID++
	c.Unlock()
	defer func() {
		c.Lock()
		if c.requests[c.requestID] != nil {
			close(c.requests[c.requestID])
			delete(c.requests, c.requestID)
		}
		c.Unlock()
	}()

	_, err = msg.WriteToWriter(c.conn)
	if err != nil {
		return nil, err
	}

	select {
	case resp := <-ch:
		switch resp.Type {
		case "failure":
			return nil, fmt.Errorf(string(resp.Body))
		case "success":
			return resp.Body, nil
		}
	case <-time.After(time.Second * 10):
		return nil, fmt.Errorf("timeout")
	}

	return nil, fmt.Errorf("should never happen")
}

func (c *Connection) worker(conn net.Conn, directUpgrade bool) (err error) {
	disconnected := make(chan struct{})
	c.Lock()
	c.disconnected = &disconnected
	c._reconnect = true
	c._connected = true
	c.conn = conn
	c.Unlock()

	logrus.WithFields(logrus.Fields{
		"server": conn.RemoteAddr().String(),
		"client": conn.LocalAddr().String(),
	}).Infof("Cloud connected")

	defer conn.Close()
	defer func() {
		state := models.CloudState{
			Connected: false,
		}
		if err != nil {
			state.Error = err.Error()
		}
		c.store.UpdateCloudState(state)

		c.Lock()
		c.conn = nil
		c._connected = false
		c.Unlock()

		logrus.WithFields(logrus.Fields{
			"server": conn.RemoteAddr().String(),
			"client": conn.LocalAddr().String(),
			"error":  err,
		}).Warnf("Cloud disconnected")

		if c.reconnect() {
			go c.Reconnect()
		}

		close(disconnected)
	}()

	//var buffer = make([]byte, 1024)

	if directUpgrade {
		err = c.tlsWorker(conn)
		return err
	}

	c.store.UpdateCloudState(models.CloudState{
		Connected: true,
	})

	cl := c.store.GetCloud()
	msg, err := models.NewMessage("instance", models.ServerInfo{
		UUID: c.config.UUID,
		Name: c.config.Name,

		Instance: cl.Config.Instance,
		Phrase:   cl.Config.Phrase,
	})
	if err != nil {
		logrus.Error(err)
		return
	}
	msg.WriteToWriter(conn)

	for {
		d := json.NewDecoder(conn)

		var msg models.Message

		err := d.Decode(&msg)
		if err != nil {
			return err
		}

		switch msg.Type {
		case "certificate-signing-request":
			return fmt.Errorf("cloud proxy is missing authorization")
		case "upgrade":
			logrus.Trace("cloud: upgrade to TLS")
			err = c.tlsWorker(conn)
			return err
		default:

			logrus.WithFields(logrus.Fields{
				"server": conn.RemoteAddr().String(),
				"type":   msg.Type,
			}).Warn("Received unknown message type from cloud")
		}
	}

	return
}

func (c *Connection) tlsWorker(unenc_conn net.Conn) error {
	conn := tls.Server(unenc_conn, c.tlsConfig)
	err := conn.Handshake()
	if err != nil {
		logrus.Warn("cloud: tls error", err)
		return err
	}

	c.Lock()
	c.conn = conn
	c.Unlock()
	defer func() {
		c.Lock()
		c.conn = unenc_conn
		c.Unlock()
	}()

	c.store.UpdateCloudState(models.CloudState{
		Connected: true,
		Secure:    true,
	})

	logrus.Trace("cloud: TLS listening")
	for {
		d := json.NewDecoder(conn)

		msg := &models.Message{}

		err := d.Decode(msg)
		if err != nil {
			return err
		}

		if id, err := strconv.Atoi(string(msg.Request)); err == nil {
			logrus.WithFields(logrus.Fields{
				"type":      msg.Type,
				"requestID": msg.Request,
			}).Debug("Received request response")

			c.Lock()
			ch, ok := c.requests[id]
			c.Unlock()

			if ok {
				go func() {
					ch <- *msg
				}()
			}
		}

		if msg.Type != "success" && msg.Type != "failure" {
			c.processRequest(msg, conn)
		}
	}
}

func (c *Connection) processRequest(req *models.Message, conn *tls.Conn) {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%s", r)
			}

			if len(req.Request) == 0 {
				logrus.Error(err)
				return
			}

			resp, err := models.NewMessage("failure", err.Error())
			if err != nil {
				logrus.Error(err)
			}
			resp.Request = req.Request
			_, err = resp.WriteToWriter(conn)
			if err != nil {
				logrus.Error(err)
			}
		}
	}()

	data, err := c.MessageFromCloud(req, nil) // TODO: Add the authorized user

	// The message contains a request ID, so respond with the result
	if len(req.Request) > 0 {
		resp, e := models.NewMessage("success", data)
		if e != nil {
			logrus.Error(e)
		}
		if err != nil {
			resp, e = models.NewMessage("failure", err.Error())
			if e != nil {
				logrus.Error(e)
			}
		}

		resp.Request = req.Request
		_, err := resp.WriteToWriter(conn)
		if err != nil {
			logrus.Error(err)
		}
	}
}

func (c *Connection) connected() bool {
	c.RLock()
	defer c.RUnlock()

	return c._connected
}

func (c *Connection) reconnect() bool {
	c.RLock()
	defer c.RUnlock()

	return c._reconnect
}
