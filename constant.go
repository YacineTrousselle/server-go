package server

const (
	TYPE          = "tcp"
	HOST          = "localhost"
	PORT          = "5555"
	MaxPacketSize = 256
)

const (
	DataType = iota
	UnableToReadPacket
	InvalidInputError
	EndTransfert

	PacketReceived
	RequestFile
	FileNotFound
	FileData
)
