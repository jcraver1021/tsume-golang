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

	"tsumegolang/internal/labrador/mapper"
)

var (
	ErrWriteFile  = errors.New("failed to write to file")
	ErrInvalidURL = errors.New("invalid URL")
	ErrCreateDir  = errors.New("failed to create directory")
)

func pathFor(urlStr string, baseDir string, section string, ext string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidURL, err)
	}

	dirPath := filepath.Join(baseDir, section)

	pathPart := strings.Trim(parsedURL.Path, "/")
	var filename string

	if pathPart == "" {
		filename = parsedURL.Host + "." + ext
	} else {
		pathSegments := strings.Split(pathPart, "/")
		lastSegment := pathSegments[len(pathSegments)-1]

		if filepath.Ext(lastSegment) != "" {
			filename = lastSegment
		} else {
			filename = lastSegment + "." + ext
		}
	}

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("%w: %w", ErrCreateDir, err)
	}

	return filepath.Join(dirPath, filename), nil
}

// Write puts a mapped payload on disk.
func Write(payload mapper.Payload, baseDir string) (string, error) {
	ext := payload.Extension
	if ext == "" {
		ext = extensionFor(payload.URL, payload.ContentType)
	}

	filePath, err := pathFor(payload.URL, baseDir, payload.Section, ext)
	if err != nil {
		return "", err
	}

	// pathFor keeps any extension already in the URL, so swap it afterwards.
	if payload.Extension != "" {
		filePath = strings.TrimSuffix(filePath, filepath.Ext(filePath)) + "." + payload.Extension
	}

	if err := os.WriteFile(filePath, payload.Content, 0644); err != nil {
		return "", fmt.Errorf("%w: %w", ErrWriteFile, err)
	}

	return filePath, nil
}
