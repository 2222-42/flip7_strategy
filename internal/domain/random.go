package domain

import (
	"math/rand"
	"time"
)

// rnd is a package-level random source seeded once.
var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

// GetRandomInt returns a non-negative pseudo-random number in [0,n).
func GetRandomInt(n int) int {
	return rnd.Intn(n)
}

// GetRandomFloat returns a pseudo-random number in [0.0,1.0).
func GetRandomFloat() float64 {
	return rnd.Float64()
}

// Shuffle pseudo-randomizes the order of elements.
func Shuffle(n int, swap func(i, j int)) {
	rnd.Shuffle(n, swap)
}

// Perm returns, as a slice of n ints, a pseudo-random permutation of the integers [0,n).
func Perm(n int) []int {
	return rnd.Perm(n)
}
