package server

import (
	"bytes"
	"encoding/binary"
	// "errors"
	// "io"
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
	// TODO FIX
	// FIX THIS IT ONLY READS 1024 BUT WHEN MESSAGES ARE > 1024 IT THROWS
	// IT THROWS AN ERROR BECAUSE IT CANT FULLY PARSE THE JSON DATA

	// Using prefix length data stream
	// Header has 4 bytes which is the number of bytes
	// to be expected by the receiver to receive and parse once
	// it has all of the required bytes defined by the header

	// Buffer is only read 1024 bytes,
	//   - Header 4 -\
	//                > Total size 104 bytes to be read
	//   - Body 100 -/

	// its easy as extracting msgBuf := readBuf[4:msgLength]
	// this excludes reading the prefix length header
	// but what if body.length >= readBuf?
	// we can check is readBuf >= body.length?
	//   - yes? read remaining bytes
	//   - no? listen to the next arrviving bytes

	defer c.Close()
	var msgBuf bytes.Buffer
	prefBuf := make([]byte, 4)
	var msglength int

	headerPrefixLength := 4

	for {
		readBuf := make([]byte, 1024)
		_, err := c.Read(readBuf)
		if err != nil {
			log.Println("ERROR: Unable to decode header prefix length")
		}
		msglength, _ = binary.Decode(prefBuf, binary.LittleEndian, readBuf[:headerPrefixLength])
		remainingBytes := int(math.Min(float64(msglength-msgBuf.Len()-headerPrefixLength), float64(1024)))

		_, _ = msgBuf.Write(readBuf[:remainingBytes])
		if remainingBytes <= msglength {
			continue
		}
		if remainingBytes >= msglength || msgBuf.Len()-headerPrefixLength == msglength {
			ep.MessageHandler(msgBuf)
		}

		// if errors.Is(err, io.EOF) {
		// 	log.Println("NOTIF: End of file stream")
		// 	return
		// }
		// if err != nil {
		// 	log.Println("ERROR: Unable to read message")
		// 	log.Println(err.Error())
		// 	return
		// }
		// if err != nil {
		// 	log.Println("ERROR: Unable to append message stream")
		// 	log.Println(err.Error())
		// 	return
		// }
		// TODO gind a way to determine the lenfrg of the incoming message from client

	}
}
