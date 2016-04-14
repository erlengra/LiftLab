package liftAssigner

import (
	def "config"
	"log"
	"queue"
	"time"
)

type reply struct {
	cost int
	lift string
}
type order struct {
	floor  int
	button int
	timer  *time.Timer
}

func Run(costReply <-chan def.Message, numOnline *int) {
	unassigned := make(map[order][]reply)

	var timeout = make(chan *order)
	const timeoutDuration = 10 * time.Second

	for {
		select {
		case message := <-costReply:
			newOrder := order{floor: message.Floor, button: message.Button}
			newReply := reply{cost: message.Cost, lift: message.Addr}

			for oldOrder := range unassigned {
				if equal(oldOrder, newOrder) {
					newOrder = oldOrder
				}
			}

			if replyList, exist := unassigned[newOrder]; exist {
				found := false
				for _, reply := range replyList {
					if reply == newReply {
						found = true
					}
				}
				if !found {
					unassigned[newOrder] = append(unassigned[newOrder], newReply)
					newOrder.timer.Reset(timeoutDuration)
				}
			} else {
				newOrder.timer = time.NewTimer(timeoutDuration)
				unassigned[newOrder] = []reply{newReply}
				go costTimer(&newOrder, timeout)
			}
			chooseBestLift(unassigned, numOnline, false)

		case <-timeout:
			log.Println(def.ColR, "Not all costs received in time!", def.ColN)
			chooseBestLift(unassigned, numOnline, true)
		}
	}
}

func chooseBestLift(unassigned map[order][]reply, numOnline *int, orderTimedOut bool) {
	const maxInt = int(^uint(0) >> 1)
	for order, replyList := range unassigned {
		if len(replyList) == *numOnline || orderTimedOut {
			lowestCost := maxInt
			var bestLift string

			for _, reply := range replyList {
				if reply.cost < lowestCost {
					lowestCost = reply.cost
					bestLift = reply.lift
				} else if reply.cost == lowestCost {
					if reply.lift < bestLift {
						lowestCost = reply.cost
						bestLift = reply.lift
					}
				}
			}
			queue.AddRemoteOrder(order.floor, order.button, bestLift)
			order.timer.Stop()
			delete(unassigned, order)
		}
	}
}

func costTimer(newOrder *order, timeout chan<- *order) {
	<-newOrder.timer.C
	timeout <- newOrder
}

func equal(o1, o2 order) bool {
	return o1.floor == o2.floor && o1.button == o2.button
}
