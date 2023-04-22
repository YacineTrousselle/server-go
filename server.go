package server

import (
	"fmt"
	"log"
	"net"
)

type handleConnectionType func(conn net.Conn)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	packetWrapper := NewPacketWrapper(MaxPacketSize)
	for {
		data := packetWrapper.ReadAllData(conn)

		switch packetWrapper.packet.dataType {
		case RequestFile:
			filename := string(data)
			fmt.Println(filename)
		default:
			packetWrapper.SendDataType(conn, InvalidInputError)
		}
	}
}

func LaunchServer(handleConnectionType handleConnectionType) {
	if handleConnectionType == nil {
		handleConnectionType = handleConnection
	}
	listener, err := net.Listen(TYPE, HOST+":"+PORT)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print("Connection failed")
			continue
		}
		go handleConnectionType(conn)
	}
}

func LaunchClient() (net.Conn, error) {
	return net.Dial(TYPE, HOST+":"+PORT)
}
