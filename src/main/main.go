package main

import (
	"../config"
	"../fsm"
	"../eventhandler"
	"../driver"
	"../network"
	"../queue"
)


//endre slik at en heis vil returnere til en etasje hvor det er to retninger bestilt
//selv om den stopper for den ene retningsbestillingen

func main() {	

	var channel = config.SystemChannels{
		NewOrder:     make(chan bool),
		FloorReached: make(chan int),
		OutgoingMsg:  make(chan config.Message, 10),
		IncomingMsg: make(chan config.Message, 10),
		CostChan: make(chan config.Message),
		QueueNetworkComm: make(chan string),
	}


	floor := driver.Elev_Initialize()
	fsm.Initialize(channel, floor)
	network.Initialize(ch)
	eventhandler.Initialize(ch)
	queue.Initialize(ch)


	aliveKeeper := make(chan bool)
	<-aliveKeeper
}
