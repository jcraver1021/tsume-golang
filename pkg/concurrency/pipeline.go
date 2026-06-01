package concurrency

type PipelineJob struct {
	Task      Job[any, any]
	Result    JobResult[any, any]
	Validator func(any) error
}

type PipelineJobResult JobResult[any, any]

func Pipeline(in <-chan PipelineJobResult, job PipelineJob) <-chan PipelineJobResult {
	out := make(chan PipelineJobResult)

	go func() {
		defer close(out)

		for input := range in {
			if err := job.Validator(input.Input); err != nil {
				out <- PipelineJobResult{Input: input.Input, Err: err}
				continue
			}

			result := job.Task(input.Input)
			if result.Err != nil {
				out <- PipelineJobResult{Input: input.Input, Err: result.Err}
			} else {
				out <- PipelineJobResult{Input: input.Input, Output: result.Output}
			}
		}
	}()

	return out
}
