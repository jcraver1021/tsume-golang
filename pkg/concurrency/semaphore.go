package concurrency

type Semaphore struct {
	permits chan struct{}
}

func NewSemaphore(n int) *Semaphore {
	return &Semaphore{
		permits: make(chan struct{}, n),
	}
}

func (s *Semaphore) Acquire() {
	s.permits <- struct{}{}
}

func (s *Semaphore) Release() {
	<-s.permits
}

func (s *Semaphore) Try() bool {
	select {
	case s.permits <- struct{}{}:
		return true
	default:
		return false
	}
}
