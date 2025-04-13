package sparsegraph

type Graph struct {
	weights map[uint64]map[uint64]float64
}

func NewGraph() *Graph {
	return &Graph{
		weights: make(map[uint64]map[uint64]float64),
	}
}

func (g *Graph) Connect(i, j uint64, w float64) {
	iAdj, ok := g.weights[i]
	if !ok {
		g.weights[i] = make(map[uint64]float64)
		iAdj = g.weights[i]
	}

	iAdj[j] = w
}

func (g *Graph) GetEdge(i, j uint64) (float64, bool) {
	iAdj, ok := g.weights[i]
	if !ok {
		return 0, false
	}

	w, ok := iAdj[j]
	if !ok {
		return 0, false
	}

	return w, true
}
