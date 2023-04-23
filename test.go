package server

import (
	"fmt"
	"io"
	"net"
	"time"
)

func Main() {
	fmt.Println("Test start")

	go clientHandleCo()
	LaunchServer(serverHandleCo)
}

func serverHandleCo(conn net.Conn) {
	defer conn.Close()
	packetWrapper := NewPacketWrapper(MaxPacketSize, conn)
	for {
		err := packetWrapper.ReadDataType()
		if err == io.EOF {
			fmt.Println("Connection finish")
			return
		}
		switch packetWrapper.packet.dataType {
		case Test:
			packetWrapper.SendDataType(Ready)
			data := packetWrapper.ReadAllData()
		}
	}
}

func clientHandleCo() {
	time.Sleep(4 * time.Second)
	conn, err := LaunchClient()
	for err != nil {
		conn, err = LaunchClient()
	}
	fmt.Println("Client start")
	defer fmt.Println("Client end")
	defer conn.Close()
	p := NewPacketWrapper(MaxPacketSize, conn)

	s := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi ac lacus fringilla, tempor erat eget, congue lectus. Integer tempus eros vel vulputate convallis. Vivamus dui sem, blandit at tristique fermentum, rhoncus nec augue. Quisque nibh lorem, auctor pretium sollicitudin volutpat, scelerisque nec diam. Aenean vitae mattis sapien. Pellentesque tincidunt eu dolor in ultricies. Morbi libero dui, rutrum vitae mollis non, interdum ut dolor. Integer blandit eros nibh, vel ullamcorper dui blandit ut. Cras est lorem, placerat vitae nibh sit amet, ullamcorper suscipit arcu. Nullam dapibus elit elit, at placerat augue semper nec. Quisque venenatis cursus feugiat. Sed sollicitudin pellentesque ligula quis pharetra. Pellentesque eget placerat felis. Nam congue tortor vitae ex auctor convallis.0Duis viverra tempus sem, sit amet aliquam erat volutpat in. Praesent eget odio sit amet quam dapibus rhoncus cursus quis sem. Maecenas consequat blandit augue vel finibus. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Aliquam a nibh velit. Cras molestie risus vel ipsum pretium malesuada. Nulla porta mollis nulla, vitae sagittis urna ornare at. Phasellus eleifend tellus pulvinar dolor dapibus, nec vehicula sem pulvinar. Nullam dignissim sem tempus, varius ante eu, commodo arcu. Duis a ipsum vel purus auctor tempor tempus nec turpis. Vivamus mollis libero sed lectus tincidunt, sit amet pulvinar lacus aliquet. Etiam consectetur mauris sit amet risus hendrerit, at efficitur odio varius. Nullam interdum molestie augue, eu elementum nulla eleifend posuere. Proin facilisis nibh in lacus porttitor rutrum. Nullam ante ex, interdum non velit commodo, accumsan malesuada lectus. ")
	p.SendAllData(s, Test)
	fmt.Println("PacketReceived: ", p.packet.dataType == PacketReceived)

	time.Sleep(1 * time.Second)
	p.SendAllData([]byte("I'm a teapot"), Test)
	fmt.Println("PacketReceived: ", p.packet.dataType == PacketReceived)
}
