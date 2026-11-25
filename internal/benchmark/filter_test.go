package benchmark

import "testing"

func TestSelectedSize(t *testing.T) {
	for _, test := range []struct {
		filter string
		size   int
		want   bool
	}{
		{"", 100, true}, {"", 32, true}, {"100", 100, true}, {"100", 1000, false},
		{"32", 32, true}, {"1024", 32, false}, {"1000", 1000, true}, {"10", 100, false},
		{"invalid", 100, false}, {"0", 100, false}, {"-1", 100, false},
	} {
		t.Run(test.filter, func(t *testing.T) {
			t.Setenv("LESSON8_BENCH_SIZE", test.filter)
			if got := selectedSize(test.size); got != test.want {
				t.Fatalf("selectedSize(%d)=%v", test.size, got)
			}
		})
	}
}
