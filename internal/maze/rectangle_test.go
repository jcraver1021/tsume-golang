package maze_test

import (
	"testing"

	. "tsumegolang/internal/maze"
)

func TestRectangle(t *testing.T) {
	testCases := []struct {
		name   string
		width  int
		height int
		init   EdgeInit
	}{
		{
			name:   "small square",
			width:  2,
			height: 2,
			init:   ConnectConst,
		},
		{
			name:   "large square",
			width:  10,
			height: 10,
			init:   ConnectConst,
		},
		{
			name:   "tall rectangle",
			width:  3,
			height: 15,
			init:   ConnectConst,
		},
		{
			name:   "wide rectangle",
			width:  15,
			height: 3,
			init:   ConnectConst,
		},
		{
			name:   "tall line",
			width:  1,
			height: 100,
			init:   ConnectConst,
		},
		{
			name:   "wide line",
			width:  100,
			height: 1,
			init:   ConnectConst,
		},
		{
			name:   "randomized small square",
			width:  2,
			height: 2,
			init:   ConnectRandom,
		},
		{
			name:   "randomized large square",
			width:  10,
			height: 10,
			init:   ConnectRandom,
		},
		{
			name:   "disconnected small square",
			width:  2,
			height: 2,
			init:   NoConnect,
		},
		{
			name:   "disconnected large square",
			width:  10,
			height: 10,
			init:   NoConnect,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewRectangle(tc.width, tc.height, tc.init)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// We will check every pair of vertices to ensure they are connected correctly.
			// For a rectangle, each vertex should connect to its north, east, south, and west neighbors,
			// if they exist, and not to any other vertices.
			for y1 := 0; y1 < tc.height; y1++ {
				for x1 := 0; x1 < tc.width; x1++ {
					for y2 := 0; y2 < tc.height; y2++ {
						for x2 := 0; x2 < tc.width; x2++ {
							i := y1*tc.width + x1
							j := y2*tc.width + x2
							e, ok := r.Graph.GetEdge(i, j)
							if (x1 == x2 && (y1 == y2-1 || y1 == y2+1)) || // North or South neighbor
								(y1 == y2 && (x1 == x2-1 || x1 == x2+1)) { // East or West neighbor
								w := e.Weight
								switch tc.init {
								case NoConnect:
									if ok {
										t.Errorf("expected no edge between %d and %d", i, j)
									}
								case ConnectConst:
									if !ok {
										t.Errorf("expected edge between %d and %d to exist", i, j)
									}
									if w != DefaultWeight {
										t.Errorf("expected weight of edge between %d and %d to be %f, got %f", i, j, DefaultWeight, w)
									}
								case ConnectRandom:
									if !ok {
										t.Errorf("expected edge between %d and %d to exist", i, j)
									}
									if w < 0.0 || w >= 1.0 {
										t.Errorf("expected weight of edge between %d and %d to be in [0.0, 1.0), got %f", i, j, w)
									}
								}
							} else {
								if ok {
									t.Errorf("expected no edge between %d and %d", i, j)
								}
							}
						}
					}
				}
			}
		})
	}
}

func TestRectToGraph(t *testing.T) {
	testCases := []struct {
		name string
		i    int
		j    int
		want int
	}{
		{
			name: "0,0",
			i:    0,
			j:    0,
			want: 0,
		},
		{
			name: "9,9",
			i:    9,
			j:    9,
			want: 99,
		},
		{
			name: "2,4",
			i:    2,
			j:    4,
			want: 24,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewRectangle(10, 5, NoConnect)
			if err != nil {
				t.Fatalf("(RectToGraph) unexpected error: %v", err)
			}

			got := r.RectToGraph(tc.i, tc.j)
			if tc.want != got {
				t.Errorf("(RectToGraph) want %d, got %d", tc.want, got)
			}
		})
	}
}
