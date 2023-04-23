package server

import (
	"io"
	"log"
	"net"
)

type handleConnectionType func(conn net.Conn)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	packetWrapper := NewPacketWrapper(MaxPacketSize, conn)
	for {
		err := packetWrapper.ReadDataType()
		if err == io.EOF {
			return
		}

		switch packetWrapper.packet.dataType {
		case RequestFile:
			data := packetWrapper.ReadAllData()
			log.Println("data read:", string(data))
			//file, err := os.ReadFile(filename)
			//if err != nil {
			//	packetWrapper.SendDataType(FileNotFound)
			//} else {
			//	packetWrapper.SendAllData(file, FileData)
			//}
		default:
			packetWrapper.SendDataType(InvalidInputError)
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
