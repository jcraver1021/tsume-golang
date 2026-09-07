// Package store names a downloaded payload and writes it. The URL's suffix wins
// over Content-Type, and a mapper's forced extension overrides both.
package store

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"tsumegolang/internal/labrador/mapper"
)

var (
	ErrWriteFile     = errors.New("failed to write to file")
	ErrInvalidURL    = errors.New("invalid URL")
	ErrCreateDir     = errors.New("failed to create directory")
	ErrPathCollision = errors.New("two downloads want the same file")
)

// A Writer holds one run's output directory and the paths written into it, so
// no download silently replaces another.
type Writer struct {
	baseDir string
	mu      sync.Mutex
	written map[string]string
}

func NewWriter(baseDir string) *Writer {
	return &Writer{baseDir: baseDir, written: map[string]string{}}
}

// Write puts a mapped payload on disk.
func (w *Writer) Write(payload mapper.Payload) (string, error) {
	ext := payload.Extension
	if ext == "" {
		ext = extensionFor(payload.URL, payload.ContentType)
	}

	filePath, err := w.pathFor(payload, ext)
	if err != nil {
		return "", err
	}

	// pathFor keeps any extension already in the URL, so swap it afterwards.
	if payload.Extension != "" {
		filePath = strings.TrimSuffix(filePath, filepath.Ext(filePath)) + "." + payload.Extension
	}

	if err := w.claim(filePath, payload.URL); err != nil {
		return "", err
	}
	if err := os.WriteFile(filePath, payload.Content, 0644); err != nil {
		return "", fmt.Errorf("%w: %w", ErrWriteFile, err)
	}

	return filePath, nil
}

// claim covers what PlanFilenames cannot foresee: a mapper rewriting two
// extensions into one.
func (w *Writer) claim(path, url string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if owner, taken := w.written[path]; taken {
		return fmt.Errorf("%w: %s was already written for %s", ErrPathCollision, path, owner)
	}
	w.written[path] = url
	return nil
}

func (w *Writer) pathFor(payload mapper.Payload, ext string) (string, error) {
	parsed, err := url.Parse(payload.URL)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidURL, err)
	}

	dirPath := filepath.Join(w.baseDir, payload.Section)
	pathPart := strings.Trim(parsed.Path, "/")

	segment := payload.Filename
	if segment == "" {
		segment = defaultSegment(parsed, pathPart)
	}

	var filename string
	switch {
	case pathPart == "":
		// A host is not a filename, so it always takes an extension.
		filename = segment + "." + ext
	case filepath.Ext(segment) != "":
		filename = segment
	default:
		filename = segment + "." + ext
	}

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("%w: %w", ErrCreateDir, err)
	}

	return filepath.Join(dirPath, filename), nil
}

func defaultSegment(parsed *url.URL, pathPart string) string {
	if pathPart == "" {
		return parsed.Host
	}
	segments := strings.Split(pathPart, "/")
	return segments[len(segments)-1]
}
