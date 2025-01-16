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
	bytesRead, err := c.Read(readBuf)
	if err != nil {
		log.Println(err.Error())
		return
	}
	endpointMsg, err := ParseMessage(readBuf[:bytesRead])
	if err != nil {
		log.Printf("Unable to parse message")
		log.Println(err.Error())
		c.Write([]byte("Unable to parse message"))
		return
	}

	// type assertion switch statement for different processing
	switch msg := endpointMsg.(type) {
	case EPMessage:
		ep.HandleEPMessage(msg)
	case Queue:
		ep.HandleQueueAssert(msg)
	default:
		fmt.Println("Unidentified type")
		c.Write([]byte("Unidentified type: Types should consist of EPMessage | Queue "))
	}
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

/*
 Create an entry in a hashmap for a new Message queue
*/

func (ep Endpoint) HandleQueueAssert(m Queue) {
	fmt.Printf("Message is of type: %s\n", m.Type)
	fmt.Printf("Creating/Asserting Queue with Route: %+v \n", m.QueueHeader.Name)
}

/*
Route the EPMessage to the appropriate channel described in the current
EPMessages' header data
Use the Route property of the EPMessage to locate the appropriate Route within
the Route Map
*/

func (ep Endpoint) HandleEPMessage(m EPMessage) {
	fmt.Printf("Message is of type: %s\n", m.Type)
	fmt.Printf("Message: %+v \n", m)
	fmt.Printf("Send message to Route: %+v \n", m.Route)
}

// Unmasrshalling returns a generic interface not `interface{}`
// need to hardcode type assertion using internal property "Type"

func ParseMessage(b []byte) (interface{}, error) {
	var temp map[string]interface{}
	if err := json.Unmarshal(b, &temp); err != nil {
		return nil, err
	}

	switch temp["Type"] {
	case "EPMessage":
		var epMsg EPMessage
		err := json.Unmarshal(b, &epMsg)
		if err != nil {
			return nil, err
		}
		return epMsg, nil
	case "Queue":
		var q Queue
		err := json.Unmarshal(b, &q)
		if err != nil {
			return nil, err
		}
		return q, nil
	default:
		return temp, fmt.Errorf("Not of any known message type")
	}
}
