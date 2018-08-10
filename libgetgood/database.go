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
	_, err := conn.db.Exec("CREATE TABLE IF NOT EXISTS requests (id INTEGER PRIMARY KEY ASC, status INTEGER, uri TEXT)")
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
