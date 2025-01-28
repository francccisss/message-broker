package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"message-broker/internal/message_queue"
	msgType "message-broker/internal/types"
	"message-broker/internal/utils"
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
	endpointMsg, err := utils.MessageParser(msgBuf.Bytes())
	// ERROR Handling
	if err != nil {
		log.Printf("ERROR: Unable to parse message")
		log.Println(err.Error())
		return
	}

	// type assertion switch statement for different processing

	defer ep.Mux.Unlock()
	switch msg := endpointMsg.(type) {
	case msgType.EPMessage:
		ep.Mux.Lock()
		ep.handleEPMessage(msg)
	case msgType.Consumer:
		ep.Mux.Lock()
		ep.handleConsumers(msg)
	case msgType.Queue:
		ep.Mux.Lock()
		ep.handleQueueAssert(msg)
	default:
		fmt.Println("ERROR: Unidentified type")
		// ep.Conn.Write([]byte("ERROR: Unidentified type: Types should consist of EPMessage | Queue "))
	}
	return

}

func (ep Endpoint) handleConsumers(msg msgType.Consumer) {
	log.Println("NOTIF: Consumer Message received")
	table := mq.GetMessageQueueTable()
	r, exists := table[msg.Route]
	if !exists {
		log.Printf("ERROR: Message queue does not exist with specified route: %s\n", msg.Route)
		return
	}
	r.Connections = append(r.Connections, ep.Conn)

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
	table := mq.GetMessageQueueTable()
	route, exists := table[q.Name]
	if !exists {
		table[q.Name] = &mq.MessageQueue{
			Type:        q.Type,
			Name:        q.Name,
			Durable:     q.Durable,
			Connections: []net.Conn{},
			// can change default size of message queue buffer size
			Queue: make(chan []byte, 50),
		}
		fmt.Printf("NOTIF: MESSAGE QUEUE CREATED: %s\n", q.Name)
		go route.ListenMessages()
		return
	}
	go route.ListenMessages()
}

/*
Handling Endpoint Messages
  - Route the EPMessage to the appropriate channel described in the current
    EPMessages' header data
  - Use the Route property of the EPMessage to locate the appropriate Route
    within the Route Map
*/
func (ep Endpoint) handleEPMessage(msg msgType.EPMessage) error {
	log.Println("NOTIF: EP Message received")
	table := mq.GetMessageQueueTable()
	msq, exists := table[msg.Route]
	if !exists {
		return fmt.Errorf("ERROR: A message for route: %s does not exist, either specify an existing route or create one using `AssertQueue`", msg.Route)
	}

	m, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("ERROR: Unable to marshal client message for delivery")
	}
	appendedMsg, err := utils.AppendPrefixLength(m)
	msq.Queue <- appendedMsg
	return nil
}
