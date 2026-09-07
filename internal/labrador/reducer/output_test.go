package reducer_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"tsumegolang/internal/labrador/operation"
	. "tsumegolang/internal/labrador/reducer"
)

// writer builds a reducer that writes fixed content to a fixed filename, so
// validation and running can be exercised without a real artifact.
func writer(name, filename, content string) Reducer {
	return Reducer{
		Name:     name,
		Artifact: filename,
		Reduce: func(_ []operation.Record, out *Output) (string, error) {
			path, err := out.Write(filename, []byte(content))
			if err != nil {
				return "", err
			}
			return "wrote " + path, nil
		},
	}
}

func TestOutputWrite(t *testing.T) {
	testCases := []struct {
		name        string
		runs        []string // content written by one reducer per run
		blockedDir  bool
		wantContent string
		wantErr     error
	}{
		{
			name:        "writes under the output directory",
			runs:        []string{"content"},
			wantContent: "content",
		},
		{
			name:        "a later run refreshes the artifact",
			runs:        []string{"first run", "second run"},
			wantContent: "second run",
		},
		{
			name:       "an uncreatable directory is reported",
			runs:       []string{"content"},
			blockedDir: true,
			wantErr:    ErrCreateDir,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outputDir := t.TempDir()
			if tc.blockedDir {
				// A path whose parent is a regular file cannot be created.
				blocked := filepath.Join(outputDir, "notadir")
				if err := os.WriteFile(blocked, []byte("x"), 0644); err != nil {
					t.Fatalf("setting up: %v", err)
				}
				outputDir = filepath.Join(blocked, "sub")
			}

			var err error
			for _, content := range tc.runs {
				_, err = Run([]Reducer{writer("only", "artifact.txt", content)}, nil, Options{OutputDir: outputDir})
			}

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Run() = %v", err)
			}

			got, readErr := os.ReadFile(filepath.Join(outputDir, "artifact.txt"))
			if readErr != nil {
				t.Fatalf("artifact.txt missing: %v", readErr)
			}
			if string(got) != tc.wantContent {
				t.Errorf("content = %q, want %q", got, tc.wantContent)
			}
		})
	}
}

func TestOutputPath(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
	}{
		{name: "a plain filename", filename: "probe.txt"},
		{name: "a nested filename", filename: filepath.Join("reports", "probe.txt")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outputDir := t.TempDir()
			var got string

			probe := Reducer{
				Name: "probe", Artifact: tc.filename,
				Reduce: func(_ []operation.Record, out *Output) (string, error) {
					got = out.Path(tc.filename)
					return "", nil
				},
			}

			if _, err := Run([]Reducer{probe}, nil, Options{OutputDir: outputDir}); err != nil {
				t.Fatalf("Run() = %v", err)
			}
			if want := filepath.Join(outputDir, tc.filename); got != want {
				t.Errorf("Path() = %q, want %q", got, want)
			}
		})
	}
}
