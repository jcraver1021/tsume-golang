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
	errDownloadFailed = errors.New("record failed")
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

func TestRenderMarkdownIndexUsesRelativePaths(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "index.md")

	records := []operation.Record{
		{
			Section:  "Chapter 1",
			URL:      "https://example.com/page1",
			FilePath: filepath.Join(tmpDir, "Chapter 1", "page1.html"),
			Success:  true,
		},
	}

	raw := RenderMarkdownIndex(records, indexPath)

	want := fmt.Sprintf("- [https://example.com/page1](%s)", filepath.Join("Chapter 1", "page1.html"))
	if !strings.Contains(string(raw), want) {
		t.Errorf("index should contain %q, got:\n%s", want, raw)
	}
}

func TestRenderMarkdownIndexGroupsSectionsOnce(t *testing.T) {
	records := []operation.Record{
		{Section: "Chapter 1", URL: "https://example.com/page1", FilePath: "page1.html", Success: true},
		{Section: "Chapter 1", URL: "https://example.com/page2", FilePath: "page2.html", Success: true},
		{Section: "Chapter 1", URL: "https://example.com/page3", FilePath: "page3.html", Success: true},
	}

	content := generateIndex(t, records)

	if got := strings.Count(content, "## Chapter 1"); got != 1 {
		t.Errorf("'## Chapter 1' appears %d times, want 1", got)
	}
	for _, record := range records {
		if !strings.Contains(content, record.URL) {
			t.Errorf("index should contain URL %q", record.URL)
		}
	}
}

func TestRenderMarkdownIndexReportsErrorMessages(t *testing.T) {
	records := []operation.Record{
		{Section: "Chapter 1", URL: "https://example.com/timeout", Error: errTimeout},
	}

	content := generateIndex(t, records)

	if !strings.Contains(content, "Error:") {
		t.Error("index should label failures with 'Error:'")
	}
	if !strings.Contains(content, "timeout") {
		t.Error("index should include the error message")
	}
}

func TestRenderMarkdownIndexHandlesMissingError(t *testing.T) {
	records := []operation.Record{
		{Section: "Chapter 1", URL: "https://example.com/x", Success: false},
	}

	content := generateIndex(t, records)

	if !strings.Contains(content, "unknown error") {
		t.Error("a failure with no error should still be explained")
	}
}
