package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type Cert struct {
	CN     string
	X509   *x509.Certificate
	TLS    *tls.Certificate
	Signed bool
}

func LoadOrCreate(CN string, ca *CA) (*Cert, error) {
	cert := &Cert{CN: CN}
	return cert, cert.LoadOrCreate(ca)
}

func (cert *Cert) LoadOrCreate(ca *CA) error {
	err := cert.Load()
	if err == nil {
		cert.Signed = true
		return nil
	}

	return cert.Create(ca)
}

func (cert *Cert) Load() error {
	certTLS, err := tls.LoadX509KeyPair(cert.CN+".crt", cert.CN+".key")
	if err != nil {
		return err
	}
	certX509, err := x509.ParseCertificate(certTLS.Certificate[0])
	if err != nil {
		return err
	}

	cert.TLS = &certTLS
	cert.X509 = certX509
	return nil
}

func (cert *Cert) Create(ca *CA) error {
	// Create a 10year cert
	recipe := &x509.Certificate{
		SerialNumber: big.NewInt(1653),
		Subject: pkix.Name{
			Organization: []string{"stampzilla-go"},
			CommonName:   cert.CN,
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
	pub := &priv.PublicKey
	certBytes, err := x509.CreateCertificate(rand.Reader, recipe, ca.X509, pub, priv)
	if err != nil {
		return err
	}

	// Write public key
	certOut, err := os.Create(cert.CN + ".crt")
	if err != nil {
		return err
	}
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if err != nil {
		return err
	}
	certOut.Close()
	logrus.Info("Wrote " + cert.CN + ".crt\n")

	// Write private key
	keyOut, err := os.OpenFile(cert.CN+".key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	err = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	if err != nil {
		return err
	}
	keyOut.Close()
	logrus.Info("Wrote " + cert.CN + ".key\n")

	return cert.Load()
}
