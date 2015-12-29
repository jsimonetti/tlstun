package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"net/http"

	"github.com/jsimonetti/tlstun/shared"
)

// dbCertInfo is here to pass the certificates content
// from the database around
type dbCertInfo struct {
	ID          int
	Fingerprint string
	Type        int
	Name        string
	Certificate string
}

var clientCerts []x509.Certificate

func readSavedClientCAList() {
	clientCerts = []x509.Certificate{}

	dbCerts, err := dbCertsGet(db)
	if err != nil {
		shared.Log("daemon", "error", fmt.Sprintf("Error reading certificates from database: %s", err))
		return
	}

	for _, dbCert := range dbCerts {
		certBlock, _ := pem.Decode([]byte(dbCert.Certificate))
		cert, err := x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			shared.Log("daemon", "error", fmt.Sprintf("Error reading certificate for %s: %s", dbCert.Name, err))
			continue
		}
		clientCerts = append(clientCerts, *cert)
	}
}

func certGenerateFingerprint(cert *x509.Certificate) string {
	return fmt.Sprintf("%x", sha256.Sum256(cert.Raw))
}

func saveCert(host string, cert *x509.Certificate) error {
	baseCert := new(dbCertInfo)
	baseCert.Fingerprint = certGenerateFingerprint(cert)
	baseCert.Type = 1
	baseCert.Name = host
	baseCert.Certificate = string(
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}),
	)

	return dbCertSave(db, baseCert)
}

func isTrustedClient(r *http.Request) bool {
	if r.TLS == nil {
		return false
	}
	for i := range r.TLS.PeerCertificates {
		if checkTrustState(*r.TLS.PeerCertificates[i]) {
			return true
		}
	}
	return false
}

func checkTrustState(cert x509.Certificate) bool {
	for _, v := range clientCerts {
		if bytes.Compare(cert.Raw, v.Raw) == 0 {
			shared.Log("daemon", "debug", "Found cert")
			return true
		}
		shared.Log("daemon", "debug", "Client cert != key")
	}
	return false
}

func readMyCert() (string, string, error) {
	certf := "server.crt"
	keyf := "server.key"
	shared.Log("daemon", "info", fmt.Sprintf("Looking for existing certificates cert: %s, key: %s", certf, keyf))

	err := shared.FindOrGenCert(certf, keyf)

	return certf, keyf, err
}

// dbCertsGet returns all certificates from the DB as CertBaseInfo objects.
func dbCertsGet(db *sql.DB) (certs []*dbCertInfo, err error) {
	rows, err := dbQuery(
		db,
		"SELECT id, fingerprint, type, name, certificate FROM certificates",
	)
	if err != nil {
		return certs, err
	}

	defer rows.Close()

	for rows.Next() {
		cert := new(dbCertInfo)
		rows.Scan(
			&cert.ID,
			&cert.Fingerprint,
			&cert.Type,
			&cert.Name,
			&cert.Certificate,
		)
		certs = append(certs, cert)
	}

	return certs, nil
}

// dbCertSave stores a CertBaseInfo object in the db,
// it will ignore the ID field from the dbCertInfo.
func dbCertSave(db *sql.DB, cert *dbCertInfo) error {
	tx, err := dbBegin(db)
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
			INSERT INTO certificates (
				fingerprint,
				type,
				name,
				certificate
			) VALUES (?, ?, ?, ?)`,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		cert.Fingerprint,
		cert.Type,
		cert.Name,
		cert.Certificate,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	return txCommit(tx)
}

func PasswordCheck(password string) bool {
	// No password set
	if registerPass == "" {
		return false
	}

	if !bytes.Equal([]byte(password), []byte(registerPass)) {
		shared.Log("daemon", "error", fmt.Sprintf("Bad password received: %s != %s", password, registerPass))
		return false
	}
	shared.Log("daemon", "info", "Verified the admin password")
	return true
}
