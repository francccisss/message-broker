# message-broker

My own implementation of a messaging system and message queue protocol

A message broker or a messaging system is an application integration pattern for inter-process communication for different services,
these services or application endpoints' communication is facilitaed by using an application layer protocol.

> This is not to be used in anywhere aside from just testing and playing around with it, this is just my way of implementing of how I understand
> how a message broker works and how I could make my own message queue protocol.

### How it works

This implementation uses a message queue system, the same way `amqp` works but a simplified version, the way it works is users are able to connect to the message broker
using network socket via tcp, the `Connect(address string) Connection` which returns a `Connection` interface which is used to create channels and streams,
`Connect(address string) Connection` also spawns a Message demultiplexer, which is used to demultiplex different messages to different channels by reading the message's
`StreamID`.

You can find the client side here:
API: https://github.com/francccisss/message-broker-endpoint-api

#### Mudem (Multiplexer/Demultiplexer)

The `Mudem` is implemented on the application endpoints responsible for reading incoming tcp streams from the server, assembles it, parses it and then reads the message's `StreamID` that points to a specific stream that holds,
a pointer to a channel's channel buffer, calling `Consume()` reads the available message in the channel buffer of the channel.

#### Endpoints

Endpoints also known as clients can use `Connect(address string)` to connect to message broker, `Connection.CreateChannel()` creates a new stream in the stream pool and adds the new channel,
`Channel.AssertQueue()` for checking if a queue already exists and adds one if it doesnt, `Channel.DeliverMessage()`, and `Channel.Consume()`. These are methods that abstracts away the communication
between one or more endpoints.

#### Message Queue

When endpoints call `Channel.AssertQueue`, a message of type `Queue` (I know doesnt seem inuitive might change it), the message broker reads the message's `MessageType` and calls a handler for the specific message type, checks the `Name` field of the Queue to be asserted, if a message queue already exists it registers the connection in the `map[string]*mq.ConsumerConnection`, a `ConsumerConnection` represents a Connection that is registered to a message queue, but if a Queue does not exist a new message queue will be created, adding the new registered connection and then spawns a message listener that will listen for incoming messages that are dispatched from endpoints via `Channel.DeliverMessage()`, which will then be stored in a Queue data structure within the message queue, each message will be dispatched for every clients that are registered in the queue.
