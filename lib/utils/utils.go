// Package utils provides utilities for krpc-go.
package utils

// NewIDGenerator returns a function that generates unique integer IDs.
func NewIDGenerator() func() int {
	i := -1
	return func() int {
		i += 1
		return i
	}
}

// SlicesEqual checks if two slices have the same values.
func SlicesEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
