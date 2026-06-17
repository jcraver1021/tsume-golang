package labrador_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "tsumegolang/internal/labrador"
)

func TestGenerateMarkdownIndex(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "labrador-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		name           string
		records        []DownloadRecord
		wantSections   []string
		wantSuccessful int
		wantFailed     int
		wantTotal      int
	}{
		{
			name: "all successful downloads",
			records: []DownloadRecord{
				{
					Section:  "Chapter 1",
					URL:      "https://example.com/page1",
					FilePath: filepath.Join(tmpDir, "page1.html"),
					Success:  true,
				},
				{
					Section:  "Chapter 1",
					URL:      "https://example.com/page2",
					FilePath: filepath.Join(tmpDir, "page2.html"),
					Success:  true,
				},
				{
					Section:  "Chapter 2",
					URL:      "https://example.com/page3",
					FilePath: filepath.Join(tmpDir, "page3.html"),
					Success:  true,
				},
			},
			wantSections:   []string{"Chapter 1", "Chapter 2"},
			wantSuccessful: 3,
			wantFailed:     0,
			wantTotal:      3,
		},
		{
			name: "mixed success and failure",
			records: []DownloadRecord{
				{
					Section:  "Chapter 1",
					URL:      "https://example.com/page1",
					FilePath: filepath.Join(tmpDir, "page1.html"),
					Success:  true,
				},
				{
					Section: "Chapter 1",
					URL:     "https://example.com/page2",
					Success: false,
					Error:   ErrDownloadFailed,
				},
				{
					Section:  "Chapter 2",
					URL:      "https://example.com/page3",
					FilePath: filepath.Join(tmpDir, "page3.html"),
					Success:  true,
				},
			},
			wantSections:   []string{"Chapter 1", "Chapter 2"},
			wantSuccessful: 2,
			wantFailed:     1,
			wantTotal:      3,
		},
		{
			name: "all failed downloads",
			records: []DownloadRecord{
				{
					Section: "Chapter 1",
					URL:     "https://example.com/page1",
					Success: false,
					Error:   ErrDownloadFailed,
				},
				{
					Section: "Chapter 1",
					URL:     "https://example.com/page2",
					Success: false,
					Error:   ErrTimeout,
				},
			},
			wantSections:   []string{"Chapter 1"},
			wantSuccessful: 0,
			wantFailed:     2,
			wantTotal:      2,
		},
		{
			name:           "empty records",
			records:        []DownloadRecord{},
			wantSections:   []string{},
			wantSuccessful: 0,
			wantFailed:     0,
			wantTotal:      0,
		},
		{
			name: "different file types",
			records: []DownloadRecord{
				{
					Section:  "Documents",
					URL:      "https://example.com/doc.pdf",
					FilePath: filepath.Join(tmpDir, "doc.pdf"),
					Success:  true,
				},
				{
					Section:  "Documents",
					URL:      "https://example.com/image.png",
					FilePath: filepath.Join(tmpDir, "image.png"),
					Success:  true,
				},
				{
					Section:  "Data",
					URL:      "https://example.com/data.json",
					FilePath: filepath.Join(tmpDir, "data.json"),
					Success:  true,
				},
			},
			wantSections:   []string{"Documents", "Data"},
			wantSuccessful: 3,
			wantFailed:     0,
			wantTotal:      3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			indexPath := filepath.Join(tmpDir, "index-"+tc.name+".md")
			err := GenerateMarkdownIndex(tc.records, indexPath)
			if err != nil {
				t.Fatalf("GenerateMarkdownIndex() error = %v", err)
			}

			content, err := os.ReadFile(indexPath)
			if err != nil {
				t.Fatalf("Failed to read markdown file: %v", err)
			}

			contentStr := string(content)

			if !strings.Contains(contentStr, "# Download Index") {
				t.Error("Markdown should contain '# Download Index' header")
			}

			if !strings.Contains(contentStr, "Generated:") {
				t.Error("Markdown should contain 'Generated:' timestamp")
			}

			totalLine := "**Total Downloads**: " + string(rune(tc.wantTotal+'0'))
			if tc.wantTotal < 10 && !strings.Contains(contentStr, totalLine) {
				t.Errorf("Markdown should contain total downloads count, want %d", tc.wantTotal)
			}

			for _, section := range tc.wantSections {
				expectedHeader := "## " + section
				if !strings.Contains(contentStr, expectedHeader) {
					t.Errorf("Markdown should contain section header %q", expectedHeader)
				}
			}

			for _, record := range tc.records {
				if record.Success {
					if !strings.Contains(contentStr, record.URL) {
						t.Errorf("Markdown should contain successful URL %q", record.URL)
					}
				} else {
					if !strings.Contains(contentStr, "❌") {
						t.Error("Markdown should contain failure marker ❌ for failed downloads")
					}
					if !strings.Contains(contentStr, record.URL) {
						t.Errorf("Markdown should contain failed URL %q", record.URL)
					}
				}
			}
		})
	}
}

