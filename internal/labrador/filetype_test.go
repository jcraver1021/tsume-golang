package labrador_test

import (
	"testing"

	. "tsumegolang/internal/labrador"
)

func TestDetermineFileExtension(t *testing.T) {
	testCases := []struct {
		name        string
		url         string
		contentType string
		want        string
	}{
		{
			name:        "URL with PDF extension",
			url:         "https://example.com/document.pdf",
			contentType: "",
			want:        "pdf",
		},
		{
			name:        "URL with PDF extension and query params",
			url:         "https://example.com/document.pdf?download=true&version=2",
			contentType: "",
			want:        "pdf",
		},
		{
			name:        "URL with PNG extension and fragment",
			url:         "https://example.com/image.png#section",
			contentType: "",
			want:        "png",
		},
		{
			name:        "URL with multiple extensions",
			url:         "https://example.com/archive.tar.gz",
			contentType: "",
			want:        "gz",
		},
		{
			name:        "URL without extension, HTML content type",
			url:         "https://example.com/page",
			contentType: "text/html",
			want:        "html",
		},
		{
			name:        "URL without extension, HTML content type with charset",
			url:         "https://example.com/page",
			contentType: "text/html; charset=utf-8",
			want:        "html",
		},
		{
			name:        "URL without extension, PDF content type",
			url:         "https://example.com/document",
			contentType: "application/pdf",
			want:        "pdf",
		},
		{
			name:        "URL without extension, JSON content type",
			url:         "https://api.example.com/data",
			contentType: "application/json",
			want:        "json",
		},
		{
			name:        "URL without extension, JPEG content type",
			url:         "https://example.com/photo",
			contentType: "image/jpeg",
			want:        "jpg",
		},
		{
			name:        "URL without extension, PNG content type",
			url:         "https://example.com/screenshot",
			contentType: "image/png",
			want:        "png",
		},
		{
			name:        "URL without extension, unknown content type",
			url:         "https://example.com/unknown",
			contentType: "application/x-custom",
			want:        "html",
		},
		{
			name:        "URL without extension, no content type",
			url:         "https://example.com/page",
			contentType: "",
			want:        "html",
		},
		{
			name:        "URL with unknown extension",
			url:         "https://example.com/file.xyz",
			contentType: "",
			want:        "html",
		},
		{
			name:        "URL with Go file extension",
			url:         "https://example.com/main.go",
			contentType: "",
			want:        "go",
		},
		{
			name:        "URL with JSON extension",
			url:         "https://example.com/config.json",
			contentType: "",
			want:        "json",
		},
		{
			name:        "URL with SVG extension",
			url:         "https://example.com/logo.svg",
			contentType: "",
			want:        "svg",
		},
		{
			name:        "URL extension takes precedence over content type",
			url:         "https://example.com/file.pdf",
			contentType: "text/html",
			want:        "pdf",
		},
		{
			name:        "Empty URL and content type",
			url:         "",
			contentType: "",
			want:        "html",
		},
		{
			name:        "URL with uppercase extension",
			url:         "https://example.com/FILE.PDF",
			contentType: "",
			want:        "pdf",
		},
		{
			name:        "URL with mixed case extension",
			url:         "https://example.com/image.PnG",
			contentType: "",
			want:        "png",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := DetermineFileExtension(tc.url, tc.contentType)
			if got != tc.want {
				t.Errorf("DetermineFileExtension(%q, %q) = %q; want %q", tc.url, tc.contentType, got, tc.want)
			}
		})
	}
}
