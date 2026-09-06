package reducer

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"tsumegolang/internal/labrador"
)

var ErrWriteManifest = errors.New("failed to write manifest file")

type manifestEntry struct {
	URL      string `json:"url"`
	FilePath string `json:"file_path,omitempty"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

type manifestSection struct {
	Name    string          `json:"name"`
	Entries []manifestEntry `json:"entries"`
}

type manifest struct {
	Generated string            `json:"generated"`
	Total     int               `json:"total"`
	Succeeded int               `json:"succeeded"`
	Failed    int               `json:"failed"`
	Sections  []manifestSection `json:"sections"`
}

var manifestJSON = Reducer{
	Name: "manifest-json",
	Reduce: func(records []labrador.DownloadRecord, outputDir string) (string, error) {
		if err := ensureOutputDir(outputDir); err != nil {
			return "", err
		}

		manifestPath := filepath.Join(outputDir, "manifest.json")
		if err := GenerateJSONManifest(records, manifestPath); err != nil {
			return "", err
		}
		return fmt.Sprintf("Manifest generated at: %s", manifestPath), nil
	},
}

func GenerateJSONManifest(records []labrador.DownloadRecord, outputPath string) error {
	succeeded, failed := countOutcomes(records)
	result := manifest{
		Generated: time.Now().Format(time.RFC3339),
		Total:     len(records),
		Succeeded: succeeded,
		Failed:    failed,
		Sections:  []manifestSection{},
	}

	bySection, sections := groupBySection(records)
	for _, section := range sections {
		entries := make([]manifestEntry, 0, len(bySection[section]))
		for _, record := range bySection[section] {
			entry := manifestEntry{
				URL:      record.URL,
				FilePath: record.FilePath,
				Success:  record.Success,
			}
			if record.Error != nil {
				entry.Error = record.Error.Error()
			}
			entries = append(entries, entry)
		}
		result.Sections = append(result.Sections, manifestSection{Name: section, Entries: entries})
	}

	encoded, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWriteManifest, err)
	}

	if err := os.WriteFile(outputPath, append(encoded, '\n'), 0644); err != nil {
		return fmt.Errorf("%w: %w", ErrWriteManifest, err)
	}

	return nil
}
