package reducer_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tsumegolang/internal/labrador"
	. "tsumegolang/internal/labrador/reducer"
)

func generateIndex(t *testing.T, records []labrador.DownloadRecord) string {
	t.Helper()

	indexPath := filepath.Join(t.TempDir(), "index.md")
	if err := GenerateMarkdownIndex(records, indexPath); err != nil {
		t.Fatalf("GenerateMarkdownIndex() = %v", err)
	}

	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("index not written: %v", err)
	}
	return string(content)
}

func TestGenerateMarkdownIndex(t *testing.T) {
	testCases := []struct {
		name           string
		records        []labrador.DownloadRecord
		wantSections   []string
		wantSuccessful int
		wantFailed     int
	}{
		{
			name: "all successful downloads",
			records: []labrador.DownloadRecord{
				{Section: "Chapter 1", URL: "https://example.com/page1", FilePath: "page1.html", Success: true},
				{Section: "Chapter 1", URL: "https://example.com/page2", FilePath: "page2.html", Success: true},
				{Section: "Chapter 2", URL: "https://example.com/page3", FilePath: "page3.html", Success: true},
			},
			wantSections:   []string{"Chapter 1", "Chapter 2"},
			wantSuccessful: 3,
		},
		{
			name: "mixed success and failure",
			records: []labrador.DownloadRecord{
				{Section: "Chapter 1", URL: "https://example.com/page1", FilePath: "page1.html", Success: true},
				{Section: "Chapter 1", URL: "https://example.com/page2", Error: labrador.ErrDownloadFailed},
				{Section: "Chapter 2", URL: "https://example.com/page3", FilePath: "page3.html", Success: true},
			},
			wantSections:   []string{"Chapter 1", "Chapter 2"},
			wantSuccessful: 2,
			wantFailed:     1,
		},
		{
			name: "all failed downloads",
			records: []labrador.DownloadRecord{
				{Section: "Chapter 1", URL: "https://example.com/page1", Error: labrador.ErrDownloadFailed},
				{Section: "Chapter 1", URL: "https://example.com/page2", Error: labrador.ErrTimeout},
			},
			wantSections: []string{"Chapter 1"},
			wantFailed:   2,
		},
		{
			name:         "empty records",
			records:      []labrador.DownloadRecord{},
			wantSections: []string{},
		},
		{
			name: "different file types",
			records: []labrador.DownloadRecord{
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

func TestGenerateMarkdownIndexUsesRelativePaths(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "index.md")

	records := []labrador.DownloadRecord{
		{
			Section:  "Chapter 1",
			URL:      "https://example.com/page1",
			FilePath: filepath.Join(tmpDir, "Chapter 1", "page1.html"),
			Success:  true,
		},
	}

	if err := GenerateMarkdownIndex(records, indexPath); err != nil {
		t.Fatalf("GenerateMarkdownIndex() = %v", err)
	}

	raw, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("index not written: %v", err)
	}

	want := fmt.Sprintf("- [https://example.com/page1](%s)", filepath.Join("Chapter 1", "page1.html"))
	if !strings.Contains(string(raw), want) {
		t.Errorf("index should contain %q, got:\n%s", want, raw)
	}
}

func TestGenerateMarkdownIndexGroupsSectionsOnce(t *testing.T) {
	records := []labrador.DownloadRecord{
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

func TestGenerateMarkdownIndexReportsErrorMessages(t *testing.T) {
	records := []labrador.DownloadRecord{
		{Section: "Chapter 1", URL: "https://example.com/timeout", Error: labrador.ErrTimeout},
	}

	content := generateIndex(t, records)

	if !strings.Contains(content, "Error:") {
		t.Error("index should label failures with 'Error:'")
	}
	if !strings.Contains(content, "timeout") {
		t.Error("index should include the error message")
	}
}

func TestGenerateMarkdownIndexHandlesMissingError(t *testing.T) {
	records := []labrador.DownloadRecord{
		{Section: "Chapter 1", URL: "https://example.com/x", Success: false},
	}

	content := generateIndex(t, records)

	if !strings.Contains(content, "unknown error") {
		t.Error("a failure with no error should still be explained")
	}
}
