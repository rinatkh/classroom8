package demo

import "testing"

func TestBufferTrace(t *testing.T) {
	frames := BufferTrace()
	want := []string{"", "Go ", "Go API", "API", "APIabcdefghijkl", "", "OK", ""}
	if len(frames) != len(want) {
		t.Fatal("frame count")
	}
	for i, f := range frames {
		text := []byte{}
		for _, value := range f.Values {
			text = append(text, byte(value))
		}
		if string(text) != want[i] || f.Size != len(want[i]) || f.Capacity < f.Size || f.Title == "" || f.Note == "" {
			t.Fatalf("frame %d: %+v", i, f)
		}
	}
	if frames[5].Capacity != frames[4].Capacity || frames[7].Capacity != frames[4].Capacity || frames[0].Capacity != 0 {
		t.Fatal("retained capacity")
	}
	first := BufferTrace()
	first[1].Values[0] = 0
	if BufferTrace()[1].Values[0] != 'G' {
		t.Fatal("aliased snapshots")
	}
}
