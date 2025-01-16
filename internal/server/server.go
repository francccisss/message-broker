package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

const HEADER_SIZE = 1024

type Server struct {
	addr string
	port string
}

func NewServer(addr string, port string) Server {
	return Server{addr: addr, port: port}
}

func (s Server) ServeTCP() (net.Listener, error) {
	ln, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		log.Println("Unable to start TCP server.")
		return nil, err
	}
	log.Printf("Server Success Listen to %s:%s", "localhost", s.port)
	return ln, nil
}

// Read Headers
func HandleConnections(c net.Conn) {

	defer c.Close()
	ep := Endpoint{}
	readBuf := make([]byte, 1024)
	_, err := c.Read(readBuf)

	if err != nil {
		log.Println(err.Error())
		return
	}

	endpointMsg, err := ParseMessage(readBuf)
	if err != nil {
		log.Printf("Unable to parse message")
		log.Println(err.Error())
		c.Write([]byte("Unable to parse message"))
		return
	}

	// type assertion switch statement for different processing
	switch msg := endpointMsg.(type) {
	case EPMessage:
		fmt.Printf("Message is of type: %s\n", msg.Type)
		fmt.Printf("Message: %+v \n", msg)
		ep.HandleEPMessage(msg)

		// Creating/Asserting message queue
	case Queue:
		fmt.Printf("Message is of type: %s\n", msg.Type)
		fmt.Printf("Message: %+v \n", msg)
		ep.HandleQueueAssert(msg)

	default:
		fmt.Println("Unidentified type")
		c.Write([]byte("Unidentified type: Types should consist of Send | Assert | Receive"))
	}
}

func AssertInterfaceType[T any](incomingMsg interface{}) (T, error) {
	var v T
	if message, ok := incomingMsg.(T); ok {
		return message, nil
	}
	return v, fmt.Errorf("Unable to assert type of message")
}

/*
 Endpoint is an abstraction of a connected application using an endpoint API
 on the client side for connecting to the server
*/

type Endpoint struct {
}

type EPHandler interface {
	HandleQueueAssert()
	HandleEPMessage()
}

func (ep Endpoint) HandleQueueAssert(m Queue) {

}

func (ep Endpoint) HandleEPMessage(m EPMessage) {

}

func ParseMessage(b []byte) (interface{}, error) {
	var v interface{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return v, err
	}
	return v, err
}
