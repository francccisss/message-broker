package queue

type Queue []interface{}

// Take in any value to be pushed into the enqueued
func (q *Queue) Enqueue(i interface{}) {
	*q = append(*q, i)
}

func (q *Queue) Dequeue() interface{} {
	if len(*q) == 0 {
		return []byte{}
	}
	front := (*q)[0]
	*q = (*q)[1:]
	return front
}
