package fsm

import (
	"time"
)

func doorTimer(timeout chan<- bool, reset <-chan bool) {
	const doorOpenTime = 3 * time.Second
	timer := time.NewTimer(0)
	timer.Stop()

	for {
		select {
		case <-reset:
			timer.Reset(doorOpenTime)
		case <-timer.C:
			timer.Stop()
			timeout <- true
		}
	}
}
