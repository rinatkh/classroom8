package linkedlist

import (
	"slices"
	"testing"

	"github.com/rinatkh/classroom8/internal/testutil"
)

func build(values []int) *List[int] {
	l := new(List[int])
	for _, v := range values {
		l.Append(v)
	}
	return l
}
func assertList(t *testing.T, l *List[int], want []int) {
	t.Helper()
	if HasCycle(l.head) {
		t.Fatal("unexpected cycle")
	}
	if l.size != len(want) || !slices.Equal(l.Values(), want) {
		t.Fatalf("got %v len=%d want %v", l.Values(), l.size, want)
	}
	count := 0
	var last *Node[int]
	for n := l.head; n != nil; n = n.Next {
		if n.owner != l {
			t.Fatal("owner")
		}
		last = n
		count++
	}
	if count != l.size || last != l.tail || (l.tail != nil && l.tail.Next != nil) {
		t.Fatal("head/tail/size invariant")
	}
	if (l.Head() == nil) != (l.Len() == 0) || (l.Tail() == nil) != (l.Len() == 0) {
		t.Fatal("empty invariant")
	}
}
func TestListContracts(t *testing.T) {
	for _, tc := range testutil.Cases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("Append", func(t *testing.T) {
				l := build(tc.Values)
				n := l.Append(91)
				if n != l.Tail() {
					t.Fatal("tail handle")
				}
				assertList(t, l, append(slices.Clone(tc.Values), 91))
			})
			t.Run("Prepend", func(t *testing.T) {
				l := build(tc.Values)
				n := l.Prepend(91)
				if n != l.Head() {
					t.Fatal("head handle")
				}
				assertList(t, l, append([]int{91}, tc.Values...))
			})
			t.Run("Find", func(t *testing.T) {
				l := build(tc.Values)
				for _, target := range append(slices.Clone(tc.Values), 1234567) {
					n := l.Find(target)
					index := slices.Index(tc.Values, target)
					if index < 0 {
						if n != nil {
							t.Fatal("missing")
						}
					} else {
						at, _ := l.At(index)
						if n != at {
							t.Fatal("must find first")
						}
					}
				}
				assertList(t, l, tc.Values)
			})
			t.Run("At", func(t *testing.T) {
				l := build(tc.Values)
				for index := -1; index <= len(tc.Values); index++ {
					n, ok := l.At(index)
					valid := index >= 0 && index < len(tc.Values)
					if ok != valid {
						t.Fatal("bounds")
					}
					if ok && n.Value != tc.Values[index] {
						t.Fatal("index")
					}
				}
			})
			t.Run("InsertAfter", func(t *testing.T) {
				for _, pos := range []int{0, len(tc.Values) / 2, len(tc.Values) - 1} {
					l := build(tc.Values)
					at, valid := l.At(pos)
					n, ok := l.InsertAfter(at, 91)
					if ok != valid {
						t.Fatal("handle validation")
					}
					want := slices.Clone(tc.Values)
					if ok {
						want = append(want, 0)
						copy(want[pos+2:], want[pos+1:])
						want[pos+1] = 91
						if at.Next != n {
							t.Fatal("link")
						}
					}
					assertList(t, l, want)
				}
			})
			t.Run("DeleteAfter", func(t *testing.T) {
				l := build(append([]int{90}, tc.Values...))
				victim := l.head.Next
				value, ok := l.DeleteAfter(l.head)
				if ok != (len(tc.Values) > 0) {
					t.Fatal("status")
				}
				want := tc.Values
				if ok {
					if value != want[0] || victim.Next != nil || victim.owner != nil || victim.Value != 0 {
						t.Fatal("detach")
					}
					want = want[1:]
				}
				assertList(t, l, append([]int{90}, want...))
			})
			t.Run("RemoveFirst", func(t *testing.T) {
				targets := append(slices.Clone(tc.Values), 1234567)
				for _, value := range targets {
					l := build(tc.Values)
					index := slices.Index(tc.Values, value)
					victim := l.Find(value)
					ok := l.RemoveFirst(value)
					if ok != (index >= 0) {
						t.Fatal("status")
					}
					want := slices.Clone(tc.Values)
					if ok {
						want = append(want[:index], want[index+1:]...)
						if victim.Next != nil || victim.owner != nil || victim.Value != 0 {
							t.Fatal("retained node")
						}
					}
					assertList(t, l, want)
				}
			})
			t.Run("PopFront", func(t *testing.T) {
				l := build(tc.Values)
				for _, v := range tc.Values {
					old := l.head
					got, ok := l.PopFront()
					if !ok || got != v || old.Next != nil || old.owner != nil || old.Value != 0 {
						t.Fatal("FIFO/detach")
					}
				}
				got, ok := l.PopFront()
				if ok || got != 0 {
					t.Fatal("empty")
				}
				assertList(t, l, nil)
			})
			t.Run("Reverse", func(t *testing.T) {
				l := build(tc.Values)
				oldHead, oldTail := l.head, l.tail
				l.Reverse()
				want := slices.Clone(tc.Values)
				slices.Reverse(want)
				assertList(t, l, want)
				if l.head != oldTail || l.tail != oldHead {
					t.Fatal("head/tail swap")
				}
				l.Reverse()
				assertList(t, l, tc.Values)
			})
			t.Run("RawReverse", func(t *testing.T) {
				var head *Node[int]
				for i := len(tc.Values) - 1; i >= 0; i-- {
					head = &Node[int]{Value: tc.Values[i], Next: head}
				}
				head = Reverse(head)
				var got []int
				for n := head; n != nil; n = n.Next {
					got = append(got, n.Value)
				}
				want := slices.Clone(tc.Values)
				slices.Reverse(want)
				if !slices.Equal(got, want) {
					t.Fatal("reverse chain")
				}
			})
			t.Run("MetadataAndSnapshot", func(t *testing.T) {
				l := build(tc.Values)
				assertList(t, l, tc.Values)
				snapshot := l.Values()
				if len(snapshot) > 0 {
					snapshot[0] = 91
					assertList(t, l, tc.Values)
				}
			})
		})
	}
}
func TestHandleRejection(t *testing.T) {
	for _, name := range []string{"nil", "foreign_empty", "foreign_first", "foreign_tail", "removed_head", "removed_middle", "removed_tail", "deleted_after", "own_tail_no_next", "zero_owner", "other_list_handle"} {
		t.Run(name, func(t *testing.T) {
			l := build([]int{1, 2, 3})
			other := build([]int{7, 8})
			var at *Node[int]
			switch name {
			case "foreign_empty":
				other = new(List[int])
				at = other.Head()
			case "foreign_first", "other_list_handle":
				at = other.head
			case "foreign_tail":
				at = other.tail
			case "removed_head":
				at = l.head
				l.RemoveFirst(1)
			case "removed_middle":
				at = l.Find(2)
				l.RemoveFirst(2)
			case "removed_tail":
				at = l.tail
				l.RemoveFirst(3)
			case "deleted_after":
				at = l.head.Next
				l.DeleteAfter(l.head)
			case "own_tail_no_next":
				at = l.tail
			case "zero_owner":
				at = &Node[int]{Value: 9}
			}
			before := l.Values()
			if name != "own_tail_no_next" {
				if _, ok := l.InsertAfter(at, 5); ok {
					t.Fatal("foreign/stale accepted")
				}
			}
			if _, ok := l.DeleteAfter(at); ok {
				t.Fatal("invalid delete accepted")
			}
			assertList(t, l, before)
		})
	}
}
func TestHasCycle(t *testing.T) {
	cases := []struct {
		name     string
		n, entry int
	}{
		{"nil", 0, -1}, {"one_acyclic", 1, -1}, {"self", 1, 0}, {"two_acyclic", 2, -1}, {"two_head", 2, 0}, {"two_tail", 2, 1}, {"three_head", 3, 0}, {"three_middle", 3, 1}, {"three_tail", 3, 2}, {"even_acyclic", 10, -1}, {"odd_acyclic", 11, -1}, {"long_prefix", 20, 18}, {"long_loop", 20, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nodes := make([]*Node[int], tc.n)
			for i := range nodes {
				nodes[i] = &Node[int]{Value: i % 3}
				if i > 0 {
					nodes[i-1].Next = nodes[i]
				}
			}
			var head *Node[int]
			if tc.n > 0 {
				head = nodes[0]
				if tc.entry >= 0 {
					nodes[tc.n-1].Next = nodes[tc.entry]
				}
			}
			before := make([]*Node[int], tc.n)
			for i, n := range nodes {
				before[i] = n.Next
			}
			if HasCycle(head) != (tc.entry >= 0) {
				t.Fatal("cycle result")
			}
			for i, n := range nodes {
				if n.Next != before[i] {
					t.Fatal("input changed")
				}
			}
		})
	}
}

