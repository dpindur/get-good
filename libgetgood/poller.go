package libgetgood

import (
	"sync"
	"time"

	. "github.com/dpindur/get-good/logger"
)

type Poller struct {
	running     bool
	wg          *sync.WaitGroup
	haltChan    chan int
	db          *DBConn
	errChan     chan *WorkerError
	requestChan chan *Request
}

func StartPoller(wg *sync.WaitGroup, db *DBConn, errChan chan *WorkerError, requestChan chan *Request) *Poller {
	haltChan := make(chan int)
	poller := &Poller{true, wg, haltChan, db, errChan, requestChan}
	wg.Add(1)
	go poller.work()
	return poller
}

func (poller *Poller) Stop() {
	Logger.Debugf("Sending database poller stop signal")
	poller.haltChan <- 0
}

func (poller *Poller) work() {
	defer poller.wg.Done()

	Logger.Debugf("Starting database poller")
	running := true
	for running {
		select {
		case <-poller.haltChan:
			running = false
			break
		default:
			time.Sleep(1 * time.Second)
			err := poller.pollDatabase()
			if err != nil {
				running = false
				poller.errChan <- &WorkerError{"poller", err}
			}
			break
		}
	}
	Logger.Debugf("Database poller stopped")
}

func (poller *Poller) pollDatabase() error {
	Logger.Debugf("Polling")
	requests, err := poller.db.GetIncompleteRequests()
	if err != nil {
		return err
	}

	// Adding a request to the queue will block if the queue is full
	// in this case we pause the poller for a bit and then return.
	// Any incomplete requests we were looking at will just be
	// picked up during the next poll
	for _, url := range requests {
		select {
		case poller.requestChan <- &Request{url}:
			err := poller.db.SetRequestInflight(url)
			if err != nil {
				return err
			}
			break
		default:
			Logger.Infof("Request queue full, pausing poller for five seconds...")
			time.Sleep(5 * time.Second)
			return nil
		}
	}

	return nil
}
