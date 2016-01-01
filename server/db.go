package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/mattn/go-sqlite3"
	log "gopkg.in/inconshreveable/log15.v2"
)

const DB_CURRENT_VERSION int = 1

const CURRENT_SCHEMA string = `
CREATE TABLE IF NOT EXISTS certificates (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    fingerprint VARCHAR(255) NOT NULL,
    type INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    certificate TEXT NOT NULL,
    UNIQUE (fingerprint)
);
CREATE TABLE IF NOT EXISTS schema (
id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
version INTEGER NOT NULL,
updated_at DATETIME NOT NULL,
UNIQUE (version)
);`

func createDb(db *sql.DB) (err error) {
	_, err = db.Exec(CURRENT_SCHEMA)
	if err != nil {
		return err
	}

	// To make the schema creation indempotent, only insert the schema version
	// if there isn't one already.
	latestVersion := dbGetSchema(db)

	if latestVersion == 0 {
		// There isn't an entry for schema version, let's put it in.
		insertStmt := `INSERT INTO schema (version, updated_at) values (?, strftime("%s"));`
		_, err = db.Exec(insertStmt, DB_CURRENT_VERSION)
		if err != nil {
			return err
		}
	}

	return nil
}

func dbGetSchema(db *sql.DB) (v int) {
	arg1 := []interface{}{}
	arg2 := []interface{}{&v}
	q := "SELECT max(version) FROM schema"
	err := dbQueryRowScan(db, q, arg1, arg2)
	if err != nil {
		return 0
	}
	return v
}

// Create a database connection object and return it.
func initializeDbObject(d *Daemon, path string) error {
	var openPath string
	var err error

	timeout := 5 // TODO - make this command-line configurable?

	// These are used to tune the transaction BEGIN behavior instead of using the
	// similar "locking_mode" pragma (locking for the whole database connection).
	openPath = fmt.Sprintf("%s?_busy_timeout=%d&_txlock=exclusive", path, timeout*1000)

	// Open the database. If the file doesn't exist it is created.
	db, err := sql.Open("sqlite3", openPath)
	if err != nil {
		return err
	}

	// Table creation is indempotent, run it every time
	err = createDb(db)
	if err != nil {
		return fmt.Errorf("Error creating database: %s", err)
	}

	// Run PRAGMA statements now since they are *per-connection*.
	db.Exec("PRAGMA foreign_keys=ON;") // This allows us to use ON DELETE CASCADE

	v := dbGetSchema(db)

	if v != DB_CURRENT_VERSION {
		err = dbUpdate(db, v)
		if err != nil {
			return err
		}
	}

	d.db = db
	return nil
}

func dbQueryRowScan(db *sql.DB, q string, args []interface{}, outargs []interface{}) error {
	for {
		err := db.QueryRow(q, args...).Scan(outargs...)
		if err == nil {
			return nil
		}
		if isNoMatchError(err) {
			return err
		}
		if !isDbLockedError(err) {
			return err
		}
		d.log.Error("DbQueryRowScan", log.Ctx{"query": q, "args": args, "reason": "DB was locked"})
		time.Sleep(1 * time.Second)
	}
}

func isNoMatchError(err error) bool {
	if err == nil {
		return false
	}
	if err.Error() == "sql: no rows in result set" {
		return true
	}
	return false
}

func isDbLockedError(err error) bool {
	if err == nil {
		return false
	}
	if err == sqlite3.ErrLocked || err == sqlite3.ErrBusy {
		return true
	}
	if err.Error() == "database is locked" {
		return true
	}
	return false
}

func dbBegin(db *sql.DB) (*sql.Tx, error) {
	for {
		tx, err := db.Begin()
		if err == nil {
			return tx, nil
		}
		if !isDbLockedError(err) {
			d.log.Error("DbBegin: error", log.Ctx{"error": err})
			return nil, err
		}
		d.log.Debug("DB was locked")
		time.Sleep(1 * time.Second)
	}
}

func txCommit(tx *sql.Tx) error {
	for {
		err := tx.Commit()
		if err == nil {
			return nil
		}
		if !isDbLockedError(err) {
			d.log.Error("Txcommit: error", log.Ctx{"error": err})
			return err
		}
		d.log.Debug("Txcommit: db was locked")
		time.Sleep(1 * time.Second)
	}
}

func dbQuery(db *sql.DB, q string, args ...interface{}) (*sql.Rows, error) {
	for {
		result, err := db.Query(q, args...)
		if err == nil {
			return result, nil
		}
		if !isDbLockedError(err) {
			d.log.Error("DbQuery error", log.Ctx{"query": q, "error": err})
			return nil, err
		}
		d.log.Debug("DbQuery", log.Ctx{"query": q, "args": args, "reason": "DB was locked"})
		time.Sleep(1 * time.Second)
	}
}

func dbUpdate(db *sql.DB, prevVersion int) error {
	if prevVersion < 0 || prevVersion > DB_CURRENT_VERSION {
		return fmt.Errorf("Bad database version: %d", prevVersion)
	}
	if prevVersion == DB_CURRENT_VERSION {
		return nil
	}

	return nil
}
