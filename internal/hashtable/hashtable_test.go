package hashtable

import (
	"fmt"
	"hash/fnv"
	"slices"
	"strings"
	"testing"
)

var keys = []string{"", "a", "b", "ab", "ba", "same", "-1", "0", "000", "русский", "世界", "key with space", "\x00", strings.Repeat("x", 1000), "last"}

func invariant(t *testing.T, table *Table, want map[string]int) {
	t.Helper()
	if table.Len() != len(want) {
		t.Fatal("size")
	}
	seen := map[string]bool{}
	for index, head := range table.buckets {
		count := 0
		for e := head; e != nil; e = e.next {
			count++
			if count > table.size {
				t.Fatal("chain cycle")
			}
			if seen[e.key] || int(e.hash%uint64(table.Capacity())) != index || want[e.key] != e.value {
				t.Fatal("bucket/duplicate/value")
			}
			seen[e.key] = true
		}
	}
	if len(seen) != len(want) {
		t.Fatal("missing entry")
	}
	for k, v := range want {
		if got, ok := table.Get(k); !ok || got != v {
			t.Fatal("lookup")
		}
	}
}
func TestHashString(t *testing.T) {
	for _, k := range keys {
		t.Run(fmt.Sprintf("key=%q", k), func(t *testing.T) {
			f := fnv.New64a()
			f.Write([]byte(k))
			if hashString(k) != f.Sum64() {
				t.Fatal("FNV-1a oracle")
			}
		})
	}
}
func TestTableOperations(t *testing.T) {
	for _, key := range keys {
		t.Run(fmt.Sprintf("key=%q", key), func(t *testing.T) {
			for _, collision := range []bool{false, true} {
				t.Run(map[bool]string{false: "hasher=fnv", true: "hasher=constant"}[collision], func(t *testing.T) {
					var table *Table
					if collision {
						table = NewWithHasher(4, func(string) uint64 { return 1 })
					} else {
						table = New(4)
					}
					want := map[string]int{}
					for j, k := range keys {
						table.Set(k, j-7)
						want[k] = j - 7
					}
					invariant(t, table, want)
					table.Set(key, -91)
					want[key] = -91
					invariant(t, table, want)
					if _, ok := table.Get("absent-key"); ok || table.Delete("absent-key") {
						t.Fatal("missing")
					}
					snapshot := table.Snapshot()
					for _, bucket := range snapshot {
						if len(bucket) > 0 {
							bucket[0] = Pair{"altered", 99}
						}
					}
					invariant(t, table, want)
					if !table.Delete(key) || table.Delete(key) {
						t.Fatal("delete once")
					}
					delete(want, key)
					invariant(t, table, want)
					table.Set(key, 0)
					want[key] = 0
					if v, ok := table.Get(key); !ok || v != 0 {
						t.Fatal("zero is present")
					}
					invariant(t, table, want)
					for _, k := range keys {
						table.Delete(k)
					}
					invariant(t, table, map[string]int{})
					if table.LoadFactor() != 0 || table.Capacity() < 4 {
						t.Fatal("empty capacity")
					}
				})
			}
		})
	}
}
func TestResize(t *testing.T) {
	for _, n := range []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 12, 24, 100, 1000} {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			table := New(4)
			want := map[string]int{}
			for i := 0; i < n; i++ {
				key := fmt.Sprint(i)
				table.Set(key, i)
				want[key] = i
				invariant(t, table, want)
				if table.LoadFactor() > 0.75 {
					t.Fatal("threshold")
				}
			}
			oldCap := table.Capacity()
			for k, v := range want {
				table.Set(k, v+1)
				want[k] = v + 1
			}
			if table.Capacity() != oldCap {
				t.Fatal("update resized")
			}
			invariant(t, table, want)
		})
	}
}
func TestZeroAndCapacity(t *testing.T) {
	for _, n := range []int{-99, -1, 0, 1, 2, 3, 4, 5, 8, 16, 31, 100} {
		t.Run(fmt.Sprintf("cap=%d", n), func(t *testing.T) {
			table := New(n)
			if table.Capacity() != max(n, 4) || table.Len() != 0 || table.LoadFactor() != 0 {
				t.Fatal("constructor")
			}
			var zero Table
			if _, ok := zero.Get("a"); ok || zero.Delete("a") || len(zero.Snapshot()) != 0 || zero.LoadFactor() != 0 {
				t.Fatal("zero")
			}
			zero.Set("a", n)
			invariant(t, &zero, map[string]int{"a": n})
		})
	}
}
func TestDeleteUnlinksEntry(t *testing.T) {
	for _, key := range keys {
		t.Run(fmt.Sprintf("key=%q", key), func(t *testing.T) {
			table := NewWithHasher(4, func(string) uint64 { return 0 })
			table.Set("other", 1)
			table.Set(key, 2)
			victim := table.buckets[0]
			if !table.Delete(key) || victim.key != "" || victim.next != nil || victim.value != 0 {
				t.Fatal("stale references")
			}
			flat := []string{}
			for _, bucket := range table.Snapshot() {
				for _, p := range bucket {
					flat = append(flat, p.Key)
				}
			}
			if !slices.Equal(flat, []string{"other"}) {
				t.Fatal("chain retained")
			}
		})
	}
}
