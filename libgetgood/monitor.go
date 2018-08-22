package libgetgood

import (
	"sync"
	"time"

	. "github.com/dpindur/get-good/logger"
)

type Monitor struct {
	running          bool
	wg               *sync.WaitGroup
	haltChan         chan int
	db               *DBConn
	errChan          chan *WorkerError
	bustCompleteChan chan int
}

func StartMonitor(wg *sync.WaitGroup, db *DBConn, errChan chan *WorkerError, bustCompleteChan chan int) *Monitor {
	haltChan := make(chan int)
	monitor := &Monitor{true, wg, haltChan, db, errChan, bustCompleteChan}
	wg.Add(1)
	go monitor.work()
	return monitor
}

func (monitor *Monitor) Stop() {
	Logger.Debugf("Sending monitor stop signal")
	monitor.haltChan <- 0
}

func (monitor *Monitor) work() {
	defer monitor.wg.Done()

	Logger.Debugf("Starting monitor")
	running := true
	for running {
		select {
		case <-monitor.haltChan:
			running = false
			break
		default:
			time.Sleep(3 * time.Second)
			err := monitor.checkRemainingRequests()
			if err != nil {
				running = false
				monitor.errChan <- &WorkerError{"monitor", err}
			}
			break
		}
	}
	Logger.Debugf("Monitor stopped")
}

func (monitor *Monitor) checkRemainingRequests() error {
	remainingReqs, err := monitor.db.GetRemainingRequestCount()
	if err != nil {
		return err
	}

	Logger.Infof("Remaining requests: %v", remainingReqs)
	if remainingReqs == 0 {
		monitor.bustCompleteChan <- 0
	}

	return nil
}
