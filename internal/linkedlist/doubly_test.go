package linkedlist

import (
	"slices"
	"testing"

	"github.com/rinatkh/classroom8/internal/testutil"
)

func assertDouble(t *testing.T, l *Doubly[int], want []int) {
	t.Helper()
	if l.Len() != len(want) || !slices.Equal(l.Values(), want) {
		t.Fatalf("got %v want %v", l.Values(), want)
	}
	var previous *DNode[int]
	count := 0
	for n := l.Head(); n != nil; n = n.Next {
		if n.Prev != previous || n.owner != l {
			t.Fatal("forward/owner")
		}
		previous = n
		count++
		if count > len(want) {
			t.Fatal("cycle")
		}
	}
	if previous != l.Tail() || count != len(want) {
		t.Fatal("tail/size")
	}
	i := len(want) - 1
	for n := l.Tail(); n != nil; n = n.Prev {
		if i < 0 || n.Value != want[i] {
			t.Fatal("reverse links")
		}
		i--
	}
	if i != -1 {
		t.Fatal("lost prefix")
	}
}
func TestDoublyContracts(t *testing.T) {
	for _, tc := range testutil.Cases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("Append", func(t *testing.T) {
				var l Doubly[int]
				for _, v := range tc.Values {
					l.Append(v)
				}
				l.Append(91)
				assertDouble(t, &l, append(slices.Clone(tc.Values), 91))
			})
			t.Run("Prepend", func(t *testing.T) {
				var l Doubly[int]
				for i := len(tc.Values) - 1; i >= 0; i-- {
					l.Prepend(tc.Values[i])
				}
				l.Prepend(91)
				assertDouble(t, &l, append([]int{91}, tc.Values...))
			})
			t.Run("Find", func(t *testing.T) {
				var l Doubly[int]
				for _, v := range tc.Values {
					l.Append(v)
				}
				for _, target := range append(slices.Clone(tc.Values), 1234567) {
					n := l.Find(target)
					i := slices.Index(tc.Values, target)
					if (n != nil) != (i >= 0) {
						t.Fatal("found")
					}
					if n != nil {
						p := l.head
						for j := 0; j < i; j++ {
							p = p.Next
						}
						if n != p {
							t.Fatal("first")
						}
					}
				}
				assertDouble(t, &l, tc.Values)
			})
			t.Run("Remove", func(t *testing.T) {
				for pos := -1; pos < len(tc.Values); pos++ {
					var l Doubly[int]
					var victim *DNode[int]
					for i, v := range tc.Values {
						n := l.Append(v)
						if i == pos {
							victim = n
						}
					}
					ok := l.Remove(victim)
					want := slices.Clone(tc.Values)
					if ok != (pos >= 0) {
						t.Fatal("status")
					}
					if ok {
						want = append(want[:pos], want[pos+1:]...)
						if victim.Prev != nil || victim.Next != nil || victim.owner != nil || victim.Value != 0 || l.Remove(victim) {
							t.Fatal("detach/stale")
						}
					}
					assertDouble(t, &l, want)
					other := new(Doubly[int])
					if l.Remove(other.Append(8)) {
						t.Fatal("foreign")
					}
				}
			})
			t.Run("MetadataAndSnapshot", func(t *testing.T) {
				var l Doubly[int]
				for _, v := range tc.Values {
					l.Append(v)
				}
				assertDouble(t, &l, tc.Values)
				out := l.Values()
				if len(out) > 0 {
					out[0] = 91
					assertDouble(t, &l, tc.Values)
				}
			})
		})
	}
}
