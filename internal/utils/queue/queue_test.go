package queue_test

import (
	"fmt"
	"message-broker/internal/utils/queue"
	"testing"
)

func TestQueue(t *testing.T) {
	q := queue.Queue{}
	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(4)
	q.Enqueue(10)
	fmt.Println(q.GetItems())
	q.Dequeue()
	fmt.Println(q.GetItems())
}
