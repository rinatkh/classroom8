package stack

import (
	"fmt"
	"math/rand"
	"slices"
	"testing"

	"github.com/rinatkh/classroom8/internal/testutil"
)

type contract interface {
	Push(int)
	Pop() (int, bool)
	Peek() (int, bool)
	Len() int
}

func TestContracts(t *testing.T) {
	factories := []struct {
		name string
		new  func() contract
	}{{"slice", func() contract { return new(Slice[int]) }}, {"linked", func() contract { return new(Linked[int]) }}}
	for _, f := range factories {
		t.Run(f.name, func(t *testing.T) {
			for _, tc := range testutil.Cases() {
				t.Run(tc.Name, func(t *testing.T) {
					s := f.new()
					if v, ok := s.Pop(); ok || v != 0 {
						t.Fatal("empty removal")
					}
					if v, ok := s.Peek(); ok || v != 0 {
						t.Fatal("empty peek")
					}
					for i, v := range tc.Values {
						s.Push(v)
						if s.Len() != i+1 {
							t.Fatal("length after insert")
						}
						// Последний Push должен немедленно стать вершиной.
						want := v
						if got, ok := s.Peek(); !ok || got != want {
							t.Fatal("peek")
						}
					}
					want := slices.Clone(tc.Values)
					slices.Reverse(want)
					for i, v := range want {
						if got, ok := s.Peek(); !ok || got != v {
							t.Fatal("peek mutation")
						}
						if got, ok := s.Pop(); !ok || got != v {
							t.Fatalf("got %d want %d", got, v)
						}
						if s.Len() != len(want)-i-1 {
							t.Fatal("length after removal")
						}
					}
					if _, ok := s.Pop(); ok {
						t.Fatal("empty again")
					}
					s.Push(-7)
					if got, ok := s.Pop(); !ok || got != -7 {
						t.Fatal("reuse")
					}
				})
			}
		})
	}
}
func TestClearsReferences(t *testing.T) {
	for n := 1; n <= 12; n++ {
		t.Run(fmt.Sprintf("len=%d", n), func(t *testing.T) {
			var s Slice[*int]
			x := 7
			for i := 0; i < n; i++ {
				s.Push(&x)
			}
			backing := s.values[:cap(s.values)]
			for i := 0; i < n; i++ {
				s.Pop()
				if backing[n-i-1] != nil {
					t.Fatal("unused tail retains reference")
				}
			}
			for _, p := range backing {
				if p != nil {
					t.Fatal("reference remains in backing array")
				}
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	for _, tc := range testutil.Cases() {
		t.Run(tc.Name, func(t *testing.T) {
			var a Slice[int]
			var b Linked[int]
			for _, v := range tc.Values {
				a.Push(v)
				b.Push(v)
			}
			if a.IsEmpty() != (len(tc.Values) == 0) || b.IsEmpty() != a.IsEmpty() {
				t.Fatal("is empty")
			}
		})
	}
}
func TestValidParentheses(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		want       bool
	}{{"empty", "", true}, {"pair", "()", true}, {"nested", "({[]})", true}, {"adjacent", "()[]{}", true}, {"opening", "(", false}, {"closing", ")", false}, {"crossed", "([)]", false}, {"unfinished", "(()", false}, {"extra", "())", false}, {"letters", "abc", false}, {"unicode", "я", false}, {"spaces", "( )", false}, {"square", "[[]]", true}, {"deep", "(((())))", true}, {"wrong_close", "{]", false}} {
		t.Run(tc.name, func(t *testing.T) {
			if ValidParentheses(tc.text) != tc.want {
				t.Fatal("parentheses")
			}
		})
	}
}

func TestMixedSequences(t *testing.T) {
	for seed := int64(1); seed <= 12; seed++ {
		t.Run(fmt.Sprint(seed), func(t *testing.T) {
			a, b := new(Slice[int]), new(Linked[int])
			model := []int{}
			r := rand.New(rand.NewSource(seed))
			for step := 0; step < 1000; step++ {
				if r.Intn(2) == 0 {
					v := r.Intn(11) - 5
					a.Push(v)
					b.Push(v)
					model = append(model, v)
				} else {
					x, ok := a.Pop()
					y, valid := b.Pop()
					if ok != valid || ok != (len(model) > 0) || x != y {
						t.Fatal("status/result")
					}
					if ok {
						index := 0
						if "stack" == "stack" {
							index = len(model) - 1
						}
						if x != model[index] {
							t.Fatal("order")
						}
						if index == 0 {
							model = model[1:]
						} else {
							model = model[:index]
						}
					}
				}
				if a.Len() != len(model) || b.Len() != len(model) {
					t.Fatal("length")
				}
			}
		})
	}
}
