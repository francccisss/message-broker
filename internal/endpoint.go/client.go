package client

import (
	"fmt"
	protocol "message-broker/internal"
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
 Create an entry in a hashmap for a new Message queue
*/

func (ep Endpoint) HandleQueueAssert(m protocol.Queue) {
	fmt.Printf("Message is of type: %s\n", m.Type)
	fmt.Printf("Creating/Asserting Queue with Route: %+v \n", m.QueueHeader.Name)
}

/*
Route the EPMessage to the appropriate channel described in the current
EPMessages' header data
Use the Route property of the EPMessage to locate the appropriate Route within
the Route Map
*/

func (ep Endpoint) HandleEPMessage(m protocol.EPMessage) {
	fmt.Printf("Message is of type: %s\n", m.Type)
	fmt.Printf("Send message to Route: %+v \n", m.Route)
}
