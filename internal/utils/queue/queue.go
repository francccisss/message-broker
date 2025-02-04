package queue

type Queue [][]byte

// Take in any value to be pushed into the enqueued
func (q *Queue) Enqueue(i []byte) {
	*q = append(*q, i)
}

func (q *Queue) Dequeue() []byte {
	if len(*q) == 0 {
		return []byte{}
	}
	front := (*q)[0]
	*q = (*q)[1:]
	return front
}
