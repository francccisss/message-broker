package client

import (
	"fmt"
	protocol "message-broker/internal"
	router "message-broker/internal/router"
	"net"
)

/*
Endpoint is an abstraction of a connected application using an endpoint API
on the client side for connecting to the server

The Interface is used to handle different message types which is defined in the
message's Header `Type` property
*/
type Endpoint struct {
}

type EPHandler interface {
	HandleQueueAssert()
	HandleEPMessage()
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
func (ep Endpoint) HandleQueueAssert(q protocol.Queue) {
	fmt.Printf("Message is of type: %s\n", q.Type)
	table := router.GetRouteTable()
	_, exists := table[q.Name]
	if !exists {
		table[q.Name] = router.Route{
			Type:        q.Type,
			Name:        q.Name,
			Durable:     q.Durable,
			Connections: []*net.Conn{},
		}
	}
	fmt.Printf("Creating/Asserting Queue with Route: %+v \n", q.Name)
}

/*
Handling Endpoint Messages
  - Route the EPMessage to the appropriate channel described in the current
    EPMessages' header data
  - Use the Route property of the EPMessage to locate the appropriate Route
    within the Route Map
*/
func (ep Endpoint) HandleEPMessage(m protocol.EPMessage) error {
	fmt.Printf("Message is of type: %s\n", m.MessageType)
	fmt.Printf("Send message to Route: %+v \n", m.Route)
	table := router.GetRouteTable()
	_, exists := table[m.Route]
	if !exists {
		return fmt.Errorf("ERROR: A message for route: %s does not exist, either specify an existing route or create one using `AssertQueue`", m.Route)
	}
	return nil

}
