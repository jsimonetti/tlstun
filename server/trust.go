package server

import (
	"bytes"
	"crypto/sha512"
	"crypto/x509"
	"fmt"
	"net/http"

	"github.com/boltdb/bolt"
)

func (s *server) isTrusted(r *http.Request) bool {
	if r.TLS == nil {
		return false
	}
	for i := range r.TLS.PeerCertificates {
		if s.checkTrustState(r.TLS.PeerCertificates[i]) {
			return true
		}
	}
	return false
}

func (s *server) checkTrustState(cert *x509.Certificate) bool {
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("certificates"))
		if b == nil {
			return fmt.Errorf("Untrusted")
		}
		v := b.Get(certGenerateFingerprint(cert))
		if bytes.Compare(cert.Raw, v) == 0 {
			s.log.Print("Found trusted certificate")
			return nil
		}
		s.log.Print("Client certificate is not trusted")
		return fmt.Errorf("Untrusted")
	})

	if err == nil {
		return true
	}
	return false
}

func (s *server) saveCert(cert *x509.Certificate) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("certificates"))
		if err != nil {
			return err
		}
		return b.Put(certGenerateFingerprint(cert), cert.Raw)
	})
	return err
}

func (s *server) passwordCheck(password string) bool {
	// No password set
	if s.registerPass == "" {
		return false
	}

	if !bytes.Equal([]byte(password), []byte(s.registerPass)) {
		s.log.Print("Bad password received")
		return false
	}
	s.log.Print("Verified the admin password")

	return true
}

func certGenerateFingerprint(cert *x509.Certificate) []byte {
	sum := sha512.Sum512(cert.Raw)
	return sum[:]
}

func TrustedResponse() string {
	return "It Works and you have a trusted cert!"
}

func UnTrustedResponse() string {
	return "It Works!"
}
