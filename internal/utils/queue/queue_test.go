package queue_test

import (
	"fmt"
	"message-broker/internal/utils/queue"
	"testing"
)

func TestQueue(t *testing.T) {
	q := queue.Queue{}
	value, err := q.Dequeue()
	fmt.Println(value)
	if value == nil {
		t.Fatal(err.Error())
	}
}
