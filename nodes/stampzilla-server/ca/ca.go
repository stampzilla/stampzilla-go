package ca

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	mrand "math/rand"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/store"
)

const storagePath = "certificates"

type CA struct {
	X509   *x509.Certificate
	TLS    *tls.Certificate
	CAX509 *x509.Certificate
	CATLS  *tls.Certificate

	Store *store.Store

	sync.Mutex
}

func LoadOrCreate(names ...string) (*CA, error) {
	name := "ca"
	if len(names) == 1 {
		name = names[0]
	}
	cert := &CA{}
	return cert, cert.LoadOrCreate(name)
}

func (ca *CA) LoadOrCreate(name string) error {
	err := ca.Load(name)
	if err == nil {
		return nil
	}

	if name == "ca" {
		return ca.CreateCA()
	}
	return ca.CreateCertificate(name)
}

func (ca *CA) Load(name string) error {
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		os.Mkdir(storagePath, 0755)
	}

	certTLS, err := tls.LoadX509KeyPair(path.Join(storagePath, name+".crt"), path.Join(storagePath, name+".key"))
	if err != nil {
		return err
	}
	certX509, err := x509.ParseCertificate(certTLS.Certificate[0])
	if err != nil {
		return err
	}

	if name == "ca" {
		ca.CATLS = &certTLS
		ca.CAX509 = certX509
		return nil
	}
	ca.TLS = &certTLS
	ca.X509 = certX509

	return nil
}

