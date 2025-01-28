package queue

import "fmt"

type Queue struct {
	array [][]byte
}

// Take in any value to be pushed into the enqueued
func (q *Queue) Enqueue(i []byte) {
	q.array = append(q.array, i)
}

func (q *Queue) GetItems() [][]byte {
	return q.array
}

func (q *Queue) Dequeue() ([]byte, error) {
	if len(q.array) == 0 {
		return nil, fmt.Errorf("ERROR: Queue is empty")
	}
	front := q.array[0]
	q.array = q.array[1:]
	return front, nil
}
