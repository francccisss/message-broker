package main

import (
	"log"
	"message-broker/internal/server"
)

func main() {
	serverState := make(chan error)
	s := server.NewServer("", "5671", serverState)
	s.ServeTCP()
	s.ListenIncomingSegments()
	state := <-serverState
	// I dont know what im doing but I made it so
	// that when the server returns an error it
	// would instead pass it to the server state
	// to handle it in a single place using channels
	log.Println("NOTIF: Exiting application")
	if state != nil {
		log.Panicln(state.Error())
	}
	log.Println("NOTIF: Server shutdown")

}
