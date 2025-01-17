package server_test

import (
	"log"
	"message-broker/internal/server"
	"net"
	"testing"
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
	log.Println("Start Connecting")
	select {
	case state := <-serverState:
		if state == false {
			t.Fatalf("Unable to listen to client connections")
		}
	}
}
