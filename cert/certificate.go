package cert

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"os"
	"time"
)

type Certificate struct {
	x509.Certificate
	Bytes []byte
	Key   *rsa.PrivateKey
}

func (c *Certificate) FromFile(certf, keyf string) error {
	var err error
	c.Bytes, err = ioutil.ReadFile(certf)
	if err != nil {
		return err
	}
	keyBytes, err := ioutil.ReadFile(keyf)
	if err != nil {
		return err
	}

	certpem, _ := pem.Decode(c.Bytes)
	cert, err := x509.ParseCertificate(certpem.Bytes)
	if err != nil {
		return err
	}
	c.Certificate = *cert

	keypem, _ := pem.Decode(keyBytes)
	c.Key, err = x509.ParsePKCS1PrivateKey(keypem.Bytes)
	return err
}

func (c *Certificate) CertString() string {
	certOut := bytes.NewBufferString("")
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: c.Bytes})
	return certOut.String()
}

func (c *Certificate) CertToFile(filename string) error {
	certOut, err := os.OpenFile(filename, os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer certOut.Close()
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: c.Bytes})
	return err
}

func (c *Certificate) KeyString() string {
	keyOut := bytes.NewBufferString("")
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(c.Key)})
	return keyOut.String()
}

func (c *Certificate) KeyToFile(filename string) error {
	keyOut, err := os.OpenFile(filename, os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer keyOut.Close()
	err = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(c.Key)})
	return err
}

func CreateCaCertificate() (cert Certificate, err error) {
	validFrom := time.Now()
	validTo := validFrom.Add(10 * 365 * 24 * time.Hour) // 10 years valid

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return
	}

	cert.Certificate = x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"TLSTun"},
			CommonName:   "TLSTun Server CA",
		},
		NotBefore:             validFrom,
		NotAfter:              validTo,
		BasicConstraintsValid: true,
		IsCA:        true,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	if cert.Key, err = rsa.GenerateKey(rand.Reader, 4096); err != nil {
		return
	}
	cert.Bytes, err = x509.CreateCertificate(rand.Reader, &cert.Certificate, &cert.Certificate, &cert.Key.PublicKey, cert.Key)
	return cert, err
}

func CreateServerCertificate(ca Certificate, name string) (cert Certificate, err error) {
	validFrom := time.Now()
	validTo := validFrom.Add(10 * 365 * 24 * time.Hour) // 10 years valid

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return
	}

	cert.Certificate = x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"TLSTun"},
			CommonName:   name,
		},
		NotBefore:             validFrom,
		NotAfter:              validTo,
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	if cert.Key, err = rsa.GenerateKey(rand.Reader, 4096); err != nil {
		return
	}

	cert.Bytes, err = x509.CreateCertificate(rand.Reader, &cert.Certificate, &ca.Certificate, &cert.Key.PublicKey, ca.Key)
	return cert, err
}

func CreateClientCertificate(ca Certificate, name string) (cert Certificate, err error) {
	validFrom := time.Now()
	validTo := validFrom.Add(10 * 365 * 24 * time.Hour) // 10 years valid

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return
	}

	cert.Certificate = x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"TLSTun"},
			CommonName:   name,
		},
		NotBefore:             validFrom,
		NotAfter:              validTo,
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	if cert.Key, err = rsa.GenerateKey(rand.Reader, 4096); err != nil {
		return
	}

	cert.Bytes, err = x509.CreateCertificate(rand.Reader, &cert.Certificate, &ca.Certificate, &cert.Key.PublicKey, ca.Key)
	return cert, err
}
