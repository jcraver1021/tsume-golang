package concurrency

type I any
type O any

type JobStatus int

const (
	StatusSkipped JobStatus = iota
	StatusSuccess
	StatusError
)

type JobResult[I, O any] struct {
	Input  I
	Output O
	Status JobStatus
	Err    error
}

type Job[I any, O any] func(I) JobResult[I, O]

type Service[I any] func(I) error
