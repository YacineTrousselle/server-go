package old

import (
	"encoding/binary"
	"errors"
	"github.com/YacineTrousselle/server-go"
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
	packetWrapper.SendDataType(conn, server.EndTransfert)
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
	dataReceived := make([]byte, packetWrapper.maxSize)
	currentLength := uint32(0)
	for {
		packetLength, err := conn.Read(buffer)
		if packetLength == 0 {
			return nil
		}
		copy(buffer, dataReceived[currentLength:currentLength+uint32(packetLength)])
		currentLength = currentLength + uint32(packetLength)
		if err != nil {
			return err
		}
		if currentLength == packetWrapper.maxSize {
			break
		}
	}
	dataType := binary.LittleEndian.Uint32(dataReceived[:4])
	dataSize := binary.LittleEndian.Uint32(dataReceived[4:8])
	data := dataReceived[8 : 8+dataSize]

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
			packetWrapper.SendDataType(conn, server.UnableToReadPacket)
			continue
		}
		if packetWrapper.packet.dataType == server.EndTransfert {
			break
		}
		copy(data[8+currentSize:8+currentSize+packetWrapper.packet.dataSize], packetWrapper.packet.data[8:packetWrapper.packet.dataSize])
		currentSize = currentSize + packetWrapper.packet.dataSize
		packetWrapper.SendDataType(conn, server.PacketReceived)
	}

	return data
}
