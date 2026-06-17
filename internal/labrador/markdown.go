package labrador

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrWriteMarkdown = fmt.Errorf("failed to write markdown file")
)

type DownloadRecord struct {
	Section  string
	URL      string
	FilePath string
	Success  bool
	Error    error
}

func GenerateMarkdownIndex(records []DownloadRecord, outputPath string) error {
	var sb strings.Builder

	sb.WriteString("# Download Index\n\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC1123)))

	sectionMap := make(map[string][]DownloadRecord)
	for _, record := range records {
		sectionMap[record.Section] = append(sectionMap[record.Section], record)
	}

	successCount := 0
	failCount := 0
	for _, record := range records {
		if record.Success {
			successCount++
		} else {
			failCount++
		}
	}

	sb.WriteString(fmt.Sprintf("**Total Downloads**: %d | **Successful**: %d | **Failed**: %d\n\n", len(records), successCount, failCount))
	sb.WriteString("---\n\n")

	for _, section := range getSortedSections(sectionMap) {
		sb.WriteString(fmt.Sprintf("## %s\n\n", section))

		for _, record := range sectionMap[section] {
			if record.Success {
				relPath, err := filepath.Rel(filepath.Dir(outputPath), record.FilePath)
				if err != nil {
					relPath = record.FilePath
				}
				sb.WriteString(fmt.Sprintf("- [%s](%s)\n", record.URL, relPath))
			} else {
				errorMsg := "unknown error"
				if record.Error != nil {
					errorMsg = record.Error.Error()
				}
				sb.WriteString(fmt.Sprintf("- ❌ %s (Error: %s)\n", record.URL, errorMsg))
			}
		}
		sb.WriteString("\n")
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWriteMarkdown, err)
	}
	defer file.Close()

	_, err = file.WriteString(sb.String())
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWriteMarkdown, err)
	}

	return nil
}

func getSortedSections(sectionMap map[string][]DownloadRecord) []string {
	sections := make([]string, 0, len(sectionMap))
	for section := range sectionMap {
		sections = append(sections, section)
	}
	return sections
}
