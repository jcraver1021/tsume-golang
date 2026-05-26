package monotonic_test

import (
	"testing"

	. "tsumegolang/pkg/algo/monotonic"
)

func TestGetNextSmallerElements(t *testing.T) {
	testCases := []struct {
		name  string
		input []int
		want  []int
	}{
		{
			name:  "Increasing sequence",
			input: []int{1, 2, 3, 4, 5},
			want:  []int{NoValue, NoValue, NoValue, NoValue, NoValue},
		},
		{
			name:  "Decreasing sequence",
			input: []int{5, 4, 3, 2, 1},
			want:  []int{1, 2, 3, 4, NoValue},
		},
		{
			name:  "Mixed sequence",
			input: []int{3, 2, 4, 1, 5},
			want:  []int{1, 3, 3, NoValue, NoValue},
		},
		{
			name:  "All elements are the same",
			input: []int{3, 3, 3, 3},
			want:  []int{NoValue, NoValue, NoValue, NoValue},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := GetNextSmallerElements(tc.input)
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("GetNextSmallerElements(%v) = %v, want %v", tc.input, got, tc.want)
					break
				}
			}
		})
	}
}

func TestGetNextGreaterElements(t *testing.T) {
	testCases := []struct {
		name  string
		input []int
		want  []int
	}{
		{
			name:  "Increasing sequence",
			input: []int{1, 2, 3, 4, 5},
			want:  []int{1, 2, 3, 4, NoValue},
		},
		{
			name:  "Decreasing sequence",
			input: []int{5, 4, 3, 2, 1},
			want:  []int{NoValue, NoValue, NoValue, NoValue, NoValue},
		},
		{
			name:  "Mixed sequence",
			input: []int{3, 2, 4, 1, 5},
			want:  []int{2, 2, 4, 4, NoValue},
		},
		{
			name:  "All elements are the same",
			input: []int{3, 3, 3, 3},
			want:  []int{NoValue, NoValue, NoValue, NoValue},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := GetNextGreaterElements(tc.input)
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("GetNextGreaterElements(%v) = %v, want %v", tc.input, got, tc.want)
					break
				}
			}
		})
	}
}

func TestGetPreviousSmallerElements(t *testing.T) {
	testCases := []struct {
		name  string
		input []int
		want  []int
	}{
		{
			name:  "Increasing sequence",
			input: []int{1, 2, 3, 4, 5},
			want:  []int{NoValue, 0, 1, 2, 3},
		},
		{
			name:  "Decreasing sequence",
			input: []int{5, 4, 3, 2, 1},
			want:  []int{NoValue, NoValue, NoValue, NoValue, NoValue},
		},
		{
			name:  "Mixed sequence",
			input: []int{3, 2, 4, 1, 5},
			want:  []int{NoValue, NoValue, 1, NoValue, 3},
		},
		{
			name:  "All elements are the same",
			input: []int{3, 3, 3, 3},
			want:  []int{NoValue, NoValue, NoValue, NoValue},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := GetPreviousSmallerElements(tc.input)
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("GetPreviousSmallerElements(%v) = %v, want %v", tc.input, got, tc.want)
					break
				}
			}
		})
	}
}

func TestGetPreviousGreaterElements(t *testing.T) {
	testCases := []struct {
		name  string
		input []int
		want  []int
	}{
		{
			name:  "Increasing sequence",
			input: []int{1, 2, 3, 4, 5},
			want:  []int{NoValue, NoValue, NoValue, NoValue, NoValue},
		},
		{
			name:  "Decreasing sequence",
			input: []int{5, 4, 3, 2, 1},
			want:  []int{NoValue, 0, 1, 2, 3},
		},
		{
			name:  "Mixed sequence",
			input: []int{3, 2, 4, 1, 5},
			want:  []int{NoValue, 0, NoValue, 2, NoValue},
		},
		{
			name:  "All elements are the same",
			input: []int{3, 3, 3, 3},
			want:  []int{NoValue, NoValue, NoValue, NoValue},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := GetPreviousGreaterElements(tc.input)
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("GetPreviousGreaterElements(%v) = %v, want %v", tc.input, got, tc.want)
					break
				}
			}
		})
	}
}

func TestDailyTemperatureExample(t *testing.T) {
	input := []int{73, 74, 75, 71, 69, 72, 76, 73}
	want := []int{1, 1, 4, 2, 1, 1, NoValue, NoValue}

	nextGreater := GetNextGreaterElements(input)
	got := make([]int, len(nextGreater))
	for i, idx := range nextGreater {
		if idx != NoValue {
			got[i] = idx - i
		} else {
			got[i] = NoValue
		}
	}

	for i := range got {
		if got[i] != want[i] {
			t.Errorf("DailyTemperatureExample: GetNextGreaterElements(%v) = %v, want %v", input, got, want)
			break
		}
	}
}

func TestLargestRectangleInHistogramExample(t *testing.T) {
	testCases := []struct {
		name  string
		input []int
		want  int
	}{
		{
			name:  "LC Example 1",
			input: []int{2, 1, 5, 6, 2, 3},
			want:  10,
		},
		{
			name:  "LC Example 2",
			input: []int{2, 4},
			want:  4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prevSmaller := GetPreviousSmallerElements(tc.input)
			nextSmaller := GetNextSmallerElements(tc.input)

			maxArea := 0
			for i := range tc.input {
				height := tc.input[i]
				width := 1

				if prevSmaller[i] != NoValue {
					width += i - prevSmaller[i] - 1
				}
				if nextSmaller[i] != NoValue {
					width += nextSmaller[i] - i - 1
				}

				area := height * width
				if area > maxArea {
					maxArea = area
				}
			}

			if maxArea != tc.want {
				t.Errorf("LargestRectangleInHistogramExample: max area for %v = %v, want %v", tc.input, maxArea, tc.want)
			}
		})
	}
}
