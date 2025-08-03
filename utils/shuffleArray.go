package utils

import (
	"math/rand"
	"time"
)

func ShuffleArray[T any](array []T) []T {
	rand.Seed(time.Now().UnixNano())
	shuffled := make([]T, len(array))
	perm := rand.Perm(len(array))
	for i, v := range perm {
		shuffled[v] = array[i]
	}
	return shuffled
}
