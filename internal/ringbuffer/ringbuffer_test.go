package ringbuffer

import (
	"fmt"
	"math/rand"
	"slices"
	"testing"
)

func TestCapacity(t *testing.T) {
	for _, n := range []int{-100, -2, -1, 0, 1, 2, 3, 4, 7, 16, 31, 1000} {
		t.Run(fmt.Sprintf("cap=%d", n), func(t *testing.T) {
			r, err := New[int](n)
			if n <= 0 {
				if err != ErrCapacity || r != nil {
					t.Fatal("invalid capacity")
				}
				return
			}
			if err != nil || r.Cap() != n || r.Len() != 0 || !r.IsEmpty() || r.IsFull() {
				t.Fatal("constructor")
			}
		})
	}
}

// Каждая последовательность проверяет все публичные операции и wrap-around,
// а oracle — простой срез. Пустое/full различаем по len, а не по head==tail.
func TestRingModel(t *testing.T) {
	for _, capacity := range []int{1, 2, 3, 4, 5, 7, 8, 9, 16, 31, 64, 100} {
		t.Run(fmt.Sprintf("cap=%d", capacity), func(t *testing.T) {
			r, _ := New[int](capacity)
			var want []int
			random := rand.New(rand.NewSource(int64(capacity)))
			for step := 0; step < 1000; step++ {
				before := r.Snapshot()
				if random.Intn(3) != 0 {
					v := random.Intn(9) - 4
					ok := r.Enqueue(v)
					if ok != (len(want) < capacity) {
						t.Fatal("full policy")
					}
					if ok {
						want = append(want, v)
					} else if !slices.Equal(r.Snapshot().Slots, before.Slots) {
						t.Fatal("rejected write changed data")
					}
				} else {
					got, ok := r.Dequeue()
					if ok != (len(want) > 0) {
						t.Fatal("empty policy")
					}
					if ok {
						if got != want[0] {
							t.Fatal("FIFO")
						}
						want = want[1:]
					} else if got != 0 {
						t.Fatal("empty zero")
					}
				}
				state := r.Snapshot()
				if r.Len() != len(want) || r.Cap() != capacity || r.IsEmpty() != (len(want) == 0) || r.IsFull() != (len(want) == capacity) || !slices.Equal(state.Values, want) {
					t.Fatal("metadata/FIFO")
				}
				if state.Head < 0 || state.Head >= capacity || state.Tail < 0 || state.Tail >= capacity || state.Tail != (state.Head+state.Len)%capacity {
					t.Fatal("indices")
				}
				got, ok := r.Front()
				if ok != (len(want) > 0) || (ok && got != want[0]) {
					t.Fatal("front")
				}
				if len(state.Slots) > 0 {
					state.Slots[0] = 12345
					if r.Snapshot().Slots[0] == 12345 {
						t.Fatal("snapshot alias")
					}
				}
			}
		})
	}
}
func TestReferenceCleanup(t *testing.T) {
	for cap := 1; cap <= 12; cap++ {
		t.Run(fmt.Sprintf("cap=%d", cap), func(t *testing.T) {
			r, _ := New[*int](cap)
			x := 5
			for round := 0; round < 5; round++ {
				for i := 0; i < cap; i++ {
					if !r.Enqueue(&x) {
						t.Fatal("enqueue")
					}
				}
				if r.Enqueue(&x) {
					t.Fatal("overflow")
				}
				for i := 0; i < cap; i++ {
					oldHead := r.head
					r.Dequeue()
					if r.slots[oldHead] != nil {
						t.Fatal("reference retained after one dequeue")
					}
				}
				for _, p := range r.slots {
					if p != nil {
						t.Fatal("stale reference")
					}
				}
				if r.head != r.tail {
					t.Fatal("empty")
				}
			}
		})
	}
}
func TestZeroValue(t *testing.T) {
	var r Ring[int]
	if r.Enqueue(1) || r.IsFull() || !r.IsEmpty() {
		t.Fatal("disabled zero")
	}
	if v, ok := r.Dequeue(); ok || v != 0 {
		t.Fatal("dequeue")
	}
	if _, ok := r.Front(); ok {
		t.Fatal("front")
	}
	s := r.Snapshot()
	if s.Len != 0 || s.Cap != 0 {
		t.Fatal("snapshot")
	}
}
