// Package linkedlist показывает узлы, связи и цену поиска позиции.
// List после начала использования нельзя копировать: узлы хранят владельца.
// Структуры рассчитаны на последовательные операции, без внутреннего Mutex.
package linkedlist

// Node — отдельный объект в памяти. Next не является соседним индексом массива.
// Value и Next доступны для изучения. У узлов, принадлежащих List, вызывающий
// читает связи, но меняет их только методами List. Иначе инварианты разрушатся.
type Node[T comparable] struct {
	Value T
	Next  *Node[T]
	owner *List[T]
}

// List хранит head/tail и точную длину; zero value готов к работе.
// head==nil эквивалентно tail==nil и size==0. tail.Next всегда nil.
type List[T comparable] struct {
	head, tail *Node[T]
	size       int
}

func (l *List[T]) Head() *Node[T] { return l.head }
func (l *List[T]) Tail() *Node[T] { return l.tail }
func (l *List[T]) Len() int       { return l.size }

// Append выполняет O(1): tail избавляет от поиска конца. Новый узел выделяется.
func (l *List[T]) Append(value T) *Node[T] {
	node := &Node[T]{Value: value, owner: l}
	if l.tail == nil {
		l.head = node
	} else {
		l.tail.Next = node
	}
	l.tail = node
	l.size++
	return node
}

// Prepend обновляет head, а у пустого списка также tail. Время O(1).
func (l *List[T]) Prepend(value T) *Node[T] {
	node := &Node[T]{Value: value, Next: l.head, owner: l}
	l.head = node
	if l.tail == nil {
		l.tail = node
	}
	l.size++
	return node
}

// Find возвращает первое совпадение; O(n), дополнительная память O(1).
func (l *List[T]) Find(value T) *Node[T] {
	for node := l.head; node != nil; node = node.Next {
		if node.Value == value {
			return node
		}
	}
	return nil
}

// At проходит от head: индекс НЕ даёт O(1), как в срезе. Некорректный индекс — false.
func (l *List[T]) At(index int) (*Node[T], bool) {
	if index < 0 || index >= l.size {
		return nil, false
	}
	node := l.head
	for i := 0; i < index; i++ {
		node = node.Next
	}
	return node, true
}

// InsertAfter требует живой handle этого списка. owner проверяет принадлежность
// за O(1); поиск узла по значению/индексу оплачивается отдельно. nil — не prepend.
func (l *List[T]) InsertAfter(at *Node[T], value T) (*Node[T], bool) {
	if at == nil || at.owner != l {
		return nil, false
	}
	node := &Node[T]{Value: value, Next: at.Next, owner: l}
	at.Next = node
	if l.tail == at {
		l.tail = node
	}
	l.size++
	return node, true
}

// DeleteAfter также O(1), если предшественник известен. Отсоединённый узел
// очищается: внешняя ссылка на него не удержит остаток списка и его владельца.
func (l *List[T]) DeleteAfter(at *Node[T]) (T, bool) {
	if at == nil || at.owner != l || at.Next == nil {
		var zero T
		return zero, false
	}
	node := at.Next
	at.Next = node.Next
	if l.tail == node {
		l.tail = at
	}
	value := node.Value
	var zero T
	node.Value = zero
	node.Next = nil
	node.owner = nil
	l.size--
	return value, true
}
func (l *List[T]) PopFront() (T, bool) {
	if l.head == nil {
		var zero T
		return zero, false
	}
	node := l.head
	l.head = node.Next
	l.size--
	if l.size == 0 {
		l.tail = nil
	}
	value := node.Value
	var zero T
	node.Value = zero
	node.Next = nil
	node.owner = nil
	return value, true
}

// RemoveFirst удаляет первое совпадение. В singly list сначала нужен
// предшественник: весь вызов O(n), хотя изменение уже найденной связи O(1).
func (l *List[T]) RemoveFirst(value T) bool {
	if l.head == nil {
		return false
	}
	if l.head.Value == value {
		l.PopFront()
		return true
	}
	for previous := l.head; previous.Next != nil; previous = previous.Next {
		if previous.Next.Value == value {
			l.DeleteAfter(previous)
			return true
		}
	}
	return false
}

// Reverse меняет связи, не значения. Старый head становится tail.
func (l *List[T]) Reverse() { l.tail = l.head; l.head = Reverse(l.head) }

// Reverse работает с конечной ациклической цепочкой. next нужно сохранить ДО
// перезаписи current.Next, иначе потеряем ещё не обработанный хвост. O(n)/O(1).
// Для узлов внутри List используйте метод List.Reverse, чтобы обновить metadata.
func Reverse[T comparable](head *Node[T]) *Node[T] { return reverse(head, nil) }
func reverse[T comparable](head *Node[T], observe func(*Node[T], *Node[T])) *Node[T] {
	var previous *Node[T]
	current := head
	if observe != nil {
		observe(previous, current)
	}
	for current != nil {
		next := current.Next
		current.Next = previous
		previous = current
		current = next
		if observe != nil {
			observe(previous, current)
		}
	}
	return previous
}

// ReverseState сохраняет физические узлы в исходном порядке и снимок их связей.
// -1 означает nil. Для trace память O(n²), для обычного Reverse остаётся O(1).
type ReverseState[T comparable] struct {
	Values            []T
	NextIndices       []int
	Previous, Current int
}

// ReverseTrace требует ациклическую цепочку, как и Reverse.
func ReverseTrace[T comparable](head *Node[T]) (*Node[T], []ReverseState[T]) {
	var nodes []*Node[T]
	indices := map[*Node[T]]int{}
	for p := head; p != nil; p = p.Next {
		indices[p] = len(nodes)
		nodes = append(nodes, p)
	}
	index := func(p *Node[T]) int {
		if p == nil {
			return -1
		}
		return indices[p]
	}
	var states []ReverseState[T]
	result := reverse(head, func(previous, current *Node[T]) {
		state := ReverseState[T]{Values: make([]T, len(nodes)), NextIndices: make([]int, len(nodes)), Previous: index(previous), Current: index(current)}
		for i, p := range nodes {
			state.Values[i] = p.Value
			state.NextIndices[i] = index(p.Next)
		}
		states = append(states, state)
	})
	return result, states
}

// HasCycle не меняет связи. Если цикл есть, быстрый указатель догонит медленный.
// Проверяем fast и fast.Next до второго шага. O(n), дополнительная память O(1).
func HasCycle[T comparable](head *Node[T]) bool {
	slow, fast := head, head
	for fast != nil && fast.Next != nil {
		slow = slow.Next
		fast = fast.Next.Next
		if slow == fast {
			return true
		}
	}
	return false
}

// Values — независимый снимок; O(n) времени и памяти. Не для циклической цепочки.
func (l *List[T]) Values() []T {
	result := make([]T, 0, l.size)
	for n := l.head; n != nil; n = n.Next {
		result = append(result, n.Value)
	}
	return result
}
