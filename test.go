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

	s := "go.mod"
	data, _ := client.RequestFile(s)
	log.Println("client:", string(data))
}
