package utils

import (
	"encoding/json"
	"fmt"
	msgType "message-broker/internal/types"
)

// Given an empty interface where it can store any value of and be represented as any type,
// we need to assert that its of some known type by matching the "MessageType" of the incoming message,
// once the "MessageType" of the message is known, we can then Unmarashal the message into the specified
// known type that matched the "MessageType"
func MessageParser(b []byte) (interface{}, error) {
	var temp map[string]interface{}
	if err := json.Unmarshal(b, &temp); err != nil {
		return nil, err
	}

	switch temp["MessageType"] {
	case "EPMessage":
		var epMsg msgType.EPMessage
		err := json.Unmarshal(b, &epMsg)
		if err != nil {
			return nil, err
		}
		fmt.Printf("NOTIF: Parsing %s message\n", epMsg.MessageType)
		return epMsg, nil
	case "Consumer":
		var cons msgType.Consumer
		err := json.Unmarshal(b, &cons)
		if err != nil {
			return nil, err
		}
		fmt.Printf("NOTIF: Parsing %s message\n", cons.MessageType)
		return cons, nil
	case "Queue":
		var q msgType.Queue
		err := json.Unmarshal(b, &q)
		if err != nil {
			return nil, err
		}
		fmt.Printf("NOTIF: Parsing %s message\n", q.MessageType)
		return q, nil
	default:
		return temp, fmt.Errorf("ERROR: Not of any known message type")
	}
}
