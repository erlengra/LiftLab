package fsm

import (
	"time"
)

func Timer(timeout chan<- bool, reset <-chan bool, get_time <- chan bool) {
	const duration = 3 * time.Second
	timer := time.NewTimer(0)
	timer.Stop()
	for {
		select {
		case timer.Now():
			log.Println(timer.Now())
		case <-timer.C:
			timer.Stop()
			timeout <- true
		case <-reset:
			timer.Reset(duration)
		
		}
	}
}
