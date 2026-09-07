package reducer_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tsumegolang/internal/labrador/operation"
	. "tsumegolang/internal/labrador/reducer"
)

func testRecords() []operation.Record {
	return []operation.Record{
		{Section: "Beta", URL: "https://example.com/b", FilePath: "Beta/b.html", Success: true},
		{Section: "Alpha", URL: "https://example.com/a", FilePath: "Alpha/a.html", Success: true},
		{Section: "Alpha", URL: "https://example.com/gone", Success: false, Error: errors.New("404")},
	}
}

func mustResolve(t *testing.T, names ...string) []Reducer {
	t.Helper()

	reducers, err := Resolve(names)
	if err != nil {
		t.Fatalf("Resolve(%v) = %v", names, err)
	}
	return reducers
}

func failer(name, filename string, err error) Reducer {
	return Reducer{
		Name:     name,
		Artifact: filename,
		Reduce:   func([]operation.Record, *Output) (string, error) { return "", err },
	}
}

func TestRegistryInvariants(t *testing.T) {
	artifacts := map[string]string{}

	for _, name := range Names() {
		t.Run(name, func(t *testing.T) {
			reducer, err := Lookup(name)
			if err != nil {
				t.Fatalf("Lookup(%q) = %v", name, err)
			}
			if reducer.Name != name {
				t.Errorf("Name = %q, want %q — registry key must match", reducer.Name, name)
			}
			if reducer.Reduce == nil {
				t.Error("Reduce is nil")
			}
			if reducer.Artifact == "" {
				t.Error("Artifact is empty; collisions cannot be detected without it")
			}
			if owner, clash := artifacts[reducer.Artifact]; clash {
				t.Errorf("Artifact %q is also claimed by %q", reducer.Artifact, owner)
			}
			artifacts[reducer.Artifact] = name
		})
	}
}

func TestResolve(t *testing.T) {
	testCases := []struct {
		name            string
		names           []string
		want            []string
		wantErr         error
		wantErrContains []string
	}{
		{name: "empty flag yields no reducers", names: []string{""}, want: []string{}},
		{name: "whitespace yields no reducers", names: []string{"   "}, want: []string{}},
		{name: "single reducer", names: []string{"markdown-index"}, want: []string{"markdown-index"}},
		{
			name:  "both reducers, order preserved",
			names: []string{"manifest-json", "markdown-index"},
			want:  []string{"manifest-json", "markdown-index"},
		},
		{
			name:  "surrounding whitespace is tolerated",
			names: []string{" markdown-index ", "", " manifest-json"},
			want:  []string{"markdown-index", "manifest-json"},
		},
		{
			name:            "unknown name is rejected",
			names:           []string{"collate"},
			wantErr:         ErrUnknownReducer,
			wantErrContains: append([]string{`"collate"`}, Names()...),
		},
		{
			name:            "repeating a reducer is rejected",
			names:           []string{"markdown-index", "manifest-json", "markdown-index"},
			wantErr:         ErrDuplicateReducer,
			wantErrContains: []string{`"markdown-index"`, "positions 1 and 3"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Resolve(tc.names)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
				if got != nil {
					t.Error("reducers should be nil when resolution fails")
				}
				for _, want := range tc.wantErrContains {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("error %q does not contain %q", err, want)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("Resolve() = %v", err)
			}

			if len(got) != len(tc.want) {
				t.Fatalf("got %d reducers, want %d", len(got), len(tc.want))
			}
			for i, reducer := range got {
				if reducer.Name != tc.want[i] {
					t.Errorf("reducers[%d] = %q, want %q", i, reducer.Name, tc.want[i])
				}
			}
		})
	}
}

