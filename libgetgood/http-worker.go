package libgetgood

import (
	"net/http"
	"sync"

	. "github.com/dpindur/get-good/logger"
)

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
	client       *http.Client
}

func StartHttpWorker(wg *sync.WaitGroup, db *DBConn, requestChan chan *Request, responseChan chan *Response) *HttpWorker {
	haltChan := make(chan int)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
	}
	httpWorker := &HttpWorker{true, wg, haltChan, db, requestChan, responseChan, client}
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
	res, err := worker.client.Get(request.Url)
	success := false
	if err != nil {
		Logger.Warnf("Error requesting %v", request.Url)
		Logger.Warnf("%v", err)
	} else {
		success = true
	}
	worker.responseChan <- &Response{success, request.Url, res}
}
