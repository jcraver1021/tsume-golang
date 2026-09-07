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

func TestRenderJSONManifest(t *testing.T) {
	testCases := []struct {
		name          string
		records       []operation.Record
		wantTotal     int
		wantSucceeded int
		wantFailed    int
		wantSections  []string
	}{
		{
			name:          "counts and sorted sections",
			records:       testRecords(),
			wantTotal:     3,
			wantSucceeded: 2,
			wantFailed:    1,
			wantSections:  []string{"Alpha", "Beta"},
		},
		{
			name:         "no records is still valid JSON",
			records:      nil,
			wantSections: []string{},
		},
		{
			name:          "one successful record",
			records:       []operation.Record{{Section: "Only", URL: "u", FilePath: "f", Success: true}},
			wantTotal:     1,
			wantSucceeded: 1,
			wantSections:  []string{"Only"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := RenderJSONManifest(tc.records)
			if err != nil {
				t.Fatalf("RenderJSONManifest() = %v", err)
			}

			var got decodedManifest
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("manifest is not valid JSON: %v", err)
			}

			if got.Total != tc.wantTotal || got.Succeeded != tc.wantSucceeded || got.Failed != tc.wantFailed {
				t.Errorf("total/succeeded/failed = %d/%d/%d, want %d/%d/%d",
					got.Total, got.Succeeded, got.Failed, tc.wantTotal, tc.wantSucceeded, tc.wantFailed)
			}
			if got.Generated == "" {
				t.Error("generated timestamp is empty")
			}

			sections := make([]string, len(got.Sections))
			for i, section := range got.Sections {
				sections[i] = section.Name
			}
			if strings.Join(sections, ",") != strings.Join(tc.wantSections, ",") {
				t.Errorf("sections = %v, want %v", sections, tc.wantSections)
			}
		})
	}
}

func TestRenderJSONManifestRecordsFailures(t *testing.T) {
	raw, err := RenderJSONManifest(testRecords())
	if err != nil {
		t.Fatalf("RenderJSONManifest() = %v", err)
	}

	var got decodedManifest
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("manifest is not valid JSON: %v", err)
	}

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

func TestManifestRoundTrip(t *testing.T) {
	testCases := []struct {
		name    string
		records []operation.Record
	}{
		{name: "no records", records: nil},
		{name: "successes and a failure", records: testRecords()},
		{
			name: "a record with every field set",
			records: []operation.Record{
				{Section: "S/Nested", URL: "https://x/y", FilePath: "/abs/y.html", Success: true},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := RenderJSONManifest(tc.records)
			if err != nil {
				t.Fatalf("RenderJSONManifest() = %v", err)
			}

			got, err := ParseJSONManifest(raw)
			if err != nil {
				t.Fatalf("ParseJSONManifest() = %v", err)
			}
			if len(got) != len(tc.records) {
				t.Fatalf("got %d records, want %d", len(got), len(tc.records))
			}

			// Rendering groups by section, so compare as a set keyed by URL.
			byURL := map[string]operation.Record{}
			for _, record := range got {
				byURL[record.URL] = record
			}

			for _, want := range tc.records {
				record, ok := byURL[want.URL]
				if !ok {
					t.Errorf("URL %q missing after the round trip", want.URL)
					continue
				}
				if record.Section != want.Section || record.FilePath != want.FilePath || record.Success != want.Success {
					t.Errorf("record for %q = %+v, want %+v", want.URL, record, want)
				}
				if errText(record.Error) != errText(want.Error) {
					t.Errorf("error for %q = %q, want %q", want.URL, errText(record.Error), errText(want.Error))
				}
			}
		})
	}
}

func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func TestLoadJSONManifest(t *testing.T) {
	testCases := []struct {
		name    string
		write   string
		absent  bool
		wantErr error
	}{
		{name: "missing file", absent: true, wantErr: ErrReadManifest},
		{name: "malformed JSON", write: "{not json", wantErr: ErrParseManifest},
		{name: "empty file", write: "", wantErr: ErrParseManifest},
		{name: "valid manifest", write: `{"sections":[{"name":"S","entries":[{"url":"u","success":true}]}]}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "manifest.json")
			if !tc.absent {
				if err := os.WriteFile(path, []byte(tc.write), 0644); err != nil {
					t.Fatalf("setting up: %v", err)
				}
			}

			_, err := LoadJSONManifest(path)
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("LoadJSONManifest() = %v", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestLoadJSONManifestReadsWhatTheReducerWrote(t *testing.T) {
	outputDir := t.TempDir()

	if _, err := Run(mustResolve(t, "manifest-json"), testRecords(), Options{OutputDir: outputDir}); err != nil {
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

func TestManifestJSONSourceHandling(t *testing.T) {
	testCases := []struct {
		name        string
		sourceIsOwn bool
		wantSkip    bool
	}{
		// Rewriting the source would only restamp it, and a failed write would
		// take it with it.
		{name: "declines to rewrite the manifest it was loaded from", sourceIsOwn: true, wantSkip: true},
		{name: "writes when the source is elsewhere"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outputDir := t.TempDir()
			artifact := filepath.Join(outputDir, "manifest.json")

			source := filepath.Join(t.TempDir(), "manifest.json")
			if tc.sourceIsOwn {
				source = artifact
			}

			report, err := Run(mustResolve(t, "manifest-json"), testRecords(), Options{OutputDir: outputDir, Source: source})
			if err != nil {
				t.Fatalf("Run() = %v", err)
			}

			if tc.wantSkip {
				if len(report.Skips) != 1 || !strings.Contains(report.Skips[0], "loaded from") {
					t.Fatalf("skips = %v, want one explaining the source", report.Skips)
				}
				if _, err := os.Stat(artifact); !os.IsNotExist(err) {
					t.Error("the source manifest should not have been written")
				}
				return
			}

			if len(report.Summaries) != 1 || len(report.Skips) != 0 {
				t.Errorf("report = %+v, want one summary and no skips", report)
			}
			if _, err := os.Stat(artifact); err != nil {
				t.Errorf("manifest not written: %v", err)
			}
		})
	}
}
