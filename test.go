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
	time.Sleep(5 * time.Second)
	client, _ := LaunchClient()
	defer client.Close()

	s := "ching chong"
	log.Println("client:", s)
	client.RequestFile(s)
}
