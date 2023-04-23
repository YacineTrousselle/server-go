package server

import (
	"net"
)

type Client struct {
	packetWrapper *PacketWrapper
}

type ClientInterface interface {
	RequestFile(filename string) (file []byte, err error)
}

func LaunchClient() (Client, error) {
	conn, err := net.Dial(TYPE, HOST+":"+PORT)
	if err != nil {
		return Client{nil}, err
	}
	client := Client{NewPacketWrapper(MaxPacketSize, conn)}

	return client, nil
}

func (client Client) RequestFile(filename string) (file []byte, err error) {
	data := []byte(filename)
	client.packetWrapper.SendAllData(data, RequestFile)

	return nil, nil
}
