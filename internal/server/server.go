package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	protocol "message-broker/internal"
	client "message-broker/internal/endpoint.go"
	"net"
)

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
	ep := client.Endpoint{}
	readBuf := make([]byte, 1024)
	bytesRead, err := c.Read(readBuf)
	if errors.Is(err, io.EOF) {
		log.Println("Error: Abrupt client disconnect")
		return
	}
	if err != nil {
		log.Println("Error: Unable to read message")
		log.Println(err.Error())
		return
	}
	endpointMsg, err := ParseMessage(readBuf[:bytesRead])
	if err != nil {
		log.Printf("Error: Unable to parse message")
		log.Println(err.Error())
		c.Write([]byte("Error: Unable to parse message"))
		return
	}

	// type assertion switch statement for different processing
	switch msg := endpointMsg.(type) {
	case protocol.EPMessage:
		ep.HandleEPMessage(msg)
	case protocol.Queue:
		ep.HandleQueueAssert(msg)
	default:
		fmt.Println("Error: Unidentified type")
		c.Write([]byte("Error: Unidentified type: Types should consist of EPMessage | Queue "))
	}
}

// Unmasrshalling returns a generic interface not `interface{}`
// need to hardcode type assertion using internal property "Type"

func ParseMessage(b []byte) (interface{}, error) {
	var temp map[string]interface{}
	if err := json.Unmarshal(b, &temp); err != nil {
		return nil, err
	}

	switch temp["MessageType"] {
	case "EPMessage":
		var epMsg protocol.EPMessage
		err := json.Unmarshal(b, &epMsg)
		if err != nil {
			return nil, err
		}
		return epMsg, nil
	case "Queue":
		var q protocol.Queue
		err := json.Unmarshal(b, &q)
		if err != nil {
			return nil, err
		}
		return q, nil
	default:
		return temp, fmt.Errorf("Not of any known message type")
	}
}
