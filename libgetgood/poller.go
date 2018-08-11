package libgetgood

import (
	"log"
	"sync"
)

type Poller struct {
	running  bool
	wg       *sync.WaitGroup
	haltChan chan int
	db       *DBConn
	errChan  chan *WorkerError
}

func StartPoller(wg *sync.WaitGroup, db *DBConn, errChan chan *WorkerError) *Poller {
	haltChan := make(chan int)
	poller := &Poller{true, wg, haltChan, db, errChan}
	wg.Add(1)
	go poller.work()
	return poller
}

func (poller *Poller) Stop() {
	log.Printf("Sending database poller stop signal\n")
	poller.haltChan <- 0
}

func (poller *Poller) work() {
	defer poller.wg.Done()

	log.Printf("Starting database poller\n")
	running := true
	for running {
		select {
		case <-poller.haltChan:
			running = false
			break
		}
	}
	log.Printf("Database poller stopped\n")
}
