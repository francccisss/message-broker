package client

import (
	"fmt"
	"log"
	protocol "message-broker/internal"
	router "message-broker/internal/router"
	"message-broker/internal/utils/queue"
	"net"
	"sync"
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
	HandleQueueAssert(protocol.Queue)
	HandleEPMessage(protocol.EPMessage)
	SendMessageToRoute()
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
	fmt.Printf("Creating/Asserting Queue with Route: %+v \n", q.Name)
	fmt.Printf("Message Queue is of type: %s\n", q.Type)
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
}

/*
Handling Endpoint Messages
  - Route the EPMessage to the appropriate channel described in the current
    EPMessages' header data
  - Use the Route property of the EPMessage to locate the appropriate Route
    within the Route Map
*/
func (ep Endpoint) HandleEPMessage(m protocol.EPMessage) error {
	fmt.Printf("Send message to Route: %+v \n", m.Route)
	fmt.Printf("Message type is of type: %s\n", m.MessageType)
	table := router.GetRouteTable()
	route, exists := table[m.Route]
	if !exists {
		return fmt.Errorf("ERROR: A message for route: %s does not exist, either specify an existing route or create one using `AssertQueue`", m.Route)
	}

	// Queue up messages
	route.MessageQueue.Enqueue(m.Body)
	return nil
}

// This is a go routine that will that should take in
// Only send a message if there is a consumer, and if there is a message in the message queue
// when new message is created place inside the messagequeue,
func (ep Endpoint) SendMessageToRoute(route router.Route, m *sync.Mutex) error {

	m.Lock()
	defer m.Unlock()
	if len(route.Connections) < 0 {
		for i, connection := range route.Connections {
			_ = *connection
			// TESTING
			// Sending messages concurrently to different connections
			go func(index int, messages queue.Queue) {
				for range len(messages.GetItems()) {
					log.Println(route.MessageQueue.Dequeue().([]byte))
				}
			}(i, route.MessageQueue)
			// Dequeuing Byte array
			// _, err := c.Write(route.MessageQueue.Dequeue().([]byte))
			// if err != nil {
			// 	log.Println("ERROR: Unable to write to consumer")
			// 	return err
			// }

		}
	}
	return nil
}
