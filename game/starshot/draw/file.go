package draw

import (
	"fmt"
	"image/color"
	"os"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"
)

var (
	hexColorPattern          = regexp.MustCompile(`^#([0-9a-fA-F]{8})$`)
	ErrInvalidHexColorFormat = fmt.Errorf("invalid hex color format")
)

// hexColor is a wrapper around color.RGBA to facilitate YAML unmarshaling from hex strings
type hexColor color.RGBA

func (h *hexColor) UnmarshalYAML(unmarshal func(any) error) error {
	var hexStr string
	if err := unmarshal(&hexStr); err != nil {
		return err
	}

	if !hexColorPattern.MatchString(hexStr) {
		return fmt.Errorf("%w: %s", ErrInvalidHexColorFormat, hexStr)
	}

	var r, g, b, a uint64
	r, _ = strconv.ParseUint(hexStr[1:3], 16, 8)
	g, _ = strconv.ParseUint(hexStr[3:5], 16, 8)
	b, _ = strconv.ParseUint(hexStr[5:7], 16, 8)
	a, _ = strconv.ParseUint(hexStr[7:9], 16, 8)

	*h = hexColor{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
	return nil
}

// colorMatrixYAML is the intermediate structure for unmarshaling from YAML
type colorMatrixYAML struct {
	Matrix             []string                          `yaml:"matrix"`
	ColorCodesHex      map[string]hexColor               `yaml:"color_codes"`
	AnimationSequences map[string]*animationSequenceYAML `yaml:"animation_sequences"`
}

// animationSequenceYAML is the intermediate structure for animation sequences
type animationSequenceYAML struct {
	FramesStr     string `yaml:"frames"`
	FrameDuration int    `yaml:"frame_duration"`
}

func ColorMatrixFromFile(path string) (*ColorMatrix, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ColorMatrixFromBytes(data)
}

func ColorMatrixFromBytes(data []byte) (*ColorMatrix, error) {
	var yamlData colorMatrixYAML
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Parse color codes
	colorCodes := make(ColorMap)
	for keyStr, hexColor := range yamlData.ColorCodesHex {
		key, err := fromString(keyStr)
		if err != nil {
			return nil, fmt.Errorf("invalid color key %q: %w", keyStr, err)
		}
		colorCodes[key] = color.RGBA(hexColor)
	}

	// Parse matrix rows (each row is a string of color keys)
	matrix := make([][]ColorKey, len(yamlData.Matrix))
	for i, rowStr := range yamlData.Matrix {
		matrix[i] = make([]ColorKey, len(rowStr))
		for j, char := range rowStr {
			key, err := fromString(string(char))
			if err != nil {
				return nil, fmt.Errorf("invalid color key at matrix[%d][%d]: %w", i, j, err)
			}
			matrix[i][j] = key
		}
	}

	// Parse animation sequences
	animationSequences := make(map[ColorKey]*AnimationSequence)
	for keyStr, animSeqYAML := range yamlData.AnimationSequences {
		key, err := fromString(keyStr)
		if err != nil {
			return nil, fmt.Errorf("invalid animation sequence key %q: %w", keyStr, err)
		}

		// Build the color map for this animation from the frames string
		animColorMap := make(ColorMap)
		frames := make([]ColorKey, len(animSeqYAML.FramesStr))
		for i, char := range animSeqYAML.FramesStr {
			frameKey, err := fromString(string(char))
			if err != nil {
				return nil, fmt.Errorf("invalid frame key in animation %q at position %d: %w", keyStr, i, err)
			}
			frames[i] = frameKey

			// Add to the animation's color map if not already present
			if _, exists := animColorMap[frameKey]; !exists {
				// Check if the color is in the main color codes
				if colorValue, exists := colorCodes[frameKey]; exists {
					animColorMap[frameKey] = colorValue
				} else {
					return nil, fmt.Errorf("animation frame key %q not found in color_codes", string(frameKey))
				}
			}
		}

		animationSequences[key] = NewAnimationSequence(&animColorMap, frames, animSeqYAML.FrameDuration)
	}

	return NewColorMatrix(matrix, &colorCodes, animationSequences)
}
