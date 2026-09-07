package reducer_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"tsumegolang/internal/labrador/operation"
	. "tsumegolang/internal/labrador/reducer"
)

var (
	errDownloadFailed = errors.New("download failed")
	errTimeout        = errors.New("timeout waiting for result")
)

func generateIndex(t *testing.T, records []operation.Record) string {
	t.Helper()
	return string(RenderMarkdownIndex(records, filepath.Join(t.TempDir(), "index.md")))
}

func TestRenderMarkdownIndex(t *testing.T) {
	testCases := []struct {
		name           string
		records        []operation.Record
		wantSections   []string
		wantSuccessful int
		wantFailed     int
		indexPath      string // where the index will sit, for relative links
		wantContains   []string
	}{
		{
			name: "all successful downloads",
			records: []operation.Record{
				{Section: "Chapter 1", URL: "https://example.com/page1", FilePath: "page1.html", Success: true},
				{Section: "Chapter 1", URL: "https://example.com/page2", FilePath: "page2.html", Success: true},
				{Section: "Chapter 2", URL: "https://example.com/page3", FilePath: "page3.html", Success: true},
			},
			wantSections:   []string{"Chapter 1", "Chapter 2"},
			wantSuccessful: 3,
		},
		{
			name: "mixed success and failure",
			records: []operation.Record{
				{Section: "Chapter 1", URL: "https://example.com/page1", FilePath: "page1.html", Success: true},
				{Section: "Chapter 1", URL: "https://example.com/page2", Error: errDownloadFailed},
				{Section: "Chapter 2", URL: "https://example.com/page3", FilePath: "page3.html", Success: true},
			},
			wantSections:   []string{"Chapter 1", "Chapter 2"},
			wantSuccessful: 2,
			wantFailed:     1,
		},
		{
			name: "all failed downloads",
			records: []operation.Record{
				{Section: "Chapter 1", URL: "https://example.com/page1", Error: errDownloadFailed},
				{Section: "Chapter 1", URL: "https://example.com/page2", Error: errTimeout},
			},
			wantSections: []string{"Chapter 1"},
			wantFailed:   2,
		},
		{
			name:         "empty records",
			records:      []operation.Record{},
			wantSections: []string{},
		},
		{
			name: "failures explain themselves",
			records: []operation.Record{
				{Section: "Chapter 1", URL: "https://example.com/timeout", Error: errTimeout},
				{Section: "Chapter 1", URL: "https://example.com/silent"},
			},
			wantSections: []string{"Chapter 1"},
			wantFailed:   2,
			wantContains: []string{"Error: timeout waiting for result", "Error: unknown error"},
		},
		{
			name: "a section is headed once however many URLs it holds",
			records: []operation.Record{
				{Section: "Chapter 1", URL: "https://example.com/page1", FilePath: "page1.html", Success: true},
				{Section: "Chapter 1", URL: "https://example.com/page2", FilePath: "page2.html", Success: true},
				{Section: "Chapter 1", URL: "https://example.com/page3", FilePath: "page3.html", Success: true},
			},
			wantSections:   []string{"Chapter 1"},
			wantSuccessful: 3,
			wantContains:   []string{"## Chapter 1"},
		},
		{
			name: "different file types",
			records: []operation.Record{
				{Section: "Documents", URL: "https://example.com/doc.pdf", FilePath: "doc.pdf", Success: true},
				{Section: "Documents", URL: "https://example.com/image.png", FilePath: "image.png", Success: true},
				{Section: "Data", URL: "https://example.com/data.json", FilePath: "data.json", Success: true},
			},
			wantSections:   []string{"Data", "Documents"},
			wantSuccessful: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content := generateIndex(t, tc.records)

			for _, want := range tc.wantContains {
				if !strings.Contains(content, want) {
					t.Errorf("index should contain %q, got:\n%s", want, content)
				}
			}
			if strings.Count(content, "## Chapter 1") > 1 {
				t.Error("a section should be headed exactly once")
			}

			if !strings.Contains(content, "# Download Index") {
				t.Error("index should contain the '# Download Index' header")
			}
			if !strings.Contains(content, "Generated:") {
				t.Error("index should contain a 'Generated:' timestamp")
			}

			wantTotals := fmt.Sprintf("**Total Downloads**: %d | **Successful**: %d | **Failed**: %d",
				len(tc.records), tc.wantSuccessful, tc.wantFailed)
			if !strings.Contains(content, wantTotals) {
				t.Errorf("index should contain %q", wantTotals)
			}

			for _, section := range tc.wantSections {
				if !strings.Contains(content, "## "+section) {
					t.Errorf("index should contain section header %q", "## "+section)
				}
			}

			for _, record := range tc.records {
				if !strings.Contains(content, record.URL) {
					t.Errorf("index should contain URL %q", record.URL)
				}
				if !record.Success && !strings.Contains(content, "❌") {
					t.Error("index should mark failed downloads with ❌")
				}
			}
		})
	}
}

// Links are relative to where the index itself sits.
func TestRenderMarkdownIndexUsesRelativePaths(t *testing.T) {
	tmpDir := t.TempDir()

	testCases := []struct {
		name     string
		filePath string
		want     string
	}{
		{
			name:     "a file beside the index",
			filePath: filepath.Join(tmpDir, "page1.html"),
			want:     "page1.html",
		},
		{
			name:     "a file in a section directory",
			filePath: filepath.Join(tmpDir, "Chapter 1", "page1.html"),
			want:     filepath.Join("Chapter 1", "page1.html"),
		},
		{
			name:     "a file outside the output tree",
			filePath: filepath.Join(tmpDir, "..", "elsewhere.html"),
			want:     filepath.Join("..", "elsewhere.html"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			records := []operation.Record{{
				Section:  "Chapter 1",
				URL:      "https://example.com/page1",
				FilePath: tc.filePath,
				Success:  true,
			}}

			got := string(RenderMarkdownIndex(records, filepath.Join(tmpDir, "index.md")))
			want := fmt.Sprintf("- [https://example.com/page1](%s)", tc.want)
			if !strings.Contains(got, want) {
				t.Errorf("index should contain %q, got:\n%s", want, got)
			}
		})
	}
}
