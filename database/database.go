// /home/krylon/go/src/github.com/blicero/uptimed/database/database.go
// -*- mode: go; coding: utf-8; -*-
// Created on 01. 06. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-05 20:44:53 krylon>

// Package database provides the persistence layer of the application.
// Internally it uses an SQLite database, but the methods it exposes are
// high-level operations that are database-agnostic, and the rest of the
// application is not affected by the specific database engine used.
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/blicero/krylib"
	"github.com/blicero/uptimed/common"
	"github.com/blicero/uptimed/database/query"
	"github.com/blicero/uptimed/logdomain"
	_ "github.com/mattn/go-sqlite3" // Import the database driver
)

var (
	openLock sync.Mutex
	idCnt    int64
)

// ErrInvalidValue indicates that one or more parameters passed to a method
// had values that are invalid for that operation.
var ErrInvalidValue = errors.New("Invalid value for parameter")

// ErrNotFound indicates that a search operation has not yielded any results.
var ErrNotFound = errors.New("Nothing was found")

// If a query returns an error and the error text is matched by this regex, we
// consider the error as transient and try again after a short delay.
var retryPat = regexp.MustCompile("(?i)database is (?:locked|busy)")

// worthARetry returns true if an error returned from the database
// is matched by the retryPat regex.
func worthARetry(e error) bool {
	return retryPat.MatchString(e.Error())
} // func worthARetry(e error) bool

// retryDelay is the amount of time we wait before we repeat a database
// operation that failed due to a transient error.
const retryDelay = 25 * time.Millisecond

func waitForRetry() {
	time.Sleep(retryDelay)
} // func waitForRetry()

// Database wraps the connection to the underlying data store and
// associated state.
type Database struct {
	id      int64
	db      *sql.DB
	log     *log.Logger
	path    string
	queries map[query.ID]*sql.Stmt
	hosts   map[string]int64
}

// Open opens a Database. If the database specified by the path does not exist,
// yet, it is created and initialized.
func Open(path string) (*Database, error) {
	var (
		err      error
		dbExists bool
		db       = &Database{
			path:    path,
			queries: make(map[query.ID]*sql.Stmt),
			hosts:   make(map[string]int64),
		}
	)

	openLock.Lock()
	defer openLock.Unlock()
	idCnt++
	db.id = idCnt

	if db.log, err = common.GetLogger(logdomain.Database); err != nil {
		return nil, err
	} else if common.Debug {
		db.log.Printf("[DEBUG] Open database %s\n", path)
	}

	var connstring = fmt.Sprintf("%s?_locking=NORMAL&_journal=WAL&_fk=1&recursive_triggers=0",
		path)

	if dbExists, err = krylib.Fexists(path); err != nil {
		db.log.Printf("[ERROR] Failed to check if %s already exists: %s\n",
			path,
			err.Error())
		return nil, err
	} else if db.db, err = sql.Open("sqlite3", connstring); err != nil {
		db.log.Printf("[ERROR] Failed to open %s: %s\n",
			path,
			err.Error())
		return nil, err
	}

	if !dbExists {
		if err = db.initialize(); err != nil {
			var e2 error
			if e2 = db.db.Close(); e2 != nil {
				db.log.Printf("[CRITICAL] Failed to close database: %s\n",
					e2.Error())
				return nil, e2
			} else if e2 = os.Remove(path); e2 != nil {
				db.log.Printf("[CRITICAL] Failed to remove database file %s: %s\n",
					db.path,
					e2.Error())
			}
			return nil, err
		}
		db.log.Printf("[INFO] Database at %s has been initialized\n",
			path)
	}

	return db, nil
} // func Open(path string) (*Database, error)

