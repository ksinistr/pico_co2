// Package fifo provides a fixed-size, no-allocation FIFO queue for int16 values.
package fifo

// FIFO16 implements a circular buffer queue of int16 with runtime capacity up to MaxCapacity.
type FIFO16 struct {
	buf      []int16
	capacity int // actual usable capacity (≤ MaxCapacity)
	head     int // index of the oldest element
	tail     int // index to write the next element
	count    int // number of elements stored
}

// NewFIFO16 creates a FIFO with the given capacity (1..MaxCapacity).
func NewFIFO16(capacity int) *FIFO16 {
	return &FIFO16{
		buf:      make([]int16, capacity),
		capacity: capacity,
	}
}

// Reset clears the queue back to empty state.
func (q *FIFO16) Reset() {
	q.head, q.tail, q.count = 0, 0, 0
}

// Len returns the number of elements in the queue.
func (q *FIFO16) Len() int {
	return q.count
}

// IsEmpty reports whether the queue has no elements.
func (q *FIFO16) IsEmpty() bool {
	return q.count == 0
}

// IsFull reports whether the queue has reached its capacity.
func (q *FIFO16) IsFull() bool {
	return q.count == q.capacity
}

// Enqueue adds v at the tail. If full, it drops the oldest to make space.
func (q *FIFO16) Enqueue(v int16) {
	if q.count == q.capacity {
		// drop oldest
		q.head = (q.head + 1) % q.capacity
		q.count--
	}

	q.buf[q.tail] = v
	q.tail = (q.tail + 1) % q.capacity
	q.count++
}

// Dequeue removes and returns the oldest element; ok=false if empty.
func (q *FIFO16) Dequeue() (int16, bool) {
	if q.count == 0 {
		return 0, false
	}

	v := q.buf[q.head]
	q.head = (q.head + 1) % q.capacity
	q.count--

	return v, true
}

// PeekAll calls fn(v) for each element from oldest to newest, without removing.
func (q *FIFO16) PeekAll(fn func(int16)) {
	idx := q.head
	for range q.count {
		fn(q.buf[idx])
		idx = (idx + 1) % q.capacity
	}
}

// Contiguous returns a single slice containing all queue elements in order.
// To achieve this, it may rearrange the internal buffer to make the elements
// contiguous. This is an in-place operation that avoids allocations.
// The returned slice is a view into the queue's internal buffer and should
// not be modified. The slice is valid until the next modification of the queue.
func (q *FIFO16) Contiguous() []int16 {
	if q.count == 0 {
		return nil
	}

	if q.head < q.tail {
		return q.buf[q.head:q.tail]
	}

	// Data is wrapped. Rearrange it to be contiguous using in-place rotation.
	// The queue layout is:
	// [ tail .... head-1 | head .... end ]
	// We want to rotate it to:
	// [ head .... end | tail .... head-1 ]
	// This is done by reversing three segments:
	// 1. Reverse [0..head-1]
	// 2. Reverse [head..end]
	// 3. Reverse [0..end]
	reverse(q.buf, 0, q.head-1)
	reverse(q.buf, q.head, q.capacity-1)
	reverse(q.buf, 0, q.capacity-1)

	q.head = 0
	q.tail = q.count % q.capacity

	return q.buf[0:q.count]
}

// reverse reverses elements of s in the range [from, to] in place.
func reverse(s []int16, from, to int) {
	if from >= to {
		return
	}

	for from < to {
		s[from], s[to] = s[to], s[from]
		from++
		to--
	}
}

/*
Usage example:

import (
    "fmt"
    "time"
    "your/module/path/fifo"
)

func main() {
    q := fifo.NewFIFO16(64)

    for {
        q.Enqueue(readSensor())
        time.Sleep(100 * time.Millisecond)

        // every minute at second 0
        if time.Now().Second() == 0 {
            fmt.Println("Measurements:")
            q.PeekAll(func(v int16) {
                fmt.Println(v)
            })
        }
    }
}
*/
