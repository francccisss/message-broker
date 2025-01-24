package server_test

import (
	"log"
	"message-broker/internal/server"
	"testing"
)

const clientCount = 10

func TestServerConnections(t *testing.T) {

	serverState := make(chan error)
	s := server.NewServer("", "5671", serverState)
	s.ServeTCP()
	s.ListenIncomingSegments()
	state := <-serverState

	if state != nil {
		log.Println("Exiting application")
		log.Panicln(state.Error())
	}
	log.Println("Servert shutdown")
}

func TestEndpointDelivery(t *testing.T) {

}
