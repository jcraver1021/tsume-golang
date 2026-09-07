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
		name    string
		names   []string
		want    []string
		wantErr error
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
		{name: "unknown name is rejected", names: []string{"collate"}, wantErr: ErrUnknownReducer},
		{
			name:    "repeating a reducer is rejected",
			names:   []string{"markdown-index", "manifest-json", "markdown-index"},
			wantErr: ErrDuplicateReducer,
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

func TestResolveErrorsExplainWhy(t *testing.T) {
	_, err := Resolve([]string{"markdown-index", "manifest-json", "markdown-index"})
	if !errors.Is(err, ErrDuplicateReducer) {
		t.Fatalf("err = %v, want %v", err, ErrDuplicateReducer)
	}
	for _, want := range []string{`"markdown-index"`, "positions 1 and 3"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err, want)
		}
	}

	_, err = Lookup("collate")
	for _, name := range Names() {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error %q does not mention %q", err, name)
		}
	}
}

func TestRunProducesEveryArtifact(t *testing.T) {
	outputDir := t.TempDir()

	reducers, err := Resolve([]string{"markdown-index", "manifest-json"})
	if err != nil {
		t.Fatalf("Resolve() = %v", err)
	}

	report, err := Run(reducers, testRecords(), Options{OutputDir: outputDir})
	if err != nil {
		t.Fatalf("Run() = %v", err)
	}
	if len(report.Summaries) != 2 {
		t.Fatalf("summaries = %d, want 2", len(report.Summaries))
	}

	for _, artifact := range []string{"index.md", "manifest.json"} {
		if _, err := os.Stat(filepath.Join(outputDir, artifact)); err != nil {
			t.Errorf("%s not written: %v", artifact, err)
		}
	}
}

// Nothing else creates the output directory when every download failed, but the
// run still owes the operator its reports.
func TestRunCreatesMissingOutputDir(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "nested", "downloads")

	reducers, err := Resolve([]string{"markdown-index", "manifest-json"})
	if err != nil {
		t.Fatalf("Resolve() = %v", err)
	}
	if _, err := Run(reducers, testRecords(), Options{OutputDir: outputDir}); err != nil {
		t.Fatalf("Run() = %v", err)
	}

	for _, artifact := range []string{"index.md", "manifest.json"} {
		if _, err := os.Stat(filepath.Join(outputDir, artifact)); err != nil {
			t.Errorf("%s not written: %v", artifact, err)
		}
	}
}

func TestRunWithNoReducersDoesNothing(t *testing.T) {
	outputDir := t.TempDir()

	report, err := Run(nil, testRecords(), Options{OutputDir: outputDir})
	if err != nil {
		t.Fatalf("Run(nil) = %v", err)
	}
	if len(report.Summaries) != 0 || len(report.Skips) != 0 {
		t.Errorf("report = %+v, want nothing", report)
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("reading output dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("output dir has %d entries, want none", len(entries))
	}
}

// A failing reducer must not deny the others their artifacts.
func TestRunContinuesPastFailuresAndAggregatesThem(t *testing.T) {
	outputDir := t.TempDir()
	boom := errors.New("boom")
	bang := errors.New("bang")

	reducers := []Reducer{
		{Name: "first-failure", Artifact: "a.txt",
			Reduce: func([]operation.Record, *Output) (string, error) { return "", boom }},
		writer("survivor", "b.txt", "written"),
		{Name: "second-failure", Artifact: "c.txt",
			Reduce: func([]operation.Record, *Output) (string, error) { return "", bang }},
	}

	report, err := Run(reducers, testRecords(), Options{OutputDir: outputDir})

	if len(report.Summaries) != 1 || !strings.HasPrefix(report.Summaries[0], "wrote ") {
		t.Errorf("summaries = %v, want the survivor's line only", report.Summaries)
	}
	if !errors.Is(err, boom) || !errors.Is(err, bang) {
		t.Fatalf("err = %v, want both failures joined", err)
	}
	for _, want := range []string{"first-failure", "second-failure"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not name %q", err, want)
		}
	}

	if _, err := os.Stat(filepath.Join(outputDir, "b.txt")); err != nil {
		t.Errorf("survivor's artifact missing: %v", err)
	}
}

func TestValidate(t *testing.T) {
	testCases := []struct {
		name     string
		reducers []Reducer
		wantErr  error
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
			name:     "different reducers claiming one artifact",
			reducers: []Reducer{writer("a", "same.txt", ""), writer("b", "same.txt", "")},
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
		})
	}
}

func TestValidateIncompatibilityErrorNamesBoth(t *testing.T) {
	err := Validate([]Reducer{writer("alpha", "same.txt", ""), writer("beta", "same.txt", "")})
	if !errors.Is(err, ErrIncompatibleReducer) {
		t.Fatalf("err = %v, want %v", err, ErrIncompatibleReducer)
	}

	for _, want := range []string{`"alpha"`, `"beta"`, `"same.txt"`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err, want)
		}
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

func TestRunRecordsSkipsSeparatelyFromFailures(t *testing.T) {
	outputDir := t.TempDir()
	boom := errors.New("boom")

	reducers := []Reducer{
		writer("wrote-it", "a.txt", "content"),
		skipper("declined", "b.txt", "b.txt is newer than this run"),
		{Name: "broke", Artifact: "c.txt",
			Reduce: func([]operation.Record, *Output) (string, error) { return "", boom }},
	}

	report, err := Run(reducers, testRecords(), Options{OutputDir: outputDir})

	if len(report.Summaries) != 1 {
		t.Errorf("summaries = %v, want one", report.Summaries)
	}
	if len(report.Skips) != 1 {
		t.Fatalf("skips = %v, want one", report.Skips)
	}
	if !strings.Contains(report.Skips[0], "declined") || !strings.Contains(report.Skips[0], "b.txt is newer than this run") {
		t.Errorf("skip line = %q, want it to name the reducer and its reason", report.Skips[0])
	}

	// A skip is not a failure, so only the broken reducer should surface.
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want it to wrap boom", err)
	}
	if strings.Contains(err.Error(), "declined") {
		t.Errorf("error %q should not mention the skipped reducer", err)
	}
}

// A run where every reducer declines is a clean run, not a failed one.
func TestRunWithOnlySkipsSucceeds(t *testing.T) {
	outputDir := t.TempDir()

	reducers := []Reducer{
		skipper("one", "a.txt", "nothing to do"),
		skipper("two", "b.txt", "nothing to do"),
	}

	report, err := Run(reducers, testRecords(), Options{OutputDir: outputDir})
	if err != nil {
		t.Fatalf("Run() = %v, want nil — skipping is not failing", err)
	}
	if len(report.Skips) != 2 || len(report.Summaries) != 0 {
		t.Errorf("report = %+v, want two skips and no summaries", report)
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("reading output dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("output dir has %d entries, want none", len(entries))
	}
}

// A reducer that skips still declares an artifact, so it still participates in
// compatibility checking — whether it runs is not knowable at validation time.
func TestValidateCountsSkippingReducers(t *testing.T) {
	err := Validate([]Reducer{
		writer("writes", "same.txt", ""),
		skipper("might-skip", "same.txt", "maybe"),
	})
	if !errors.Is(err, ErrIncompatibleReducer) {
		t.Fatalf("err = %v, want %v", err, ErrIncompatibleReducer)
	}
}
