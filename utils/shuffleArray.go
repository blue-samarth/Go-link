package utils

import (
	"math/rand"
	"time"
)

func ShuffleArray[T any](array []T) []T {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	shuffled := make([]T, len(array))
	perm := r.Perm(len(array))
	for i, v := range perm {
		shuffled[v] = array[i]
	}
	return shuffled
}
