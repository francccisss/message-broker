package utils

import (
	"bytes"
	"encoding/binary"
	"log"
)

func AppendPrefixLength(b []byte) ([]byte, error) {

	prefixBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(prefixBuf, uint32(len(b)))
	var bufWriter bytes.Buffer

	_, err := bufWriter.Write(prefixBuf)
	if err != nil {
		log.Println("ERROR: Unable write prefix to buffer")
		return []byte{}, err
	}

	value := binary.LittleEndian.Uint32(bufWriter.Bytes())
	log.Printf("Message Body Size to be sent: %d\n", value)
	log.Printf("Message Body Size to be sent in slice: %+v\n", prefixBuf)

	_, err = bufWriter.Write(b)
	if err != nil {
		log.Println("ERROR: Unable write message body to buffer")
		return []byte{}, err
	}

	log.Printf("Total: %+v\n", bufWriter.Len())
	return bufWriter.Bytes(), nil
}
