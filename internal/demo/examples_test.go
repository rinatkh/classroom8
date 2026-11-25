package demo

import (
	"strings"
	"testing"
)

func TestExamples(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   func() string
		want string
	}{
		{"list", LinkedList, "reverse=[3 9 2]"}, {"stack", Stack, "pop=3/3 peek=2 len=2"}, {"queue", Queue, "dequeue=1/1 front=2 len=2"}, {"buffer", Buffer, "original=\"API\""}, {"ring", Ring, "slots=[40 20 30] values=[20 30 40] head=1 tail=1 len=3 cap=3"}, {"hash", Hash, "get-a=9/true len=4 cap=8 load=0.50"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(tc.fn(), tc.want) {
				t.Fatal(tc.fn())
			}
		})
	}
}
func TestTraces(t *testing.T) {
	scenes := Traces()
	if len(scenes) != 6 {
		t.Fatal("scene count")
	}
	for _, s := range scenes {
		for i, f := range s.Frames {
			if f.Title == "" {
				t.Fatalf("%s frame %d", s.ID, i)
			}
			if s.ID == "ring" && (f.Size < 0 || f.Size > f.Capacity || f.Tail != (f.Head+f.Size)%f.Capacity) {
				t.Fatal("ring trace")
			}
			if s.ID == "hash" {
				count := 0
				for _, chain := range f.Buckets {
					count += len(chain)
				}
				if count != f.Size || f.Capacity != len(f.Buckets) {
					t.Fatal("hash trace")
				}
			}
		}
	}
}
