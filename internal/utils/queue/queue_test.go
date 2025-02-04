package queue_test

import (
	"fmt"
	"message-broker/internal/utils/queue"
	"testing"
)

func TestQueue(t *testing.T) {
	var q queue.Queue
	q.Enqueue([]byte("2"))
	q.Enqueue([]byte("3"))
	q.Dequeue()
	fmt.Println(q)
}
