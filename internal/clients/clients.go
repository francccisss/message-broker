package clients

import "net"

type MsgClient struct {
	Conn         net.Conn
	ChannelCount uint32
}
