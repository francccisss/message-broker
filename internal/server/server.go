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
	readBuf := make([]byte, 1024)
	_, err := c.Read(readBuf)

	if err != nil {
		log.Println(err.Error())
		return
	}

	endpointMessage, err := ParseMessage[any](readBuf)
	if err != nil {
		log.Printf("Unable to parse message")
		log.Println(err.Error())
		c.Write([]byte("Unable to parse message"))
		return
	}

	// type assertion switch statement for different processing

	fmt.Printf("Message: %+v \n", endpointMessage)
}

func ParseMessage[T any](b []byte) (T, error) {
	var v T
	err := json.Unmarshal(b, &v)
	if err != nil {
		return v, err
	}
	return v, err
}
