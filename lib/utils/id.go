package utils

// NewIDGenerator returns a function that generates unique integer IDs.
func NewIDGenerator() func() int {
	i := -1
	return func() int {
		i += 1
		return i
	}
}
