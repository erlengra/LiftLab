package queue

import (
	def "config"
	"log"
	"time"
)

func (q *queue) startTimer(floor, button int) {
	const orderTimeout = 30 * time.Second

	q.queue_table[floor][button].timer = time.NewTimer(orderTimeout)
	<-q.queue_table[floor][button].timer.C
	OrderTimeoutChan <- def.Keypress{Button: button, Floor: floor}
}

func (q *queue) stopTimer(floor, button int) {
	if q.queue_table[floor][button].timer != nil {
		q.queue_table[floor][button].timer.Stop()
	}
}

func (q *queue) isEmpty() bool {
	for f := 0; f < def.N_Floors; f++ {
		for b := 0; b < def.N_Buttons; b++ {
			if q.queue_table[f][b].active {
				return false
			}
		}
	}
	return true
}

func (q *queue) setOrder(floor, button int, status Status) {
	if q.isOrder(floor, button) == status.active {
		return
	}
	q.queue_table[floor][button] = status
	def.SyncLightsChan <- true
	takeBackup <- true
	printQueues()
}

func (q *queue) isOrder(floor, button int) bool {
	return q.queue_table[floor][button].active
}

func (q *queue) isOrdersAbove(floor int) bool {
	for f := floor + 1; f < def.N_Floors; f++ {
		for b := 0; b < def.N_Buttons; b++ {
			if q.isOrder(f, b) {
				return true
			}
		}
	}
	return false
}

func (q *queue) isOrdersBelow(floor int) bool {
	for f := 0; f < floor; f++ {
		for b := 0; b < def.N_Buttons; b++ {
			if q.isOrder(f, b) {
				return true
			}
		}
	}
	return false
}

func (q *queue) chooseDirection(floor, dir int) int {
	if q.isEmpty() {
		return def.DirStop
	}
	switch dir {
	case def.DirDown:
		if q.isOrdersBelow(floor) && floor > 0 {
			return def.DirDown
		} else {
			return def.DirUp
		}
	case def.DirUp:
		if q.isOrdersAbove(floor) && floor < def.N_Floors-1 {
			return def.DirUp
		} else {
			return def.DirDown
		}
	case def.DirStop:
		if q.isOrdersAbove(floor) {
			return def.DirUp
		} else if q.isOrdersBelow(floor) {
			return def.DirDown
		} else {
			return def.DirStop
		}
	default:
		def.CloseConnectionChan <- true
		def.Restart.Run()
		log.Printf("%sChooseDirection(): called with invalid direction %d, returning stop%s\n", def.ColR, dir, def.ColN)
		return 0
	}
}

func (q *queue) shouldStop(floor, dir int) bool {
	switch dir {
	case def.DirDown:
		return q.isOrder(floor, def.BtnDown) ||
			q.isOrder(floor, def.BtnInside) ||
			floor == 0 ||
			!q.isOrdersBelow(floor)
	case def.DirUp:
		return q.isOrder(floor, def.BtnUp) ||
			q.isOrder(floor, def.BtnInside) ||
			floor == def.N_Floors-1 ||
			!q.isOrdersAbove(floor)
	case def.DirStop:
		return q.isOrder(floor, def.BtnDown) ||
			q.isOrder(floor, def.BtnUp) ||
			q.isOrder(floor, def.BtnInside)
	default:
		def.CloseConnectionChan <- true
		def.Restart.Run()
		log.Fatalln(def.ColR, "This direction doesn't exist", def.ColN)
	}
	return false
}

func (q *queue) deepCopy() *queue {
	queueCopy := new(queue)
	for f := 0; f < def.N_Floors; f++ {
		for b := 0; b < def.N_Buttons; b++ {
			queueCopy.queue_table[f][b] = q.queue_table[f][b]
		}
	}
	return queueCopy
}
