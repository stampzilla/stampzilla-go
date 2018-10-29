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

type CA struct {
	X509 *x509.Certificate
	TLS  *tls.Certificate
}

func (ca *CA) LoadOrCreate() error {
	err := ca.Load()
	if err == nil {
		return nil
	}

	return ca.Create()
}

func (ca *CA) Load() error {
	certTLS, err := tls.LoadX509KeyPair("ca.crt", "ca.key")
	if err != nil {
		return err
	}
	certX509, err := x509.ParseCertificate(certTLS.Certificate[0])
	if err != nil {
		return err
	}

	ca.TLS = &certTLS
	ca.X509 = certX509
	return nil
}

func (ca *CA) Create() error {
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
	pub := &priv.PublicKey
	certBytes, err := x509.CreateCertificate(rand.Reader, recipe, recipe, pub, priv)
	if err != nil {
		return err
	}

	crt := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	key := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	certTLS, err := tls.X509KeyPair(crt, key)
	if err != nil {
		return err
	}

	// Write public key
	certOut, err := os.Create("ca.crt")
	if err != nil {
		return err
	}
	_, err = certOut.Write(crt)
	if err != nil {
		return err
	}
	certOut.Close()
	logrus.Info("Wrote ca.crt\n")

	// Write private key
	keyOut, err := os.OpenFile("ca.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	_, err = keyOut.Write(key)
	if err != nil {
		return err
	}
	keyOut.Close()
	logrus.Info("Wrote ca.key\n")

	certX509, err := x509.ParseCertificate(certTLS.Certificate[0])
	if err != nil {
		return err
	}

	ca.TLS = &certTLS
	ca.X509 = certX509

	return nil
}
