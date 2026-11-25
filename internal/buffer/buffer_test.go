package buffer

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"testing"
)

var samples = []string{"", "a", "ab", "abc", "aaaa", "negative:-7", "with space", "\x00\x00", "язык Go", "世界", "0123456789", "long:abcdefghijklmnopqrstuvwxyz", "\n\t", "equal equal"}

func TestLinearContract(t *testing.T) {
	for _, text := range samples {
		t.Run(fmt.Sprintf("input=%q", text), func(t *testing.T) {
			var b Linear
			input := []byte(text)
			n, err := b.Write(input)
			if n != len(input) || err != nil || b.Len() != len(input) || b.Cap() < b.Len() {
				t.Fatal("write/metadata")
			}
			if len(input) > 0 {
				input[0] ^= 1
			}
			if string(b.Bytes()) != text {
				t.Fatal("write aliases input")
			}
			snap := b.Bytes()
			if len(snap) > 0 {
				snap[0] ^= 1
			}
			if string(b.Bytes()) != text {
				t.Fatal("snapshot alias")
			}
			var reference bytes.Buffer
			reference.WriteString(text)
			for _, width := range []int{0, 1, 2, 3, 100} {
				out, want := make([]byte, width), make([]byte, width)
				n, e := b.Read(out)
				m, f := reference.Read(want)
				if n != m || e != f || !slices.Equal(out, want) || b.Len() != reference.Len() {
					t.Fatal("Reader contract")
				}
			}
			b.Write([]byte(text))
			backing := b.data[:len(b.data)]
			oldCap := b.Cap()
			b.Reset()
			if b.Len() != 0 || b.Cap() != oldCap {
				t.Fatal("reset retains capacity")
			}
			for _, v := range backing {
				if v != 0 {
					t.Fatal("reset clearing")
				}
			}
			if _, e := b.Read(make([]byte, 1)); e != io.EOF {
				t.Fatal("EOF")
			}
			if n, e := b.Read(nil); n != 0 || e != nil {
				t.Fatal("zero read")
			}
		})
	}
}
func TestWriteReadCycles(t *testing.T) {
	for _, width := range []int{1, 2, 3, 4, 5, 7, 8, 16, 31, 64, 127, 256} {
		t.Run(fmt.Sprintf("read-chunk=%d", width), func(t *testing.T) {
			var got Linear
			var want bytes.Buffer
			for round := 0; round < 500; round++ {
				p := bytes.Repeat([]byte{byte(round)}, width)
				got.Write(p)
				want.Write(p)
				n := width/2 + 1
				a, b := make([]byte, n), make([]byte, n)
				x, e := got.Read(a)
				y, f := want.Read(b)
				if x != y || e != f || !bytes.Equal(a, b) || !bytes.Equal(got.Bytes(), want.Bytes()) {
					t.Fatal("cycle/compaction")
				}
			}
		})
	}
}
