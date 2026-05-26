package monotonic

import (
	"cmp"

	"tsumegolang/pkg/ds/basic"
	"tsumegolang/pkg/ds/slices"
)

const (
	NoValue = -1
)

func getNextInconsistentElements[T cmp.Ordered](input []T, compare func(a, b T) bool) []int {
	stack := basic.NewStack[int]()
	result := make([]int, len(input))
	for i := range result {
		result[i] = NoValue
	}

	for i, v := range input {
		for !stack.IsEmpty() {
			topIndex, _ := stack.Peek()
			if compare(input[topIndex], v) {
				break
			}
			stack.Pop()
			result[topIndex] = i
		}
		stack.Push(i)
	}

	return result
}

func GetNextSmallerElements[T cmp.Ordered](input []T) []int {
	return getNextInconsistentElements(input, func(a, b T) bool {
		return cmp.Compare(a, b) <= 0
	})
}

func GetNextGreaterElements[T cmp.Ordered](input []T) []int {
	return getNextInconsistentElements(input, func(a, b T) bool {
		return cmp.Compare(a, b) >= 0
	})
}

func GetPreviousSmallerElements[T cmp.Ordered](input []T) []int {
	result := getNextInconsistentElements(slices.Reverse(input), func(a, b T) bool {
		return cmp.Compare(a, b) <= 0
	})

	for i := range result {
		if result[i] != NoValue {
			result[i] = len(input) - 1 - result[i]
		}
	}

	return slices.Reverse(result)
}

func GetPreviousGreaterElements[T cmp.Ordered](input []T) []int {
	result := getNextInconsistentElements(slices.Reverse(input), func(a, b T) bool {
		return cmp.Compare(a, b) >= 0
	})

	for i := range result {
		if result[i] != NoValue {
			result[i] = len(input) - 1 - result[i]
		}
	}

	return slices.Reverse(result)
}
