package slices

func Reverse[S ~[]T, T any](input S) []T {
	result := make([]T, len(input))

	for i, v := range input {
		result[len(input)-1-i] = v
	}

	return result
}
