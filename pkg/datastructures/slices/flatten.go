package slices

func FlattenAppend[T any](slices [][]T) []T {
	var res []T
	for _, a := range slices {
		res = append(res, a...)
	}
	return res
}

func FlattenAllocate[T any](slices [][]T) []T {
	n := 0
	for _, a := range slices {
		n += len(a)
	}

	res := make([]T, n)
	offset := 0
	for _, a := range slices {
		copy(res[offset:], a)
		offset += len(a)
	}

	return res
}
