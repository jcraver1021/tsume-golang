package sprite_test

import (
	"errors"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	. "tsumegolang/pkg/tool/sprite"
)

func TestSpriteFromBytes(t *testing.T) {
	t.Run("ValidStaticSprite", func(t *testing.T) {
		yaml := `
matrix:
  - '001'
  - '010'
  - '001'
color_codes:
  '0': '#00000000'
  '1': '#FF0000FF'
`
		s, err := SpriteFromBytes([]byte(yaml))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.Width() != 3 || s.Height() != 3 {
			t.Fatalf("expected 3×3, got %d×%d", s.Width(), s.Height())
		}
		pixels := s.Render()
		red := color.RGBA{R: 255, A: 255}
		transparent := color.RGBA{}
		if pixels[0][0] != transparent {
			t.Errorf("[0][0]: want transparent, got %v", pixels[0][0])
		}
		if pixels[0][2] != red {
			t.Errorf("[0][2]: want red, got %v", pixels[0][2])
		}
		if pixels[1][1] != red {
			t.Errorf("[1][1]: want red, got %v", pixels[1][1])
		}
	})

	t.Run("ValidAnimatedSprite", func(t *testing.T) {
		yaml := `
matrix:
  - 'g'
color_codes:
  'a': '#FF0000FF'
  'b': '#0000FFFF'
animation_sequences:
  'g':
    frames: 'ab'
    frame_duration: 1
`
		s, err := SpriteFromBytes([]byte(yaml))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		red := color.RGBA{R: 255, A: 255}
		blue := color.RGBA{B: 255, A: 255}

		if got := s.Render()[0][0]; got != red {
			t.Errorf("frame 0: want red, got %v", got)
		}
		s.Advance()
		if got := s.Render()[0][0]; got != blue {
			t.Errorf("frame 1: want blue, got %v", got)
		}
		s.Advance()
		if got := s.Render()[0][0]; got != red {
			t.Errorf("frame 2 (wrapped): want red, got %v", got)
		}
	})

	t.Run("InvalidYAML", func(t *testing.T) {
		_, err := SpriteFromBytes([]byte(":::"))
		if err == nil {
			t.Fatal("expected error for invalid YAML, got nil")
		}
	})

	t.Run("InvalidHexColor", func(t *testing.T) {
		yaml := `
matrix:
  - '1'
color_codes:
  '1': 'notacolor'
`
		_, err := SpriteFromBytes([]byte(yaml))
		if err == nil {
			t.Fatal("expected error for bad hex color, got nil")
		}
		if !errors.Is(err, ErrInvalidHexColorFormat) {
			t.Errorf("want ErrInvalidHexColorFormat in chain, got: %v", err)
		}
	})

	t.Run("AnimationKeyCollidesWithColorKey", func(t *testing.T) {
		yaml := `
matrix:
  - '1'
color_codes:
  '1': '#FF0000FF'
animation_sequences:
  '1':
    frames: '1'
    frame_duration: 1
`
		_, err := SpriteFromBytes([]byte(yaml))
		if err == nil {
			t.Fatal("expected error for key collision, got nil")
		}
	})

	t.Run("FrameKeyNotInColorCodes", func(t *testing.T) {
		yaml := `
matrix:
  - 'g'
color_codes:
  'a': '#FF0000FF'
animation_sequences:
  'g':
    frames: 'az'
    frame_duration: 1
`
		_, err := SpriteFromBytes([]byte(yaml))
		if err == nil {
			t.Fatal("expected error for frame key not in color_codes, got nil")
		}
	})

	t.Run("DanglingMatrixKey", func(t *testing.T) {
		yaml := `
matrix:
  - '1x'
color_codes:
  '1': '#FF0000FF'
`
		_, err := SpriteFromBytes([]byte(yaml))
		if err == nil {
			t.Fatal("expected error for matrix key not in color_codes or animation_sequences, got nil")
		}
	})

	t.Run("EmptyMatrix", func(t *testing.T) {
		yaml := `
matrix: []
color_codes:
  '0': '#00000000'
`
		_, err := SpriteFromBytes([]byte(yaml))
		if !errors.Is(err, ErrInvalidSpriteMatrix) {
			t.Errorf("want ErrInvalidSpriteMatrix, got: %v", err)
		}
	})

	t.Run("ZeroFrameDuration", func(t *testing.T) {
		yaml := `
matrix:
  - 'g'
color_codes:
  'a': '#FF0000FF'
animation_sequences:
  'g':
    frames: 'a'
    frame_duration: 0
`
		_, err := SpriteFromBytes([]byte(yaml))
		if !errors.Is(err, ErrInvalidFrameDuration) {
			t.Errorf("want ErrInvalidFrameDuration, got: %v", err)
		}
	})
}

func TestSpriteFromFile(t *testing.T) {
	t.Run("ValidFile", func(t *testing.T) {
		yaml := `
matrix:
  - '01'
color_codes:
  '0': '#00000000'
  '1': '#FFFFFFFF'
`
		path := filepath.Join(t.TempDir(), "test.yaml")
		if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
			t.Fatalf("writing temp file: %v", err)
		}
		s, err := SpriteFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.Width() != 2 || s.Height() != 1 {
			t.Errorf("expected 2×1, got %d×%d", s.Width(), s.Height())
		}
	})

	t.Run("FileNotFound", func(t *testing.T) {
		_, err := SpriteFromFile("/does/not/exist.yaml")
		if err == nil {
			t.Fatal("expected error for missing file, got nil")
		}
	})
}
