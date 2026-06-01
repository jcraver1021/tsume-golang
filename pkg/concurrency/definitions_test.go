package concurrency_test

import (
	"errors"
	"testing"

	. "tsumegolang/pkg/concurrency"
)

func TestJob(t *testing.T) {
	testCases := []struct {
		name  string
		input int
		want  JobResult[int, int]
	}{
		{
			name:  "Job with even input",
			input: 4,
			want:  JobResult[int, int]{Input: 4, Output: 16, Err: nil},
		},
		{
			name:  "Job with odd input",
			input: 3,
			want:  JobResult[int, int]{Input: 3, Output: 0, Err: errors.New("odd number")},
		},
	}

	j := func(x int) (int, error) {
		if x%2 == 0 {
			return x * x, nil
		}
		return 0, errors.New("odd number")
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := j(tc.input)
			if err != nil && err.Error() != tc.want.Err.Error() {
				t.Errorf("expected error '%v', got '%v'", tc.want.Err, err)
			}
			if result != tc.want.Output {
				t.Errorf("expected output %d, got %d", tc.want.Output, result)
			}
		})
	}
}

func TestService(t *testing.T) {
	testCases := []struct {
		name  string
		input int
		want  error
	}{
		{
			name:  "Service with even input",
			input: 4,
			want:  nil,
		},
		{
			name:  "Service with odd input",
			input: 3,
			want:  errors.New("odd number"),
		},
	}

	s := func(x int) error {
		if x%2 == 0 {
			return nil
		}
		return errors.New("odd number")
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := s(tc.input)
			if err != nil && err.Error() != tc.want.Error() {
				t.Errorf("expected error '%v', got '%v'", tc.want, err)
			}
		})
	}
}
