// Package config reads the YAML describing an operation. Section names double
// as directory paths, so the document's shape is the output tree's shape.
package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	ErrCantOpenFile = errors.New("failed to open file")
	ErrParseYAML    = errors.New("failed to parse YAML")
)

type Section struct {
	Name string // may contain "/" to nest directories
	URLs []string
}

func isValidURL(url string) bool {
	if url == "" {
		return false
	}
	if len(url) >= 7 && url[:7] == "http://" {
		return true
	}
	if len(url) >= 8 && url[:8] == "https://" {
		return true
	}
	return false
}

// Load reads a YAML document, dropping non-HTTP(S) URLs and empty sections.
func Load(filename string) ([]Section, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCantOpenFile, err)
	}
	defer file.Close()

	var sectionsMap map[string][]string
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&sectionsMap); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParseYAML, err)
	}

	sections := make([]Section, 0, len(sectionsMap))
	for name, urls := range sectionsMap {
		validURLs := make([]string, 0, len(urls))
		for _, url := range urls {
			if isValidURL(url) {
				validURLs = append(validURLs, url)
			}
		}
		if len(validURLs) > 0 {
			sections = append(sections, Section{
				Name: name,
				URLs: validURLs,
			})
		}
	}

	return sections, nil
}