func TestRun(t *testing.T) {
	boom := errors.New("boom")
	bang := errors.New("bang")

	testCases := []struct {
		name          string
		reducers      []Reducer
		nested        bool // nothing has created the output directory yet
		wantSummaries int
		wantSkips     int
		wantSkipText  []string
		wantErrs      []error
		wantFiles     []string
	}{
		{
			name:          "every reducer writes its artifact",
			reducers:      mustResolve(t, "markdown-index", "manifest-json"),
			wantSummaries: 2,
			wantFiles:     []string{"index.md", "manifest.json"},
		},
		{
			name:          "the output directory is created if absent",
			reducers:      mustResolve(t, "markdown-index", "manifest-json"),
			nested:        true,
			wantSummaries: 2,
			wantFiles:     []string{"index.md", "manifest.json"},
		},
		{
			name: "no reducers write nothing",
		},
		{
			name: "a failure does not deny the others their artifacts",
			reducers: []Reducer{
				failer("first-failure", "a.txt", boom),
				writer("survivor", "b.txt", "written"),
				failer("second-failure", "c.txt", bang),
			},
			wantSummaries: 1,
			wantErrs:      []error{boom, bang},
			wantFiles:     []string{"b.txt"},
		},
		{
			name: "a skip is neither a summary nor a failure",
			reducers: []Reducer{
				writer("wrote-it", "a.txt", "content"),
				skipper("declined", "b.txt", "b.txt is newer than this run"),
				failer("broke", "c.txt", boom),
			},
			wantSummaries: 1,
			wantSkips:     1,
			wantSkipText:  []string{"declined", "b.txt is newer than this run"},
			wantErrs:      []error{boom},
			wantFiles:     []string{"a.txt"},
		},
		{
			name: "a run where everything declines is still a clean run",
			reducers: []Reducer{
				skipper("one", "a.txt", "nothing to do"),
				skipper("two", "b.txt", "nothing to do"),
			},
			wantSkips:    2,
			wantSkipText: []string{"one", "two"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outputDir := t.TempDir()
			if tc.nested {
				outputDir = filepath.Join(outputDir, "nested", "downloads")
			}

			report, err := Run(tc.reducers, testRecords(), Options{OutputDir: outputDir})

			if len(report.Summaries) != tc.wantSummaries {
				t.Errorf("summaries = %v, want %d", report.Summaries, tc.wantSummaries)
			}
			if len(report.Skips) != tc.wantSkips {
				t.Errorf("skips = %v, want %d", report.Skips, tc.wantSkips)
			}
			for _, want := range tc.wantSkipText {
				if !strings.Contains(strings.Join(report.Skips, "\n"), want) {
					t.Errorf("skips %v do not mention %q", report.Skips, want)
				}
			}

			for _, want := range tc.wantErrs {
				if !errors.Is(err, want) {
					t.Errorf("err = %v, want it to wrap %v", err, want)
				}
			}
			if len(tc.wantErrs) == 0 && err != nil {
				t.Errorf("err = %v, want nil", err)
			}

			entries, readErr := os.ReadDir(outputDir)
			if readErr != nil {
				t.Fatalf("reading output dir: %v", readErr)
			}
			got := make([]string, len(entries))
			for i, entry := range entries {
				got[i] = entry.Name()
			}
			if strings.Join(got, ",") != strings.Join(tc.wantFiles, ",") {
				t.Errorf("wrote %v, want %v", got, tc.wantFiles)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	testCases := []struct {
		name            string
		reducers        []Reducer
		wantErr         error
		wantErrContains []string
	}{
		{name: "no reducers", reducers: nil},
		{name: "single reducer", reducers: []Reducer{writer("a", "a.txt", "")}},
		{
			name:     "distinct names and artifacts",
			reducers: []Reducer{writer("a", "a.txt", ""), writer("b", "b.txt", "")},
		},
		{
			name:     "the same reducer twice",
			reducers: []Reducer{writer("a", "a.txt", ""), writer("a", "a.txt", "")},
			wantErr:  ErrDuplicateReducer,
		},
		{
			name:            "different reducers claiming one artifact",
			reducers:        []Reducer{writer("alpha", "same.txt", ""), writer("beta", "same.txt", "")},
			wantErr:         ErrIncompatibleReducer,
			wantErrContains: []string{`"alpha"`, `"beta"`, `"same.txt"`},
		},
		{
			name:     "a reducer that may decline still collides",
			reducers: []Reducer{writer("writes", "same.txt", ""), skipper("might-skip", "same.txt", "maybe")},
			wantErr:  ErrIncompatibleReducer,
		},
		{
			name: "a collision anywhere in the set is caught",
			reducers: []Reducer{
				writer("a", "a.txt", ""),
				writer("b", "b.txt", ""),
				writer("c", "a.txt", ""),
			},
			wantErr: ErrIncompatibleReducer,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.reducers)
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			for _, want := range tc.wantErrContains {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not contain %q", err, want)
				}
			}
		})
	}
}

// Validate is only sound if every reducer writes exactly what it declares, so
// pin that for the whole registry.
func TestRegistryReducersWriteExactlyTheirArtifact(t *testing.T) {
	for _, name := range Names() {
		t.Run(name, func(t *testing.T) {
			reducer, err := Lookup(name)
			if err != nil {
				t.Fatalf("Lookup(%q) = %v", name, err)
			}

			outputDir := t.TempDir()
			if _, err := Run([]Reducer{reducer}, testRecords(), Options{OutputDir: outputDir}); err != nil {
				t.Fatalf("Run() = %v", err)
			}

			entries, err := os.ReadDir(outputDir)
			if err != nil {
				t.Fatalf("reading output dir: %v", err)
			}
			if len(entries) != 1 {
				names := make([]string, len(entries))
				for i, entry := range entries {
					names[i] = entry.Name()
				}
				t.Fatalf("wrote %v, want only %q", names, reducer.Artifact)
			}
			if entries[0].Name() != reducer.Artifact {
				t.Errorf("wrote %q, want the declared artifact %q", entries[0].Name(), reducer.Artifact)
			}
		})
	}
}

// skipper builds a reducer that declines with a reason instead of writing.
func skipper(name, filename, reason string) Reducer {
	return Reducer{
		Name:     name,
		Artifact: filename,
		Reduce: func([]operation.Record, *Output) (string, error) {
			return "", fmt.Errorf("%w: %s", ErrSkipped, reason)
		},
	}
}
