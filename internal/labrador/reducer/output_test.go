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

func TestOutputWritesUnderTheOutputDir(t *testing.T) {
	outputDir := t.TempDir()

	if _, err := Run([]Reducer{writer("probe", "artifact.txt", "content")}, nil, Options{OutputDir: outputDir}); err != nil {
		t.Fatalf("Run() = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(outputDir, "artifact.txt"))
	if err != nil {
		t.Fatalf("artifact.txt missing: %v", err)
	}
	if string(content) != "content" {
		t.Errorf("content = %q, want %q", content, "content")
	}
}

func TestOutputPathIsRelativeToTheOutputDir(t *testing.T) {
	outputDir := t.TempDir()
	var got string

	probe := Reducer{
		Name: "probe", Artifact: "probe.txt",
		Reduce: func(_ []operation.Record, out *Output) (string, error) {
			got = out.Path("probe.txt")
			return "", nil
		},
	}

	if _, err := Run([]Reducer{probe}, nil, Options{OutputDir: outputDir}); err != nil {
		t.Fatalf("Run() = %v", err)
	}
	if want := filepath.Join(outputDir, "probe.txt"); got != want {
		t.Errorf("Path() = %q, want %q", got, want)
	}
}

// Re-running labrador over the same directory refreshes its artifacts, as it
// always has.
func TestOutputOverwritesAcrossRuns(t *testing.T) {
	outputDir := t.TempDir()

	if _, err := Run([]Reducer{writer("only", "artifact.txt", "first run")}, nil, Options{OutputDir: outputDir}); err != nil {
		t.Fatalf("first Run() = %v", err)
	}
	if _, err := Run([]Reducer{writer("only", "artifact.txt", "second run")}, nil, Options{OutputDir: outputDir}); err != nil {
		t.Fatalf("second Run() = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(outputDir, "artifact.txt"))
	if err != nil {
		t.Fatalf("artifact.txt missing: %v", err)
	}
	if string(content) != "second run" {
		t.Errorf("content = %q, want %q", content, "second run")
	}
}

func TestOutputReportsWriteFailure(t *testing.T) {
	// A path whose parent is a regular file cannot be created.
	blocked := filepath.Join(t.TempDir(), "notadir")
	if err := os.WriteFile(blocked, []byte("x"), 0644); err != nil {
		t.Fatalf("setting up: %v", err)
	}

	_, err := Run([]Reducer{writer("probe", "artifact.txt", "content")}, nil, Options{OutputDir: filepath.Join(blocked, "sub")})
	if !errors.Is(err, ErrCreateDir) {
		t.Fatalf("err = %v, want %v", err, ErrCreateDir)
	}
}
