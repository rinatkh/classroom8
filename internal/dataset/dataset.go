// Package dataset продолжает воспроизводимый подход занятия 7.
package dataset

import (
	"fmt"
	"math/rand"
)

const Seed int64 = 7

func GenerateRandomSlice(n int) []int {
	if n <= 0 {
		return nil
	}
	r := rand.New(rand.NewSource(Seed))
	out := make([]int, n)
	for i := range out {
		out[i] = r.Intn(2000001) - 1000000
	}
	return out
}

// Keys уникальны, затем переставлены с seed 7: всем структурам — те же string-ключи.
func Keys(n int) []string {
	if n <= 0 {
		return nil
	}
	r := rand.New(rand.NewSource(Seed))
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("key-%08d", i)
	}
	r.Shuffle(n, func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}
