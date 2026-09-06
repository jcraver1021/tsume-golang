package reducer_test

import (
	"errors"
	"strings"
	"testing"

	"tsumegolang/internal/labrador"
	. "tsumegolang/internal/labrador/reducer"
)

func testRecords() []labrador.DownloadRecord {
	return []labrador.DownloadRecord{
		{Section: "Beta", URL: "https://example.com/b", FilePath: "Beta/b.html", Success: true},
		{Section: "Alpha", URL: "https://example.com/a", FilePath: "Alpha/a.html", Success: true},
		{Section: "Alpha", URL: "https://example.com/gone", Success: false, Error: errors.New("404")},
	}
}

func TestRegistryInvariants(t *testing.T) {
	for _, name := range Names() {
		t.Run(name, func(t *testing.T) {
			reducer, has, err := Lookup(name)
			if err != nil || !has {
				t.Fatalf("Lookup(%q) = %v, %v", name, has, err)
			}
			if reducer.Name != name {
				t.Errorf("Name = %q, want %q — registry key must match", reducer.Name, name)
			}
			if reducer.Reduce == nil {
				t.Error("Reduce is nil")
			}
		})
	}
}

func TestLookup(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    string
		wantHas bool
		wantErr error
	}{
		{name: "empty means no reducer", input: "", wantHas: false},
		{name: "whitespace means no reducer", input: "   ", wantHas: false},
		{name: "markdown index", input: "markdown-index", want: "markdown-index", wantHas: true},
		{name: "manifest json", input: " manifest-json ", want: "manifest-json", wantHas: true},
		{name: "unknown name is rejected", input: "collate", wantErr: ErrUnknownReducer},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reducer, has, err := Lookup(tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
				for _, name := range Names() {
					if !strings.Contains(err.Error(), name) {
						t.Errorf("error %q does not mention %q", err, name)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("Lookup() = %v", err)
			}
			if has != tc.wantHas {
				t.Fatalf("has = %v, want %v", has, tc.wantHas)
			}
			if has && reducer.Name != tc.want {
				t.Errorf("Name = %q, want %q", reducer.Name, tc.want)
			}
		})
	}
}

// Every reducer must survive a run in which nothing was downloaded, because
// that is exactly when no other code has created the output directory.
func TestReducersCreateMissingOutputDir(t *testing.T) {
	for _, name := range Names() {
		t.Run(name, func(t *testing.T) {
			reducer, _, err := Lookup(name)
			if err != nil {
				t.Fatalf("Lookup(%q) = %v", name, err)
			}

			outputDir := t.TempDir() + "/nested/downloads"
			summary, err := reducer.Reduce(testRecords(), outputDir)
			if err != nil {
				t.Fatalf("Reduce() = %v", err)
			}
			if !strings.Contains(summary, outputDir) {
				t.Errorf("summary = %q, want it to name the artifact under %q", summary, outputDir)
			}
		})
	}
}

func TestReducersHandleEmptyRecords(t *testing.T) {
	for _, name := range Names() {
		t.Run(name, func(t *testing.T) {
			reducer, _, err := Lookup(name)
			if err != nil {
				t.Fatalf("Lookup(%q) = %v", name, err)
			}
			if _, err := reducer.Reduce(nil, t.TempDir()); err != nil {
				t.Fatalf("Reduce(nil) = %v", err)
			}
		})
	}
}
