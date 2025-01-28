package client

import (
	"bytes"
	"fmt"
	"log"
	router "message-broker/internal/router"
	msgType "message-broker/internal/types"
	parser "message-broker/internal/utils"
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
	Mux  *sync.Mutex
	Conn net.Conn
}

type EPHandler interface {
	handleQueueAssert(msgType.Queue)
	handleEPMessage(msgType.EPMessage)
	handleConsumers(msgType.Consumer)
	sendMessageToRoute()
}

func (ep Endpoint) MessageHandler(msgBuf bytes.Buffer) {
	endpointMsg, err := parser.MessageParser(msgBuf.Bytes())
	// ERROR Handling
	if err != nil {
		log.Printf("ERROR: Unable to parse message")
		log.Println(err.Error())
		return
	}

	// type assertion switch statement for different processing
	switch msg := endpointMsg.(type) {
	case msgType.EPMessage:
		ep.Mux.Lock()
		ep.handleEPMessage(msg)
		ep.Mux.Unlock()
	case msgType.Consumer:
		ep.Mux.Lock()
		ep.handleConsumers(msg)
		ep.Mux.Unlock()
	case msgType.Queue:
		ep.Mux.Lock()
		ep.handleQueueAssert(msg)
		ep.Mux.Unlock()
	default:
		fmt.Println("ERROR: Unidentified type")
		// ep.Conn.Write([]byte("ERROR: Unidentified type: Types should consist of EPMessage | Queue "))
	}
	return

}

func (ep Endpoint) handleConsumers(msg msgType.Consumer) {
	log.Println("NOTIF: Consumer Message received")
	table := router.GetRouteTable()
	r, exists := table[msg.Route]
	if !exists {
		log.Printf("ERROR: Message queue does not exist with specified route: %s\n", msg.Route)
		return
	}
	r.Connections = append(r.Connections, ep.Conn)

	ep.sendMessageToRoute(r)
	log.Printf("NOTIF: Register consumer in route: %s\n", msg.Route)
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
func (ep Endpoint) handleQueueAssert(q msgType.Queue) {
	log.Println("NOTIF: Queue Message received")
	table := router.GetRouteTable()
	_, exists := table[q.Name]
	if !exists {
		table[q.Name] = &router.Route{
			Type:        q.Type,
			Name:        q.Name,
			Durable:     q.Durable,
			Connections: []net.Conn{},
		}
		fmt.Printf("NOTIF: MESSAGE QUEUE CREATED: %s\n", q.Name)
	}
}

/*
Handling Endpoint Messages
  - Route the EPMessage to the appropriate channel described in the current
    EPMessages' header data
  - Use the Route property of the EPMessage to locate the appropriate Route
    within the Route Map
*/
func (ep Endpoint) handleEPMessage(m msgType.EPMessage) error {
	log.Println("NOTIF: EP Message received")
	table := router.GetRouteTable()
	route, exists := table[m.Route]
	if !exists {
		return fmt.Errorf("ERROR: A message for route: %s does not exist, either specify an existing route or create one using `AssertQueue`", m.Route)
	}

	// Queue up messages
	route.MessageQueue.Enqueue(m.Body)
	go ep.sendMessageToRoute(route)
	return nil
}

// This is a go routine that will that should take in
// Only send a message if there is a consumer, and if there is a message in the message queue
// when new message is created place inside the messagequeue,
func (ep Endpoint) sendMessageToRoute(route *router.Route) {
	log.Printf("Number of connections in the current route: \nRoute: %s, \nConnections: %d, \nPending Messages in Queue: %d", route.Name, len(route.Connections), len(route.MessageQueue.GetItems()))
	for i := range route.Connections {
		// TESTING
		// Sending messages concurrently to different connections
		go func(index int, messages queue.Queue) {
			for range len(messages.GetItems()) {
				value, err := route.MessageQueue.Dequeue()
				if err != nil {
					log.Println(err.Error())
					return
				}
				log.Printf("Message for route %s: %s", route.Name, string(value))
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
