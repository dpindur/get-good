package libgetgood

import (
	"log"
	"net/http"
	"sync"
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
	log.Printf("Sending http worker stop signal\n")
	worker.haltChan <- 0
}

func (worker *HttpWorker) work() {
	defer worker.wg.Done()

	log.Printf("Starting http worker\n")
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
	log.Printf("Http worker stopped\n")
}

func (worker *HttpWorker) processRequest(request *Request) {
	log.Printf("Processing request %v\n", request.Url)
	res, err := worker.client.Get(request.Url)
	success := false
	if err != nil {
		log.Printf("Error requesting %v\n", request.Url)
		log.Printf("%v\n", err)
	} else {
		success = true
	}
	worker.responseChan <- &Response{success, request.Url, res}
}
