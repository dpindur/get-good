package libgetgood

import (
	"strings"
	"sync"

	. "github.com/dpindur/get-good/logger"
)

type Updater struct {
	running      bool
	wg           *sync.WaitGroup
	haltChan     chan int
	db           *DBConn
	errChan      chan *WorkerError
	requestChan  chan *Request
	responseChan chan *Response
	words        []string
	extensions   []string
}

type Request struct {
	Url string
}

func StartUpdater(wg *sync.WaitGroup, db *DBConn, errChan chan *WorkerError, responseChan chan *Response, words []string, extensions []string) *Updater {
	haltChan := make(chan int)
	requestChan := make(chan *Request)
	updater := &Updater{true, wg, haltChan, db, errChan, requestChan, responseChan, words, extensions}
	wg.Add(1)
	go updater.work()
	return updater
}

func (updater *Updater) EnqueueRequest(req *Request) {
	updater.requestChan <- req
}

func (updater *Updater) Stop() {
	Logger.Debugf("Sending database updater stop signal")
	updater.haltChan <- 0
}

func (updater *Updater) work() {
	defer updater.wg.Done()

	Logger.Debugf("Starting database updater")
	running := true
	for running {
		select {
		case r := <-updater.requestChan:
			err := updater.addURLs(r.Url)
			if err != nil {
				running = false
				updater.errChan <- &WorkerError{"updater", err}
			}
			break
		case r := <-updater.responseChan:
			err := updater.handleResponse(r)
			if err != nil {
				running = false
				updater.errChan <- &WorkerError{"updater", err}
			}
			break
		case <-updater.haltChan:
			running = false
			break
		}
	}
	Logger.Debugf("Database updater stopped")
}

func (updater *Updater) addURLs(baseURL string) error {
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	requests := make([]string, 0)

	for _, word := range updater.words {
		for _, ext := range updater.extensions {
			request := baseURL + word + ext
			requests = append(requests, request)
		}
	}

	return updater.db.AddRequests(requests)
}

func (updater *Updater) handleResponse(res *Response) error {
	if res.Success == false {
		err := updater.db.SetRequestFailed(res.Url)
		return err
	}

	err := updater.db.SetRequestCompleted(res.Url, res.Response.StatusCode)
	if err != nil {
		return err
	}

	// If response is successful, add recursive urls
	if res.Response.StatusCode == 200 {
		Logger.Infof("Successful response for %v", res.Url)
		updater.addURLs(res.Url)
	}

	return nil
}
