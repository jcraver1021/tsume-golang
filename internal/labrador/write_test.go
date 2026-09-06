package labrador_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	. "tsumegolang/internal/labrador"
	"tsumegolang/internal/labrador/mapper"
)

func TestWritePayloadSectionBased(t *testing.T) {
	testCases := []struct {
		name         string
		payload      mapper.Payload
		wantDir      string
		wantFilename string
	}{
		{
			name:         "root URL in simple section",
			payload:      mapper.Payload{URL: "https://example.com", Section: "Chapter 1", Content: []byte("<html>test</html>"), ContentType: "text/html"},
			wantDir:      "Chapter 1",
			wantFilename: "example.com.html",
		},
		{
			name:         "page with path in simple section",
			payload:      mapper.Payload{URL: "https://example.com/docs/guide", Section: "Chapter 2", Content: []byte("<html>guide</html>"), ContentType: "text/html"},
			wantDir:      "Chapter 2",
			wantFilename: "guide.html",
		},
		{
			name:         "PDF in nested section",
			payload:      mapper.Payload{URL: "https://example.com/manual.pdf", Section: "Documents/PDFs", Content: []byte("PDF content"), ContentType: "application/pdf"},
			wantDir:      filepath.Join("Documents", "PDFs"),
			wantFilename: "manual.pdf",
		},
		{
			name:         "JSON in deeply nested section",
			payload:      mapper.Payload{URL: "https://api.example.com/v1/data", Section: "API/v1/Responses", Content: []byte(`{"key": "value"}`), ContentType: "application/json"},
			wantDir:      filepath.Join("API", "v1", "Responses"),
			wantFilename: "data.json",
		},
		{
			name:         "PNG with existing extension",
			payload:      mapper.Payload{URL: "https://example.com/logo.png", Section: "Images", Content: []byte("PNG binary data"), ContentType: "image/png"},
			wantDir:      "Images",
			wantFilename: "logo.png",
		},
		{
			name:         "file with extension already in URL",
			payload:      mapper.Payload{URL: "https://example.com/config.json", Section: "Config Files", Content: []byte(`{"test": true}`), ContentType: "application/json"},
			wantDir:      "Config Files",
			wantFilename: "config.json",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			baseDir := t.TempDir()

			gotPath, err := WritePayload(tc.payload, baseDir)
			if err != nil {
				t.Fatalf("WritePayload() = %v", err)
			}

			wantPath := filepath.Join(baseDir, tc.wantDir, tc.wantFilename)
			if gotPath != wantPath {
				t.Fatalf("path = %q, want %q", gotPath, wantPath)
			}

			gotContent, err := os.ReadFile(gotPath)
			if err != nil {
				t.Fatalf("reading written file: %v", err)
			}
			if string(gotContent) != string(tc.payload.Content) {
				t.Errorf("content = %q, want %q", gotContent, tc.payload.Content)
			}

			if _, err := os.Stat(filepath.Join(baseDir, tc.wantDir)); err != nil {
				t.Errorf("directory %q not created: %v", tc.wantDir, err)
			}
		})
	}
}

func TestWritePayloadBinaryContent(t *testing.T) {
	binaryContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D}
	payload := mapper.Payload{
		URL:         "https://example.com/image.png",
		Section:     "Binary Files",
		Content:     binaryContent,
		ContentType: "image/png",
	}

	gotPath, err := WritePayload(payload, t.TempDir())
	if err != nil {
		t.Fatalf("WritePayload() = %v", err)
	}

	gotContent, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}
	if !bytes.Equal(gotContent, binaryContent) {
		t.Errorf("content = % X, want % X", gotContent, binaryContent)
	}
}

func TestWritePayloadForcedExtensionReplacesURLSuffix(t *testing.T) {
	testCases := []struct {
		name       string
		payload    mapper.Payload
		wantSuffix string
	}{
		{
			name:       "forced extension overrides URL suffix",
			payload:    mapper.Payload{URL: "https://example.com/page.html", Section: "Docs", Content: []byte("text"), ContentType: "text/plain", Extension: "txt"},
			wantSuffix: filepath.Join("Docs", "page.txt"),
		},
		{
			name:       "forced extension applies to extensionless URL",
			payload:    mapper.Payload{URL: "https://example.com/guide", Section: "Docs", Content: []byte("text"), ContentType: "text/plain", Extension: "txt"},
			wantSuffix: filepath.Join("Docs", "guide.txt"),
		},
		{
			name:       "without a forced extension the URL suffix wins",
			payload:    mapper.Payload{URL: "https://example.com/page.html", Section: "Docs", Content: []byte("<p>x</p>"), ContentType: "text/html"},
			wantSuffix: filepath.Join("Docs", "page.html"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			baseDir := t.TempDir()

			got, err := WritePayload(tc.payload, baseDir)
			if err != nil {
				t.Fatalf("WritePayload() = %v", err)
			}

			want := filepath.Join(baseDir, tc.wantSuffix)
			if got != want {
				t.Fatalf("path = %q, want %q", got, want)
			}

			content, err := os.ReadFile(got)
			if err != nil {
				t.Fatalf("reading written file: %v", err)
			}
			if string(content) != string(tc.payload.Content) {
				t.Errorf("content = %q, want %q", content, tc.payload.Content)
			}
		})
	}
}
