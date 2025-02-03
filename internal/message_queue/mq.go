package mq

import (
	"fmt"
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
	Connections []net.Conn
	Durable     bool
	Queue       chan []byte
	Notif       chan uint32
	m           *sync.Mutex
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

	// only read from channel buffer if there are connections in the route
	// else do nothing if empty

	// Using Notif to check if connecttions exists before sending out messages

	for range <-mq.Notif {

		fmt.Println("NOTIF: New message received")
		if len(mq.Connections) > 0 {
			fmt.Println("NOTIF: There are 0 connections")
			continue
		}
		fmt.Println("NOTIF: Sending messages")
		message := <-mq.Queue
		fmt.Printf("NOTIF: Total messages to be sent%d", len(mq.Queue))
		for _, c := range mq.Connections {
			go func() {
				_, err := c.Write(message)
				if err != nil {
					fmt.Println("ERROR: Unable to write to consumer")
					return
				}
				fmt.Printf("NOTIF: Message sent for route %s: %s", mq.Name, string(message))
			}()
		}
	}

}

func (mq *MessageQueue) Log() {
	fmt.Printf("Messsage Queue Stats: \n |-Route: %s, \n |-Connections: %d, \n |-Pending Messages in Queue: %d\n", mq.Name, len(mq.Connections), len(mq.Queue))

}
