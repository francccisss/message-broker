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
			go server.HandleConnections(conn)
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

			endpointMessage := server.EPMessage{
				Type:       "EPMessage",
				Route:      "Somewhere",
				HeaderSize: 1024,
				Body:       "This is a message sent by a client",
			}

			b, err := json.Marshal(endpointMessage)

			if err != nil {
				log.Println(err.Error())
				return
			}
			_, err = conn.Write(b)
			if err != nil {
				log.Println(err.Error())
				return
			}
		}(i)
	}
	wg.Wait()
	log.Printf("Done")
}

func TestProtocolParsing(t *testing.T) {
	log.Println("Start protocol parsing test")
	endpointMessage := server.EPMessage{
		Type:       "EPMessage",
		Route:      "Somewhere",
		HeaderSize: 1024,
		Body:       "This is a message sent by a client",
	}

	b, err := json.Marshal(endpointMessage)
	if err != nil {
		t.Fatalf("Unable to Marshal endpoint message")
	}
	// Should return an abstract message with a Type
	endpointMsg, err := server.ParseMessage(b)
	if err != nil {
		log.Println(err.Error())
		t.Fatalf("Unable to Marshal endpoint message")

	}
	ep := server.Endpoint{}
	fmt.Println(endpointMsg)
	// Message Dispatcher
	switch msg := endpointMsg.(type) {
	case server.EPMessage:
		ep.HandleEPMessage(msg)
	case server.Queue:
		ep.HandleQueueAssert(msg)
	default:
		t.Fatalf("Not any type")
	}

}
