package libgetgood

import (
	"log"
	"sync"
)

type Updater struct {
	running  bool
	wg       *sync.WaitGroup
	haltChan chan int
	db       *DBConn
}

func StartUpdater(wg *sync.WaitGroup, db *DBConn) *Updater {
	haltChan := make(chan int)
	updater := &Updater{true, wg, haltChan, db}
	wg.Add(1)
	go updater.work()
	return updater
}

func (updater *Updater) Stop() {
	log.Printf("Sending database updater stop signal\n")
	updater.haltChan <- 0
}

func (updater *Updater) work() {
	defer updater.wg.Done()

	log.Printf("Starting database updater\n")
	running := true
	for running {
		select {
		case <-updater.haltChan:
			running = false
			break
		}
	}
	log.Printf("Database updater stopped\n")
}
