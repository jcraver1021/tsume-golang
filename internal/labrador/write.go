package labrador

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
	ErrCantCreateFile = errors.New("failed to create file")
	ErrWriteFile      = errors.New("failed to write to file")
	ErrInvalidURL     = errors.New("invalid URL")
	ErrCreateDir      = errors.New("failed to create directory")
)

func ConvertUrlToFilename(urlStr string) string {
	replacer := strings.NewReplacer("http://", "", "https://", "", "/", "_", ":", "_")
	filename := replacer.Replace(urlStr)
	return filename + ".html"
}

func buildFilePath(urlStr string, baseDir string, section string, ext string) (string, error) {
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

func WriteToFile(urlStr string, content []byte, contentType string, baseDir string, section string) (string, error) {
	ext := DetermineFileExtension(urlStr, contentType)
	filePath, err := buildFilePath(urlStr, baseDir, section, ext)
	if err != nil {
		return "", err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCantCreateFile, err)
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrWriteFile, err)
	}

	return filePath, nil
}

// WritePayload writes a mapped payload, honouring an extension a mapper forced
// in place of the one DetermineFileExtension would infer from the URL.
func WritePayload(payload mapper.Payload, baseDir string) (string, error) {
	ext := payload.Extension
	if ext == "" {
		ext = DetermineFileExtension(payload.URL, payload.ContentType)
	}

	filePath, err := buildFilePath(payload.URL, baseDir, payload.Section, ext)
	if err != nil {
		return "", err
	}

	// buildFilePath keeps an extension already present in the URL's last
	// segment, so a forced extension has to be swapped in afterwards.
	if payload.Extension != "" {
		filePath = strings.TrimSuffix(filePath, filepath.Ext(filePath)) + "." + payload.Extension
	}

	if err := os.WriteFile(filePath, payload.Content, 0644); err != nil {
		return "", fmt.Errorf("%w: %w", ErrWriteFile, err)
	}

	return filePath, nil
}
