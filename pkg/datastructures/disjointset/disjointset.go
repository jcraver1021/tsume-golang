package disjointset

type dsNode[T comparable] struct {
	parent T
	rank   int
}

type DisjointSet[T comparable] struct {
	forest map[T]*dsNode[T]
}

func New[T comparable]() *DisjointSet[T] {
	return &DisjointSet[T]{forest: make(map[T]*dsNode[T])}
}

func (ds *DisjointSet[T]) Add(e T) {
	if _, ok := ds.forest[e]; ok {
		return
	}
	ds.forest[e] = &dsNode[T]{parent: e, rank: 0}
}

func (ds *DisjointSet[T]) Find(e T) (T, error) {
	if _, ok := ds.forest[e]; !ok {
		// This gets the zero value of T.
		// If T is a pointer, this will be nil.
		var t T
		return t, ErrMissingElement
	}

	// Find the root (first element who is its own parent).
	root := e
	for root != ds.forest[root].parent {
		root = ds.forest[root].parent
	}

	// Compress the path for faster future lookups.
	// We search the same direction again to keep constant space complexity.
	for e != root {
		parent := ds.forest[e].parent
		ds.forest[e].parent = root
		e = parent
	}

	return root, nil
}

func (ds *DisjointSet[T]) Union(e1, e2 T) error {
	root1, err := ds.Find(e1)
	if err != nil {
		return err
	}
	root2, err := ds.Find(e2)
	if err != nil {
		return err
	}

	// If they are already in the same set, do nothing.
	if root1 == root2 {
		return nil
	}

	node1 := ds.forest[root1]
	node2 := ds.forest[root2]

	if node1.rank < node2.rank {
		node1.parent = root2
	} else if node1.rank > node2.rank {
		node2.parent = root1
	} else {
		node2.parent = root1
		node1.rank++
	}

	return nil
}
