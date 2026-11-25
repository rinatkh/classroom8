// Package demo связывает исполняемые примеры и пошаговые схемы с рабочим кодом.
package demo

import (
	"bytes"
	"fmt"

	"github.com/rinatkh/classroom8/internal/buffer"
	"github.com/rinatkh/classroom8/internal/hashtable"
	"github.com/rinatkh/classroom8/internal/linkedlist"
	"github.com/rinatkh/classroom8/internal/queue"
	"github.com/rinatkh/classroom8/internal/ringbuffer"
	"github.com/rinatkh/classroom8/internal/stack"
)

func LinkedList() string {
	var l linkedlist.List[int]
	for _, v := range []int{1, 2, 3} {
		l.Append(v)
	}
	l.InsertAfter(l.Find(2), 9)
	inserted := l.Values()
	l.RemoveFirst(1)
	l.Reverse()
	var d linkedlist.Doubly[int]
	d.Append(1)
	handle := d.Append(2)
	d.Append(3)
	d.Remove(handle)
	return fmt.Sprintf("insert=%v\nreverse=%v\ndoubly=%v", inserted, l.Values(), d.Values())
}
func Stack() string {
	var a stack.Slice[int]
	var b stack.Linked[int]
	for _, v := range []int{1, 2, 3} {
		a.Push(v)
		b.Push(v)
	}
	x, _ := a.Pop()
	y, _ := b.Pop()
	top, _ := a.Peek()
	return fmt.Sprintf("pop=%d/%d peek=%d len=%d\nvalid=%t crossed=%t", x, y, top, a.Len(), stack.ValidParentheses("([])"), stack.ValidParentheses("([)]"))
}
func Queue() string {
	var a queue.Slice[int]
	var b queue.Linked[int]
	for _, v := range []int{1, 2, 3} {
		a.Enqueue(v)
		b.Enqueue(v)
	}
	x, _ := a.Dequeue()
	y, _ := b.Dequeue()
	front, _ := a.Front()
	return fmt.Sprintf("dequeue=%d/%d front=%d len=%d\nslice dequeue shifts; linked dequeue detaches head", x, y, front, a.Len())
}
func Buffer() string {
	var a buffer.Linear
	var b bytes.Buffer
	a.Write([]byte("Go "))
	a.Write([]byte("API"))
	b.WriteString("Go API")
	p := make([]byte, 3)
	a.Read(p)
	snapshot := a.Bytes()
	snapshot[0] = 'X'
	return fmt.Sprintf("read=%q unread=%q std=%q\nsnapshot-copy=%q original=%q", p, a.Bytes(), b.Bytes(), snapshot, a.Bytes())
}
func Ring() string {
	r, _ := ringbuffer.New[int](3)
	for _, v := range []int{10, 20, 30} {
		r.Enqueue(v)
	}
	accepted := r.Enqueue(99)
	v, _ := r.Dequeue()
	r.Enqueue(40)
	s := r.Snapshot()
	return fmt.Sprintf("overflow-accepted=%t removed=%d\nslots=%v values=%v head=%d tail=%d len=%d cap=%d", accepted, v, s.Slots, s.Values, s.Head, s.Tail, s.Len, s.Cap)
}
func Hash() string {
	t := hashtable.New(4)
	for i, k := range []string{"a", "e", "i"} {
		t.Set(k, i+1)
	}
	before := t.Snapshot()
	t.Set("m", 4)
	t.Set("a", 9)
	v, ok := t.Get("a")
	return fmt.Sprintf("before=%v\nafter=%v\nget-a=%d/%t len=%d cap=%d load=%.2f", before, t.Snapshot(), v, ok, t.Len(), t.Capacity(), t.LoadFactor())
}
