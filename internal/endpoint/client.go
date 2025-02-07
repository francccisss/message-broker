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

	"github.com/google/uuid"
)

/*
Endpoint is an abstraction of a connected application using an endpoint API
on the client side for connecting to the server

The Interface is used to handle different message types which is defined in the
message's Header `Type` property
*/
type Endpoint struct {
	Mux                   *sync.Mutex
	Conn                  net.Conn
	IncomingConsumerQueue queue.Queue
}

type EPHandler interface {
	handleQueueAssert(msgType.Queue)
	handleEPMessage(msgType.EPMessage)
	handleConsumers(msgType.Consumer)
	sendMessageToRoute()
}

func (ep Endpoint) MessageHandler(msgBuf bytes.Buffer) {
	// utils.MessageParser(msgBuf.Bytes())
	endpointMsg, err := utils.MessageParser(msgBuf.Bytes())
	// ERROR Handling
	if err != nil {
		fmt.Printf("ERROR: Unable to parse message")
		fmt.Println(err.Error())
		return
	}

	// // type assertion switch statement for different processing
	switch msg := endpointMsg.(type) {

	case msgType.Queue:
		ep.Mux.Lock()
		ep.handleQueueAssert(msg)
		ep.Mux.Unlock()
	case msgType.Consumer:
		ep.Mux.Lock()
		ep.handleConsumers(msg)
		ep.Mux.Unlock()
	case msgType.EPMessage:
		ep.Mux.Lock()
		// Signals ep message handler to store new message once
		// a queue has been created
		ep.handleEPMessage(msg)
		ep.Mux.Unlock()
	default:
		fmt.Println("ERROR: Unidentified type")
		// ep.Conn.Write([]byte("ERROR: Unidentified type: Types should consist of EPMessage | Queue "))
	}
	return

}

// TODO Fix Consumer message is received before a message queue could be created
// need to make sure that a message queue should first be created before assigning
// a consumer to a message queue
// or instead of registering a consumer within the server using Consume(), we could just include
// the register with queue assertion since queue asssertion already needs to be
// defined with a route along with the the StreamID from a channel, on the client side when calling consume
// we could register that separately and the mudem would still work
// Sol#2 Create a queue of consumer message and EPMessage's

func (ep *Endpoint) handleConsumers(msg msgType.Consumer) {
	fmt.Println("NOTIF: Consumer Message received")
	// messageQueueTable := mq.GetMessageQueueTable()
	// msq, exists := messageQueueTable[msg.Route]
	// if !exists {
	// 	fmt.Printf("ERROR: Message queue does not exist with specified route: %s\n", msg.Route)
	// 	return
	// }
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
func (ep Endpoint) handleQueueAssert(msg msgType.Queue) {
	fmt.Printf("NOTIF: Creating Queue \"%s\" \n", msg.Name)
	newConnectionID := uuid.NewString()
	table := mq.GetMessageQueueTable()
	msq, exists := table[msg.Name]

	if exists {
		// Just register connection to route
		fmt.Printf("NOTIF: Message queue exists: %s\n", msg.Name)
		fmt.Println("NOTIF: Registering client connection to route...")
		msq.Connections[newConnectionID] = &mq.ConsumerConnection{
			Conn:         ep.Conn,
			ConnectionID: newConnectionID,
		}
		msq.ConnectionIDs = append(msq.ConnectionIDs, newConnectionID)
		fmt.Printf("NOTIF: Connection successfully registerd to route \"%s\"\n", msg.Name)
		msq.Log()
		msq.Notif <- struct{}{}
		return
	}
	fmt.Printf("NOTIF: Creating a new Message Queue: %s\n", msg.Name)
	var m sync.Mutex
	newMq :=
		&mq.MessageQueue{
			Type:          msg.Type,
			Name:          msg.Name,
			Durable:       msg.Durable,
			Connections:   map[string]*mq.ConsumerConnection{},
			ConnectionIDs: []string{},
			Notif:         make(chan struct{}),
			Queue:         queue.Queue{},
			M:             &m,
		}
	table[msg.Name] = newMq
	fmt.Printf("NOTIF: New Message queue created: %s\n", msg.Name)
	go newMq.ListenMessages()
	// insert new connection to the route
	// Might optimize this by storing client connectios separately
	// and create an array of routes a connection is listening to
	// and when a message is to be sent to clients then we could
	// just look at each connections []route field and check
	// if the connection matches the route the message is to be sent to
	// then send it to that route

	fmt.Println("NOTIF: Registering client connection to route...")
	newMq.Connections[newConnectionID] = &mq.ConsumerConnection{
		Conn:         ep.Conn,
		ConnectionID: newConnectionID,
	}
	newMq.ConnectionIDs = append(newMq.ConnectionIDs, newConnectionID)
	fmt.Printf("NOTIF: Connection successfully registerd to route \"%s\"\n", msg.Name)
	newMq.Log()
	newMq.Notif <- struct{}{}
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

	msq.M.Lock()
	m, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("ERROR: Unable to marshal client message for delivery")
	}
	appendedMsg, err := utils.AppendPrefixLength(m)
	if err != nil {
		return err
	}
	msq.Queue.Enqueue(appendedMsg)
	msq.M.Unlock()
	fmt.Println("NOTIF: New message in queue")
	msq.Log()
	msq.Notif <- struct{}{}
	return nil
}
