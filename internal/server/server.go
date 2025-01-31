package server

import (
	"bufio"
	"bytes"
	"encoding/binary"
	// "fmt"
	"log"
	"math"
	client "message-broker/internal/endpoint"
	"net"
	"sync"
)

const (
	HEADER_SIZE       = 4
	DEFAULT_READ_SIZE = 50
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
		go HandleIncomingRequests(c)
	}
}

/*
Logic body for handling different message types and distributing to different performers
  - Endpoint messages will be handled by 'HandleEPMessage()'
  - Queue assertion will be handled by 'HandleQueueAssert()'
*/

func HandleIncomingRequests(c net.Conn) {
	var mux sync.Mutex
	connBufReader := bufio.NewReader(c)

	ep := client.Endpoint{Mux: &mux, Conn: c}
	defer c.Close()

	// fixed sized header length to extract from message stream
	// Outer loop will always take 4 iterations
	headerBuf := make([]byte, HEADER_SIZE)
	readSize := DEFAULT_READ_SIZE
	for {
		var msgBuf bytes.Buffer
		_, err := c.Read(headerBuf)
		if err != nil {
			log.Println("ERROR: Unable to decode header prefix length")
			return
		}

		expectedMsgLength := int(binary.LittleEndian.Uint32(headerBuf[:HEADER_SIZE]))
		log.Printf("Prefix Length Receieved: %d\n", expectedMsgLength)
		for {
			bodyBuf := make([]byte, readSize)
			_, err := connBufReader.Read(bodyBuf)
			if err != nil {
				log.Printf("ERROR: Unable to read the incoming message body ")
				break
			}

			// store bytes from stream up to the current readsize length into the
			// msgBuf (msgBuf is the current accumulated requested stream from client)
			_, err = msgBuf.Write(bodyBuf[:])
			if err != nil {
				log.Printf("ERROR: Unable to write incoming bytes to the message buffer ")
				break
			}

			// Updates the readsize for the next stream of bytes to be captured in bulk
			// This formula returns the minimum int between the two, if there is space
			// to fit the stream of bytes in the bodyBuf then return current readSize which is the DEFAULT_READ_SIZE
			// else if current readSize is greater than the remaining bytes left from the expected message
			// return n bytes up to the length of the remaining bytes of the current message.
			readSize = int(math.Min(float64(expectedMsgLength-msgBuf.Len()), float64(readSize)))

			log.Printf("CURRENT TOTAL IN MSGBUF: %+v\n", msgBuf.Len())
			log.Printf("REMAINING BYTES NEEDED: %+v\n", expectedMsgLength-msgBuf.Len())
			// finishes the current stream request
			if msgBuf.Len() == expectedMsgLength {
				go ep.MessageHandler(msgBuf)
				readSize = DEFAULT_READ_SIZE
				break
			}
		}
	}
}
