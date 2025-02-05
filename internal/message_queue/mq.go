package mq

import (
	"errors"
	"fmt"
	"message-broker/internal/utils/queue"
	"net"
	"sync"
)

/*
Router is responsible for routing messages based on each messages'
`Route` property the Header to route the message to the appropriate
TCP Connection, the application endpoint will be the one responsible for
handling the multiplexing/demultiplexing using a `message dispatcher`
*/

type MessageQueue struct {
	Type string
	Name string
	// TODO Need to handle disconnected clients
	// - Messages can only be pushed if there are connections
	// - Messages enqueued and sent will not be availble afterwards
	Connections   map[string]*ConsumerConnection
	ConnectionIDs []string
	Durable       bool
	Queue         queue.Queue
	Notif         chan struct{}
	M             *sync.Mutex
}
type ConsumerConnection struct {
	Conn         net.Conn
	ConnectionID string
}

var table = map[string]*MessageQueue{}

type P2P interface{}
type PubSub interface{}

func GetMessageQueueTable() map[string]*MessageQueue {
	return table
}

/*
This is a go routine Message Listener that will ONLY send a message
if a consumer exists, and will only queue up messages until there is
> 0 consumers in the mq.Connections field

If a message queue already exists, connections will be stored to the existing
one, else it will create a new queue when AssertQueue is called by the client

When a new message queue is created via AssertQueue in the client side,
a new Listener will be spawned to listen to incoming messages, and pushed
to consumers that are defined in the mq.Connections, the Connections field
in a message queue stores an array of client connections, where messages
will be sent for consumption

When there are client > 0 connected, message is sent
When there are no connections message stays in the queue
*/
func (mq *MessageQueue) ListenMessages() {
	fmt.Println("NOTIF: New Listener spawned")
	mq.Log()

	for range mq.Notif {
		// Handling disconnected clients
		disconnectedClients := []string{}

		// Listener is notified when there are new consumers, and or there are new messages
		// when message queue was full wit messages, and connections are present
		// the first notificaion will send all of the messages, so every subsequent
		// notifications wont do anything

		fmt.Println("NOTIF: Checking connections before sending messages...")
		fmt.Printf("NOTIF: Current Connections %d\n", len(mq.Connections))
		if len(mq.Connections) < 1 {
			fmt.Println("NOTIF: There are 0 connections")
			continue
		}

		if len(mq.Queue) < 1 {
			fmt.Println("NOTIF: There are 0 messages in queue")
			continue
		}
		fmt.Printf("NOTIF: Total messages to be sent %d\n", len(mq.Queue))

		fmt.Printf("NOTIF: Sending all %d messages\n", len(mq.Queue))
		for i := range mq.Queue {
			// TODO ADD MUTEX LOCK IF RACE CONDITION OCCURS
			message := mq.Queue.Dequeue()

			// TODO This does not work btw
			for _, connID := range mq.ConnectionIDs {
				conn, exists := mq.Connections[connID]
				fmt.Printf("TEST_NOTIF: Sending Message #%d\n", i)
				go sendMessage(conn, message, disconnectedClients) // might cause race condition
				fmt.Printf("TEST_NOTIF: Message sent for route %s: %+v\n", mq.Name, message[:4])

				if !exists {
					fmt.Println("ERROR: Connection does not exist, please remove it")
					return
				}
			}
		}
		fmt.Println("TEST_NOTIF: All messages has been sent")
		mq.Log()

		// // TODO Update ConnectionIDs[] removing dead connections
		// for _, deadConnID := range disconnectedClients {
		// 	delete(mq.Connections, deadConnID)
		// }
	}
}

func sendMessage(c *ConsumerConnection, message []byte, disconnectedClients []string) {
	_, err := c.Conn.Write(message)
	if err != nil {
		fmt.Println("ERROR: Unable to write to consumer")
		if errors.Is(err, net.ErrClosed) {
			fmt.Println(err.Error())
			disconnectedClients = append(disconnectedClients, c.ConnectionID)
		}
		return
	}
	fmt.Println("TEST_NOTIF: Message sent")
}

func (mq *MessageQueue) Log() {
	fmt.Printf("Messsage Queue Stats: \n |-Route: %s, \n |-Connections: %d, \n |-Pending Messages in Queue: %d\n", mq.Name, len(mq.Connections), len(mq.Queue))
}
