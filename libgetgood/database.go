package libgetgood

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"sync"
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
	_, err := conn.db.Exec("CREATE TABLE IF NOT EXISTS requests (id INTEGER PRIMARY KEY ASC, status INTEGER, uri TEXT, httpStatus INTEGER)")
	return err
}

func (conn *DBConn) Clear() error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	_, err := conn.db.Exec("DELETE FROM requests")
	return err
}

func (conn *DBConn) AddRequest(uri string) error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	_, err := conn.db.Exec("INSERT INTO requests (status, uri) VALUES (?, ?)", Unprocessed, uri)
	return err
}

func (conn *DBConn) RequestExists(uri string) (bool, error) {
	var tmp string
	err := conn.db.QueryRow("SELECT uri FROM requests WHERE uri = ?", uri).Scan(&tmp)
	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}

func (conn *DBConn) GetIncompleteRequests() ([]string, error) {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	rows, err := conn.db.Query("SELECT uri FROM requests WHERE status = ? LIMIT 50", Unprocessed)
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

func (conn *DBConn) SetRequestInflight(uri string) error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	_, err := conn.db.Exec("UPDATE requests SET status = ? WHERE uri = ?", Inflight, uri)
	return err
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
