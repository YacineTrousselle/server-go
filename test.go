package server

import (
	"fmt"
	"log"
	"time"
)

func Main() {
	fmt.Println("Test start")

	go clientStart()
	LaunchServer(nil)
}

func clientStart() {
	time.Sleep(2 * time.Second)
	client, _ := LaunchClient()
	defer client.packetWrapper.conn.Close()
	defer log.Println("I'm die. Thank you forever.")
	s := "ching chong"
	log.Println("client:", s)
	client.RequestFile(s)
}
