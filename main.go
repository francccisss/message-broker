package main

import (
	"fmt"
	"log"
	"message-broker/internal/server"
)

func main() {
	serverState := make(chan error)
	s := server.NewServer("", "5671", serverState)
	s.ServeTCP()
	s.ListenConnections()
	state := <-serverState
	// I dont know what im doing but I made it so
	// that when the server returns an error it
	// would instead pass it to the server state
	// to handle it in a single place using channels
	fmt.Println("NOTIF: Exiting application")
	if state != nil {
		log.Panicln(state.Error())
	}
	fmt.Println("NOTIF: Server shutdown")

}
