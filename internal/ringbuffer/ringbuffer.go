// Package ringbuffer хранит ограниченную FIFO-очередь в одном массиве.
package ringbuffer

import "errors"

var ErrCapacity = errors.New("capacity must be positive")

// Ring не увеличивает ёмкость. Политика full: отклонить новый элемент (false).
// head — следующее чтение, tail — следующая запись. size различает full/empty,
// когда head==tail. Структуру с данными не копируем: backing array был бы общим.
type Ring[T any] struct {
	slots            []T
	head, tail, size int
}

func New[T any](capacity int) (*Ring[T], error) {
	if capacity <= 0 {
		return nil, ErrCapacity
	}
	return &Ring[T]{slots: make([]T, capacity)}, nil
}
func (r *Ring[T]) Len() int      { return r.size }
func (r *Ring[T]) Cap() int      { return len(r.slots) }
func (r *Ring[T]) IsEmpty() bool { return r.size == 0 }
func (r *Ring[T]) IsFull() bool  { return r.Cap() > 0 && r.size == r.Cap() }

// Enqueue/Dequeue/Front выполняют O(1), после New не выделяют память.
// Zero value — отключённый буфер; Enqueue возвращает false, деления на 0 нет.
func (r *Ring[T]) Enqueue(value T) bool {
	if r.Cap() == 0 || r.IsFull() {
		return false
	}
	r.slots[r.tail] = value
	r.tail = (r.tail + 1) % r.Cap()
	r.size++
	return true
}
func (r *Ring[T]) Dequeue() (T, bool) {
	if r.IsEmpty() {
		var zero T
		return zero, false
	}
	value := r.slots[r.head]
	var zero T
	r.slots[r.head] = zero
	r.head = (r.head + 1) % r.Cap()
	r.size--
	return value, true
}
func (r *Ring[T]) Front() (T, bool) {
	if r.IsEmpty() {
		var zero T
		return zero, false
	}
	return r.slots[r.head], true
}

// State — независимый снимок для схемы и тестов, O(cap) времени/памяти.
// Физический порядок Slots отличается от логического FIFO-порядка Values.
type State[T any] struct {
	Head, Tail, Len, Cap int
	Slots, Values        []T
}

func (r *Ring[T]) Snapshot() State[T] {
	slots := append([]T(nil), r.slots...)
	values := make([]T, 0, r.size)
	for i := 0; i < r.size; i++ {
		values = append(values, r.slots[(r.head+i)%r.Cap()])
	}
	return State[T]{r.head, r.tail, r.size, r.Cap(), slots, values}
}
