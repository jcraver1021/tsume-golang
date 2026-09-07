// Package reducer defines the aggregate artifacts labrador can produce from a
// completed run. Any number of reducers may run; each folds the same records
// into its own artifact, independently of the others.
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

	// ErrSkipped lets a reducer decline the run rather than fail it. Wrap it
	// with the reason: fmt.Errorf("%w: index.md is newer than this run",
	// ErrSkipped).
	ErrSkipped = errors.New("skipped")
)

type Reducer struct {
	Name string
	// Artifact is the file this reducer writes, relative to the output
	// directory. Two reducers claiming the same artifact are incompatible, and
	// Validate rejects the pair. The package tests assert that each reducer
	// writes exactly what it declares here, which is what makes that check
	// sound rather than advisory.
	Artifact string
	// Reduce folds every record of the run into its artifact and returns a
	// one-line summary for the operator. Returning an error wrapping ErrSkipped
	// declines the run without failing it.
	Reduce func(records []operation.Record, out *Output) (string, error)
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

// Resolve turns flag names into the set of reducers a run will apply, rejecting
// a set that cannot coexist. Blank entries are dropped, so an empty -reduce flag
// means no aggregate artifact.
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

// Validate reports whether a set of reducers can run together. It is called
// before any download starts, so an unrunnable set costs nothing but the
// milliseconds spent finding out.
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

// Options is the environment a reduce phase runs in.
type Options struct {
	OutputDir string
	// Source is the file the records were loaded from, empty for a fresh
	// download run.
	Source string
}

// Report is what one reduce phase produced: a line for each reducer that wrote
// its artifact, and a line for each that declined.
type Report struct {
	Summaries []string
	Skips     []string
}

// Run applies every reducer to the same records, in order. A failing reducer
// does not stop the rest: Run returns what the others produced alongside every
// failure joined into one error. A reducer that declines via ErrSkipped is
// recorded as a skip, not a failure.
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
