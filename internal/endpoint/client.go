package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"message-broker/internal/message_queue"
	msgType "message-broker/internal/types"
	"message-broker/internal/utils"
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
	endpointMsg, err := utils.MessageParser(msgBuf.Bytes())
	// ERROR Handling
	if err != nil {
		fmt.Printf("ERROR: Unable to parse message")
		fmt.Println(err.Error())
		return
	}

	// type assertion switch statement for different processing

	switch msg := endpointMsg.(type) {
	case msgType.Consumer:
		ep.Mux.Lock()
		ep.handleConsumers(msg)
		ep.Mux.Unlock()
	case msgType.EPMessage:
		ep.Mux.Lock()
		ep.handleEPMessage(msg)
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
	fmt.Println("NOTIF: Consumer Message received")
	messageQueueTable := mq.GetMessageQueueTable()
	msq, exists := messageQueueTable[msg.Route]
	if !exists {
		fmt.Printf("ERROR: Message queue does not exist with specified route: %s\n", msg.Route)
		return
	}
	msq.Connections = append(msq.Connections, ep.Conn)
	fmt.Printf("NOTIF: Register consumer in route: %s\n", msq.Name)
	msq.Log()
	msq.Notif <- struct{}{}
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
	fmt.Println("NOTIF: Queue Message received")
	table := mq.GetMessageQueueTable()
	_, exists := table[q.Name]
	if !exists {
		newMq :=
			&mq.MessageQueue{
				Type:        q.Type,
				Name:        q.Name,
				Durable:     q.Durable,
				Connections: []net.Conn{},
				Notif:       make(chan struct{}),
				Queue:       queue.Queue{},
			}
		table[q.Name] = newMq
		fmt.Printf("NOTIF: MESSAGE QUEUE CREATED: %s\n", q.Name)
		go newMq.ListenMessages()
		return
	}
	fmt.Printf("NOTIF: MESSAGE QUEUE ALREADY EXISTS: %s\n", q.Name)
}

/*
Handling Endpoint Messages
  - Route the EPMessage to the appropriate channel described in the current
    EPMessages' header data
  - Use the Route property of the EPMessage to locate the appropriate Route
    within the Route Map
*/
func (ep Endpoint) handleEPMessage(msg msgType.EPMessage) error {
	fmt.Println("NOTIF: EP Message received")
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
	if err != nil {
		return err
	}
	msq.Queue.Enqueue(appendedMsg)
	fmt.Println("NOTIF: New message in queue")
	msq.Log()
	msq.Notif <- struct{}{}
	return nil
}
