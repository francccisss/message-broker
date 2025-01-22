package router

import (
	"message-broker/internal/utils/queue"
	"net"
)

/*
Router is responsible for routing messages based on each messages'
`Route` property the Header to route the message to the appropriate
TCP Connection, the application endpoint will be the one responsible for
handling the multiplexing/demultiplexing using a `message dispatcher`
*/

type Route struct {
	Type         string
	Name         string
	Connections  []*net.Conn
	MessageQueue queue.Queue
	Durable      bool
}

var table = map[string]Route{}

type P2P interface{}
type PubSub interface{}

func GetRouteTable() map[string]Route {
	return table
}
