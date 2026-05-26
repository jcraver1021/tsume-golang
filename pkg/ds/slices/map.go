package slices

func Map[S ~[]T, T any, U any](input S, f func(T) U) []U {
	result := make([]U, len(input))
	for i, v := range input {
		result[i] = f(v)
	}
	return result
}
