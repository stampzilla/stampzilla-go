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
	"net"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"github.com/onrik/logrus/filename"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
)

func main() {
	err := os.MkdirAll("./certs", 0700)
	if err != nil {
		logrus.Fatal(err)
	}

	filenameHook := filename.NewHook()
	logrus.AddHook(filenameHook)

	pool := NewPool()
	go pool.Start()
	webserver := NewWebserver(pool)
	go webserver.Start()

	listener, _ := net.Listen("tcp", "0.0.0.0:1337")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("server: accept: %s", err)
			break
		}
		go handleConnection(conn, pool, webserver)
	}
}

func handleConnection(conn net.Conn, pool *Pool, webserver *Webserver) (err error) {
	log.Printf("server: accepted from %s", conn.RemoteAddr())
	defer conn.Close()
	defer log.Printf("server: closed from %s", conn.RemoteAddr())

	var instance models.ServerInfo

	for {
		d := json.NewDecoder(conn)

		var msg models.Message

		err = d.Decode(&msg)
		if err != nil {
			logrus.Error(err)
			return err
		}

		switch msg.Type {
		case "instance":
			err = json.Unmarshal(msg.Body, &instance)
			if err != nil {
				logrus.Error(err)
				return err
			}

			var config *tls.Config
			config, err = loadCerts(instance.UUID)
			if config == nil {
				// Request certificates
				var csr []byte
				csr, err = generateCSR(instance.UUID)
				if err != nil {
					logrus.Error(err)
					return err
				}

				var resp *models.Message
				resp, err = models.NewMessage("certificate-signing-request", models.Request{
					Type:    "cloud",
					Version: "v0.0", // TODO
					CSR:     string(csr),
				})
				if err != nil {
					logrus.Error(err)
					return err
				}
				resp.WriteToWriter(conn)
			} else {
				// Everything is ready, upgrade to TLS
				var resp *models.Message
				resp, err = models.NewMessage("upgrade", nil)
				if err != nil {
					logrus.Error(err)
					return err
				}
				resp.WriteToWriter(conn)

				handleTLSConnection(config, conn, instance, pool, webserver)
			}
		case "approved-certificate-signing-request":
			var cert string
			err = json.Unmarshal(msg.Body, &cert)
			if err != nil {
				logrus.Error(err)
				return err
			}
			err = ioutil.WriteFile("./certs/"+instance.UUID+".crt", []byte(cert), 0644)
			if err != nil {
				logrus.Error(err)
				return err
			}
		case "certificate-authority":
			var cert string
			err = json.Unmarshal(msg.Body, &cert)
			if err != nil {
				logrus.Error(err)
				return err
			}
			err = ioutil.WriteFile("./certs/"+instance.UUID+"-ca.crt", []byte(cert), 0644)
			if err != nil {
				logrus.Error(err)
				return err
			}

			var config *tls.Config
			config, err = loadCerts(instance.UUID)
			if err != nil {
				logrus.Error(err)
				return err
			}

			var resp *models.Message
			resp, err = models.NewMessage("upgrade", nil)
			if err != nil {
				logrus.Error(err)
				return err
			}
			resp.WriteToWriter(conn)

			handleTLSConnection(config, conn, instance, pool, webserver)
		}
	}
}

func loadCerts(id string) (config *tls.Config, err error) {
	var cert tls.Certificate
	cert, err = tls.LoadX509KeyPair("./certs/"+id+".crt", "./certs/"+id+".key")
	if err != nil {
		return nil, err
	}

	CA_Pool := x509.NewCertPool()
	var rawCaCert []byte
	rawCaCert, err = ioutil.ReadFile("./certs/" + id + "-ca.crt")
	if err != nil {
		return nil, err
	}
	CA_Pool.AppendCertsFromPEM(rawCaCert)

	return &tls.Config{
		RootCAs:      CA_Pool,
		Certificates: []tls.Certificate{cert},
		ServerName:   "localhost",
	}, nil
}

