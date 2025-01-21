package router

import (
	"fmt"
	protocol "message-broker/internal"
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

/*
When a route is matched within the RouteTable a type of Route will be accessible
  - Sender will be passed in as an argument to send back some response or ack
    back to the sender.
  - Route contains the connections within the route including the sender itself
  - Each Route will contain a Queue which is a multi-dimensional array of bytes containing
    each item in the queue messages
  - an error is thrown if no route matched with the message Route
*/
func RouteMessage(m protocol.EPMessage) error {
	table := GetRouteTable()
	_, exists := table[m.Route]
	if !exists {
		return fmt.Errorf("ERROR: A message for route: %s does not exist, either specify an existing route or create one using `AssertQueue`", m.Route)
	}
	return nil
}

func CreateQueue(q protocol.Queue) {
	table := GetRouteTable()
	_, exists := table[q.Name]
	if !exists {
		table[q.Name] = Route{
			Type:        q.Type,
			Name:        q.Name,
			Durable:     q.Durable,
			Connections: []*net.Conn{},
		}
	}
}
