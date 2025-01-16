package protocol

/*

A Protocol defines the set of rules for two endpoints to communicate, client to server,
server to client, or client to client through C -> S -> C.

Max Header size 1kb
Body undecided size?

Mesaging system doesnt need to know that Im creating a new "Channel"
There are 4 types of messages for now:
  - Connect
  - Create/Assert
  - Send
  - Receive

Header for Sending
  - Can have default exchange (no routing) bound to everything
  - Route string
    ?- Specify different model
*/

// To describe the queue to be created
type QueueHeader struct {
	Name    string
	Type    string
	Durable bool
}

type Queue struct {
	Type string
	QueueHeader
	HeaderSize int
	Body       string
}

type EPMessage struct {
	Type       string
	Route      string
	HeaderSize int
	Body       string
}