func handleTLSConnection(config *tls.Config, unenc_conn net.Conn, instance models.ServerInfo, pool *Pool, webserver *Webserver) (err error) {
	logrus.Info("Upgrade to TLS")
	conn := tls.Client(unenc_conn, config)
	err = conn.Handshake()
	if err != nil {
		logrus.Error(err)
		return
	}
	logrus.Info("TLS ACTIVE")

	client := &Client{
		ID:       instance.UUID,
		Name:     instance.Name,
		Instance: instance.Instance,
		Phrase:   instance.Phrase,
		Conn:     conn,
		Pool:     pool,

		hub:      webserver.hub,
		requests: make(map[int]chan models.Message),
	}

	logrus.WithFields(logrus.Fields{
		"uuid":     instance.UUID,
		"name":     instance.Name,
		"instance": instance.Instance,
		"phrase":   instance.Phrase,
	}).Info("Client connected")

	if i, _ := pool.GetByID(instance.UUID); i != nil {
		conn.Close()
		return fmt.Errorf("id already connected")
	}

	if i, _ := pool.GetByInstance(instance.Name); i != nil {
		conn.Close()
		return fmt.Errorf("instance already connected")
	}

	pool.Register <- client
	defer func() {
		pool.Unregister <- client
		logrus.WithFields(logrus.Fields{
			"uuid":     instance.UUID,
			"name":     instance.Name,
			"instance": instance.Instance,
		}).Info("Client disconnected")
	}()

	var m *models.Message
	m, err = models.NewMessage("subscribe", []string{"devices", "nodes"})
	m.WriteToWriter(conn)

	for {
		d := json.NewDecoder(conn)

		var msg models.Message

		err := d.Decode(&msg)
		if err != nil {
			return err
		}

		switch msg.Type {
		case "success":
			webserver.HandleResponse(msg, client)
		case "failure":
			webserver.HandleResponse(msg, client)
		case "request":
			go client.handleOutgoingRequest(msg)
		case "nodes":
			raw, _ := msg.Encode()
			client.SetNodes(raw)
		case "devices":
			raw, _ := msg.Encode()
			client.SetDevices(raw)
		default:
			spew.Dump(msg)
		}
	}
}

func (c *Client) handleOutgoingRequest(msg models.Message) {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%s", r)
			}

			if len(msg.Request) == 0 {
				logrus.Error(err)
				return
			}

			resp, err := models.NewMessage("failure", err.Error())
			if err != nil {
				logrus.Error(err)
			}
			resp.Request = msg.Request
			resp.WriteToWriter(c.Conn)
		}
	}()

	respBody, err := doOutgoingRequest(msg.Body)

	var resp *models.Message
	if err != nil {
		resp, err = models.NewMessage("failure", err.Error())
	} else {
		resp, err = models.NewMessage("success", respBody)
	}
	if err != nil {
		logrus.Error(err)
		return
	}

	resp.Request = msg.Request
	resp.WriteToWriter(c.Conn)
}

func generateCSR(id string) ([]byte, error) {
	subj := pkix.Name{
		CommonName:         uuid.New().String(),
		Organization:       []string{"stampzilla-go"},
		OrganizationalUnit: []string{"cloud"},
	}

	template := x509.CertificateRequest{
		Subject:            subj,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	priv, err := loadOrGenerateKey(id)
	if err != nil {
		return nil, err
	}

	csrBytes, _ := x509.CreateCertificateRequest(rand.Reader, &template, priv)
	d := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})

	return d, nil
}

func loadOrGenerateKey(id string) (*rsa.PrivateKey, error) {
	data, err := ioutil.ReadFile("./certs/" + id + ".key")
	if err != nil {
		if os.IsNotExist(err) {
			return generateKey(id)
		}
		return nil, err
	}
	block, _ := pem.Decode(data)
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func generateKey(id string) (*rsa.PrivateKey, error) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	keyOut, err := os.OpenFile("./certs/"+id+".key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	err = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()
	return priv, err
}

func doOutgoingRequest(body json.RawMessage) (interface{}, error) {
	fr, err := models.ParseForwardedRequest(body)
	if err != nil {
		return nil, err
	}
	req, err := fr.ParseRequest()
	if err != nil {
		return nil, err
	}

	// Make the request!
	u, err := url.Parse("https://homegraph.googleapis.com" + req.RequestURI)
	req.Host = "homegraph.googleapis.com"
	//u, err := url.Parse("https://" + req.Host + req.RequestURI)
	if err != nil {
		panic(err)
	}
	req.URL = u
	req.RequestURI = ""

	// GOOGLE CLIENT
	ctx := context.Background()
	googleclient, _, err := transport.NewHTTPClient(ctx, option.WithCredentialsFile("./credentials/google-assistant.json"), option.WithScopes("https://www.googleapis.com/auth/homegraph"))
	if err != nil {
		return nil, err
	}

	resp, err := googleclient.Do(req)
	if err != nil {
		return nil, err
	}

	// Send the result back to the requester
	dump, err := httputil.DumpResponse(resp, true)
	return dump, err
}
