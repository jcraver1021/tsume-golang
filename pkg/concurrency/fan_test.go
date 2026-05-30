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

	workFn := func(x int) (int, error) {
		if x%2 == 0 {
			return x * x, nil
		}
		return 0, errors.New("odd number")
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

			outChannels, errChannel := FanOut(in, tc.nWorkers, workFn)
			results := FanIn(outChannels...)

			expectedResults := map[int]bool{4: true, 16: true}
			resultCount := 0
			errCount := 0

			// Read from both channels concurrently to avoid deadlock
			for {
				select {
				case result, ok := <-results:
					if !ok {
						results = nil
					} else {
						if !expectedResults[result] {
							t.Errorf("Unexpected result: %d", result)
						}
						resultCount++
					}
				case err, ok := <-errChannel:
					if !ok {
						errChannel = nil
					} else {
						if err.Err.Error() != "odd number" {
							t.Errorf("Unexpected error: %v", err.Err)
						}
						errCount++
					}
				}

				if results == nil && errChannel == nil {
					break
				}
			}

			if resultCount != 2 {
				t.Errorf("Expected 2 results, got %d", resultCount)
			}
			if errCount != 3 {
				t.Errorf("Expected 3 errors, got %d", errCount)
			}
		})
	}
}
