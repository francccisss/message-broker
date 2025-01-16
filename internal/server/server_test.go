package server_test

import (
	"encoding/json"
	"fmt"
	"log"
	"message-broker/internal/server"
	"net"
	"sync"
	"testing"
	"time"
)

const clientCount = 10

func TestServerConnections(t *testing.T) {

	s := server.NewServer("", "8080")
	serverState := make(chan bool)
	ln, err := s.ServeTCP()
	if err != nil {
		t.Fatalf(err.Error())
	}

	go func(s net.Listener, state chan bool) {
		for {
			conn, err := ln.Accept()
			if err != nil {
				state <- false
			}
			log.Printf("Connected %s", conn.RemoteAddr())
		}
	}(ln, serverState)

	log.Println("Waiting for server")
	dr := time.Duration(3) * time.Millisecond
	time.Sleep(dr)
	log.Println("Start Connecting")
	go createClients(clientCount)
	select {
	case state := <-serverState:
		if state == false {
			t.Fatalf("Unable to listen to client connections")
		}
	}
}

func TestProtocolParsing(t *testing.T) {
	log.Println("Start protocol parsing test")
	endpointMessage := server.Send{
		Type:       "Send",
		Route:      "Somewhere",
		HeaderSize: 1024,
		Body:       "This is a message sent by a client",
	}

	b, err := json.Marshal(endpointMessage)
	if err != nil {
		t.Fatalf("Unable to Marshal endpoint message")
	}
	// Should return an abstract message with a Type
	endpointMsg, err := server.ParseMessage[struct{ Type string }](b)
	if err != nil {
		log.Println(err.Error())
		t.Fatalf("Unable to Marshal endpoint message")

	}
	// Message Dispatcher
	switch endpointMsg.Type {
	case "Send":
		fmt.Printf("Message is of type: %s\n", endpointMsg.Type)
		fmt.Printf("Message: %+v \n", endpointMessage)

	case "Assert":
		fmt.Printf("Message is of type: %s\n", endpointMsg.Type)
		fmt.Printf("Message: %+v \n", endpointMessage)

	case "Receive":
		fmt.Printf("Message is of type: %s\n", endpointMsg.Type)
		fmt.Printf("Message: %+v \n", endpointMessage)

	default:
		fmt.Println("Not any type")
	}

}

func createClients(num int) {
	var wg sync.WaitGroup
	for i := range num {
		wg.Add(i)
		go func(i int) {
			defer wg.Done()

			conn, err := net.Dial("tcp", "localhost:8080")
			log.Printf("Client connected %d", i)
			if err != nil {
				log.Fatalf(err.Error())
			}

			defer conn.Close()
			t := time.Duration(i) * time.Millisecond
			log.Printf("Time to sleep %d", t)
			time.Sleep(t)
			log.Printf("Wake up thread")
		}(i)
	}
	wg.Wait()
	log.Printf("Done")
}
