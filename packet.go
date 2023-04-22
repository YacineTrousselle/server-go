package server

import (
	"encoding/binary"
	"errors"
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
}

func NewPacketWrapper(maxSize uint32) *PacketWrapper {
	packet := &Packet{
		dataType: 0,
		dataSize: 0,
		data:     make([]byte, maxSize-8),
	}
	packetWrapper := &PacketWrapper{
		packet:  packet,
		maxSize: maxSize,
	}
	return packetWrapper
}

type PacketInterface interface {
	WriteDataInPacket(data []byte, dataType uint32) error
	sendData(conn net.Conn) error
	SendAllData(data []byte, dataType uint32, conn net.Conn)
	SendDataType(packetWrapper *PacketWrapper, conn net.Conn, dateType uint32) error
	readData(conn net.Conn) error
	ReadAllData(conn net.Conn) []byte
}

func (packetWrapper *PacketWrapper) WriteDataInPacket(data []byte, dataType uint32) error {
	lenData := uint32(len(data))
	if lenData > packetWrapper.maxSize-8 {
		return errors.New("data size is higher than the max")
	}
	packetWrapper.packet.dataSize = lenData
	packetWrapper.packet.dataType = dataType
	copy(data, packetWrapper.packet.data[:lenData])

	return nil
}

func (packetWrapper *PacketWrapper) SendAllData(data []byte, dataType uint32, conn net.Conn) {
	startBlock := uint32(0)
	endBlock := uint32(0)
	for {
		if startBlock == uint32(len(data)) {
			break
		}
		if startBlock+packetWrapper.maxSize-8 > uint32(len(data)) {
			endBlock = uint32(len(data))
		} else {
			endBlock = endBlock + packetWrapper.maxSize - 8
		}
		packetWrapper.WriteDataInPacket(data[startBlock:endBlock], dataType)
		packetWrapper.sendData(conn)
		startBlock = endBlock
	}
	packetWrapper.SendDataType(conn, EndTransfert)
}

func (packetWrapper *PacketWrapper) sendData(conn net.Conn) error {
	buffer := make([]byte, packetWrapper.maxSize)
	binary.LittleEndian.PutUint32(buffer[0:], packetWrapper.packet.dataType)
	binary.LittleEndian.PutUint32(buffer[4:], packetWrapper.packet.dataSize)
	copy(buffer[8:], packetWrapper.packet.data)

	_, err := conn.Write(buffer)
	if err != nil {
		return err
	}

	return nil
}

func (packetWrapper *PacketWrapper) SendDataType(conn net.Conn, dateType uint32) error {
	err := packetWrapper.WriteDataInPacket([]byte{}, dateType)
	if err != nil {
		return err
	}
	err = packetWrapper.sendData(conn)

	return err
}

func (packetWrapper *PacketWrapper) readData(conn net.Conn) error {
	buffer := make([]byte, packetWrapper.maxSize)
	for {
		packetLength, err := conn.Read(buffer)
		if err != nil {
			return err
		}
		if uint32(packetLength) == packetWrapper.maxSize {
			break
		}
	}
	dataType := binary.LittleEndian.Uint32(buffer[:4])
	dataSize := binary.LittleEndian.Uint32(buffer[4:8])
	data := buffer[8 : 8+dataSize]

	err := packetWrapper.WriteDataInPacket(data, dataType)
	if err != nil {
		return err
	}

	return nil
}

func (packetWrapper *PacketWrapper) ReadAllData(conn net.Conn) []byte {
	var data []byte
	currentSize := uint32(0)
	for {
		err := packetWrapper.readData(conn)
		if err != nil {
			packetWrapper.SendDataType(conn, UnableToReadPacket)
			continue
		}
		if packetWrapper.packet.dataType == EndTransfert {
			break
		}
		copy(data[currentSize:currentSize+packetWrapper.packet.dataSize], packetWrapper.packet.data[:packetWrapper.packet.dataSize])
		currentSize = currentSize + packetWrapper.packet.dataSize
		packetWrapper.SendDataType(conn, PacketReceived)
	}

	return data
}
