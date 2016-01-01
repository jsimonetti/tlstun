package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"net/http"

	log "gopkg.in/inconshreveable/log15.v2"
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

func readSavedClientCAList(d *Daemon) {
	d.clientCerts = []x509.Certificate{}

	dbCerts, err := dbCertsGet(d.db)
	if err != nil {
		d.log.Error("Error reading certificates from database", log.Ctx{"error": err})
		return
	}

	for _, dbCert := range dbCerts {
		certBlock, _ := pem.Decode([]byte(dbCert.Certificate))
		cert, err := x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			d.log.Error("Error reading certificate", log.Ctx{"name": dbCert.Name, "error": err})
			continue
		}
		d.clientCerts = append(d.clientCerts, *cert)
	}
}

func certGenerateFingerprint(cert *x509.Certificate) string {
	return fmt.Sprintf("%x", sha256.Sum256(cert.Raw))
}

func saveCert(d *Daemon, host string, cert *x509.Certificate) error {
	baseCert := new(dbCertInfo)
	baseCert.Fingerprint = certGenerateFingerprint(cert)
	baseCert.Type = 1
	baseCert.Name = host
	baseCert.Certificate = string(
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}),
	)

	return dbCertSave(d.db, baseCert)
}

func (d *Daemon) isTrustedClient(r *http.Request) bool {
	if r.TLS == nil {
		return false
	}
	for i := range r.TLS.PeerCertificates {
		if d.CheckTrustState(*r.TLS.PeerCertificates[i]) {
			return true
		}
	}
	return false
}

func (d *Daemon) CheckTrustState(cert x509.Certificate) bool {
	for _, v := range d.clientCerts {
		if bytes.Compare(cert.Raw, v.Raw) == 0 {
			d.log.Debug("Found cert")
			return true
		}
		d.log.Debug("Client cert != key")
	}
	return false
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

func PasswordCheck(d *Daemon, password string) bool {
	// No password set
	if registerPass == "" {
		return false
	}

	if !bytes.Equal([]byte(password), []byte(registerPass)) {
		d.log.Error("Bad password received")
		return false
	}
	d.log.Info("Verified the admin password")
	return true
}
