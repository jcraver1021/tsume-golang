package slices_test

import (
	"slices"
	"testing"

	. "tsumegolang/pkg/ds/slices"
)

func TestMap(t *testing.T) {
	testCases := []struct {
		name     string
		input    []int
		function func(int) int
		want     []int
	}{
		{
			name:     "Identity function",
			input:    []int{1, 2, 3, 4, 5},
			function: func(x int) int { return x },
			want:     []int{1, 2, 3, 4, 5},
		},
		{
			name:     "Square function",
			input:    []int{1, 2, 3, 4, 5},
			function: func(x int) int { return x * x },
			want:     []int{1, 4, 9, 16, 25},
		},
		{
			name:     "Negation function",
			input:    []int{1, -2, 3, -4, 5},
			function: func(x int) int { return -x },
			want:     []int{-1, 2, -3, 4, -5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Map(tc.input, tc.function)
			if !slices.Equal(got, tc.want) {
				t.Errorf("Expected %v, got %v", tc.want, got)
			}
		})
	}
}
