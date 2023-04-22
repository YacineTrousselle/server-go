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
		err := packetWrapper.ReadData(conn)
		if err != nil {
			log.Print("Unable to read packet")
			packetWrapper.SendDataType(conn, UnableToReadPacket)
			continue
		}

		switch packetWrapper.packet.dataType {
		case RequestFile:
			filename, err := packetWrapper.ReadAllData(conn)
			if err != nil {
				packetWrapper.SendDataType(conn, UnableToReadPacket)
			}
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
