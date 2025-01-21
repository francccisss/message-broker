package router

import (
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
	MessageQueue [][]byte
	Durable      bool
}

type P2P interface{}
type PubSub interface{}

var Table map[string]Route

func GetRouteTable() map[string]Route {
	return Table
}
