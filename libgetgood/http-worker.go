package libgetgood

import (
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	. "github.com/dpindur/get-good/logger"
)

var TotalRequestCount int

type Response struct {
	Success  bool
	Url      string
	Response *http.Response
}

type HttpWorker struct {
	running      bool
	wg           *sync.WaitGroup
	haltChan     chan int
	db           *DBConn
	requestChan  chan *Request
	responseChan chan *Response
}

var client *http.Client

func StartHttpWorker(wg *sync.WaitGroup, db *DBConn, requestChan chan *Request, responseChan chan *Response, timeOut int) *HttpWorker {
	ConfigureClient(timeOut)
	haltChan := make(chan int, 1)
	httpWorker := &HttpWorker{true, wg, haltChan, db, requestChan, responseChan}
	wg.Add(1)
	go httpWorker.work()
	return httpWorker
}

func (worker *HttpWorker) Stop() {
	Logger.Debugf("Sending http worker stop signal")
	worker.haltChan <- 0
}

func (worker *HttpWorker) work() {
	defer worker.wg.Done()

	Logger.Debugf("Starting http worker")
	running := true
	for running {
		select {
		case <-worker.haltChan:
			running = false
			break
		case request := <-worker.requestChan:
			worker.processRequest(request)
			break
		default:
			break
		}
	}
	Logger.Debugf("Http worker stopped")
}

func (worker *HttpWorker) processRequest(request *Request) {
	Logger.Debugf("Http worker requesting %v", request.Url)

	res, err := client.Get(request.Url)
	success := false
	if err != nil {
		Logger.Warnf("Error requesting %v", request.Url)
		Logger.Warnf("%v", err)
	} else {
		success = true
		ioutil.ReadAll(res.Body)
		res.Body.Close()
	}
	TotalRequestCount++
	worker.responseChan <- &Response{success, request.Url, res}
}

func ConfigureClient(timeout int) {
	if client == nil {
		client = &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
			Transport: &http.Transport{
				MaxIdleConns:        200,
				MaxIdleConnsPerHost: 200,
			},
			Timeout: time.Duration(timeout) * time.Second,
		}
	}
}

func CleanupClient() {
	if client != nil {
		tr := client.Transport.(*http.Transport)
		tr.CloseIdleConnections()
	}
}
