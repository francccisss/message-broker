package server

import (
	"bytes"
	"encoding/binary"
	"log"
	"math"
	client "message-broker/internal/endpoint"
	"net"
	"sync"
)

type Server struct {
	addr        string
	port        string
	ln          net.Listener
	serverState chan error
}

func NewServer(addr string, port string, state chan error) Server {
	return Server{addr: addr, port: port, serverState: state}
}

func (s *Server) ServeTCP() {
	ln, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		log.Println("ERROR: Unable to start TCP server.")
		s.serverState <- err
	}
	s.ln = ln
	log.Printf("NOTIF: Server Success Listen to %s:%s", "localhost", s.port)
}
func (s *Server) ListenIncomingSegments() {
	log.Println("NOTIF: Accepting connections")
	for {
		c, err := s.ln.Accept()
		log.Println("NOTIF: Client connected")
		if err != nil {
			s.serverState <- err
			log.Println("ERROR: Unable to accept new TCP connections")
		}
		go HandleConnections(c)
	}
}

/*
Logic body for handling different message types and distributing to different performers
  - Endpoint messages will be handled by 'HandleEPMessage()'
  - Queue assertion will be handled by 'HandleQueueAssert()'
*/
func HandleConnections(c net.Conn) {
	var mux sync.Mutex
	ep := client.Endpoint{Mux: &mux, Conn: c}
	defer c.Close()

	// fixed sized header length to extract from message stream
	// Outer loop will always take 4 iterations

	var msgBuf bytes.Buffer
	for {
		var _ bytes.Buffer
		headerBuf := make([]byte, 4)
		_, err := c.Read(headerBuf)
		if err != nil {
			log.Println("ERROR: Unable to decode header prefix length")
			return
		}
		// encode to little endian
		expectedMsgLength := int(binary.LittleEndian.Uint32(headerBuf[:4]))

		log.Printf("Prefix Length Receieved: %d\n", expectedMsgLength)
		for {
			bodyBuf := make([]byte, 1024)
			_, err := c.Read(bodyBuf)
			if err != nil {
				log.Printf("ERROR: Unable to read the incoming message body ")
				break
			}
			remainingBytes := int(math.Min(float64(expectedMsgLength-msgBuf.Len()), float64(1024)))
			// Writes the from the minimum value of remainingBytes into the buffer up to
			// 1024 that is to be read into the bodyBuf
			_, err = msgBuf.Write(bodyBuf[:remainingBytes])
			if err != nil {
				log.Printf("ERROR: Unable to append bytes to the message buffer ")
				break
			}
			log.Printf("Total in slice: %+v\n", msgBuf.Bytes()[:expectedMsgLength])
			if msgBuf.Len() == expectedMsgLength {
				log.Printf("NOTIF: Receieved all values: %d", msgBuf.Len())
				go ep.MessageHandler(msgBuf)
				break
			}
		}
	}
}
