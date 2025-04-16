package sparsegraph

import (
	"testing"
)

func TestConnect(t *testing.T) {
	testCases := []struct {
		name    string
		n       int
		i, j    int
		w       float64
		wantErr error
	}{
		{
			name:    "valid connection",
			n:       3,
			i:       0,
			j:       1,
			w:       1.5,
			wantErr: nil,
		},
		{
			name:    "invalid node index",
			n:       3,
			i:       -1,
			j:       1,
			w:       1.5,
			wantErr: ErrNoSuchNode,
		},
		{
			name:    "invalid node index 2",
			n:       3,
			i:       0,
			j:       3,
			w:       1.5,
			wantErr: ErrNoSuchNode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g, err := NewGraph(tc.n)
			if err != nil {
				t.Fatalf("NewGraph() failed: %v", err)
			}

			err = g.Connect(tc.i, tc.j, tc.w)
			if err != tc.wantErr {
				t.Errorf("Connect(%d, %d, %f) = %v; want error? %v", tc.i, tc.j, tc.w, err, tc.wantErr)
			}

			w, ok := g.GetEdge(tc.i, tc.j)
			if err == nil && (!ok || w != tc.w) {
				t.Errorf("GetEdge(%d, %d) = %f; want %f", tc.i, tc.j, w, tc.w)
			}

			_, ok = g.GetEdge(tc.j, tc.i)
			if err == nil && ok {
				t.Errorf("GetEdge(%d, %d) should not exist after Connect(%d, %d, %f)", tc.j, tc.i, tc.i, tc.j, tc.w)
			}
		})
	}
}

func TestConnectBidirectional(t *testing.T) {
	testCases := []struct {
		name    string
		n       int
		i, j    int
		w       float64
		wantErr error
	}{
		{
			name:    "valid bidirectional connection",
			n:       3,
			i:       0,
			j:       1,
			w:       1.5,
			wantErr: nil,
		},
		{
			name:    "invalid node index",
			n:       3,
			i:       -1,
			j:       1,
			w:       1.5,
			wantErr: ErrNoSuchNode,
		},
		{
			name:    "invalid node index 2",
			n:       3,
			i:       0,
			j:       3,
			w:       1.5,
			wantErr: ErrNoSuchNode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g, err := NewGraph(tc.n)
			if err != nil {
				t.Fatalf("NewGraph() failed: %v", err)
			}

			err = g.ConnectBidirectional(tc.i, tc.j, tc.w)
			if err != tc.wantErr {
				t.Errorf("ConnectBidirectional(%d, %d, %f) = %v; want error? %v", tc.i, tc.j, tc.w, err, tc.wantErr)
			}

			w, ok := g.GetEdge(tc.i, tc.j)
			if err == nil && (!ok || w != tc.w) {
				t.Errorf("GetEdge(%d, %d) = %f; want %f", tc.i, tc.j, w, tc.w)
			}

			w, ok = g.GetEdge(tc.j, tc.i)
			if err == nil && (!ok || w != tc.w) {
				t.Errorf("GetEdge(%d, %d) = %f; want %f", tc.j, tc.i, w, tc.w)
			}
		})
	}
}
