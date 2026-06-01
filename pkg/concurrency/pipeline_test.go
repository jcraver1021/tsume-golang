package concurrency_test

import (
	"errors"
	"testing"

	. "tsumegolang/pkg/concurrency"
)

func TestPipeline(t *testing.T) {
	testCases := []struct {
		name            string
		input           []int
		expectedSuccess int
		expectedErrors  int
	}{
		{
			name:            "Pipeline with valid inputs",
			input:           []int{1, 2, 3, 4, 5},
			expectedSuccess: 5,
			expectedErrors:  0,
		},
		{
			name:            "Pipeline with some invalid inputs",
			input:           []int{1, -2, 3, -4, 5},
			expectedSuccess: 3,
			expectedErrors:  2,
		},
	}

	job := PipelineJob{
		Task: func(input any) JobResult[any, any] {
			x := input.(int)
			if x < 0 {
				return JobResult[any, any]{Input: x, Err: errors.New("negative number"), Status: StatusError}
			}
			return JobResult[any, any]{Input: x, Output: x * x, Status: StatusSuccess}
		},
		Validator: func(input any) error {
			x := input.(int)
			if x < 0 {
				return errors.New("negative number")
			}
			return nil
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			in := make(chan PipelineJobResult)
			go func() {
				defer close(in)
				for _, v := range tc.input {
					in <- PipelineJobResult{Input: v}
				}
			}()

			results := Pipeline(in, job)

			expectedResults := map[int]bool{1: true, 4: true, 9: true, 16: true, 25: true}
			successCount := 0
			errCount := 0

			for result := range results {
				if result.Err != nil {
					errCount++
					continue
				}

				if !expectedResults[result.Output.(int)] {
					t.Errorf("Unexpected result: %d", result.Output)
				} else {
					successCount++
				}
			}

			if successCount != tc.expectedSuccess {
				t.Errorf("Expected %d successful results, got %d", tc.expectedSuccess, successCount)
			}
			if errCount != tc.expectedErrors {
				t.Errorf("Expected %d errors, got %d", tc.expectedErrors, errCount)
			}
		})
	}
}
