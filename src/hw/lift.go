package hw

import (
	def "config"
	"errors"
	"log"
)

var lampChannelMatrix = [def.N_Floors][def.N_Buttons]int{
	{LIGHT_UP1, LIGHT_DOWN1, LIGHT_COMMAND1},
	{LIGHT_UP2, LIGHT_DOWN2, LIGHT_COMMAND2},
	{LIGHT_UP3, LIGHT_DOWN3, LIGHT_COMMAND3},
	{LIGHT_UP4, LIGHT_DOWN4, LIGHT_COMMAND4},
}
var buttonChannelMatrix = [def.N_Floors][def.N_Buttons]int{
	{BUTTON_UP1, BUTTON_DOWN1, BUTTON_COMMAND1},
	{BUTTON_UP2, BUTTON_DOWN2, BUTTON_COMMAND2},
	{BUTTON_UP3, BUTTON_DOWN3, BUTTON_COMMAND3},
	{BUTTON_UP4, BUTTON_DOWN4, BUTTON_COMMAND4},
}

func Init() (int, error) {
	if !ioInit() {
		return -1, errors.New("Hardware driver: ioInit() failed!")
	}

	for f := 0; f < def.N_Floors; f++ {
		if f != 0 {
			SetButtonLamp(f, def.BtnDown, false)
		}
		if f != def.N_Floors-1 {
			SetButtonLamp(f, def.BtnUp, false)
		}
		SetButtonLamp(f, def.BtnInside, false)
	}

	SetStopLamp(false)
	SetDoorLamp(false)

	SetMotorDir(def.DirDown)
	floor := Floor()
	for floor == -1 {
		floor = Floor()
	}
	SetMotorDir(def.DirStop)
	SetFloorLamp(floor)

	log.Println(def.ColG, "Hardware initialised.", def.ColN)
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
	if floor < 0 || floor >= def.N_Floors {
		log.Printf("Error: Floor %d out of range!\n", floor)
		log.Println("No floor indicator will be set.")
		return
	}

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
	if floor < 0 || floor >= def.N_Floors {
		log.Printf("Error: Floor %d out of range!\n", floor)
		return false
	}
	if button < 0 || button >= def.N_Buttons {
		log.Printf("Error: Button %d out of range!\n", button)
		return false
	}
	if button == def.BtnUp && floor == def.N_Floors-1 {
		log.Println("Button up from top floor does not exist!")
		return false
	}
	if button == def.BtnDown && floor == 0 {
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
	if floor < 0 || floor >= def.N_Floors {
		log.Printf("Error: Floor %d out of range!\n", floor)
		return
	}
	if button == def.BtnUp && floor == def.N_Floors-1 {
		log.Println("Button up from top floor does not exist!")
		return
	}
	if button == def.BtnDown && floor == 0 {
		log.Println("Button down from ground floor does not exist!")
		return
	}
	if button != def.BtnUp &&
		button != def.BtnDown &&
		button != def.BtnInside {
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
