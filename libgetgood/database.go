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
	_, err := conn.db.Exec("CREATE TABLE IF NOT EXISTS requests (id INTEGER PRIMARY KEY ASC, status INTEGER, uri TEXT, path TEXT)")
	return err
}

func (conn *DBConn) Clear() error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	_, err := conn.db.Exec("DELETE FROM requests")
	return err
}

func (conn *DBConn) AddRequest(uri string, path string) error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	_, err := conn.db.Exec("INSERT INTO requests (status, uri, path) VALUES (?, ?, ?)", Unprocessed, uri, path)
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
