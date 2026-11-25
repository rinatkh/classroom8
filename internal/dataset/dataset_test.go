package dataset

import (
	"fmt"
	"slices"
	"testing"
)

func TestDatasets(t *testing.T) {
	for _, n := range []int{-1, 0, 1, 2, 3, 7, 16, 31, 100, 512, 1000, 10000} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			a, b := GenerateRandomSlice(n), GenerateRandomSlice(n)
			if len(a) != max(n, 0) || !slices.Equal(a, b) {
				t.Fatal("random reproducibility")
			}
			keys, again := Keys(n), Keys(n)
			if len(keys) != max(n, 0) || !slices.Equal(keys, again) {
				t.Fatal("keys reproducibility")
			}
			seen := map[string]bool{}
			for _, k := range keys {
				if seen[k] {
					t.Fatal("duplicate key")
				}
				seen[k] = true
			}
			if len(a) > 0 {
				a[0] = 2000001
				if b[0] == a[0] {
					t.Fatal("shared data")
				}
				keys[0] = "different"
				if again[0] == keys[0] {
					t.Fatal("shared keys")
				}
			}
		})
	}
}
