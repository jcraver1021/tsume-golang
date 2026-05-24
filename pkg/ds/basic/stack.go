package basic

type Stack[T any] struct {
	elements []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{
		elements: make([]T, 0),
	}
}

func (s *Stack[T]) Push(element T) {
	// Note: we do not need to check for capacity here, as the underlying slice will automatically resize as needed
	s.elements = append(s.elements, element)
}

func (s *Stack[T]) Pop() (T, bool) {
	if len(s.elements) == 0 {
		var zero T
		return zero, false
	}
	element := s.elements[len(s.elements)-1]
	s.elements = s.elements[:len(s.elements)-1]
	return element, true
}