func TestReverseTrace(t *testing.T) {
	for _, tc := range testutil.Cases() {
		t.Run(tc.Name, func(t *testing.T) {
			var head *Node[int]
			for i := len(tc.Values) - 1; i >= 0; i-- {
				head = &Node[int]{Value: tc.Values[i], Next: head}
			}
			head, states := ReverseTrace(head)
			if len(states) != len(tc.Values)+1 {
				t.Fatal("states")
			}
			var got []int
			for p := head; p != nil; p = p.Next {
				got = append(got, p.Value)
			}
			want := slices.Clone(tc.Values)
			slices.Reverse(want)
			if !slices.Equal(got, want) {
				t.Fatal("reverse")
			}
			for step, state := range states {
				if !slices.Equal(state.Values, tc.Values) {
					t.Fatal("physical node values changed")
				}
				for i, next := range state.NextIndices {
					expected := i + 1
					if i == len(tc.Values)-1 {
						expected = -1
					}
					if i < step {
						expected = i - 1
					}
					if next != expected {
						t.Fatal("snapshot links")
					}
				}
			}
			if len(tc.Values) > 0 {
				states[len(states)-1].NextIndices[0] = 99
				if states[0].NextIndices[0] == 99 {
					t.Fatal("snapshot alias")
				}
			}
		})
	}
}
