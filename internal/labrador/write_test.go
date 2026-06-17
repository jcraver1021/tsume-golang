package labrador_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tsumegolang/internal/labrador"
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
			got := labrador.ConvertUrlToFilename(tc.url)
			if got != tc.want {
				t.Errorf("ConvertUrlToFilename(%q) = %q; want %q", tc.url, got, tc.want)
			}
		})
	}
}

func TestWriteToFile_FlatMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "labrador-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		name        string
		url         string
		content     []byte
		contentType string
		wantFile    string
	}{
		{
			name:        "HTML file",
			url:         "https://example.com/page",
			content:     []byte("<html>test</html>"),
			contentType: "text/html",
			wantFile:    "example.com_page.html",
		},
		{
			name:        "PDF from URL extension",
			url:         "https://example.com/document.pdf",
			content:     []byte("PDF content"),
			contentType: "",
			wantFile:    "example.com_document.pdf.pdf",
		},
		{
			name:        "JSON from content type",
			url:         "https://api.example.com/data",
			content:     []byte(`{"key": "value"}`),
			contentType: "application/json",
			wantFile:    "api.example.com_data.json",
		},
		{
			name:        "PNG from URL extension",
			url:         "https://example.com/logo.png",
			content:     []byte("PNG binary data"),
			contentType: "image/png",
			wantFile:    "example.com_logo.png.png",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotPath, err := labrador.WriteToFile(tc.url, tc.content, tc.contentType, tmpDir, labrador.OrgModeFlat)
			if err != nil {
				t.Fatalf("WriteToFile() error = %v", err)
			}

			if !strings.HasSuffix(gotPath, tc.wantFile) {
				t.Errorf("WriteToFile() path = %q; want suffix %q", gotPath, tc.wantFile)
			}

			gotContent, err := os.ReadFile(gotPath)
			if err != nil {
				t.Fatalf("Failed to read written file: %v", err)
			}

			if string(gotContent) != string(tc.content) {
				t.Errorf("File content = %q; want %q", string(gotContent), string(tc.content))
			}
		})
	}
}

func TestWriteToFile_DomainMode(t *testing.T) {
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
		wantDir      string
		wantFilename string
	}{
		{
			name:         "root page",
			url:          "https://example.com",
			content:      []byte("<html>test</html>"),
			contentType:  "text/html",
			wantDir:      "example.com",
			wantFilename: "index.html",
		},
		{
			name:         "page with path",
			url:          "https://example.com/docs/guide",
			content:      []byte("<html>guide</html>"),
			contentType:  "text/html",
			wantDir:      "example.com",
			wantFilename: "docs_guide.html",
		},
		{
			name:         "PDF file",
			url:          "https://example.com/manual.pdf",
			content:      []byte("PDF content"),
			contentType:  "application/pdf",
			wantDir:      "example.com",
			wantFilename: "manual.pdf.pdf",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotPath, err := labrador.WriteToFile(tc.url, tc.content, tc.contentType, tmpDir, labrador.OrgModeDomain)
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

func TestWriteToFile_PathMode(t *testing.T) {
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
		wantDir      string
		wantFilename string
	}{
		{
			name:         "root page",
			url:          "https://example.com",
			content:      []byte("<html>test</html>"),
			contentType:  "text/html",
			wantDir:      "example.com",
			wantFilename: "index.html",
		},
		{
			name:         "nested path",
			url:          "https://example.com/docs/api/reference",
			content:      []byte("<html>reference</html>"),
			contentType:  "text/html",
			wantDir:      filepath.Join("example.com", "docs", "api"),
			wantFilename: "reference.html",
		},
		{
			name:         "PDF in nested path",
			url:          "https://example.com/downloads/manual.pdf",
			content:      []byte("PDF content"),
			contentType:  "application/pdf",
			wantDir:      filepath.Join("example.com", "downloads"),
			wantFilename: "manual.pdf",
		},
		{
			name:         "file with extension preserved",
			url:          "https://example.com/data/config.json",
			content:      []byte(`{"test": true}`),
			contentType:  "application/json",
			wantDir:      filepath.Join("example.com", "data"),
			wantFilename: "config.json",
		},
		{
			name:         "single level path",
			url:          "https://example.com/about",
			content:      []byte("<html>about</html>"),
			contentType:  "text/html",
			wantDir:      "example.com",
			wantFilename: "about.html",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotPath, err := labrador.WriteToFile(tc.url, tc.content, tc.contentType, tmpDir, labrador.OrgModePath)
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

	gotPath, err := labrador.WriteToFile("https://example.com/image.png", binaryContent, "image/png", tmpDir, labrador.OrgModeFlat)
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
