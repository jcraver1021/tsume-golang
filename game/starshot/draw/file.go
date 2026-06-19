package draw

import (
	"fmt"
	"image/color"
	"strconv"
	"regexp"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	hexColorPattern = regexp.MustCompile(`^#([0-9a-fA-F]{8})$`)
	ErrInvalidHexColorFormat = fmt.Errorf("invalid hex color format")
)

type hexColor color.RGBA


  func (h *hexColor) UnmarshalYAML(unmarshal func(interface{}) error) error {
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
	Matrix             [][]int                    `yaml:"matrix"`
	ColorCodesHex      map[int]hexColor           `yaml:"color_codes"`
	AnimationSequences map[int]*animationSequenceYAML `yaml:"animation_sequences"`
}

// animationSequenceYAML is the intermediate structure for animation sequences
type animationSequenceYAML struct {
	FramesHex     []hexColor `yaml:"frames"`
	FrameDuration int        `yaml:"frame_duration"`
}

func ColorMatrixFromFile(path string) (*ColorMatrix, error) {
	// Read the YAML file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal into intermediate structure with HexColor
	var yamlData colorMatrixYAML
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Convert ColorCodesHex (map[int]HexColor) to ColorCodes (map[int]color.RGBA)
	colorCodes := make(map[int]color.RGBA)
	for code, hexColor := range yamlData.ColorCodesHex {
		colorCodes[code] = color.RGBA(hexColor)
	}

	// Convert AnimationSequences from HexColor to color.RGBA
	animationSequences := make(map[int]*AnimationSequence)
	for code, animSeqYAML := range yamlData.AnimationSequences {
		// Convert HexColor frames to color.RGBA frames
		frames := make([]color.RGBA, len(animSeqYAML.FramesHex))
		for i, hexColor := range animSeqYAML.FramesHex {
			frames[i] = color.RGBA(hexColor)
		}

		animationSequences[code] = NewAnimationSequence(frames, animSeqYAML.FrameDuration)
	}

	// Create and return the ColorMatrix using the existing constructor
	return NewColorMatrix(yamlData.Matrix, colorCodes, animationSequences)
}
  