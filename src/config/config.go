package config

import (
	"os/exec"
	"os"
	"fmt"
	"time"
)

const N_Buttons = 3
const N_Floors = 4
const MOTOR_SPEED = 2800
const MessageSize = 1024
const LocalListenPort = 37103
const BroadcastListenPort = 37104
const NetworkTimeoutPeriod = 2 * time.Second
const (
	BtnUp int = iota
	BtnDown
	BtnInside
)
const (
	DirDown int = iota - 1
	DirStop
	DirUp
)
const (
	Alive int = iota + 1
	NewOrder
	CompleteOrder
	Cost
)

var Laddr string
var SyncLightsChan = make(chan bool)
var CloseConnectionChan = make(chan bool)

type Keypress struct {
	Button int
	Floor  int
}

type Message struct {
	Category int
	Floor    int
	Button   int
	Cost     int
	Addr     string `json:"-"`
}

type UdpConnection struct {
	Addr  string
	Timer *time.Timer
}

type SystemChannels struct {
	NewOrder     chan bool
	FloorReached chan int
	DoorTimeout  chan bool
	DoorTimerReset chan bool
	OutgoingMsg chan Message
	IncomingMsg chan Message
	CostChan chan Message
	QueueNetworkComm chan string
}

func CheckError(err error) {
    if err  != nil {
        fmt.Println("Error: " , err)
        os.Exit(0)
    }
}
