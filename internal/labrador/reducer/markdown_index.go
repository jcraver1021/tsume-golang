package reducer

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"tsumegolang/internal/labrador/operation"
)

const artifactIndex = "index.md"

var markdownIndex = Reducer{
	Name:     "markdown-index",
	Artifact: artifactIndex,
	Reduce: func(records []operation.Record, out *Output) (string, error) {
		content := RenderMarkdownIndex(records, out.Path(artifactIndex))

		path, err := out.Write(artifactIndex, content)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Index generated at: %s", path), nil
	},
}

// RenderMarkdownIndex needs outputPath to make each link relative to where the
// index itself will sit.
func RenderMarkdownIndex(records []operation.Record, outputPath string) []byte {
	var sb strings.Builder

	sb.WriteString("# Download Index\n\n")
	fmt.Fprintf(&sb, "Generated: %s\n\n", time.Now().Format(time.RFC1123))

	successCount, failCount := operation.CountOutcomes(records)
	fmt.Fprintf(&sb, "**Total Downloads**: %d | **Successful**: %d | **Failed**: %d\n\n", len(records), successCount, failCount)
	sb.WriteString("---\n\n")

	for _, group := range operation.GroupBySection(records) {
		fmt.Fprintf(&sb, "## %s\n\n", group.Section)

		for _, record := range group.Records {
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

	return []byte(sb.String())
}