func TestGenerateMarkdownIndex_RelativePaths(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "labrador-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	records := []DownloadRecord{
		{
			Section:  "Chapter 1",
			URL:      "https://example.com/page1",
			FilePath: filepath.Join(tmpDir, "downloads", "page1.html"),
			Success:  true,
		},
	}

	indexPath := filepath.Join(tmpDir, "index.md")
	err = GenerateMarkdownIndex(records, indexPath)
	if err != nil {
		t.Fatalf("GenerateMarkdownIndex() error = %v", err)
	}

	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("Failed to read markdown file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "[https://example.com/page1]") {
		t.Error("Markdown should contain URL as link text")
	}
}

func TestGenerateMarkdownIndex_MultipleURLsPerSection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "labrador-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	records := []DownloadRecord{
		{
			Section:  "Chapter 1",
			URL:      "https://example.com/page1",
			FilePath: filepath.Join(tmpDir, "page1.html"),
			Success:  true,
		},
		{
			Section:  "Chapter 1",
			URL:      "https://example.com/page2",
			FilePath: filepath.Join(tmpDir, "page2.html"),
			Success:  true,
		},
		{
			Section:  "Chapter 1",
			URL:      "https://example.com/page3",
			FilePath: filepath.Join(tmpDir, "page3.html"),
			Success:  true,
		},
	}

	indexPath := filepath.Join(tmpDir, "index.md")
	err = GenerateMarkdownIndex(records, indexPath)
	if err != nil {
		t.Fatalf("GenerateMarkdownIndex() error = %v", err)
	}

	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("Failed to read markdown file: %v", err)
	}

	contentStr := string(content)

	chapterCount := strings.Count(contentStr, "## Chapter 1")
	if chapterCount != 1 {
		t.Errorf("Section 'Chapter 1' should appear once as header, got %d times", chapterCount)
	}

	for i := 1; i <= 3; i++ {
		url := "https://example.com/page" + string(rune(i+'0'))
		if !strings.Contains(contentStr, url) {
			t.Errorf("Markdown should contain URL %q", url)
		}
	}
}

func TestGenerateMarkdownIndex_ErrorMessages(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "labrador-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	customError := ErrTimeout
	records := []DownloadRecord{
		{
			Section: "Chapter 1",
			URL:     "https://example.com/timeout",
			Success: false,
			Error:   customError,
		},
	}

	indexPath := filepath.Join(tmpDir, "index.md")
	err = GenerateMarkdownIndex(records, indexPath)
	if err != nil {
		t.Fatalf("GenerateMarkdownIndex() error = %v", err)
	}

	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("Failed to read markdown file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Error:") {
		t.Error("Markdown should contain 'Error:' label for failed downloads")
	}

	if !strings.Contains(contentStr, "timeout") {
		t.Error("Markdown should contain error message details")
	}
}
