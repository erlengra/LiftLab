package statemachine

import (
	"../config"
	"log"
	"../queue"
	"../driver"
)

const (
	idle int = iota
	moving
	doorOpen
)

var state int
var floor int
var dir int

func Initialize(channel config.SystemChannels, startFloor int) {
	state = idle
	dir = config.DirStop
	floor = startFloor
	channel.CloseDoor = make(chan bool)
	channel.DoorTimerReset = make(chan bool)

	go Timer(channel.CloseDoor, channel.DoorTimerReset)
	go EventForwarder(channel.)

	log.Println("--------------Statemachine initialised--------------")
}

func EventForwarder(channel. config.SystemChannels) {
	for {
		select {
		case <-channel.NewOrder:
			EventNewOrder(channel)
		case floor := <-channel.AtFloor:
			EventAtFloor(channel, floor)
		case <-channel.CloseDoor:
			EventCloseDoor(channel)
		}
	}
}

func EventNewOrder(channel config.SystemChannels) {
	switch state {
	case idle:
		dir = queue.ChooseDirection(floor, dir)
		if queue.StopElevator(floor, dir) {
			channel.DoorTimerReset <- true
			queue.RemoveOrdersAt(floor, channel.OutgoingMsg)
			driver.Elev_Set_Door_Open_Lamp(true)
			state = doorOpen
		} else {
			driver.Elev_Set_Motor_Direction(dir)
			state = moving
		}
	case moving:
		log.Println("Nothing to be done in state: moving")
	case doorOpen:
		if queue.StopElevator(floor, dir) {
			channel.DoorTimerReset <- true
			queue.RemoveOrdersAt(floor, channel.OutgoingMsg)
		}
	default:
		config.CloseConnectionChan <- true
		config.Restart.Run()
		log.Println("invalid state")
	}
}

func EventAtFloor(channel config.SystemChannels, newFloor int) {
	log.Printf("AtFloor triggered while in state ")
	floor = newFloor
	driver.Elev_Set_Floor_Indicator(floor)
	switch state {
	case moving:
		if queue.StopElevator(floor, dir) {
			log.Printf("Should stop")
			channel.DoorTimerReset <- true
			queue.RemoveOrdersAt(floor, channel.OutgoingMsg)
			driver.Elev_Set_Door_Open_Lamp(true)
			dir = config.DirStop
			driver.Elev_Set_Motor_Direction(dir)
			state = doorOpen
		}
	default:
		config.CloseConnectionChan <- true
		config.Restart.Run()

	}
}

func EventCloseDoor(channel. config.SystemChannels) {
	switch state {
	case doorOpen:
		driver.Elev_Set_Door_Open_Lamp(false)
		dir = queue.ChooseDirection(floor, dir)
		driver.Elev_Set_Motor_Direction(dir)
		if dir == config.DirStop {
			state = idle
		} else {
			state = moving
		}
	default:
		config.CloseConnectionChan <- true
		config.Restart.Run()
	}
}
