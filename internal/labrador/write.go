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

type OrganizationMode string

const (
	OrgModeFlat   OrganizationMode = "flat"
	OrgModeDomain OrganizationMode = "domain"
	OrgModePath   OrganizationMode = "path"
)

func ConvertUrlToFilename(urlStr string) string {
	replacer := strings.NewReplacer("http://", "", "https://", "", "/", "_", ":", "_")
	filename := replacer.Replace(urlStr)
	return filename + ".html"
}

func buildFilePath(urlStr string, baseDir string, mode OrganizationMode, ext string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidURL, err)
	}

	var dirPath string
	var filename string

	switch mode {
	case OrgModeFlat:
		dirPath = baseDir
		replacer := strings.NewReplacer("http://", "", "https://", "", "/", "_", ":", "_")
		baseFilename := replacer.Replace(urlStr)
		filename = baseFilename + "." + ext
	case OrgModeDomain:
		dirPath = filepath.Join(baseDir, parsedURL.Host)
		pathPart := strings.Trim(parsedURL.Path, "/")
		if pathPart == "" {
			filename = "index." + ext
		} else {
			replacer := strings.NewReplacer("/", "_", ":", "_")
			filename = replacer.Replace(pathPart) + "." + ext
		}
	case OrgModePath:
		dirPath = filepath.Join(baseDir, parsedURL.Host)
		pathPart := strings.Trim(parsedURL.Path, "/")
		if pathPart == "" {
			filename = "index." + ext
		} else {
			pathSegments := strings.Split(pathPart, "/")
			if len(pathSegments) > 1 {
				dirPath = filepath.Join(dirPath, filepath.Join(pathSegments[:len(pathSegments)-1]...))
				lastSegment := pathSegments[len(pathSegments)-1]
				if filepath.Ext(lastSegment) != "" {
					filename = lastSegment
				} else {
					filename = lastSegment + "." + ext
				}
			} else {
				if filepath.Ext(pathSegments[0]) != "" {
					filename = pathSegments[0]
				} else {
					filename = pathSegments[0] + "." + ext
				}
			}
		}
	default:
		dirPath = baseDir
		replacer := strings.NewReplacer("http://", "", "https://", "", "/", "_", ":", "_")
		baseFilename := replacer.Replace(urlStr)
		filename = baseFilename + "." + ext
	}

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("%w: %w", ErrCreateDir, err)
	}

	return filepath.Join(dirPath, filename), nil
}

func WriteToFile(urlStr string, content []byte, contentType string, baseDir string, mode OrganizationMode) (string, error) {
	ext := DetermineFileExtension(urlStr, contentType)
	filePath, err := buildFilePath(urlStr, baseDir, mode, ext)
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
