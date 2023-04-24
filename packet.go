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
	packetWrapper.ReadDataType()
	if packetWrapper.packet.dataType == InvalidInputError {
		log.Fatalln("InvalidInputError")
	}

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
			log.Println("sendData error", err)
			continue
		}

		packetWrapper.ReadDataType()
		if packetWrapper.packet.dataType == UnableToReadPacket {
			log.Println("UnableToReadPacket")
			err = errors.New("UnableToReadPacket")
			continue
		}

		startPos = endPos
	}
	packetWrapper.SendDataType(EndTransfert)
	packetWrapper.ReadDataType()
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
	packetWrapper.SendDataType(PacketReceived)

	dataType := binary.LittleEndian.Uint32(buffer[:4])
	dataSize := binary.LittleEndian.Uint32(buffer[4:8])
	data := buffer[8 : 8+dataSize]
	packetWrapper.WriteDataInPacket(data, dataType)

	return nil
}

func (packetWrapper *PacketWrapper) ReadAllData() []byte {
	var data []byte
	packetWrapper.SendDataType(Ready)

	for {
		err := packetWrapper.readData()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			packetWrapper.SendDataType(UnableToReadPacket)
			continue
		}
		if packetWrapper.packet.dataSize > 0 {
			data = append(data, packetWrapper.packet.data[:packetWrapper.packet.dataSize]...)
		}
		if packetWrapper.packet.dataType == EndTransfert {
			return data
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
