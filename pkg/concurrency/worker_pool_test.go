package concurrency_test

import (
	"errors"
	"strconv"
	"sync"
	"testing"

	. "tsumegolang/pkg/concurrency"
)

func TestWorkerPool(t *testing.T) {
	client := func(s string) (int, error) {
		result, err := strconv.Atoi(s)
		return result, err
	}
	var clientLock sync.Mutex
	job := func(input string) JobResult[string, int] {
		clientLock.Lock()
		defer clientLock.Unlock()

		result, err := client(input)
		status := StatusSuccess
		if err != nil {
			status = StatusError
		}
		return JobResult[string, int]{Input: input, Output: result, Err: err, Status: status}
	}

	testCases := []struct {
		name       string
		numWorkers int
		inputs     []string
		want       []JobResult[string, int]
	}{
		{
			name:       "Single worker, valid inputs",
			numWorkers: 1,
			inputs:     []string{"1", "2", "3"},
			want: []JobResult[string, int]{
				{Input: "1", Output: 1, Status: StatusSuccess},
				{Input: "2", Output: 2, Status: StatusSuccess},
				{Input: "3", Output: 3, Status: StatusSuccess},
			},
		},
		{
			name:       "Multiple workers, valid inputs",
			numWorkers: 3,
			inputs:     []string{"4", "5", "6"},
			want: []JobResult[string, int]{
				{Input: "4", Output: 4, Status: StatusSuccess},
				{Input: "5", Output: 5, Status: StatusSuccess},
				{Input: "6", Output: 6, Status: StatusSuccess},
			},
		},
		{
			name:       "Single worker, invalid input",
			numWorkers: 1,
			inputs:     []string{"7", "invalid", "8"},
			want: []JobResult[string, int]{
				{Input: "7", Output: 7, Status: StatusSuccess},
				{Input: "invalid", Output: 0, Status: StatusError, Err: strconv.ErrSyntax},
				{Input: "8", Output: 8, Status: StatusSuccess},
			},
		},
		{
			name:       "Multiple workers, mixed inputs",
			numWorkers: 2,
			inputs:     []string{"9", "10", "invalid", "11"},
			want: []JobResult[string, int]{
				{Input: "9", Output: 9, Status: StatusSuccess},
				{Input: "10", Output: 10, Status: StatusSuccess},
				{Input: "invalid", Output: 0, Status: StatusError, Err: strconv.ErrSyntax},
				{Input: "11", Output: 11, Status: StatusSuccess},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wp := NewWorkerPool(job, tc.numWorkers)
			wp.Start()

			var results []JobResult[string, int]
			for _, input := range tc.inputs {
				resultCh, err := wp.Submit(input)
				if err != nil {
					t.Fatalf("Submit failed: %v", err)
				}
				result := <-resultCh
				results = append(results, result)
			}

			if len(results) != len(tc.want) {
				t.Fatalf("Expected %d results, got %d", len(tc.want), len(results))
			}

			for i, got := range results {
				want := tc.want[i]
				errMatch := (got.Err == nil && want.Err == nil) ||
					(got.Err != nil && want.Err != nil && errors.Is(got.Err, want.Err))
				if got.Input != want.Input || got.Output != want.Output || !errMatch || got.Status != want.Status {
					t.Errorf("Result mismatch at index %d: got %+v, want %+v", i, got, want)
				}
			}
		})
	}
}

func TestWorkerPoolCancellation(t *testing.T) {
	job := func(input string) JobResult[string, int] {
		return JobResult[string, int]{Input: input, Output: len(input), Status: StatusSuccess}
	}

	wp := NewWorkerPool(job, 1) // Single worker to ensure one job is queued
	wp.Start()

	// Submit more jobs than workers to ensure some are queued
	resultCh1, err := wp.Submit("test1")
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	resultCh2, err := wp.Submit("test2")
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	resultCh3, err := wp.Submit("test3")
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	wp.Shutdown()

	// After shutdown, all result channels should be closed
	// Some may have completed, some may have been drained without processing
	channels := []chan JobResult[string, int]{resultCh1, resultCh2, resultCh3}
	for i, ch := range channels {
		_, ok := <-ch
		if ok {
			// Channel received a value but should still be closed afterwards
			_, stillOpen := <-ch
			if stillOpen {
				t.Errorf("Channel %d is still open after shutdown", i+1)
			}
		}
		// If ok is false, channel was closed without sending (drained during cancellation)
	}
}
