// Package reducer defines the aggregate artifacts labrador can produce from a
// completed run. Exactly one reducer runs, after every download has settled.
package reducer

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"tsumegolang/internal/labrador/operation"
)

var (
	ErrUnknownReducer = errors.New("unknown reducer")
	ErrCreateDir      = errors.New("failed to create output directory")
)

type Reducer struct {
	Name string
	// Reduce folds every record of the run into a single artifact under
	// outputDir and returns a one-line summary for the operator.
	Reduce func(records []operation.Record, outputDir string) (string, error)
}

var registry = map[string]Reducer{
	markdownIndex.Name: markdownIndex,
	manifestJSON.Name:  manifestJSON,
}

func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Lookup resolves a flag value into the single trailing reducer. A blank name
// reports false, meaning the run produces no aggregate artifact.
func Lookup(name string) (Reducer, bool, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Reducer{}, false, nil
	}

	reducer, ok := registry[name]
	if !ok {
		return Reducer{}, false, fmt.Errorf("%w: %q (available: %s)", ErrUnknownReducer, name, strings.Join(Names(), ", "))
	}
	return reducer, true, nil
}

// ensureOutputDir matters when every download failed: nothing created the
// directory, but the run still owes the operator a report.
func ensureOutputDir(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("%w: %w", ErrCreateDir, err)
	}
	return nil
}
