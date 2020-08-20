package main

import (
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
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"github.com/onrik/logrus/filename"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
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

		spew.Dump("RECEIVED", msg)
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
		Name: instance.Name,
		ID:   instance.UUID,
		Conn: conn,
		Pool: pool,

		requests: make(map[int]chan models.Message),
	}

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
	}()

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
		default:
			spew.Dump("RECEIVED TLS", msg)
		}
	}
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
