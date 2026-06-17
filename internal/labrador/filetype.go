package labrador

import (
	"path/filepath"
	"strings"
)

var contentTypeToExtension = map[string]string{
	"text/html":                "html",
	"application/pdf":          "pdf",
	"image/jpeg":               "jpg",
	"image/jpg":                "jpg",
	"image/png":                "png",
	"image/gif":                "gif",
	"image/svg+xml":            "svg",
	"image/webp":               "webp",
	"text/plain":               "txt",
	"text/markdown":            "md",
	"application/json":         "json",
	"application/xml":          "xml",
	"text/xml":                 "xml",
	"application/zip":          "zip",
	"application/gzip":         "gz",
	"application/x-tar":        "tar",
	"video/mp4":                "mp4",
	"video/webm":               "webm",
	"audio/mpeg":               "mp3",
	"audio/wav":                "wav",
	"application/octet-stream": "bin",
}

func DetermineFileExtension(url string, contentType string) string {
	urlExt := extractExtensionFromURL(url)
	if urlExt != "" {
		return urlExt
	}

	if contentType != "" {
		baseContentType := strings.Split(contentType, ";")[0]
		baseContentType = strings.TrimSpace(baseContentType)
		if ext, ok := contentTypeToExtension[baseContentType]; ok {
			return ext
		}
	}

	return "html"
}

func extractExtensionFromURL(url string) string {
	path := url
	if idx := strings.Index(url, "?"); idx != -1 {
		path = url[:idx]
	}
	if idx := strings.Index(path, "#"); idx != -1 {
		path = path[:idx]
	}

	ext := filepath.Ext(path)
	if ext != "" {
		ext = strings.TrimPrefix(ext, ".")
		ext = strings.ToLower(ext)

		knownExtensions := map[string]bool{
			"html": true, "htm": true, "pdf": true, "jpg": true, "jpeg": true,
			"png": true, "gif": true, "svg": true, "webp": true, "txt": true,
			"md": true, "json": true, "xml": true, "zip": true, "gz": true,
			"tar": true, "mp4": true, "webm": true, "mp3": true, "wav": true,
			"css": true, "js": true, "py": true, "go": true, "rs": true,
			"java": true, "c": true, "cpp": true, "h": true, "hpp": true,
		}

		if knownExtensions[ext] {
			return ext
		}
	}

	return ""
}
