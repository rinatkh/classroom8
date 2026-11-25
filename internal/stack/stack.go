// Package stack реализует один контракт LIFO поверх среза и узлов.
package stack

import "github.com/rinatkh/classroom8/internal/linkedlist"

// Slice использует конец среза. Push амортизированно O(1), при росте O(n).
// Общая память O(cap), отдельный Pop/Peek O(1), нулевая структура готова к работе.
type Slice[T any] struct{ values []T }

func (s *Slice[T]) Push(value T) { s.values = append(s.values, value) }
func (s *Slice[T]) Pop() (T, bool) {
	if len(s.values) == 0 {
		var zero T
		return zero, false
	}
	last := len(s.values) - 1
	value := s.values[last]
	var zero T
	s.values[last] = zero // Без очистки backing array продолжит удерживать ссылку.
	s.values = s.values[:last]
	return value, true
}
func (s *Slice[T]) Peek() (T, bool) {
	if len(s.values) == 0 {
		var zero T
		return zero, false
	}
	return s.values[len(s.values)-1], true
}
func (s *Slice[T]) Len() int      { return len(s.values) }
func (s *Slice[T]) IsEmpty() bool { return s.Len() == 0 }

// Linked сохраняет LIFO, но каждый Push выделяет узел. Для общности с List
// используется comparable, хотя самой идее стека сравнение значений не требуется.
type Linked[T comparable] struct{ list linkedlist.List[T] }

func (s *Linked[T]) Push(value T)   { s.list.Prepend(value) }
func (s *Linked[T]) Pop() (T, bool) { return s.list.PopFront() }
func (s *Linked[T]) Peek() (T, bool) {
	if s.list.Head() == nil {
		var zero T
		return zero, false
	}
	return s.list.Head().Value, true
}
func (s *Linked[T]) Len() int      { return s.list.Len() }
func (s *Linked[T]) IsEmpty() bool { return s.Len() == 0 }

// ValidParentheses — разобранный пример применения стека, а не решение в HW.
// Принимаются только ASCII ()[]{}; любой другой символ отвергается. O(n)/O(n).
func ValidParentheses(text string) bool {
	var s Slice[byte]
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '(', '[', '{':
			s.Push(text[i])
		case ')', ']', '}':
			opening, ok := s.Pop()
			if !ok {
				return false
			}
			if (text[i] == ')' && opening != '(') || (text[i] == ']' && opening != '[') || (text[i] == '}' && opening != '{') {
				return false
			}
		default:
			return false
		}
	}
	return s.IsEmpty()
}
