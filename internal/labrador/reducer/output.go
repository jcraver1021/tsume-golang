package reducer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrCreateDir     = errors.New("failed to create output directory")
	ErrWriteArtifact = errors.New("failed to write artifact")
)

// Output is a reducer's handle on the output directory.
type Output struct {
	dir    string
	source string
}

// Source is the file the records were loaded from, empty for a fresh download.
// A reducer whose artifact is that file can decline rather than rewrite it.
func (o *Output) Source() string {
	return o.source
}

func (o *Output) Path(filename string) string {
	return filepath.Join(o.dir, filename)
}

func (o *Output) Write(filename string, content []byte) (string, error) {
	// When every download failed nothing else created the directory, and the
	// run still owes a report.
	if err := os.MkdirAll(o.dir, 0755); err != nil {
		return "", fmt.Errorf("%w: %w", ErrCreateDir, err)
	}

	path := o.Path(filename)
	if err := os.WriteFile(path, content, 0644); err != nil {
		return "", fmt.Errorf("%w: %w", ErrWriteArtifact, err)
	}
	return path, nil
}
