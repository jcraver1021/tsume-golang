package sprite_test

import (
	"image/color"
	"testing"

	. "tsumegolang/pkg/tool/sprite"
)

func TestPalette(t *testing.T) {
	testCases := []struct {
		name      string
		input     color.RGBA
		wantAdded bool
	}{
		{
			name:      "add new color",
			input:     color.RGBA{R: 255, G: 0, B: 0, A: 255},
			wantAdded: true,
		},
		{
			name:      "add existing color",
			input:     color.RGBA{R: 255, G: 0, B: 0, A: 255},
			wantAdded: false,
		},
	}

	palette := NewPalette()
	for _, tc := range testCases {
		key, gotAdded := palette.Add(tc.input)

		if gotAdded != tc.wantAdded {
			t.Errorf("%s: gotAdded %v, want %v", tc.name, gotAdded, tc.wantAdded)
		}

		if got, ok := palette.Get(key); !ok || got != tc.input {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.input)
		}

		// Regardless, if we add the same color again, it should return the same key
		if sameKey, _ := palette.Add(tc.input); sameKey != key {
			t.Errorf("%s: expected same key %v, got %v", tc.name, key, sameKey)
		}
	}
}

func TestPaletteReserve(t *testing.T) {
	testCases := []struct {
		name        string
		preReserved []ColorKey
		input       ColorKey
		wantKey     ColorKey
		wantErr     bool
	}{
		{
			name:        "reserve new key",
			preReserved: []ColorKey{},
			input:       ColorKey("A"),
			wantKey:     ColorKey("A"),
			wantErr:     false,
		},
		{
			name:        "reserve already reserved key",
			preReserved: []ColorKey{ColorKey("A")},
			input:       ColorKey("A"),
			wantKey:     ColorKey("1"), // next available key
			wantErr:     false,
		},
		{
			name:        "reserve key in use",
			preReserved: []ColorKey{ColorKey("A")},
			input:       ColorKey("0"),
			wantKey:     ColorKey("1"), // next available key
			wantErr:     false,
		},
		{
			name:        "reserve invalid key",
			preReserved: []ColorKey{},
			input:       ColorKey("AB"),
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		palette := NewPalette()
		for _, pre := range tc.preReserved {
			palette.Reserve(pre)
		}
		palette.Add(color.RGBA{0, 0, 0, 0}) // use "0"

		gotKey, err := palette.Reserve(tc.input)
		if (err != nil) != tc.wantErr {
			t.Errorf("%s: got error %v, wantErr %v", tc.name, err, tc.wantErr)
		}
		if gotKey != tc.wantKey && !tc.wantErr {
			t.Errorf("%s: got key %v, want %v", tc.name, gotKey, tc.wantKey)
		}
	}
}