func (ca *CA) CreateCA() error {
	// Create a 10year CA cert
	recipe := &x509.Certificate{
		SerialNumber: big.NewInt(int64(mrand.Int())),
		Subject: pkix.Name{
			Organization: []string{"stampzilla-go"},
			CommonName:   "stampzilla-go CA",
			//Country:       []string{"se"},
			//Province:      []string{"PROVINCE"},
			//Locality:      []string{"CITY"},
			//StreetAddress: []string{"ADDRESS"},
			//PostalCode:    []string{"POSTAL_CODE"},
		},
		SubjectKeyId:          bigIntHash(big.NewInt(int64(mrand.Int()))),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10 years
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// Generate keys
	priv, _ := rsa.GenerateKey(rand.Reader, 2048) // key size
	certBytes, err := x509.CreateCertificate(rand.Reader, recipe, recipe, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	err = ca.ExportToDisk("ca", certBytes, priv)
	if err != nil {
		return err
	}

	return ca.Load("ca")
}

func (ca *CA) SetStore(s *store.Store) {
	ca.Store = s
	ca.Store.UpdateCertificates(ca.GetCertificates())
}

func bigIntHash(n *big.Int) []byte {
	h := sha1.New()
	h.Write(n.Bytes())
	return h.Sum(nil)
}

func (ca *CA) CreateCertificate(name string) error {
	recipe := &x509.Certificate{
		SerialNumber: big.NewInt(1653),
		Subject: pkix.Name{
			Organization: []string{"stampzilla-go"},
			CommonName:   name,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10 years
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		//IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames: []string{"localhost"},
	}

	// Generate keys
	priv, _ := rsa.GenerateKey(rand.Reader, 2048) // key size
	certBytes, err := x509.CreateCertificate(rand.Reader, recipe, ca.CAX509, &priv.PublicKey, ca.CATLS.PrivateKey)
	if err != nil {
		return err
	}

	err = ca.ExportToDisk(name, certBytes, priv)
	if err != nil {
		return err
	}

	return ca.Load(name)
}

func (ca *CA) CreateCertificateFromRequest(wr io.Writer, c string, r models.Request) error {
	pemBlock, _ := pem.Decode([]byte(r.CSR))
	if pemBlock == nil {
		return fmt.Errorf("pem.Decode failed?")
	}
	clientCSR, err := x509.ParseCertificateRequest(pemBlock.Bytes)
	if err != nil {
		return err
	}
	if err = clientCSR.CheckSignature(); err != nil {
		return err
	}

	approved := <-ca.WaitForApproval(clientCSR.Subject.CommonName, c, r)
	if approved != true {
		return fmt.Errorf("Request was not approved")
	}

	// create client certificate template
	clientCRTTemplate := x509.Certificate{
		Signature:          clientCSR.Signature,
		SignatureAlgorithm: clientCSR.SignatureAlgorithm,

		PublicKeyAlgorithm: clientCSR.PublicKeyAlgorithm,
		PublicKey:          clientCSR.PublicKey,

		SerialNumber: big.NewInt(2),
		Issuer:       ca.CAX509.Subject,
		Subject:      clientCSR.Subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0), // 10 years
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	// create client certificate from template and CA public key
	certBytes, err := x509.CreateCertificate(rand.Reader, &clientCRTTemplate, ca.CAX509, clientCRTTemplate.PublicKey, ca.CATLS.PrivateKey)
	if err != nil {
		return err
	}

	// Save the issued certificate to file
	certOut, err := os.Create(path.Join(storagePath, clientCSR.Subject.CommonName+".crt"))
	if err != nil {
		return err
	}
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if err != nil {
		return err
	}
	certOut.Close()
	logrus.Info("Wrote " + clientCSR.Subject.CommonName + ".crt\n")

	ca.Store.UpdateCertificates(ca.GetCertificates())

	return pem.Encode(wr, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
}

func (ca *CA) ExportToDisk(name string, certBytes []byte, privateKey *rsa.PrivateKey) error {
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		os.Mkdir(storagePath, 0644)
	}

	// Write public key
	certOut, err := os.Create(path.Join(storagePath, name+".crt"))
	if err != nil {
		return err
	}
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if err != nil {
		return err
	}
	certOut.Close()
	logrus.Info("Wrote " + name + ".crt\n")

	if privateKey == nil {
		return nil
	}

	// Write private key
	keyOut, err := os.OpenFile(path.Join(storagePath, name+".key"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	err = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	if err != nil {
		return err
	}
	keyOut.Close()
	logrus.Info("Wrote " + name + ".key\n")

	return nil
}

func (ca *CA) GetCertificates() []store.Certificate {

	files, err := ioutil.ReadDir(storagePath)
	if err != nil {
		log.Fatal(err)
	}

	certs := make([]store.Certificate, 0)

	for _, f := range files {
		n := strings.Split(f.Name(), ".")
		if n[len(n)-1] != "crt" {
			continue
		}

		cf, err := ioutil.ReadFile(path.Join(storagePath, f.Name()))
		if err != nil {
			logrus.Warnf("Failed to read certificate %s: %s", f.Name(), err.Error())
			continue
		}

		cpb, _ := pem.Decode(cf)

		crt, err := x509.ParseCertificate(cpb.Bytes)
		if err != nil {
			logrus.Warnf("Failed to read certificate %s: %s", f.Name(), err.Error())
			continue
		}

		sh1 := sha1.Sum(crt.Raw)
		sh256 := sha256.Sum256(crt.Raw)

		cert := store.Certificate{
			Serial:     crt.SerialNumber.String(),
			CommonName: crt.Subject.CommonName,
			IsCA:       crt.IsCA,
			//Usage      []string
			//Revoked    bool
			Issued:  crt.NotBefore,
			Expires: crt.NotAfter,

			Fingerprints: map[string]string{
				"sha1":   hex.EncodeToString(sh1[:]),
				"sha256": hex.EncodeToString(sh256[:]),
			},
		}

		if crt.IsCA {
			cert.Usage = append(cert.Usage, "ca")
		}

		for _, usage := range crt.ExtKeyUsage {
			switch usage {
			case x509.ExtKeyUsageServerAuth:
				cert.Usage = append(cert.Usage, "server")
			case x509.ExtKeyUsageClientAuth:
				cert.Usage = append(cert.Usage, "client")
			}
		}

		certs = append(certs, cert)
	}

	return certs
}

func (ca *CA) WaitForApproval(i, c string, r models.Request) chan bool {
	req := store.Request{
		Identity:   i,
		Connection: c,

		Type:    r.Type,
		Version: r.Version,

		Approved: make(chan bool),
	}

	ca.Store.AddRequest(req)

	return req.Approved
}
