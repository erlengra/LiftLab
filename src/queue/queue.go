package queue

import (
	def "config"
	"fmt"
	"log"
	"time"
)

type queue struct {
	queue_table [def.N_Floors][def.N_Buttons]Status
}

type Status struct {
	active bool
	addr   string      `json:"-"`
	timer  *time.Timer `json:"-"'
}

var inactive = Status{active: false, addr: "", timer: nil}
var local queue
var main queue
var updateLocal = make(chan bool)
var takeBackup = make(chan bool, 10)
var OrderTimeoutChan = make(chan def.Keypress)
var newOrder chan bool

func Initialize(newOrderTemp chan bool, outgoingMsg chan def.Message) {
	newOrder = newOrderTemp
	go updateLocalQueue()
	runBackup(outgoingMsg)
	log.Println(def.ColG, "Queue initialised.", def.ColN)
}

func SetOrderLocal(floor int, button int) {
	local.setOrder(floor, button, Status{true, "", nil})
	newOrder <- true
}

func AddRemoteOrder(floor, button int, addr string) {
	alreadyExist := IsRemoteOrder(floor, button)
	main.setOrder(floor, button, Status{true, addr, nil})
	if !alreadyExist {
		go main.startTimer(floor, button)
	}
	updateLocal <- true
}

func RemoveRemoteOrdersAt(floor int) {
	for b := 0; b < def.N_Buttons; b++ {
		main.stopTimer(floor, b)
		main.setOrder(floor, b, inactive)
	}
	updateLocal <- true
}

func RemoveOrdersAt(floor int, outgoingMsg chan<- def.Message) {
	for b := 0; b < def.N_Buttons; b++ {
		main.stopTimer(floor, b)
		local.setOrder(floor, b, inactive)
		main.setOrder(floor, b, inactive)
	}
	outgoingMsg <- def.Message{Category: def.CompleteOrder, Floor: floor}
}

func ShouldStop(floor, dir int) bool {
	return local.shouldStop(floor, dir)
}

func ChooseDirection(floor, dir int) int {
	return local.chooseDirection(floor, dir)
}

func IsLocalOrder(floor, button int) bool {
	return local.isOrder(floor, button)
}

func IsRemoteOrder(floor, button int) bool {
	return main.isOrder(floor, button)
}

func ReassignOrders(deadAddr string, outgoingMsg chan<- def.Message) {
	for f := 0; f < def.N_Floors; f++ {
		for b := 0; b < def.N_Buttons; b++ {
			if main.queue_table[f][b].addr == deadAddr {
				main.setOrder(f, b, inactive)
				outgoingMsg <- def.Message{Category: def.NewOrder, Floor: f, Button: b}
			}
		}
	}
}

func printQueues() {
	fmt.Printf(def.ColC)
	fmt.Println("Local   Remote")
	for f := def.N_Floors - 1; f >= 0; f-- {

		s1 := ""
		if local.isOrder(f, def.BtnUp) {
			s1 += "↑"
		} else {
			s1 += " "
		}
		if local.isOrder(f, def.BtnInside) {
			s1 += "×"
		} else {
			s1 += " "
		}
		fmt.Printf(s1)
		if local.isOrder(f, def.BtnDown) {
			fmt.Printf("↓   %d  ", f+1)
		} else {
			fmt.Printf("    %d  ", f+1)
		}

		s2 := "   "
		if main.isOrder(f, def.BtnUp) {
			fmt.Printf("↑")
			s2 += "(↑ " + main.queue_table[f][def.BtnUp].addr[12:15] + ")"
		} else {
			fmt.Printf(" ")
		}
		if main.isOrder(f, def.BtnDown) {
			fmt.Printf("↓")
			s2 += "(↓ " + main.queue_table[f][def.BtnDown].addr[12:15] + ")"
		} else {
			fmt.Printf(" ")
		}
		fmt.Printf("%s", s2)
		fmt.Println()
	}
	fmt.Printf(def.ColN)
}

func updateLocalQueue() {
	for {
		<-updateLocal
		for f := 0; f < def.N_Floors; f++ {
			for b := 0; b < def.N_Buttons; b++ {
				if main.isOrder(f, b) {
					if b != def.BtnInside && main.queue_table[f][b].addr == def.Laddr {
						if !local.isOrder(f, b) {
							local.setOrder(f, b, Status{true, "", nil})
							newOrder <- true
						}
					}
				}
			}
		}
	}
}
