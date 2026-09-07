package store_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"tsumegolang/internal/labrador/mapper"
	. "tsumegolang/internal/labrador/store"
)

func TestWriteSectionBased(t *testing.T) {
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
		{
			name: "binary content survives byte for byte",
			payload: mapper.Payload{
				URL:         "https://example.com/image.png",
				Section:     "Binary Files",
				Content:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D},
				ContentType: "image/png",
			},
			wantDir:      "Binary Files",
			wantFilename: "image.png",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			baseDir := t.TempDir()

			gotPath, err := NewWriter(baseDir).Write(tc.payload)
			if err != nil {
				t.Fatalf("Write() = %v", err)
			}

			wantPath := filepath.Join(baseDir, tc.wantDir, tc.wantFilename)
			if gotPath != wantPath {
				t.Fatalf("path = %q, want %q", gotPath, wantPath)
			}

			gotContent, err := os.ReadFile(gotPath)
			if err != nil {
				t.Fatalf("reading written file: %v", err)
			}
			if !bytes.Equal(gotContent, tc.payload.Content) {
				t.Errorf("content = % X, want % X", gotContent, tc.payload.Content)
			}

			if _, err := os.Stat(filepath.Join(baseDir, tc.wantDir)); err != nil {
				t.Errorf("directory %q not created: %v", tc.wantDir, err)
			}
		})
	}
}

func TestWriteForcedExtensionReplacesURLSuffix(t *testing.T) {
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

			got, err := NewWriter(baseDir).Write(tc.payload)
			if err != nil {
				t.Fatalf("Write() = %v", err)
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

// PlanFilenames works from URLs, so it cannot foresee a mapper collapsing two
// different extensions onto one name. The writer refuses rather than overwrite.
func TestWriterRefusesToOverwriteWithinARun(t *testing.T) {
	testCases := []struct {
		name     string
		payloads []mapper.Payload
		wantErr  error
		wantKept string
	}{
		{
			name: "two payloads mapped onto the same name",
			payloads: []mapper.Payload{
				{URL: "https://x.com/a/doc.html", Section: "S", Content: []byte("first"), Extension: "txt"},
				{URL: "https://x.com/b/doc.txt", Section: "S", Content: []byte("second"), Filename: "doc.txt"},
			},
			wantErr:  ErrPathCollision,
			wantKept: "first",
		},
		{
			name: "distinct names are both written",
			payloads: []mapper.Payload{
				{URL: "https://x.com/a", Section: "S", Content: []byte("first"), Filename: "a", ContentType: "text/html"},
				{URL: "https://x.com/b", Section: "S", Content: []byte("second"), Filename: "b", ContentType: "text/html"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			baseDir := t.TempDir()
			writer := NewWriter(baseDir)

			var firstPath string
			var lastErr error
			for i, payload := range tc.payloads {
				path, err := writer.Write(payload)
				if i == 0 {
					firstPath = path
				}
				lastErr = err
			}

			if tc.wantErr == nil {
				if lastErr != nil {
					t.Fatalf("Write() = %v", lastErr)
				}
				return
			}

			if !errors.Is(lastErr, tc.wantErr) {
				t.Fatalf("err = %v, want %v", lastErr, tc.wantErr)
			}
			if !strings.Contains(lastErr.Error(), tc.payloads[0].URL) {
				t.Errorf("error %q does not name the URL that claimed the path", lastErr)
			}

			got, err := os.ReadFile(firstPath)
			if err != nil {
				t.Fatalf("reading %s: %v", firstPath, err)
			}
			if string(got) != tc.wantKept {
				t.Errorf("content = %q, want %q — the first write must survive", got, tc.wantKept)
			}
		})
	}
}

// Workers share one Writer, so its bookkeeping has to be safe under -race.
func TestWriterIsSafeForConcurrentUse(t *testing.T) {
	writer := NewWriter(t.TempDir())

	var wg sync.WaitGroup
	errs := make(chan error, 32)
	for i := range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			payload := mapper.Payload{
				URL:         fmt.Sprintf("https://x.com/doc%d", i),
				Section:     "S",
				Filename:    fmt.Sprintf("doc%d", i),
				Content:     []byte("body"),
				ContentType: "text/html",
			}
			if _, err := writer.Write(payload); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent Write() = %v", err)
	}
}
