package labrador

import (
	"bufio"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	ErrCantOpenFile = fmt.Errorf("failed to open file")
	ErrParseFile    = fmt.Errorf("failed to parse file")
	ErrParseYAML    = fmt.Errorf("failed to parse YAML")
)

type Section struct {
	Name string
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

func ParseURLsFromTextFile(filename string) ([]string, error) {
	urls := []string{}

	// Open the text file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCantOpenFile, err)
	}
	defer file.Close()

	// Read all lines from the text file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if isValidURL(line) {
			urls = append(urls, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParseFile, err)
	}

	return urls, nil
}

func ParseSectionsFromYAML(filename string) ([]Section, error) {
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
