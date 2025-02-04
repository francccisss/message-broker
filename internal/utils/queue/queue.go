package queue

type Queue [][]byte

// Take in any value to be pushed into the enqueued
func (q *Queue) Enqueue(i []byte) {
	*q = append(*q, i)
}

func (q *Queue) GetItems() Queue {
	return *q
}

func (q *Queue) Dequeue() []byte {
	if len(*q) == 0 {
		return []byte{}
	}

	tmp := *q
	front := tmp[0]
	*q = tmp[1:]
	return front
}
