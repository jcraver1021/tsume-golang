package sparsegraph

type Graph struct {
	nodes []map[int]float64 // Maps node index to its adjacency list, with the value being the weight of the edge.
}

func NewGraph(n int) (*Graph, error) {
	if n <= 0 {
		return nil, ErrInvalidConfiguration
	}

	return &Graph{
		nodes: make([]map[int]float64, n),
	}, nil
}

func (g *Graph) checkIdx(i int) error {
	if i < 0 || i >= len(g.nodes) {
		return ErrNoSuchNode
	}

	return nil
}

func (g *Graph) Connect(i, j int, w float64) error {
	if err := g.checkIdx(i); err != nil {
		return err
	}
	if err := g.checkIdx(j); err != nil {
		return err
	}

	if g.nodes[i] == nil {
		g.nodes[i] = make(map[int]float64)
	}

	g.nodes[i][j] = w

	return nil
}

func (g *Graph) ConnectBidirectional(i, j int, w float64) error {
	if err := g.Connect(i, j, w); err != nil {
		return err
	}
	return g.Connect(j, i, w)
}

func (g *Graph) GetEdge(i, j int) (float64, bool) {
	if err := g.checkIdx(i); err != nil {
		return 0, false
	}
	if err := g.checkIdx(j); err != nil {
		return 0, false
	}

	if g.nodes[i] == nil {
		return 0, false
	}

	w, exists := g.nodes[i][j]
	return w, exists
}

func (g *Graph) GetAllEdges() [][3]float64 {
	edges := [][3]float64{}
	for i, adj := range g.nodes {
		if adj == nil {
			continue
		}
		for j, w := range adj {
			// Put the edge first to enable lexicographical ordering.
			edges = append(edges, [3]float64{w, float64(i), float64(j)})
		}
	}
	return edges
}
