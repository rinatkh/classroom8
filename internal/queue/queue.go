// Package queue сравнивает FIFO с одинаковым поведением и разной ценой dequeue.
package queue

import "github.com/rinatkh/classroom8/internal/linkedlist"

// Slice намеренно сдвигает хвост: Dequeue O(n), Enqueue амортизированно O(1).
// Это простой baseline. Очередь с offset и compaction — другой возможный вариант.
type Slice[T any] struct{ values []T }

func (q *Slice[T]) Enqueue(value T) { q.values = append(q.values, value) }
func (q *Slice[T]) Dequeue() (T, bool) {
	if len(q.values) == 0 {
		var zero T
		return zero, false
	}
	value := q.values[0]
	copy(q.values, q.values[1:])
	last := len(q.values) - 1
	var zero T
	q.values[last] = zero
	q.values = q.values[:last]
	return value, true
}
func (q *Slice[T]) Front() (T, bool) {
	if len(q.values) == 0 {
		var zero T
		return zero, false
	}
	return q.values[0], true
}
func (q *Slice[T]) Len() int { return len(q.values) }

// Linked использует tail для Enqueue и head для Dequeue, оба O(1).
// Узлы после удаления доступны для GC, если на них нет внешних ссылок.
type Linked[T comparable] struct{ list linkedlist.List[T] }

func (q *Linked[T]) Enqueue(value T)    { q.list.Append(value) }
func (q *Linked[T]) Dequeue() (T, bool) { return q.list.PopFront() }
func (q *Linked[T]) Front() (T, bool) {
	if q.list.Head() == nil {
		var zero T
		return zero, false
	}
	return q.list.Head().Value, true
}
func (q *Linked[T]) Len() int { return q.list.Len() }
