package router

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

type Route struct {
	Type string
	Name string
	// TODO Need to handle disconnected clients
	// - Messages can only be pushed if there are connections
	// - Messages enqueued and sent will not be availble afterwards
	Connections  []net.Conn
	Durable      bool
	m            *sync.Mutex
	MessageQueue chan []byte
}

var table = map[string]*Route{}

type P2P interface{}
type PubSub interface{}

func GetRouteTable() map[string]*Route {
	return table
}

// This is a go routine that will that should take in
// Only send a message if there is a consumer, and if there is a message in the message queue
// when new message is created place inside the messagequeue,
func (r Route) ListenMessages() {
	log.Printf("Route stats: \nRoute: %s, \nConnections: %d, \nPending Messages in Queue: %d", r.Name, len(r.Connections), len(r.MessageQueue))
	for _, c := range r.Connections {
		// only read from channel buffer if there are connections in the route
		message := <-r.MessageQueue
		go func() {
			r.m.Lock()
			defer r.m.Unlock()
			_, err := c.Write(message)
			if err != nil {
				log.Println("ERROR: Unable to write to consumer")
				return
			}
			log.Printf("Message sent for route %s: %s", r.Name, string(message))
		}()
	}

}