func (db *Database) initialize() error {
	var err error
	var tx *sql.Tx

	if common.Debug {
		db.log.Printf("[DEBUG] Initialize fresh database at %s\n",
			db.path)
	}

	if tx, err = db.db.Begin(); err != nil {
		db.log.Printf("[ERROR] Cannot begin transaction: %s\n",
			err.Error())
		return err
	}

	for _, q := range qInit {
		db.log.Printf("[TRACE] Execute init query:\n%s\n",
			q)
		if _, err = tx.Exec(q); err != nil {
			db.log.Printf("[ERROR] Cannot execute init query: %s\n%s\n",
				err.Error(),
				q)
			if rbErr := tx.Rollback(); rbErr != nil {
				db.log.Printf("[CANTHAPPEN] Cannot rollback transaction: %s\n",
					rbErr.Error())
				return rbErr
			}
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		db.log.Printf("[CANTHAPPEN] Failed to commit init transaction: %s\n",
			err.Error())
		return err
	}

	return nil
} // func (db *Database) initialize() error

// Close closes the database.
// If there is a pending transaction, it is rolled back.
func (db *Database) Close() error {
	// I wonder if would make more snese to panic() if something goes wrong

	var err error

	for key, stmt := range db.queries {
		if err = stmt.Close(); err != nil {
			db.log.Printf("[CRITICAL] Cannot close statement handle %s: %s\n",
				key,
				err.Error())
			return err
		}
		delete(db.queries, key)
	}

	if err = db.db.Close(); err != nil {
		db.log.Printf("[CRITICAL] Cannot close database: %s\n",
			err.Error())
	}

	db.db = nil
	return nil
} // func (db *Database) Close() error

func (db *Database) getQuery(id query.ID) (*sql.Stmt, error) {
	var (
		stmt  *sql.Stmt
		found bool
		err   error
	)

	if stmt, found = db.queries[id]; found {
		return stmt, nil
	} else if _, found = qDB[id]; !found {
		return nil, fmt.Errorf("Unknown Query %d",
			id)
	}

	db.log.Printf("[TRACE] Prepare query %s\n", id)

PREPARE_QUERY:
	if stmt, err = db.db.Prepare(qDB[id]); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto PREPARE_QUERY
		}

		db.log.Printf("[ERROR] Cannor parse query %s: %s\n%s\n",
			id,
			err.Error(),
			qDB[id])
		return nil, err
	}

	db.queries[id] = stmt
	return stmt, nil
} // func (db *Database) getQuery(query.ID) (*sql.Stmt, error)

// PerformMaintenance performs some maintenance operations on the database.
// It cannot be called while a transaction is in progress and will block
// pretty much all access to the database while it is running.
func (db *Database) PerformMaintenance() error {
	var mQueries = []string{
		"PRAGMA wal_checkpoint(TRUNCATE)",
		"VACUUM",
		"REINDEX",
		"ANALYZE",
	}
	var err error

	for _, q := range mQueries {
		if _, err = db.db.Exec(q); err != nil {
			db.log.Printf("[ERROR] Failed to execute %s: %s\n",
				q,
				err.Error())
		}
	}

	return nil
} // func (db *Database) PerformMaintenance() error

// HostGetID looks up the ID for the given hostname.
func (db *Database) HostGetID(name string) (int64, error) {
	const qid query.ID = query.HostGetID

	if hid, ok := db.hosts[name]; ok {
		return hid, nil
	}

	var (
		err  error
		stmt *sql.Stmt
		rows *sql.Rows
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return -1, err
	}

EXEC_QUERY:
	if rows, err = stmt.Query(name); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return -1, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	if rows.Next() {
		var id int64

		if err = rows.Scan(&id); err != nil {
			db.log.Printf("[ERROR] Cannot extract value from row: %s\n",
				err.Error())
			return -1, err
		}

		db.hosts[name] = id
		return id, nil
	}

	return -1, ErrNotFound
} // func (db *Database) HostGetID(name string) (int64, error)

// HostGetAll loads all hosts from the database.
func (db *Database) HostGetAll() ([]common.Host, error) {
	const qid query.ID = query.HostGetAll
	var (
		err  error
		stmt *sql.Stmt
		rows *sql.Rows
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return nil, err
	}

EXEC_QUERY:
	if rows, err = stmt.Query(); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	var hosts = make([]common.Host, 0)

	for rows.Next() {
		var h common.Host

		if err = rows.Scan(&h.ID, &h.Name); err != nil {
			db.log.Printf("[ERROR] Cannot extract values from row: %s\n",
				err.Error())
			return nil, err
		}

		db.hosts[h.Name] = h.ID
		hosts = append(hosts, h)
	}

	return hosts, nil
} // func (db *Database) HostGetAll() ([]common.Host, error)

// FIXME: Couldn't we use an upsert here to make the method idempotent?

