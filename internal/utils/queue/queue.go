package queue

type Queue struct {
	array []any
}

// Take in any value to be pushed into the enqueued
func (q *Queue) Enqueue(i any) {
	q.array = append(q.array, i)
}

func (q *Queue) GetItems() []any {
	return q.array
}

func (q *Queue) Dequeue() any {
	front := q.array[0]
	q.array = q.array[1:]
	return front
}
