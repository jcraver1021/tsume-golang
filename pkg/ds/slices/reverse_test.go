package slices_test

import (
	"slices"
	"testing"

	. "tsumegolang/pkg/ds/slices"
)

func TestReverse(t *testing.T) {
	testCases := []struct {
		name  string
		input []int
		want  []int
	}{
		{
			name:  "Empty slice",
			input: []int{},
			want:  []int{},
		},
		{
			name:  "Single element",
			input: []int{1},
			want:  []int{1},
		},
		{
			name:  "Multiple elements",
			input: []int{1, 2, 3, 4, 5},
			want:  []int{5, 4, 3, 2, 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Reverse(tc.input)
			if !slices.Equal(got, tc.want) {
				t.Errorf("Expected %v, got %v", tc.want, got)
			}
		})
	}
}
