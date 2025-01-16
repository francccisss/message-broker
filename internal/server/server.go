package server

import (
	"fmt"
	"log"
	"message-broker/internal/helper"
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

	header, body, err := extractMessage(readBuf)
	if err != nil {
		log.Printf("Unable to parse message")
		log.Println(err.Error())
		c.Write([]byte("Unable to parse message"))
		return
	}

	fmt.Printf("Header: %+v \n", header)
	fmt.Printf("Body: %+s\n", body)
}

// Generic function for extracting at a fixed size slice to be returned
func extractMessage(b []byte) (interface{}, string, error) {
	header := b[0:HEADER_SIZE]
	body := string(b[HEADER_SIZE:])
	headerUnmarshalled, err := helper.ParseIncomingHeader(header)
	if err != nil {
		return nil, "", err
	}
	return headerUnmarshalled, body, nil
}
