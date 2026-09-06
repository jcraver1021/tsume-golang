package reducer_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	. "tsumegolang/internal/labrador/reducer"
)

type decodedManifest struct {
	Generated string `json:"generated"`
	Total     int    `json:"total"`
	Succeeded int    `json:"succeeded"`
	Failed    int    `json:"failed"`
	Sections  []struct {
		Name    string `json:"name"`
		Entries []struct {
			URL      string `json:"url"`
			FilePath string `json:"file_path"`
			Success  bool   `json:"success"`
			Error    string `json:"error"`
		} `json:"entries"`
	} `json:"sections"`
}

func generateManifest(t *testing.T) decodedManifest {
	t.Helper()

	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := GenerateJSONManifest(testRecords(), path); err != nil {
		t.Fatalf("GenerateJSONManifest() = %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("manifest not written: %v", err)
	}

	var got decodedManifest
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("manifest is not valid JSON: %v", err)
	}
	return got
}

func TestGenerateJSONManifestCounts(t *testing.T) {
	got := generateManifest(t)

	if got.Total != 3 || got.Succeeded != 2 || got.Failed != 1 {
		t.Errorf("total/succeeded/failed = %d/%d/%d, want 3/2/1", got.Total, got.Succeeded, got.Failed)
	}
	if got.Generated == "" {
		t.Error("generated timestamp is empty")
	}
}

func TestGenerateJSONManifestSortsSections(t *testing.T) {
	got := generateManifest(t)

	if len(got.Sections) != 2 {
		t.Fatalf("sections = %d, want 2", len(got.Sections))
	}
	if got.Sections[0].Name != "Alpha" || got.Sections[1].Name != "Beta" {
		t.Errorf("sections = %q, %q, want Alpha, Beta", got.Sections[0].Name, got.Sections[1].Name)
	}
}

func TestGenerateJSONManifestRecordsFailures(t *testing.T) {
	got := generateManifest(t)

	failed := got.Sections[0].Entries[1]
	if failed.Success {
		t.Error("failed entry reported as successful")
	}
	if failed.Error != "404" {
		t.Errorf("error = %q, want %q", failed.Error, "404")
	}
	if failed.FilePath != "" {
		t.Errorf("file_path = %q, want it omitted for a failed download", failed.FilePath)
	}
}

func TestGenerateJSONManifestEmptyRecordsIsValidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := GenerateJSONManifest(nil, path); err != nil {
		t.Fatalf("GenerateJSONManifest(nil) = %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("manifest not written: %v", err)
	}

	var got decodedManifest
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("manifest is not valid JSON: %v", err)
	}
	if got.Total != 0 || len(got.Sections) != 0 {
		t.Errorf("total/sections = %d/%d, want 0/0", got.Total, len(got.Sections))
	}
}
