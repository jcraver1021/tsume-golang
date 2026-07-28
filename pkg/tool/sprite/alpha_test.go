package sprite

import (
	"image/color"
	"testing"
)

func TestAlphaComposite(t *testing.T) {
	testCases := []struct {
		name     string
		src, dst color.RGBA
		want     color.RGBA
	}{
		{
			name: "opaque source over transparent destination",
			src:  color.RGBA{R: 255, G: 0, B: 0, A: 255},
			dst:  color.RGBA{R: 0, G: 0, B: 255, A: 0},
			want: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		},
		{
			name: "transparent source over opaque destination",
			src:  color.RGBA{R: 255, G: 0, B: 0, A: 0},
			dst:  color.RGBA{R: 0, G: 0, B: 255, A: 255},
			want: color.RGBA{R: 0, G: 0, B: 255, A: 255},
		},
		{
			name: "semi-transparent source over opaque destination",
			src:  color.RGBA{R: 255, G: 0, B: 0, A: 128},
			dst:  color.RGBA{R: 0, G: 0, B: 255, A: 255},
			want: color.RGBA{R: 128, G: 0, B: 126, A: 255},
		},
		{
			name: "semi-transparent source over transparent destination",
			src:  color.RGBA{R: 255, G: 0, B: 0, A: 128},
			dst:  color.RGBA{R: 0, G: 0, B: 255, A: 0},
			want: color.RGBA{R: 255, G: 0, B: 0, A: 128},
		},
	}

	for _, tc := range testCases {
		got := alphaComposite(tc.src, tc.dst)
		if got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}
