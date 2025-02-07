package mq

import (
	"encoding/json"
	"errors"
	"fmt"
	"message-broker/internal/types"
	"message-broker/internal/utils"
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
	Connections     map[string]*ConsumerConnection
	ConnectionIDs   []string
	DeadConnections map[string]string
	Durable         bool
	Queue           queue.Queue
	Notif           chan struct{}
	M               *sync.Mutex
}
type ConsumerConnection struct {
	Conn         net.Conn
	ConnectionID string
	StreamID     string
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
	var wg sync.WaitGroup

	for range mq.Notif {
		// Handling disconnected clients

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
			message := mq.Queue.Dequeue()

			// This is the only way for now to remove dead connections
			for _, connID := range mq.ConnectionIDs {
				wg.Add(1)
				conn, exists := mq.Connections[connID]
				fmt.Printf("TEST_NOTIF: Sending Message #%d\n", i)
				// if for some reason a dead connection exists when it shouldn't
				// add it to disconnected clients for removal after sending messages
				if !exists {
					fmt.Println("ERROR: Connection does not exist, please remove it")
					fmt.Println("TEST_ERROR: Appending connectionID to disconnected clients array")
					mq.removeDeadConnection(connID)
					continue
				}
				go mq.sendMessage(&wg, conn, message.(types.EPMessage)) // might cause race condition
			}
		}
		fmt.Println("TEST_NOTIF: Barrier Waiting for all messages to be sent by go routines...")
		wg.Wait()
		fmt.Println("TEST_NOTIF: All messages has been sent")
		mq.Log()
	}
}

// TODO Need to be able to remove consumers from message queue if they are disconnected
func (mq *MessageQueue) sendMessage(wg *sync.WaitGroup, c *ConsumerConnection, message types.EPMessage) {
	defer wg.Done()

	message.StreamID = c.StreamID
	fmt.Println(c.StreamID)
	fmt.Println(message.StreamID)
	m, err := json.Marshal(message)
	if err != nil {
		fmt.Println("ERROR: Unable to marshal client message for delivery")
		return
	}

	appendedMsg, err := utils.AppendPrefixLength(m)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	_, err = c.Conn.Write(appendedMsg)
	if err != nil {
		fmt.Println("ERROR: Unable to write to consumer")
		if errors.Is(err, net.ErrClosed) {

			// Error returns "use of closed network connection"
			// if connection was closed but wanting to write to it
			fmt.Println(err.Error())
			// Add disconnected clients into the map for clean up
			fmt.Printf("TEST_NOTIF: Connection dead: %s\n", c.ConnectionID)
			mq.removeDeadConnection(c.ConnectionID)
		}
		return
	}
	fmt.Printf("TEST_NOTIF: Message sent for route %s:\n|-StreamID %s\n|-MsgLen %d\n", mq.Name, message.StreamID, len(message.Body))
}

// for each connectionID in connectionIDs
// if connectionID is a dead connection
// move every element in connectionIDs from dead connection's index position + 1 to dead connection's index position
// [1,2,3, dead, <- 5, 6, 7, dead, <- 9]
func (mq *MessageQueue) removeDeadConnection(deadConnID string) {
	fmt.Println("TEST_NOTIF: Cleaning up dead connection...")
	mq.M.Lock()
	defer mq.M.Unlock()
	for i, connectionID := range mq.ConnectionIDs {
		if connectionID == deadConnID {
			fmt.Printf("TEST_NOTIF: Removing dead connection id: %s\n", connectionID)
			before := mq.ConnectionIDs[:0]
			if i > 0 {
				before = mq.ConnectionIDs[:i-1]
			}
			mq.ConnectionIDs = append(before, mq.ConnectionIDs[i+1:]...)
			break
		}
	}
	fmt.Println("TEST_NOTIF: Removing dead connection net.Conn ^^^")
	delete(mq.Connections, deadConnID)
	fmt.Printf("TEST_NOTIF: Connection IDs remaining: %d\n", len(mq.ConnectionIDs))
	fmt.Printf("TEST_NOTIF: Connections remaining: %d\n", len(mq.Connections))
}

func (mq *MessageQueue) Log() {
	fmt.Printf("Messsage Queue Stats: \n |-Route: %s, \n |-Connections: %d, \n |-Pending Messages in Queue: %d\n", mq.Name, len(mq.Connections), len(mq.Queue))
}
