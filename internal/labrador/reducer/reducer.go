// Package reducer folds a completed run into report artifacts. Any number of
// reducers may run; each writes its own artifact, independently of the others.
package reducer

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"tsumegolang/internal/labrador/operation"
)

var (
	ErrUnknownReducer      = errors.New("unknown reducer")
	ErrDuplicateReducer    = errors.New("duplicate reducer")
	ErrIncompatibleReducer = errors.New("incompatible reducers")

	// ErrSkipped declines a run without failing it. Wrap it with the reason.
	ErrSkipped = errors.New("skipped")
)

// A Reducer writes one artifact. Reduce returns a one-line summary, or an error
// wrapping ErrSkipped to decline.
type Reducer struct {
	Name     string
	Artifact string // file written, relative to the output directory; must be unique across a run
	Reduce   func(records []operation.Record, out *Output) (string, error)
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

func Lookup(name string) (Reducer, error) {
	reducer, ok := registry[name]
	if !ok {
		return Reducer{}, fmt.Errorf("%w: %q (available: %s)", ErrUnknownReducer, name, strings.Join(Names(), ", "))
	}
	return reducer, nil
}

// Resolve turns flag names into a validated set, dropping blank entries.
func Resolve(names []string) ([]Reducer, error) {
	reducers := make([]Reducer, 0, len(names))

	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		reducer, err := Lookup(name)
		if err != nil {
			return nil, err
		}
		reducers = append(reducers, reducer)
	}

	if err := Validate(reducers); err != nil {
		return nil, err
	}
	return reducers, nil
}

// Validate reports whether a set of reducers can coexist. Callers run it before
// downloading, so an unrunnable set costs milliseconds rather than an operation.
func Validate(reducers []Reducer) error {
	names := make(map[string]int, len(reducers))
	artifacts := make(map[string]string, len(reducers))

	for i, reducer := range reducers {
		position := i + 1

		if previous, duplicate := names[reducer.Name]; duplicate {
			return fmt.Errorf("%w: %q appears at positions %d and %d; it would only overwrite its own artifact",
				ErrDuplicateReducer, reducer.Name, previous, position)
		}
		names[reducer.Name] = position

		if owner, taken := artifacts[reducer.Artifact]; taken {
			return fmt.Errorf("%w: %q and %q both write %q; pick one",
				ErrIncompatibleReducer, owner, reducer.Name, reducer.Artifact)
		}
		artifacts[reducer.Artifact] = reducer.Name
	}

	return nil
}

type Options struct {
	OutputDir string
	Source    string // file the records were loaded from; empty for a fresh download
}

type Report struct {
	Summaries []string // one line per reducer that wrote its artifact
	Skips     []string // one line per reducer that declined
}

// Run applies every reducer in order. A failure does not stop the rest; they
// are joined into one error. An ErrSkipped is recorded as a skip, not a failure.
func Run(reducers []Reducer, records []operation.Record, options Options) (Report, error) {
	out := &Output{dir: options.OutputDir, source: options.Source}

	var report Report
	var failures []error

	for _, reducer := range reducers {
		summary, err := reducer.Reduce(records, out)
		switch {
		case errors.Is(err, ErrSkipped):
			report.Skips = append(report.Skips, fmt.Sprintf("%s: %v", reducer.Name, err))
		case err != nil:
			failures = append(failures, fmt.Errorf("%s: %w", reducer.Name, err))
		default:
			report.Summaries = append(report.Summaries, summary)
		}
	}

	return report, errors.Join(failures...)
}
