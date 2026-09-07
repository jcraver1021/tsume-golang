package reducer

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"tsumegolang/internal/labrador/operation"
)

var (
	ErrRenderManifest = errors.New("failed to render manifest")
	ErrReadManifest   = errors.New("failed to read manifest")
	ErrParseManifest  = errors.New("failed to parse manifest")
)

const artifactManifest = "manifest.json"

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
	Name:     "manifest-json",
	Artifact: artifactManifest,
	Reduce: func(records []operation.Record, out *Output) (string, error) {
		// Rewriting the source would only restamp it, and a failed write
		// would take it with it.
		if path := out.Path(artifactManifest); out.Source() == path {
			return "", fmt.Errorf("%w: %s is the manifest this run was loaded from", ErrSkipped, path)
		}

		content, err := RenderJSONManifest(records)
		if err != nil {
			return "", err
		}

		path, err := out.Write(artifactManifest, content)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Manifest generated at: %s", path), nil
	},
}

func RenderJSONManifest(records []operation.Record) ([]byte, error) {
	succeeded, failed := operation.CountOutcomes(records)
	result := manifest{
		Generated: time.Now().Format(time.RFC3339),
		Total:     len(records),
		Succeeded: succeeded,
		Failed:    failed,
		Sections:  []manifestSection{},
	}

	for _, group := range operation.GroupBySection(records) {
		entries := make([]manifestEntry, 0, len(group.Records))
		for _, record := range group.Records {
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
		result.Sections = append(result.Sections, manifestSection{Name: group.Section, Entries: entries})
	}

	encoded, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRenderManifest, err)
	}

	return append(encoded, '\n'), nil
}

// LoadJSONManifest reads back an earlier run's records. Errors come back as
// plain values: the manifest keeps their text, not their identity.
func LoadJSONManifest(path string) ([]operation.Record, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrReadManifest, err)
	}
	return ParseJSONManifest(raw)
}

func ParseJSONManifest(raw []byte) ([]operation.Record, error) {
	var decoded manifest
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParseManifest, err)
	}

	records := []operation.Record{}
	for _, section := range decoded.Sections {
		for _, entry := range section.Entries {
			record := operation.Record{
				Section:  section.Name,
				URL:      entry.URL,
				FilePath: entry.FilePath,
				Success:  entry.Success,
			}
			if entry.Error != "" {
				record.Error = errors.New(entry.Error)
			}
			records = append(records, record)
		}
	}

	return records, nil
}
