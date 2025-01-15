package main

import (
	"log"
	"message-broker/internal/server"
)

func main() {
	s := server.NewServer("", "5671")

	ln, err := s.ServeTCP()
	if err != nil {
		log.Println("Exiting application")
		log.Panicf(err.Error())
	}

	for {
		log.Println("Accepting connections")
		_, err := ln.Accept()
		if err != nil {
			log.Printf("Unable to accept new TCP connections")
			log.Panicf(err.Error())
		}
	}
}
