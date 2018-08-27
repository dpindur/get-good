package libgetgood

import (
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type RequestStatus int

const (
	Unprocessed RequestStatus = 0
	Inflight    RequestStatus = 1
	Failed      RequestStatus = 2
	Processed   RequestStatus = 3
)

type DBConn struct {
	db    *sql.DB
	mutex *sync.Mutex
}

func (conn *DBConn) CreateSchema() error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	_, err := conn.db.Exec("CREATE TABLE IF NOT EXISTS requests (id INTEGER PRIMARY KEY ASC, status INTEGER, uri TEXT, httpStatus INTEGER, UNIQUE(uri))")
	return err
}

func (conn *DBConn) Clear() error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	_, err := conn.db.Exec("DELETE FROM requests")
	return err
}

func (conn *DBConn) AddRequests(requests []string) error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	insertURI, err := conn.db.Prepare("INSERT OR IGNORE INTO requests (status, uri) VALUES (?, ?)")
	if err != nil {
		return err
	}

	tx, err := conn.db.Begin()
	if err != nil {
		return err
	}

	for _, request := range requests {
		_, err := tx.Stmt(insertURI).Exec(Unprocessed, request)
		if err != nil {
			return tx.Rollback()
		}
	}

	return tx.Commit()
}

func (conn *DBConn) GetIncompleteRequests(batchSize int) ([]string, error) {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	rows, err := conn.db.Query("SELECT uri FROM requests WHERE status = ? LIMIT ?", Unprocessed, batchSize)
	if err != nil {
		return nil, err
	}

	requests := make([]string, 0)

	defer rows.Close()
	for rows.Next() {
		var uri string
		err = rows.Scan(&uri)
		if err != nil {
			return nil, err
		}
		requests = append(requests, uri)
	}

	return requests, nil
}

func (conn *DBConn) SetRequestsInflight(requests []string) error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	updateURI, err := conn.db.Prepare("UPDATE requests SET status = ? WHERE uri = ?")
	if err != nil {
		return err
	}

	tx, err := conn.db.Begin()
	if err != nil {
		return err
	}

	for _, request := range requests {
		_, err := tx.Stmt(updateURI).Exec(Inflight, request)
		if err != nil {
			return tx.Rollback()
		}
	}

	return tx.Commit()
}

func (conn *DBConn) SetRequestFailed(uri string) error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	_, err := conn.db.Exec("UPDATE requests SET status = ? WHERE uri = ?", Failed, uri)
	return err
}

func (conn *DBConn) SetRequestCompleted(uri string, httpStatus int) error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	_, err := conn.db.Exec("UPDATE requests SET status = ?, httpStatus = ? WHERE uri = ?", Processed, httpStatus, uri)
	return err
}

func (conn *DBConn) ResetInflightRequests() error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	_, err := conn.db.Exec("UPDATE requests SET status = ? WHERE status = ?", Unprocessed, Inflight)
	return err
}

func (conn *DBConn) ResetFailedRequests() error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	_, err := conn.db.Exec("UPDATE requests SET status = ? WHERE status = ?", Unprocessed, Failed)
	return err
}

func OpenDatabaseConnection(filename string) (*DBConn, error) {
	db, err := sql.Open("sqlite3", filename)
	mutex := &sync.Mutex{}
	return &DBConn{db, mutex}, err
}

func (conn *DBConn) CloseDatabaseConnection() error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	return conn.db.Close()
}
