package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	client "message-broker/internal/endpoint"
	"net"
	"sync"
)

const (
	HEADER_SIZE       = 4
	DEFAULT_READ_SIZE = 1024
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
		fmt.Println("ERROR: Unable to start TCP server.")
		s.serverState <- err
	}
	s.ln = ln
	fmt.Printf("NOTIF: Server Success Listen to %s:%s\n", "localhost", s.port)
}
func (s *Server) ListenConnections() {
	fmt.Println("NOTIF: Accepting connections")
	for {
		c, err := s.ln.Accept()
		fmt.Println("NOTIF: Client connected")
		if err != nil {
			s.serverState <- err
			fmt.Println("ERROR: Unable to accept new TCP connections")
		}
		go HandleIncomingRequests(c)
	}
}

/*
Logic body for handling different message types and distributing to different performers

  - Endpoint messages will be handled by 'HandleEPMessage()'

  - Queue assertion will be handled by 'HandleQueueAssert()'

    TODO remove consumer connection in the message queue if they are disconnected
*/
func HandleIncomingRequests(c net.Conn) {
	var mux sync.Mutex

	ep := client.Endpoint{Mux: &mux, Conn: c}
	defer c.Close()

	// fixed sized header length to extract from message stream
	// Outer loop will always take 4 iterations
	headerBuf := make([]byte, HEADER_SIZE)
	for {
		var msgBuf bytes.Buffer
		_, err := c.Read(headerBuf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("NOTIF: Client has abrubtly terminated the connection")
				return
			}
		}

		expectedMsgLength := int(binary.LittleEndian.Uint32(headerBuf[:HEADER_SIZE]))

		// Initial read size to check if whether incoming request exceeds
		// current read size then read up to DEFAULT_READ_SIZE else only
		// up to remaning bytes to be read instead
		// This formula returns the minimum int between the two, if there is space
		// to fit the stream of bytes in the bodyBuf then return current readSize which is the DEFAULT_READ_SIZE
		// else if current readSize is greater than the remaining bytes left from the expected message
		// return n bytes up to the length of the remaining bytes of the current message.
		currentReadSize := int(math.Min(float64(expectedMsgLength-msgBuf.Len()), float64(DEFAULT_READ_SIZE)))

		fmt.Printf("NOTIF: Prefix Length Receieved: %d\n", expectedMsgLength)
		for {

			// creates a buffer up to the calculated readSize
			bodyBuf := make([]byte, currentReadSize)

			_, err := c.Read(bodyBuf)
			if err != nil {
				fmt.Printf("ERROR: Unable to read the incoming message body ")
				break
			}

			// store bytes from stream up to the current readsize length into the
			// msgBuf (msgBuf is the current accumulated requested stream from client)
			_, err = msgBuf.Write(bodyBuf[:])
			if err != nil {
				fmt.Printf("ERROR: Unable to write incoming bytes to the message buffer ")
				break
			}

			// Updates the readSize for the next bytes to be read
			remainingBytesLen := expectedMsgLength - msgBuf.Len()
			if currentReadSize < remainingBytesLen {
				currentReadSize = remainingBytesLen
			}
			// finishes the current stream request
			if msgBuf.Len() == expectedMsgLength {
				go ep.MessageHandler(msgBuf)
				fmt.Println("NOTIF: Message sequence complete.")
				break
			}
		}
	}
}
