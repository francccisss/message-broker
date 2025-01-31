package mq

import (
	"log"
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

// This is a go routine that will that should take in
// Only send a message if there is a consumer, and if there is a message in the message queue
// when new message is created place inside the messagequeue,
//
// When a new route message queue is created, the message queue listens to new messages
// within its own route, the MessageQueue is where data is queued and pushed
// to connected clients for consumption

// When there are client > 0 connected, message is sent
// When there are no connections message stays in the queue
func (mq MessageQueue) ListenMessages() {
	log.Println("NOTIF: New Listener spawned")
	mq.Log()
	for {

		// only read from channel buffer if there are connections in the route
		// else requeue if empty
		// Using Notif to check if connecttion exists before sending out messages
		<-mq.Notif
		if len(mq.Connections) > 0 {
			log.Println("NOTIF: There are 0 connections")
			continue
		}
		log.Println("NOTIF: Sending message")
		message := <-mq.Queue
		for _, c := range mq.Connections {
			go func() {
				_, err := c.Write(message)
				if err != nil {
					log.Println("ERROR: Unable to write to consumer")
					return
				}
				log.Printf("NOTIF: Message sent for route %s: %s", mq.Name, string(message))
			}()
		}
	}

}

func (mq *MessageQueue) Log() {
	log.Printf("ROUTE_STATS: \nRoute: %s, \nConnections: %d, \nPending Messages in Queue: %d", mq.Name, len(mq.Connections), len(mq.Queue))

}
