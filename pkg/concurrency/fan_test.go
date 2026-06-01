package concurrency_test

import (
	"errors"
	"testing"

	. "tsumegolang/pkg/concurrency"
)

func TestFanOutAndFanIn(t *testing.T) {
	testCases := []struct {
		name     string
		input    []int
		nWorkers int
	}{
		{
			name:     "FanOut and FanIn with 2 workers",
			input:    []int{1, 2, 3, 4, 5},
			nWorkers: 2,
		},
		{
			name:     "FanOut and FanIn with 3 workers",
			input:    []int{1, 2, 3, 4, 5},
			nWorkers: 3,
		},
	}

	workFn := func(x int) JobResult[int, int] {
		if x%2 == 0 {
			return JobResult[int, int]{Input: x, Output: x * x, Status: StatusSuccess}
		}
		return JobResult[int, int]{Input: x, Err: errors.New("odd number"), Status: StatusError}
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			in := make(chan int)
			go func() {
				defer close(in)
				for _, v := range tc.input {
					in <- v
				}
			}()

			outChannels := FanOut(in, tc.nWorkers, workFn)
			results := FanIn(outChannels...)

			expectedResults := map[int]bool{4: true, 16: true}
			successCount := 0
			errCount := 0

			// Read results from the channel
			for result := range results {
				if result.Err != nil {
					errCount++
					continue
				}

				if !expectedResults[result.Output] {
					t.Errorf("Unexpected result: %d", result.Output)
				}
				successCount++
			}

			if successCount != 2 {
				t.Errorf("Expected 2 successful results, got %d", successCount)
			}
			if errCount != 3 {
				t.Errorf("Expected 3 errors, got %d", errCount)
			}
		})
	}
}
