package demo

import (
	"fmt"

	"github.com/rinatkh/classroom8/internal/hashtable"
	"github.com/rinatkh/classroom8/internal/linkedlist"
	"github.com/rinatkh/classroom8/internal/queue"
	"github.com/rinatkh/classroom8/internal/ringbuffer"
	"github.com/rinatkh/classroom8/internal/stack"
)

type Frame struct {
	Title, Note                                   string
	Values, Links                                 []int
	Head, Tail, Previous, Current, Size, Capacity int
	Buckets                                       [][]hashtable.Pair
}
type Scene struct {
	ID, Title string
	Frames    []Frame
}

// Traces возвращает неизменяемые после создания снимки. HTML читает этот же JSON.
func Traces() []Scene {
	scenes := []Scene{}
	var l linkedlist.List[int]
	nodes := []*linkedlist.Node[int]{l.Append(10), l.Append(20), l.Append(30)}
	linked := Scene{ID: "list", Title: "InsertAfter и удаление связи"}
	captureList := func(title, note string) {
		ids := map[*linkedlist.Node[int]]int{}
		for i, p := range nodes {
			ids[p] = i
		}
		f := Frame{Title: title, Note: note, Head: -1, Tail: -1, Previous: -1, Current: -1, Size: l.Len()}
		if l.Head() != nil {
			f.Head = ids[l.Head()]
		}
		if l.Tail() != nil {
			f.Tail = ids[l.Tail()]
		}
		for _, p := range nodes {
			f.Values = append(f.Values, p.Value)
			next := -1
			if p.Next != nil {
				next = ids[p.Next]
			}
			f.Links = append(f.Links, next)
		}
		linked.Frames = append(linked.Frames, f)
	}
	captureList("Исходные узлы", "Соседство задаёт Next, не положение на картинке")
	inserted, _ := l.InsertAfter(nodes[0], 15)
	nodes = append(nodes, inserted)
	captureList("InsertAfter(head,15)", "Новый узел физически показан справа; связь ведёт к нему")
	l.DeleteAfter(nodes[0])
	captureList("DeleteAfter(head)", "Отсоединённый узел очищен: Value=0, Next=nil, owner=nil")
	scenes = append(scenes, linked)
	head := &linkedlist.Node[int]{Value: 10, Next: &linkedlist.Node[int]{Value: 20, Next: &linkedlist.Node[int]{Value: 30}}}
	_, states := linkedlist.ReverseTrace(head)
	rev := Scene{ID: "reverse", Title: "Reverse: previous / current / next"}
	for i, s := range states {
		rev.Frames = append(rev.Frames, Frame{Title: fmt.Sprintf("После %d перенаправлений", i), Note: "Сохраняем next до изменения связи; значения узлов не двигаются", Values: s.Values, Links: s.NextIndices, Previous: s.Previous, Current: s.Current, Head: -1, Tail: -1, Size: len(s.Values)})
	}
	scenes = append(scenes, rev)
	// Дополнительный срез здесь — модель наблюдаемых значений, а не реализация стека.
	st := Scene{ID: "stack", Title: "Stack: LIFO"}
	var s stack.Slice[int]
	var model []int
	st.Frames = append(st.Frames, Frame{Title: "Пусто"})
	for _, v := range []int{10, 20, 30} {
		s.Push(v)
		model = append(model, v)
		st.Frames = append(st.Frames, Frame{Title: fmt.Sprintf("Push(%d)", v), Values: append([]int(nil), model...), Size: s.Len()})
	}
	for s.Len() > 0 {
		v, _ := s.Pop()
		if v != model[len(model)-1] {
			panic("stack model")
		}
		model = model[:len(model)-1]
		st.Frames = append(st.Frames, Frame{Title: fmt.Sprintf("Pop() = %d", v), Values: append([]int(nil), model...), Size: s.Len()})
	}
	scenes = append(scenes, st)
	qu := Scene{ID: "queue", Title: "Queue: FIFO"}
	var q queue.Linked[int]
	model = nil
	qu.Frames = append(qu.Frames, Frame{Title: "Пусто"})
	for _, v := range []int{10, 20, 30} {
		q.Enqueue(v)
		model = append(model, v)
		qu.Frames = append(qu.Frames, Frame{Title: fmt.Sprintf("Enqueue(%d)", v), Values: append([]int(nil), model...), Size: q.Len()})
	}
	for q.Len() > 0 {
		v, _ := q.Dequeue()
		if v != model[0] {
			panic("queue model")
		}
		model = model[1:]
		qu.Frames = append(qu.Frames, Frame{Title: fmt.Sprintf("Dequeue() = %d", v), Values: append([]int(nil), model...), Size: q.Len()})
	}
	scenes = append(scenes, qu)
	ring := Scene{ID: "ring", Title: "Ring: head / tail и wrap-around"}
	r, _ := ringbuffer.New[int](3)
	captureRing := func(title string) {
		s := r.Snapshot()
		ring.Frames = append(ring.Frames, Frame{Title: title, Note: "head читает; tail пишет; len различает full и empty", Values: s.Slots, Head: s.Head, Tail: s.Tail, Size: s.Len, Capacity: s.Cap})
	}
	captureRing("Пусто: head=tail=0, len=0")
	for _, v := range []int{10, 20, 30} {
		r.Enqueue(v)
		title := fmt.Sprintf("Enqueue(%d)", v)
		if v == 30 {
			title += ": tail 2→0, переход через конец"
		}
		captureRing(title)
	}
	if r.Enqueue(99) {
		panic("overflow policy")
	}
	captureRing("Enqueue(99) = false: данные не меняются")
	r.Dequeue()
	captureRing("Dequeue() = 10: slot[0] очищен")
	r.Enqueue(40)
	captureRing("Enqueue(40): tail 0→1, повторно используем slot[0]")
	r.Dequeue()
	captureRing("Dequeue() = 20")
	r.Dequeue()
	captureRing("Dequeue() = 30: head проходит через конец")
	r.Dequeue()
	captureRing("Dequeue() = 40: пусто, head=tail=1")
	scenes = append(scenes, ring)
	hash := Scene{ID: "hash", Title: "Hash Table: коллизия и rehash"}
	table := hashtable.New(4)
	captureHash := func(title string) {
		hash.Frames = append(hash.Frames, Frame{Title: title, Note: "Индекс = hash(key) % capacity; rehash меняет цепочки, не пары key/value", Buckets: table.Snapshot(), Size: table.Len(), Capacity: table.Capacity()})
	}
	captureHash("4 пустых buckets")
	for i, k := range []string{"a", "e", "i", "m"} {
		table.Set(k, i+1)
		captureHash(fmt.Sprintf("Set(%q,%d)", k, i+1))
	}
	table.Set("a", 9)
	captureHash("Set(a,9): обновление без роста Len")
	table.Delete("e")
	captureHash("Delete(e): unlink из цепочки")
	scenes = append(scenes, hash)
	return scenes
}
