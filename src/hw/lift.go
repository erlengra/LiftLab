// Package hw defines interactions with the lift hardware at the real time
// lab at The Department of Engineering Cybernetics at NTNU, Trondheim,
// Norway.
//
// This file is a golang port of elev.c from the hand out driver
// (https://github.com/TTK4145/Project)
package hw

import (
	"../config"
	"errors"
	"log"
)

var lampChannelMatrix = [config.N_Floors][config.N_Buttons]int{
	{LIGHT_UP1, LIGHT_DOWN1, LIGHT_COMMAND1},
	{LIGHT_UP2, LIGHT_DOWN2, LIGHT_COMMAND2},
	{LIGHT_UP3, LIGHT_DOWN3, LIGHT_COMMAND3},
	{LIGHT_UP4, LIGHT_DOWN4, LIGHT_COMMAND4},
}
var buttonChannelMatrix = [config.N_Floors][config.N_Buttons]int{
	{BUTTON_UP1, BUTTON_DOWN1, BUTTON_COMMAND1},
	{BUTTON_UP2, BUTTON_DOWN2, BUTTON_COMMAND2},
	{BUTTON_UP3, BUTTON_DOWN3, BUTTON_COMMAND3},
	{BUTTON_UP4, BUTTON_DOWN4, BUTTON_COMMAND4},
}

// Init initialises the lift hardware and moves the lift to a defined state.
// (Descending until it reaches a floor.)
func Init() (int, error) {
	// Init hardware
	if !ioInit() {
		return -1, errors.New("Hardware driver: ioInit() failed!")
	}

	// Zero all floor button lamps
	for f := 0; f < config.N_Floors; f++ {
		if f != 0 {
			SetButtonLamp(f, config.BtnDown, false)
		}
		if f != config.N_Floors-1 {
			SetButtonLamp(f, config.BtnUp, false)
		}
		SetButtonLamp(f, config.BtnInside, false)
	}

	SetStopLamp(false)
	SetDoorLamp(false)

	// Move to defined state
	SetMotorDir(config.DirDown)
	floor := Floor()
	for floor == -1 {
		floor = Floor()
	}
	SetMotorDir(config.DirStop)
	SetFloorLamp(floor)

	log.Println(config.ColG, "Hardware initialised.", config.ColN)
	return floor, nil
}

func SetMotorDir(dirn int) {
	if dirn == 0 {
		ioWriteAnalog(MOTOR, 0)
	} else if dirn > 0 {
		ioClearBit(MOTORDIR)
		ioWriteAnalog(MOTOR, 2800)
	} else if dirn < 0 {
		ioSetBit(MOTORDIR)
		ioWriteAnalog(MOTOR, 2800)
	}
}

func SetDoorLamp(value bool) {
	if value {
		ioSetBit(LIGHT_DOOR_OPEN)
	} else {
		ioClearBit(LIGHT_DOOR_OPEN)
	}
}

func Floor() int {
	if ioReadBit(SENSOR_FLOOR1) {
		return 0
	} else if ioReadBit(SENSOR_FLOOR2) {
		return 1
	} else if ioReadBit(SENSOR_FLOOR3) {
		return 2
	} else if ioReadBit(SENSOR_FLOOR4) {
		return 3
	} else {
		return -1
	}
}

func SetFloorLamp(floor int) {
	if floor < 0 || floor >= config.N_Floors {
		log.Printf("Error: Floor %d out of range!\n", floor)
		log.Println("No floor indicator will be set.")
		return
	}

	// Binary encoding. One light must always be on.
	if floor&0x02 > 0 {
		ioSetBit(LIGHT_FLOOR_IND1)
	} else {
		ioClearBit(LIGHT_FLOOR_IND1)
	}

	if floor&0x01 > 0 {
		ioSetBit(LIGHT_FLOOR_IND2)
	} else {
		ioClearBit(LIGHT_FLOOR_IND2)
	}
}

func ReadButton(floor int, button int) bool {
	if floor < 0 || floor >= config.N_Floors {
		log.Printf("Error: Floor %d out of range!\n", floor)
		return false
	}
	if button < 0 || button >= config.N_Buttons {
		log.Printf("Error: Button %d out of range!\n", button)
		return false
	}
	if button == config.BtnUp && floor == config.N_Floors-1 {
		log.Println("Button up from top floor does not exist!")
		return false
	}
	if button == config.BtnDown && floor == 0 {
		log.Println("Button down from ground floor does not exist!")
		return false
	}

	if ioReadBit(buttonChannelMatrix[floor][button]) {
		return true
	} else {
		return false
	}
}

func SetButtonLamp(floor int, button int, value bool) {
	if floor < 0 || floor >= config.N_Floors {
		log.Printf("Error: Floor %d out of range!\n", floor)
		return
	}
	if button == config.BtnUp && floor == config.N_Floors-1 {
		log.Println("Button up from top floor does not exist!")
		return
	}
	if button == config.BtnDown && floor == 0 {
		log.Println("Button down from ground floor does not exist!")
		return
	}
	if button != config.BtnUp &&
		button != config.BtnDown &&
		button != config.BtnInside {
		log.Printf("Invalid button %d\n", button)
		return
	}

	if value {
		ioSetBit(lampChannelMatrix[floor][button])
	} else {
		ioClearBit(lampChannelMatrix[floor][button])
	}
}

func SetStopLamp(value bool) {
	if value {
		ioSetBit(LIGHT_STOP)
	} else {
		ioClearBit(LIGHT_STOP)
	}
}

func GetObstructionSignal() bool {
	return ioReadBit(OBSTRUCTION)
}

func GetStopSignal() bool {
	return ioReadBit(STOP)
}
