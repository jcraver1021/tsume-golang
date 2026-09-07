package reducer_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tsumegolang/internal/labrador/operation"
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

	raw, err := RenderJSONManifest(testRecords())
	if err != nil {
		t.Fatalf("RenderJSONManifest() = %v", err)
	}

	var got decodedManifest
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("manifest is not valid JSON: %v", err)
	}
	return got
}

func TestRenderJSONManifestCounts(t *testing.T) {
	got := generateManifest(t)

	if got.Total != 3 || got.Succeeded != 2 || got.Failed != 1 {
		t.Errorf("total/succeeded/failed = %d/%d/%d, want 3/2/1", got.Total, got.Succeeded, got.Failed)
	}
	if got.Generated == "" {
		t.Error("generated timestamp is empty")
	}
}

func TestRenderJSONManifestSortsSections(t *testing.T) {
	got := generateManifest(t)

	if len(got.Sections) != 2 {
		t.Fatalf("sections = %d, want 2", len(got.Sections))
	}
	if got.Sections[0].Name != "Alpha" || got.Sections[1].Name != "Beta" {
		t.Errorf("sections = %q, %q, want Alpha, Beta", got.Sections[0].Name, got.Sections[1].Name)
	}
}

func TestRenderJSONManifestRecordsFailures(t *testing.T) {
	got := generateManifest(t)

	failed := got.Sections[0].Entries[1]
	if failed.Success {
		t.Error("failed entry reported as successful")
	}
	if failed.Error != "404" {
		t.Errorf("error = %q, want %q", failed.Error, "404")
	}
	if failed.FilePath != "" {
		t.Errorf("file_path = %q, want it omitted for a failed record", failed.FilePath)
	}
}

func TestRenderJSONManifestEmptyRecordsIsValidJSON(t *testing.T) {
	raw, err := RenderJSONManifest(nil)
	if err != nil {
		t.Fatalf("RenderJSONManifest(nil) = %v", err)
	}

	var got decodedManifest
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("manifest is not valid JSON: %v", err)
	}
	if got.Total != 0 || len(got.Sections) != 0 {
		t.Errorf("total/sections = %d/%d, want 0/0", got.Total, len(got.Sections))
	}
}

func TestManifestRoundTrip(t *testing.T) {
	original := testRecords()

	raw, err := RenderJSONManifest(original)
	if err != nil {
		t.Fatalf("RenderJSONManifest() = %v", err)
	}

	got, err := ParseJSONManifest(raw)
	if err != nil {
		t.Fatalf("ParseJSONManifest() = %v", err)
	}

	if len(got) != len(original) {
		t.Fatalf("got %d records, want %d", len(got), len(original))
	}

	// Rendering groups by section, so compare as a set keyed by URL.
	byURL := map[string]operation.Record{}
	for _, record := range got {
		byURL[record.URL] = record
	}

	for _, want := range original {
		record, ok := byURL[want.URL]
		if !ok {
			t.Errorf("URL %q missing after the round trip", want.URL)
			continue
		}
		if record.Section != want.Section || record.FilePath != want.FilePath || record.Success != want.Success {
			t.Errorf("record for %q = %+v, want %+v", want.URL, record, want)
		}

		switch {
		case want.Error == nil && record.Error != nil:
			t.Errorf("record for %q gained error %v", want.URL, record.Error)
		case want.Error != nil && record.Error == nil:
			t.Errorf("record for %q lost its error", want.URL)
		case want.Error != nil && record.Error.Error() != want.Error.Error():
			t.Errorf("error for %q = %q, want %q", want.URL, record.Error, want.Error)
		}
	}
}

func TestParseJSONManifestEmptyManifest(t *testing.T) {
	raw, err := RenderJSONManifest(nil)
	if err != nil {
		t.Fatalf("RenderJSONManifest(nil) = %v", err)
	}

	got, err := ParseJSONManifest(raw)
	if err != nil {
		t.Fatalf("ParseJSONManifest() = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d records, want none", len(got))
	}
}

func TestLoadJSONManifestErrors(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		_, err := LoadJSONManifest(filepath.Join(t.TempDir(), "absent.json"))
		if !errors.Is(err, ErrReadManifest) {
			t.Fatalf("err = %v, want %v", err, ErrReadManifest)
		}
	})

	t.Run("malformed JSON", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "manifest.json")
		if err := os.WriteFile(path, []byte("{not json"), 0644); err != nil {
			t.Fatalf("setting up: %v", err)
		}

		_, err := LoadJSONManifest(path)
		if !errors.Is(err, ErrParseManifest) {
			t.Fatalf("err = %v, want %v", err, ErrParseManifest)
		}
	})
}

func TestLoadJSONManifestReadsWhatTheReducerWrote(t *testing.T) {
	outputDir := t.TempDir()

	reducers, err := Resolve([]string{"manifest-json"})
	if err != nil {
		t.Fatalf("Resolve() = %v", err)
	}
	if _, err := Run(reducers, testRecords(), Options{OutputDir: outputDir}); err != nil {
		t.Fatalf("Run() = %v", err)
	}

	got, err := LoadJSONManifest(filepath.Join(outputDir, "manifest.json"))
	if err != nil {
		t.Fatalf("LoadJSONManifest() = %v", err)
	}
	if len(got) != len(testRecords()) {
		t.Errorf("got %d records, want %d", len(got), len(testRecords()))
	}
}

// Re-applying manifest-json to the manifest it was loaded from would only
// restamp the file, and a failed write would take the source with it.
func TestManifestJSONDeclinesToRewriteItsOwnSource(t *testing.T) {
	outputDir := t.TempDir()
	source := filepath.Join(outputDir, "manifest.json")

	reducers, err := Resolve([]string{"manifest-json"})
	if err != nil {
		t.Fatalf("Resolve() = %v", err)
	}

	report, err := Run(reducers, testRecords(), Options{OutputDir: outputDir, Source: source})
	if err != nil {
		t.Fatalf("Run() = %v, want a skip rather than a failure", err)
	}
	if len(report.Summaries) != 0 {
		t.Errorf("summaries = %v, want none", report.Summaries)
	}
	if len(report.Skips) != 1 || !strings.Contains(report.Skips[0], "loaded from") {
		t.Fatalf("skips = %v, want one explaining the source", report.Skips)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Error("the source manifest should not have been written")
	}
}

func TestManifestJSONWritesWhenTheSourceIsElsewhere(t *testing.T) {
	outputDir := t.TempDir()

	reducers, err := Resolve([]string{"manifest-json"})
	if err != nil {
		t.Fatalf("Resolve() = %v", err)
	}

	report, err := Run(reducers, testRecords(), Options{
		OutputDir: outputDir,
		Source:    filepath.Join(t.TempDir(), "manifest.json"),
	})
	if err != nil {
		t.Fatalf("Run() = %v", err)
	}
	if len(report.Summaries) != 1 || len(report.Skips) != 0 {
		t.Errorf("report = %+v, want one summary and no skips", report)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "manifest.json")); err != nil {
		t.Errorf("manifest not written: %v", err)
	}
}
