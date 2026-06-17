package labrador_test

import (
	"os"
	"path/filepath"
	"testing"

	. "tsumegolang/internal/labrador"
)

func TestConvertUrlToFilename(t *testing.T) {
	testCases := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "simple HTTP URL",
			url:  "http://example.com",
			want: "example.com.html",
		},
		{
			name: "simple HTTPS URL",
			url:  "https://example.com",
			want: "example.com.html",
		},
		{
			name: "URL with path",
			url:  "https://example.com/path/to/page",
			want: "example.com_path_to_page.html",
		},
		{
			name: "URL with port",
			url:  "https://example.com:8080/page",
			want: "example.com_8080_page.html",
		},
		{
			name: "URL with query params",
			url:  "https://example.com/page?query=param",
			want: "example.com_page?query=param.html",
		},
		{
			name: "URL with multiple slashes",
			url:  "https://example.com/path/to/deep/page",
			want: "example.com_path_to_deep_page.html",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := ConvertUrlToFilename(tc.url)
			if got != tc.want {
				t.Errorf("ConvertUrlToFilename(%q) = %q; want %q", tc.url, got, tc.want)
			}
		})
	}
}

func TestWriteToFile_SectionBased(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "labrador-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		name         string
		url          string
		content      []byte
		contentType  string
		section      string
		wantDir      string
		wantFilename string
	}{
		{
			name:         "root URL in simple section",
			url:          "https://example.com",
			content:      []byte("<html>test</html>"),
			contentType:  "text/html",
			section:      "Chapter 1",
			wantDir:      "Chapter 1",
			wantFilename: "example.com.html",
		},
		{
			name:         "page with path in simple section",
			url:          "https://example.com/docs/guide",
			content:      []byte("<html>guide</html>"),
			contentType:  "text/html",
			section:      "Chapter 2",
			wantDir:      "Chapter 2",
			wantFilename: "guide.html",
		},
		{
			name:         "PDF in nested section",
			url:          "https://example.com/manual.pdf",
			content:      []byte("PDF content"),
			contentType:  "application/pdf",
			section:      "Documents/PDFs",
			wantDir:      filepath.Join("Documents", "PDFs"),
			wantFilename: "manual.pdf",
		},
		{
			name:         "JSON in deeply nested section",
			url:          "https://api.example.com/v1/data",
			content:      []byte(`{"key": "value"}`),
			contentType:  "application/json",
			section:      "API/v1/Responses",
			wantDir:      filepath.Join("API", "v1", "Responses"),
			wantFilename: "data.json",
		},
		{
			name:         "PNG with existing extension",
			url:          "https://example.com/logo.png",
			content:      []byte("PNG binary data"),
			contentType:  "image/png",
			section:      "Images",
			wantDir:      "Images",
			wantFilename: "logo.png",
		},
		{
			name:         "file with extension already in URL",
			url:          "https://example.com/config.json",
			content:      []byte(`{"test": true}`),
			contentType:  "application/json",
			section:      "Config Files",
			wantDir:      "Config Files",
			wantFilename: "config.json",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotPath, err := WriteToFile(tc.url, tc.content, tc.contentType, tmpDir, tc.section)
			if err != nil {
				t.Fatalf("WriteToFile() error = %v", err)
			}

			wantPath := filepath.Join(tmpDir, tc.wantDir, tc.wantFilename)
			if gotPath != wantPath {
				t.Errorf("WriteToFile() path = %q; want %q", gotPath, wantPath)
			}

			gotContent, err := os.ReadFile(gotPath)
			if err != nil {
				t.Fatalf("Failed to read written file: %v", err)
			}

			if string(gotContent) != string(tc.content) {
				t.Errorf("File content = %q; want %q", string(gotContent), string(tc.content))
			}

			if _, err := os.Stat(filepath.Join(tmpDir, tc.wantDir)); os.IsNotExist(err) {
				t.Errorf("Expected directory %q to be created", tc.wantDir)
			}
		})
	}
}

func TestWriteToFile_BinaryContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "labrador-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binaryContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D}

	gotPath, err := WriteToFile("https://example.com/image.png", binaryContent, "image/png", tmpDir, "Binary Files")
	if err != nil {
		t.Fatalf("WriteToFile() error = %v", err)
	}

	gotContent, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	if len(gotContent) != len(binaryContent) {
		t.Errorf("Binary content length = %d; want %d", len(gotContent), len(binaryContent))
	}

	for i := range binaryContent {
		if gotContent[i] != binaryContent[i] {
			t.Errorf("Binary content[%d] = 0x%02X; want 0x%02X", i, gotContent[i], binaryContent[i])
		}
	}
}
