package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func AppendPrefixLength(b []byte) ([]byte, error) {

	prefixBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(prefixBuf, uint32(len(b)))
	var bufWriter bytes.Buffer

	_, err := bufWriter.Write(prefixBuf)
	if err != nil {
		fmt.Println("ERROR: Unable write prefix to buffer")
		return []byte{}, err
	}

	value := binary.LittleEndian.Uint32(bufWriter.Bytes())
	fmt.Printf("NOTIF: Appended: %d\n", value)

	_, err = bufWriter.Write(b)
	if err != nil {
		fmt.Println("ERROR: Unable write message body to buffer")
		return []byte{}, err
	}

	fmt.Printf("Total: %+v\n", bufWriter.Len())
	return bufWriter.Bytes(), nil
}
