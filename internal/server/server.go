package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	client "message-broker/internal/endpoint.go"
	msgType "message-broker/internal/types"
	"net"
	"sync"
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

/*
Logic body for handling different message types and distributing to different performers
  - Endpoint messages will be handled by 'HandleEPMessage()'
  - Queue assertion will be handled by 'HandleQueueAssert()'
*/
func HandleConnections(c net.Conn) {
	var mux sync.Mutex
	ep := client.Endpoint{Mux: &mux}
	readBuf := make([]byte, 1024)
	for {
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
		endpointMsg, err := MessageParser(readBuf[:bytesRead])
		if err != nil {
			log.Printf("Error: Unable to parse message")
			log.Println(err.Error())
			c.Write([]byte("Error: Unable to parse message"))
			return
		}

		// type assertion switch statement for different processing
		switch msg := endpointMsg.(type) {
		case msgType.EPMessage:
			mux.Lock()
			ep.HandleEPMessage(msg)
			mux.Unlock()
		case msgType.Queue:
			mux.Lock()
			ep.HandleQueueAssert(msg)
			mux.Unlock()
		default:
			fmt.Println("Error: Unidentified type")
			c.Write([]byte("Error: Unidentified type: Types should consist of EPMessage | Queue "))
		}
	}
}

// Given an empty interface where it can store any value of and be represented as any type,
// we need to assert that its of some known type by matching the "MessageType" of the incoming message,
// once the "MessageType" of the message is known, we can then Unmarashal the message into the specified
// known type that matched the "MessageType"
func MessageParser(b []byte) (interface{}, error) {
	var temp map[string]interface{}
	if err := json.Unmarshal(b, &temp); err != nil {
		return nil, err
	}

	switch temp["MessageType"] {
	case "EPMessage":
		var epMsg msgType.EPMessage
		err := json.Unmarshal(b, &epMsg)
		if err != nil {
			return nil, err
		}
		return epMsg, nil
	case "Queue":
		var q msgType.Queue
		err := json.Unmarshal(b, &q)
		if err != nil {
			return nil, err
		}
		return q, nil
	default:
		return temp, fmt.Errorf("Not of any known message type")
	}
}
