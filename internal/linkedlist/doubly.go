package linkedlist

// DNode содержит две связи. Их цена — дополнительный указатель на каждый узел.
type DNode[T comparable] struct {
	Value      T
	Prev, Next *DNode[T]
	owner      *Doubly[T]
}

// Doubly тоже нельзя копировать после начала использования.
type Doubly[T comparable] struct {
	head, tail *DNode[T]
	size       int
}

func (l *Doubly[T]) Head() *DNode[T] { return l.head }
func (l *Doubly[T]) Tail() *DNode[T] { return l.tail }
func (l *Doubly[T]) Len() int        { return l.size }
func (l *Doubly[T]) Append(value T) *DNode[T] {
	n := &DNode[T]{Value: value, Prev: l.tail, owner: l}
	if l.tail == nil {
		l.head = n
	} else {
		l.tail.Next = n
	}
	l.tail = n
	l.size++
	return n
}
func (l *Doubly[T]) Prepend(value T) *DNode[T] {
	n := &DNode[T]{Value: value, Next: l.head, owner: l}
	if l.head == nil {
		l.tail = n
	} else {
		l.head.Prev = n
	}
	l.head = n
	l.size++
	return n
}
func (l *Doubly[T]) Find(value T) *DNode[T] {
	for n := l.head; n != nil; n = n.Next {
		if n.Value == value {
			return n
		}
	}
	return nil
}

// Remove меняет обе соседние связи. Благодаря Prev не ищем предшественника.
// O(1) с живым handle; поиск handle по значению всё ещё O(n).
func (l *Doubly[T]) Remove(n *DNode[T]) bool {
	if n == nil || n.owner != l {
		return false
	}
	if n.Prev == nil {
		l.head = n.Next
	} else {
		n.Prev.Next = n.Next
	}
	if n.Next == nil {
		l.tail = n.Prev
	} else {
		n.Next.Prev = n.Prev
	}
	var zero T
	n.Value = zero
	n.Prev = nil
	n.Next = nil
	n.owner = nil
	l.size--
	return true
}
func (l *Doubly[T]) Values() []T {
	out := make([]T, 0, l.size)
	for n := l.head; n != nil; n = n.Next {
		out = append(out, n.Value)
	}
	return out
}
