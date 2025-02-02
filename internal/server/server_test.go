package server_test

import (
	"fmt"
	"log"
	"message-broker/internal/server"
	"testing"
)

const clientCount = 10

func TestServerConnections(t *testing.T) {

	serverState := make(chan error)
	s := server.NewServer("", "5671", serverState)
	s.ServeTCP()
	s.ListenConnections()
	state := <-serverState

	if state != nil {
		fmt.Println("Exiting application")
		log.Panicln(state.Error())
	}
	fmt.Println("Servert shutdown")
}

func TestEndpointDelivery(t *testing.T) {

}
