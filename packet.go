package server

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
)

type Packet struct {
	dataType uint32
	dataSize uint32
	data     []byte
}

type PacketWrapper struct {
	packet  *Packet
	maxSize uint32
	conn    net.Conn
}

func NewPacketWrapper(maxSize uint32, conn net.Conn) *PacketWrapper {
	packet := &Packet{
		dataType: 0,
		dataSize: 0,
		data:     make([]byte, maxSize-8),
	}
	packetWrapper := &PacketWrapper{
		packet:  packet,
		maxSize: maxSize,
		conn:    conn,
	}
	return packetWrapper
}

type PacketInterface interface {
	WriteDataInPacket(data []byte, dataType uint32) error
	sendData() error
	SendAllData(data []byte, dataType uint32)
	SendDataType(dataType uint32) error
	readData() error
	ReadAllData() []byte
	ReadDataType() error
}

func (packetWrapper *PacketWrapper) WriteDataInPacket(data []byte, dataType uint32) error {
	lenData := uint32(len(data))
	if lenData > packetWrapper.maxSize-8 {
		return errors.New("data size is higher than the max")
	}
	packetWrapper.packet.dataSize = lenData
	packetWrapper.packet.dataType = dataType
	if lenData > 0 {
		copy(packetWrapper.packet.data, data)
	}

	return nil
}

func (packetWrapper *PacketWrapper) sendData() error {
	buffer := make([]byte, packetWrapper.maxSize)
	binary.LittleEndian.PutUint32(buffer[0:], packetWrapper.packet.dataType)
	binary.LittleEndian.PutUint32(buffer[4:], packetWrapper.packet.dataSize)
	copy(buffer[8:], packetWrapper.packet.data)

	_, err := packetWrapper.conn.Write(buffer)
	if err != nil {
		return err
	}

	return nil
}

func (packetWrapper *PacketWrapper) SendAllData(data []byte, dataType uint32) {
	packetWrapper.SendDataType(dataType)
	packetWrapper.readData()
	startPos := uint32(0)
	endPos := uint32(0)
	lenData := uint32(len(data))
	var err error = nil
	for {
		if startPos == lenData {
			break
		}
		if err == nil {
			endPos = startPos + packetWrapper.maxSize - 8
			if endPos > lenData {
				endPos = lenData
			}
		}
		packetWrapper.WriteDataInPacket(data[startPos:endPos], PacketSent)
		err = packetWrapper.sendData()
		if err != nil {
			log.Println("error in send: ", err)
			continue
		}
		log.Println("A paquet has been sent", packetWrapper.packet.dataType)
		packetWrapper.readData()
		if packetWrapper.packet.dataType == UnableToReadPacket {
			log.Print("UnableToReadPacket")
			err = errors.New("UnableToReadPacket")
			continue
		}

		startPos = endPos
	}
	packetWrapper.SendDataType(EndTransfert)
	packetWrapper.readData()
}

func (packetWrapper *PacketWrapper) SendDataType(dataType uint32) error {
	err := packetWrapper.WriteDataInPacket([]byte{}, dataType)
	if err != nil {
		return err
	}

	return packetWrapper.sendData()
}

func (packetWrapper *PacketWrapper) readData() error {
	buffer := make([]byte, packetWrapper.maxSize)
	_, err := io.ReadFull(packetWrapper.conn, buffer)
	if err != nil {
		return err
	}

	dataType := binary.LittleEndian.Uint32(buffer[:4])
	dataSize := binary.LittleEndian.Uint32(buffer[4:8])
	data := buffer[8 : 8+dataSize]

	packetWrapper.WriteDataInPacket(data, dataType)

	return nil
}

func (packetWrapper *PacketWrapper) ReadAllData() []byte {
	var data []byte

	for {
		err := packetWrapper.readData()
		log.Println("A packet has been read", packetWrapper.packet.dataType)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			packetWrapper.SendDataType(UnableToReadPacket)
			continue
		}
		packetWrapper.SendDataType(PacketReceived)
		if packetWrapper.packet.dataType == EndTransfert {
			return data
		}
		if packetWrapper.packet.dataSize > 0 {
			data = append(data, packetWrapper.packet.data[:packetWrapper.packet.dataSize]...)
		}
	}
}

func (packetWrapper *PacketWrapper) ReadDataType() error {
	err := packetWrapper.readData()
	if err != nil {
		return err
	}
	return nil
}
