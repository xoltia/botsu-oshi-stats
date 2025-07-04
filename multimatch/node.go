package multimatch

import (
	"slices"
)

type node struct {
	next     [256]*node
	output   []int
	failLink *node
}

func (n *node) appendUniqueOutput(o int) {
	if !slices.Contains(n.output, o) {
		n.output = append(n.output, o)
	}
}

type nodeQueue struct {
	buffer []*node
	count  int
	head   int
	tail   int
}

func (q *nodeQueue) get(i int) *node {
	if i >= q.count {
		panic("index out of bounds")
	}
	return q.buffer[(q.head+i)%len(q.buffer)]
}

func (q *nodeQueue) grow() {
	newCap := max(32, len(q.buffer)*3/2)
	newBuf := make([]*node, newCap)
	for i := range q.count {
		newBuf[i] = q.get(i)
	}
	q.buffer = newBuf
	q.tail = q.count
	q.head = 0
}

func (q *nodeQueue) remainingCap() int {
	return len(q.buffer) - q.count
}

func (q *nodeQueue) push(n *node) {
	if q.remainingCap() == 0 {
		q.grow()
	}
	q.buffer[q.tail] = n
	q.tail = (q.tail + 1) % len(q.buffer)
	q.count++
}

func (q *nodeQueue) empty() bool {
	return q.count == 0
}

func (q *nodeQueue) pop() *node {
	n := q.get(0)
	q.head = (q.head + 1) % len(q.buffer)
	q.count--
	return n
}

func (q *nodeQueue) clear() {
	for i := range q.buffer {
		q.buffer[i] = nil
	}
	q.count = 0
}
