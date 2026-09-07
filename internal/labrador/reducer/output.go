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

// Output is a reducer's handle on the output directory. Reducers write through
// it rather than touching the filesystem directly, which keeps artifact paths
// and directory creation in one place.
type Output struct {
	dir    string
	source string
}

// Source is the file the run's records were loaded from, or empty for a fresh
// download. A reducer whose artifact is that file can decline rather than
// rewrite its own input.
func (o *Output) Source() string {
	return o.source
}

func (o *Output) Path(filename string) string {
	return filepath.Join(o.dir, filename)
}

func (o *Output) Write(filename string, content []byte) (string, error) {
	// Creating the directory here matters when every download failed: nothing
	// else created it, but the run still owes the operator a report.
	if err := os.MkdirAll(o.dir, 0755); err != nil {
		return "", fmt.Errorf("%w: %w", ErrCreateDir, err)
	}

	path := o.Path(filename)
	if err := os.WriteFile(path, content, 0644); err != nil {
		return "", fmt.Errorf("%w: %w", ErrWriteArtifact, err)
	}
	return path, nil
}
