package network

import (
	"../config"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"net"
	"strconv"
	//"../queue"
)

var laddr *net.UDPAddr
var baddr *net.UDPAddr

type udpMessage struct {
	raddr  string
	data   []byte
	length int
}

var onlineLifts = make(map[string]config.UdpConnection)
var numOnline int
var queueNotifier = make(chan string)


func NumberOfOnlineLifts() int {
	return numOnline
}

func UpdateConnections(address string, networkPackage chan config.Message) {	
	if connection, ok := onlineLifts[address]; ok {
		connection.Timer.Reset(config.NetworkTimeoutPeriod)
	} else {
		newConn := config.UdpConnection{address, time.NewTimer(config.NetworkTimeoutPeriod)}
		log.Printf("Elevator at IP %s discovered!", address[0:15])
		onlineLifts[address] = newConn
		numOnline++
		go udp_connection_timer(&newConn, networkPackage)
	}
}

func udp_connection_timer(connection *config.UdpConnection, networkPackage chan config.Message) {	
	<-connection.Timer.C
	lostAddr := connection.Addr
	numOnline--
	log.Printf("Elevator at IP %s lost!", lostAddr[0:15])
	delete(onlineLifts, lostAddr)
	queueNotifier <- lostAddr
}

func Initialize(channels config.SystemChannels) {

	var udpOutgoing = make(chan udpMessage)
	var udpIncoming = make(chan udpMessage, 10)

	queueNotifier = channels.QueueNetworkComm

	err := udp_init(config.LocalListenPort, config.BroadcastListenPort, config.MessageSize, udpOutgoing, udpIncoming)
	config.CheckError(err)

	go iAmAlive(channels.OutgoingMsg)
	go networkHandler(channels.OutgoingMsg, udpOutgoing, channels.IncomingMsg, udpIncoming)

	log.Println("--------------Network initialised-------------------")
}

func udp_init(localListenPort, broadcastListenPort, message_size int, send_channel, receive_ch chan udpMessage) (err error) {
	//Generating broadcast address
	baddr, err = net.ResolveUDPAddr("udp4", "255.255.255.255:"+strconv.Itoa(broadcastListenPort))
	if err != nil {
		return err
	}


	if GetOwnID() != "127.0.0.1" {
		//Generating localaddress
		tempConn, err := net.DialUDP("udp4", nil, baddr)
		defer tempConn.Close()
		tempAddr := tempConn.LocalAddr()

		laddr, err := net.ResolveUDPAddr("udp4", tempAddr.String())

		laddr.Port = localListenPort
		config.Laddr = laddr.String()

		//Creating local listening connections
		localListenConn, err := net.ListenUDP("udp4", laddr)
		if err != nil {
			return err
		}

		//Creating listener on broadcast connection
		broadcastListenConn, err := net.ListenUDP("udp", baddr)
		if err != nil {
			localListenConn.Close()
			return err
		}

		go udp_receive_server(localListenConn, broadcastListenConn, message_size, receive_ch)
		go udp_transmit_server(localListenConn, broadcastListenConn, send_ch)
		go udp_connection_closer(localListenConn, broadcastListenConn)

		return err
	}
	return err	
}

func udp_transmit_server(lconn, bconn *net.UDPConn, send_ch <-chan udpMessage) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("ERROR in udp_transmit_server: %s \n Closing connection.", r)
			lconn.Close()
			bconn.Close()
		}
	}()

	var err error
	var n int

	for {
		msg := <-send_ch
		if msg.raddr == "broadcast" {
			n, err = lconn.WriteToUDP(msg.data, baddr)
		} else {
			raddr, err := net.ResolveUDPAddr("udp", msg.raddr)
			if err != nil {
				fmt.Printf("Error: udp_transmit_server: could not resolve raddr\n")
				panic(err)
			}
			n, err = lconn.WriteToUDP(msg.data, raddr)
		}
		if err != nil || n < 0 {
			fmt.Printf("Error: udp_transmit_server: writing\n")
			panic(err)
		}
	}
}

func udp_receive_server(lconn, bconn *net.UDPConn, message_size int, receive_ch chan<- udpMessage) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("ERROR in udp_receive_server: %s \n Closing connection.", r)
			lconn.Close()
			bconn.Close()
		}
	}()

	bconn_rcv_ch := make(chan udpMessage)
	lconn_rcv_ch := make(chan udpMessage)

	go udp_connection_reader(lconn, message_size, lconn_rcv_ch)
	go udp_connection_reader(bconn, message_size, bconn_rcv_ch)

	for {
		select {

		case buf := <-bconn_rcv_ch:
			receive_ch <- buf

		case buf := <-lconn_rcv_ch:
			receive_ch <- buf
		}
	}
}

func udp_connection_reader(conn *net.UDPConn, message_size int, rcv_ch chan<- udpMessage) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("ERROR in udp_connection_reader: %s \n Closing connection.", r)
			conn.Close()
		}
	}()

	for {
		buf := make([]byte, message_size)
		//		fmt.Printf("udp_connection_reader: Waiting on data from UDPConn\n")
		n, raddr, err := conn.ReadFromUDP(buf)
		//		fmt.Printf("udp_connection_reader: Received %s from %s \n", string(buf), raddr.String())
		if err != nil || n < 0 {
			fmt.Printf("Error: udp_connection_reader: reading\n")
			panic(err)
		}
		rcv_ch <- udpMessage{raddr: raddr.String(), data: buf, length: n}
	}
}

func udp_connection_closer(lconn, bconn *net.UDPConn) {
	<-config.CloseConnectionChan
	lconn.Close()
	bconn.Close()
}

func GetOwnID() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Fatal(err)
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}
	return "127.0.0.1"
}

func iAmAlive(networkPackage chan<- config.Message) {
	alive := config.Message{Category: config.Alive, Floor: -1, Button: -1, Cost: -1}
	for {
		time.Sleep(500 * time.Millisecond)
		networkPackage <- alive
	}
}

func networkHandler(networkPackage <-chan config.Message, udpOutgoing chan<- udpMessage,
					incomingMsg chan<- config.Message, udpIncoming <-chan udpMessage) {
	for {
		select {
		case unencodedMsg := <-networkPackage:
			encodedMsg, err := json.Marshal(unencodedMsg)
			config.CheckError(err)

			udpOutgoing <- udpMessage{raddr: "broadcast", data: encodedMsg, length: len(encodedMsg)}

		case udpMessage := <-udpIncoming:
			var decodedMessage config.Message
			err := json.Unmarshal(udpMessage.data[:udpMessage.length], &decodedMessage)
			config.CheckError(err)

			decodedMessage.Addr = udpMessage.raddr
			incomingMsg <- decodedMessage
		}
	}
}
