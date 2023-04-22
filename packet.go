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
	SendData(conn net.Conn) error
	SendDataType(packetWrapper *PacketWrapper, conn net.Conn, dateType uint32) error
	ReadData(conn net.Conn) error
	ReadAllData(conn net.Conn) (string, error)
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

func (packetWrapper *PacketWrapper) SendData(conn net.Conn) error {
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
	err = packetWrapper.SendData(conn)

	return err
}

func (packetWrapper *PacketWrapper) ReadData(conn net.Conn) error {
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
	data := buffer[8:dataSize]

	err := packetWrapper.WriteDataInPacket(data, dataType)
	if err != nil {
		return err
	}

	return nil
}

func (packetWrapper *PacketWrapper) ReadAllData(conn net.Conn) (string, error) {
	var data []byte
	currentSize := uint32(0)
	for {
		err := packetWrapper.ReadData(conn)
		if err != nil {
			return "", err
		}
		if packetWrapper.packet.dataType == EndTransfert {
			break
		}
		copy(data[currentSize:currentSize+packetWrapper.packet.dataSize], packetWrapper.packet.data[:packetWrapper.packet.dataSize])
		currentSize = currentSize + packetWrapper.packet.dataSize
	}

	return string(data), nil
}
