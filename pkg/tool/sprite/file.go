package sprite

import (
	"fmt"
	"image/color"
	"os"
	"regexp"
	"strconv"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

var (
	hexColorPattern          = regexp.MustCompile(`^#([0-9a-fA-F]{8})$`)
	ErrInvalidHexColorFormat = fmt.Errorf("invalid hex color format: must be #RRGGBBAA")
)

type hexColor color.RGBA

func (h *hexColor) UnmarshalYAML(unmarshal func(any) error) error {
	var hexStr string
	if err := unmarshal(&hexStr); err != nil {
		return err
	}
	if !hexColorPattern.MatchString(hexStr) {
		return fmt.Errorf("%w: %s", ErrInvalidHexColorFormat, hexStr)
	}
	r, _ := strconv.ParseUint(hexStr[1:3], 16, 8)
	g, _ := strconv.ParseUint(hexStr[3:5], 16, 8)
	b, _ := strconv.ParseUint(hexStr[5:7], 16, 8)
	a, _ := strconv.ParseUint(hexStr[7:9], 16, 8)
	*h = hexColor{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
	return nil
}

type spriteFileDef struct {
	Matrix             []string                     `yaml:"matrix"`
	ColorCodes         map[string]hexColor          `yaml:"color_codes"`
	AnimationSequences map[string]*animationFileDef `yaml:"animation_sequences"`
}

type animationFileDef struct {
	Frames        string `yaml:"frames"`
	FrameDuration int    `yaml:"frame_duration"`
}

// SpriteFromFile reads a YAML sprite definition from path and returns the
// constructed Sprite. It is a thin wrapper around SpriteFromBytes.
func SpriteFromFile(path string) (*Sprite, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("sprite: reading %q: %w", path, err)
	}
	return SpriteFromBytes(data)
}

// SpriteFromBytes parses a YAML sprite definition and returns the constructed Sprite.
//
//	matrix:
//	  - '0g0'
//	  - 'g1g'
//	  - '0g0'
//	color_codes:
//	  '0': '#00000000'   # transparent
//	  '1': '#FF0000FF'   # red, static center
//	  'a': '#FF4400FF'   # glow frame 1
//	  'b': '#FF2200FF'   # glow frame 2
//	animation_sequences:
//	  'g':
//	    frames: 'aabba'
//	    frame_duration: 8
//
// Rules enforced during parsing:
//   - color_codes keys must be single ASCII characters.
//   - animation_sequences keys must be single ASCII characters and must not
//     appear in color_codes.
//   - Every key referenced in frames must exist in color_codes.
//   - Every key in matrix must exist in color_codes or animation_sequences.
//   - frame_duration must be positive.
func SpriteFromBytes(data []byte) (*Sprite, error) {
	var def spriteFileDef
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("sprite: parsing YAML: %w", err)
	}
	return buildSpriteFromDef(def)
}

func buildSpriteFromDef(def spriteFileDef) (*Sprite, error) {
	palette := NewPalette()
	for keyStr, hc := range def.ColorCodes {
		ck, err := fromString(keyStr)
		if err != nil {
			return nil, fmt.Errorf("sprite: color_codes key %q: %w", keyStr, err)
		}
		if err := palette.setKeyed(ck, color.RGBA(hc)); err != nil {
			return nil, fmt.Errorf("sprite: color_codes[%q]: %w", keyStr, err)
		}
	}

	animationSequences := map[ColorKey]*AnimationSequence{}
	for keyStr, seqDef := range def.AnimationSequences {
		ck, err := fromString(keyStr)
		if err != nil {
			return nil, fmt.Errorf("sprite: animation_sequences key %q: %w", keyStr, err)
		}
		if _, exists := def.ColorCodes[keyStr]; exists {
			return nil, fmt.Errorf("sprite: animation_sequences key %q collides with a color_codes key", keyStr)
		}
		if seqDef == nil || len(seqDef.Frames) == 0 {
			return nil, fmt.Errorf("sprite: animation_sequences[%q]: %w", keyStr, ErrEmptyFrames)
		}

		// Reserve the key so nextKey never assigns it to a color.
		reservedCk, err := palette.Reserve(ck)
		if err != nil {
			return nil, fmt.Errorf("sprite: animation_sequences key %q: %w", keyStr, err)
		}
		if reservedCk != ck {
			return nil, fmt.Errorf("sprite: animation_sequences key %q: key collision in palette", keyStr)
		}

		frames := make([]ColorKey, len(seqDef.Frames))
		for i, r := range seqDef.Frames {
			frameCk, err := fromString(string(r))
			if err != nil {
				return nil, fmt.Errorf("sprite: animation_sequences[%q] frame %d: %w", keyStr, i, err)
			}
			if _, exists := def.ColorCodes[string(r)]; !exists {
				return nil, fmt.Errorf("sprite: animation_sequences[%q] frame %d: key %q not found in color_codes", keyStr, i, string(r))
			}
			frames[i] = frameCk
		}

		seq, err := NewAnimationSequence(palette, frames, seqDef.FrameDuration)
		if err != nil {
			return nil, fmt.Errorf("sprite: animation_sequences[%q]: %w", keyStr, err)
		}
		animationSequences[ck] = seq
	}

	if len(def.Matrix) == 0 {
		return nil, ErrInvalidSpriteMatrix
	}
	matrix := make([][]ColorKey, len(def.Matrix))
	for i, rowStr := range def.Matrix {
		matrix[i] = make([]ColorKey, utf8.RuneCountInString(rowStr))
		j := 0
		for _, r := range rowStr {
			ck, err := fromString(string(r))
			if err != nil {
				return nil, fmt.Errorf("sprite: matrix[%d][%d]: %w", i, j, err)
			}
			_, inColors := def.ColorCodes[string(r)]
			_, inAnims := def.AnimationSequences[string(r)]
			if !inColors && !inAnims {
				return nil, fmt.Errorf("sprite: matrix[%d][%d]: key %q not defined in color_codes or animation_sequences", i, j, string(r))
			}
			matrix[i][j] = ck
			j++
		}
	}

	return NewSprite(matrix, palette, animationSequences)
}
