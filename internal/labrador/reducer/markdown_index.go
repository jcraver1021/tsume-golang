package reducer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tsumegolang/internal/labrador"
)

var ErrWriteMarkdown = errors.New("failed to write markdown file")

var markdownIndex = Reducer{
	Name: "markdown-index",
	Reduce: func(records []labrador.DownloadRecord, outputDir string) (string, error) {
		if err := ensureOutputDir(outputDir); err != nil {
			return "", err
		}

		indexPath := filepath.Join(outputDir, "index.md")
		if err := GenerateMarkdownIndex(records, indexPath); err != nil {
			return "", err
		}
		return fmt.Sprintf("Index generated at: %s", indexPath), nil
	},
}

func GenerateMarkdownIndex(records []labrador.DownloadRecord, outputPath string) error {
	var sb strings.Builder

	sb.WriteString("# Download Index\n\n")
	fmt.Fprintf(&sb, "Generated: %s\n\n", time.Now().Format(time.RFC1123))

	successCount, failCount := countOutcomes(records)
	fmt.Fprintf(&sb, "**Total Downloads**: %d | **Successful**: %d | **Failed**: %d\n\n", len(records), successCount, failCount)
	sb.WriteString("---\n\n")

	bySection, sections := groupBySection(records)
	for _, section := range sections {
		fmt.Fprintf(&sb, "## %s\n\n", section)

		for _, record := range bySection[section] {
			if record.Success {
				relPath, err := filepath.Rel(filepath.Dir(outputPath), record.FilePath)
				if err != nil {
					relPath = record.FilePath
				}
				fmt.Fprintf(&sb, "- [%s](%s)\n", record.URL, relPath)
			} else {
				errorMsg := "unknown error"
				if record.Error != nil {
					errorMsg = record.Error.Error()
				}
				fmt.Fprintf(&sb, "- ❌ %s (Error: %s)\n", record.URL, errorMsg)
			}
		}
		sb.WriteString("\n")
	}

	if err := os.WriteFile(outputPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("%w: %w", ErrWriteMarkdown, err)
	}

	return nil
}
