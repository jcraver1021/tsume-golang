// DisjointSet represents a Disjoint Set data structure. It does not itself
// store values represented by elements of the set, only their 0-indexed
// order of addition. Clients should therefore maintain a map of the indices
// returned to the values represented. This implementation uses is not
// thread-safe.

package disjointset

const (
	DefaultCapacity = 10
	ScaleFactor     = 2
	NotFound        = -1
)

type DisjointSet struct {
	c       int   // Initial capacity.
	s       int   // Scale factor for capacity expansion.
	parents []int // Each element's value represents the index of its parent.
	ranks   []int // Each element's value represents the rank of the tree rooted at that element.
}

// WithCapacity sets the initial capacity of the DisjointSet.
func WithCapacity(c int) func(*DisjointSet) {
	return func(ds *DisjointSet) {
		ds.c = c
	}
}

// WithScaleFactor sets the scale factor of the DisjointSet.
func WithScaleFactor(s int) func(*DisjointSet) {
	return func(ds *DisjointSet) {
		ds.s = s
	}
}

// NewDisjointSet creates a new DisjointSet with the given options.
func NewDisjointSet(opts ...func(*DisjointSet)) (*DisjointSet, error) {
	ds := &DisjointSet{
		c: DefaultCapacity,
		s: ScaleFactor,
	}
	for _, opt := range opts {
		opt(ds)
	}

	if !ds.isValid() {
		return nil, ErrInvalidConfiguration
	}

	ds.init()

	return ds, nil
}

// isValid checks if the DisjointSet is configured correctly.
func (s *DisjointSet) isValid() bool {
	return s.c > 0 && s.s > 1
}

// init initializes the set's internal arrays.
func (s *DisjointSet) init() {
	s.parents = make([]int, 0, s.c)
	s.ranks = make([]int, 0, s.c)
}

// ensureIdx returns the next available index, expanding the capacity by the scale factor if the array is full.
func (s *DisjointSet) ensureIdx() int {
	n := len(s.parents)
	if n >= cap(s.parents) {
		newParents := make([]int, 0, cap(s.parents)*s.s)
		copy(newParents, s.parents)
		s.parents = newParents
		newRanks := make([]int, 0, cap(s.ranks)*s.s)
		copy(newRanks, s.ranks)
		s.ranks = newRanks
	}

	return int(n)
}

// Add adds a new element to the DisjointSet and returns its index.
func (s *DisjointSet) Add() int {
	i := s.ensureIdx()
	s.parents = append(s.parents, i)
	s.ranks = append(s.ranks, 0)

	return i
}

// AddMany adds multiple new elements to the DisjointSet and returns the last index added.
func (s *DisjointSet) AddMany(n int) (int, error) {
	if n <= 0 {
		return NotFound, ErrInvalidRequest
	}

	for i := 0; i < n; i++ {
		s.Add()
	}

	return s.Size() - 1, nil
}

// Size returns the number of elements in the DisjointSet.
func (s *DisjointSet) Size() int {
	return len(s.parents)
}

// checkIdx checks if the given index is valid, returning an error if it is not.
func (s *DisjointSet) checkIdx(i int) error {
	if i >= len(s.parents) {
		return ErrMissingElement
	}

	return nil
}

// Find returns the root of the tree containing the given index.
func (s *DisjointSet) Find(i int) (int, error) {
	if err := s.checkIdx(i); err != nil {
		return NotFound, err
	}

	// Find the root of the tree from the given index.
	p := s.parents[i]
	for p != s.parents[p] {
		p = s.parents[p]
	}
	root := p

	// Compress the path for faster future lookups.
	// We search the same direction again to keep constant space complexity.
	p = s.parents[i]
	for p != root {
		n := s.parents[p]
		s.parents[p] = root
		p = n
	}

	return root, nil
}

// Union merges the trees containing the given indices.
func (s *DisjointSet) Union(i, j int) error {
	// No need to check index here; Find will cover it.
	root1, err := s.Find(i)
	if err != nil {
		return err
	}
	root2, err := s.Find(j)
	if err != nil {
		return err
	}

	if root1 != root2 {
		if s.ranks[root1] < s.ranks[root2] {
			s.parents[root1] = root2
		} else {
			s.parents[root2] = root1
			if s.ranks[root1] == s.ranks[root2] {
				s.ranks[root1]++
			}
		}
	}

	return nil
}
