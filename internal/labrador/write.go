package labrador

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrCantCreateFile = fmt.Errorf("failed to create file")
	ErrWriteFile      = fmt.Errorf("failed to write to file")
	ErrInvalidURL     = fmt.Errorf("invalid URL")
	ErrCreateDir      = fmt.Errorf("failed to create directory")
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
