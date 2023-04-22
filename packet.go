package main

import (
	"encoding/binary"
	"errors"
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
	ReadData(conn net.Conn) error
	SendDataType(packetWrapper *PacketWrapper, conn net.Conn, dateType uint32)
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

func (packetWrapper *PacketWrapper) SendDataType(conn net.Conn, dateType uint32) {
	err := packetWrapper.WriteDataInPacket([]byte{}, dateType)
	if err != nil {
		return
	}
	err = packetWrapper.SendData(conn)
	if err != nil {
		log.Print("Can't send the packet")
	}
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
