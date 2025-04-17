package maze

import (
	"testing"
)

func TestRectangle(t *testing.T) {
	testCases := []struct {
		name      string
		width     int
		height    int
		randomize bool
	}{
		{
			name:      "small square",
			width:     2,
			height:    2,
			randomize: false,
		},
		{
			name:      "large square",
			width:     10,
			height:    10,
			randomize: false,
		},
		{
			name:      "tall rectangle",
			width:     3,
			height:    15,
			randomize: false,
		},
		{
			name:      "wide rectangle",
			width:     15,
			height:    3,
			randomize: false,
		},
		{
			name:      "tall line",
			width:     1,
			height:    100,
			randomize: false,
		},
		{
			name:      "wide line",
			width:     100,
			height:    1,
			randomize: false,
		},
		{
			name:      "randomized small square",
			width:     2,
			height:    2,
			randomize: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g, err := Rectangle(tc.width, tc.height, tc.randomize)
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
							e, ok := g.GetEdge(i, j)
							if (x1 == x2 && (y1 == y2-1 || y1 == y2+1)) || // North or South neighbor
								(y1 == y2 && (x1 == x2-1 || x1 == x2+1)) { // East or West neighbor
								if !ok {
									t.Errorf("expected edge between %d and %d to exist", i, j)
								} else {
									w := e.Weight
									if tc.randomize && (w < 0.0 || w >= 1.0) {
										t.Errorf("expected weight of edge between %d and %d to be in [0.0, 1.0), got %f", i, j, w)
									}
									if !tc.randomize && w != DefaultWeight {
										t.Errorf("expected weight of edge between %d and %d to be %f, got %f", i, j, DefaultWeight, w)
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
