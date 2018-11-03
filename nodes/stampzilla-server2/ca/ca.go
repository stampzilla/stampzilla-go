package ca

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type CA struct {
	X509   *x509.Certificate
	TLS    *tls.Certificate
	CAX509 *x509.Certificate
	CATLS  *tls.Certificate
}

func LoadOrCreate(name string) (*CA, error) {
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
	certTLS, err := tls.LoadX509KeyPair(name+".crt", name+".key")
	if err != nil {
		return err
	}
	certX509, err := x509.ParseCertificate(certTLS.Certificate[0])
	if err != nil {
		return err
	}

	if name == "" {
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
		SerialNumber: big.NewInt(1653),
		Subject: pkix.Name{
			Organization: []string{"stampzilla-go"},
			CommonName:   "stampzilla-go CA",
			//Country:       []string{"se"},
			//Province:      []string{"PROVINCE"},
			//Locality:      []string{"CITY"},
			//StreetAddress: []string{"ADDRESS"},
			//PostalCode:    []string{"POSTAL_CODE"},
		},
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

func (ca *CA) CreateCertificate(name string) error {
	recipe := &x509.Certificate{
		SerialNumber: big.NewInt(1653),
		Subject: pkix.Name{
			Organization: []string{"stampzilla-go"},
			CommonName:   name,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10 years
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		//IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames: []string{"localhost"},
	}

	// Generate keys
	priv, _ := rsa.GenerateKey(rand.Reader, 2048) // key size
	certBytes, err := x509.CreateCertificate(rand.Reader, recipe, recipe, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	err = ca.ExportToDisk(name, certBytes, priv)
	if err != nil {
		return err
	}

	return ca.Load(name)
}

func (ca *CA) CreateCertificateFromRequest(wr io.Writer, name string, request []byte) error {
	pemBlock, _ := pem.Decode(request)
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
	return pem.Encode(wr, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
}

func (ca *CA) ExportToDisk(name string, certBytes []byte, privateKey *rsa.PrivateKey) error {

	// Write public key
	certOut, err := os.Create(name + ".crt")
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
	keyOut, err := os.OpenFile(name+".key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
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