// HostAdd adds the given hostname to the database and, if successful, returns
// the Host's ID.
func (db *Database) HostAdd(name string) (int64, error) {
	const qid query.ID = query.HostAdd
	var (
		err  error
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return 0, err
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(name); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return 0, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	if rows.Next() {
		var id int64

		if err = rows.Scan(&id); err != nil {
			db.log.Printf("[ERROR] Cannot extract return value from row: %s\n",
				err.Error())
			return 0, err
		}

		db.hosts[name] = id
		return id, err
	}

	// CANTHAPPEN
	return 0, errors.New("Something went wrong")
} // func (db *Database) HostAdd(name string) (int64, error)

// RecordAdd adds a new Record to the database.
func (db *Database) RecordAdd(r *common.Record) error {
	const qid query.ID = query.RecordAdd
	var (
		err    error
		stmt   *sql.Stmt
		hostID int64
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return err
	} else if hostID, err = db.HostGetID(r.Hostname); err != nil {
		if errors.Is(err, ErrNotFound) {
			if hostID, err = db.HostAdd(r.Hostname); err != nil {
				return err
			}
		} else {
			db.log.Printf("[ERROR] Cannot query ID for Host %s: %s\n",
				r.Hostname,
				err.Error())
			return err
		}
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(hostID, r.Timestamp.Unix(), r.Uptime.Seconds(), r.Load[0], r.Load[1], r.Load[2]); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return err
	}

	defer rows.Close() // nolint: errcheck,gosec

	rows.Next()

	if err = rows.Scan(&r.ID); err != nil {
		db.log.Printf("[ERROR] Cannot extract Record ID: %s\n",
			err.Error())
		return err
	}

	return nil
} // func (db *Database) RecordAdd(r *common.Record) error

// RecordGetByPeriod returns all Records for the given period.
func (db *Database) RecordGetByPeriod(t1, t2 time.Time) ([]common.Record, error) {
	const qid query.ID = query.RecordGetByPeriod
	var (
		err  error
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return nil, err
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(t1.Unix(), t2.Unix()); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	var records = make([]common.Record, 0)

	for rows.Next() {
		var (
			stamp, uptime int64
			r             common.Record
		)

		if err = rows.Scan(&r.ID, &r.Hostname, &stamp, &uptime, &r.Load[0], &r.Load[1], &r.Load[2]); err != nil {
			db.log.Printf("[ERROR] Cannot extract values from Rows: %s\n",
				err.Error())
			return nil, err
		}

		r.Timestamp = time.Unix(stamp, 0)
		r.Uptime = time.Second * time.Duration(uptime)

		records = append(records, r)
	}

	return records, nil
} // func (db *Database) RecordGetByPeriod(t1, t2 time.Time) ([]common.Record, error)

// RecordGetByHost loads all Records for the given Host
func (db *Database) RecordGetByHost(name string, begin time.Time) ([]common.Record, error) {
	const qid query.ID = query.RecordGetByHost
	var (
		err    error
		stmt   *sql.Stmt
		hostID int64
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return nil, err
	} else if hostID, err = db.HostGetID(name); err != nil {
		db.log.Printf("[ERROR] Unknown Host %s: %s\n",
			name,
			err.Error())
		return nil, err
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(hostID, begin.Unix()); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	var records = make([]common.Record, 0)

	for rows.Next() {
		var (
			stamp, uptime int64
			r             = common.Record{Hostname: name}
		)

		if err = rows.Scan(&r.ID, &stamp, &uptime, &r.Load[0], &r.Load[1], &r.Load[2]); err != nil {
			db.log.Printf("[ERROR] Cannot extract values from Rows: %s\n",
				err.Error())
			return nil, err
		}

		r.Timestamp = time.Unix(stamp, 0)
		r.Uptime = time.Second * time.Duration(uptime)

		records = append(records, r)
	}

	return records, nil
} // func (db *Database) RecordGetByHost(name string) ([]common.Record, error)

// RecentGetAll returns the most recent Record per Host
func (db *Database) RecentGetAll() ([]common.Record, error) {
	const qid query.ID = query.RecentGetAll
	var (
		err  error
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return nil, err
	}

	var rows *sql.Rows
EXEC_QUERY:
	if rows, err = stmt.Query(); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	var records = make([]common.Record, 0)

	for rows.Next() {
		var (
			stamp, uptime int64
			r             common.Record
		)

		if err = rows.Scan(&r.ID, &r.Hostname, &stamp, &uptime, &r.Load[0], &r.Load[1], &r.Load[2]); err != nil {
			db.log.Printf("[ERROR] Cannot extract values from Rows: %s\n",
				err.Error())
			return nil, err
		}

		r.Timestamp = time.Unix(stamp, 0)
		r.Uptime = time.Second * time.Duration(uptime)

		records = append(records, r)
	}

	return records, nil
} // func (db *Database) RecentGetAll() ([]common.Record, error)
