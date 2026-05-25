package basic

const (
	// Because we are using a ring buffer, we want to control resizing to deal with possibly copying two segments without losing order
	initialDequeCapacity = 8
)

// Deque implements a double-ended queue using a ring buffer with the "next empty slot" pattern:
// - head points to the first element (front)
// - tail points to the next empty slot (one past the back)
// - Empty when: head == tail
// - Full when: (tail + 1) % capacity == head
// - We sacrifice one slot to distinguish empty from full
type Deque[T any] struct {
	data     []T
	capacity int
	head     int // points to the FRONT (first element)
	tail     int // points to the next EMPTY slot after the BACK
}

type DequeOption[T any] func(*Deque[T])

func NewDeque[T any](opts ...DequeOption[T]) *Deque[T] {
	d := &Deque[T]{
		data:     make([]T, initialDequeCapacity),
		capacity: initialDequeCapacity,
		head:     0,
		tail:     0, // head == tail means empty
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

// Len returns the number of elements in the deque
func (d *Deque[T]) Len() int {
	// Contiguous case: tail is ahead of head, so length is tail - head
	if d.tail >= d.head {
		return d.tail - d.head
	}

	// Wrapped case: elements from head to end, plus elements from start to tail
	return d.capacity - d.head + d.tail
}

// nextIndex returns the next index in the ring buffer
func (d *Deque[T]) nextIndex(i int) int {
	// We use a conditional instead of modulo for better performance, as the branch predictor will handle the wrap-around case efficiently
	i++
	if i >= d.capacity {
		return 0
	}
	return i
}

// prevIndex returns the previous index in the ring buffer.
func (d *Deque[T]) prevIndex(i int) int {
	// Similar to nextIndex, we use a conditional to avoid modulo for better performance
	i--
	if i < 0 {
		return d.capacity - 1
	}
	return i
}

// isFull returns true if the deque is at capacity
func (d *Deque[T]) isFull() bool {
	return d.nextIndex(d.tail) == d.head
}

// isEmpty returns true if the deque has no elements
func (d *Deque[T]) isEmpty() bool {
	return d.head == d.tail
}

// resize doubles the capacity and reorders elements
func (d *Deque[T]) resize() {
	newCapacity := d.capacity * 2
	newData := make([]T, newCapacity)

	// Copy elements from head to tail in order
	size := d.Len()
	if d.tail > d.head {
		// No wrap: elements are contiguous from head to tail
		copy(newData, d.data[d.head:d.tail])
	} else if size > 0 {
		// Wrapped: elements from head to end, then from start to tail
		n := copy(newData, d.data[d.head:])
		copy(newData[n:], d.data[:d.tail])
	}

	d.data = newData
	d.capacity = newCapacity
	d.head = 0
	d.tail = size
}

// Front returns the first element
func (d *Deque[T]) Front() (T, bool) {
	if d.isEmpty() {
		var zero T
		return zero, false
	}

	return d.data[d.head], true
}

// Back returns the last element (one before tail)
func (d *Deque[T]) Back() (T, bool) {
	if d.isEmpty() {
		var zero T
		return zero, false
	}

	return d.data[d.prevIndex(d.tail)], true
}

// PushFront adds an element to the front (moves head backward, writes at new head)
func (d *Deque[T]) PushFront(value T) {
	if d.isFull() {
		d.resize()
	}

	d.head = d.prevIndex(d.head)
	d.data[d.head] = value
}

// PushBack adds an element to the back (writes at tail, moves tail forward)
func (d *Deque[T]) PushBack(value T) {
	if d.isFull() {
		d.resize()
	}

	d.data[d.tail] = value
	d.tail = d.nextIndex(d.tail)
}

// PopFront removes and returns the first element (reads head, moves head forward)
func (d *Deque[T]) PopFront() (T, bool) {
	if d.isEmpty() {
		var zero T
		return zero, false
	}

	result := d.data[d.head]
	d.head = d.nextIndex(d.head)

	return result, true
}

// PopBack removes and returns the last element (moves tail backward, reads new tail)
func (d *Deque[T]) PopBack() (T, bool) {
	if d.isEmpty() {
		var zero T
		return zero, false
	}

	d.tail = d.prevIndex(d.tail)
	result := d.data[d.tail]

	return result, true
}
