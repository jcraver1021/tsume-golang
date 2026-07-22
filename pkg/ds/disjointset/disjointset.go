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
	c       int
	s       int
	parents []int       // each element's value is the index of its parent
	ranks   map[int]int // rank of each root, used for union by rank
}

func WithCapacity(c int) func(*DisjointSet) {
	return func(ds *DisjointSet) {
		ds.c = c
	}
}

func WithScaleFactor(s int) func(*DisjointSet) {
	return func(ds *DisjointSet) {
		ds.s = s
	}
}

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

func (s *DisjointSet) isValid() bool {
	return s.c > 0 && s.s > 1
}

func (s *DisjointSet) init() {
	s.parents = make([]int, 0, s.c)
	s.ranks = make(map[int]int, s.c)
}

func (s *DisjointSet) ensureIdx() int {
	n := len(s.parents)
	if n >= cap(s.parents) {
		newParents := make([]int, 0, cap(s.parents)*s.s)
		copy(newParents, s.parents)
		s.parents = newParents
	}

	return int(n)
}

func (s *DisjointSet) Add() int {
	i := s.ensureIdx()
	s.parents = append(s.parents, i)
	s.ranks[i] = 0

	return i
}

func (s *DisjointSet) AddMany(n int) (int, error) {
	if n <= 0 {
		return NotFound, ErrInvalidRequest
	}

	for i := 0; i < n; i++ {
		s.Add()
	}

	return s.Size() - 1, nil
}

func (s *DisjointSet) Size() int {
	return len(s.parents)
}

func (s *DisjointSet) Capacity() int {
	return cap(s.parents)
}

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
			delete(s.ranks, root1)
		} else {
			s.parents[root2] = root1
			delete(s.ranks, root2)
			// Rank is only incremented if the trees are of equal rank.
			if s.ranks[root1] == s.ranks[root2] {
				s.ranks[root1]++
			}
		}
	}

	return nil
}
