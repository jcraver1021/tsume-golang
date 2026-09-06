package operation_test

import (
	"errors"
	"testing"

	. "tsumegolang/internal/labrador/operation"
)

func TestCountOutcomes(t *testing.T) {
	testCases := []struct {
		name          string
		records       []Record
		wantSucceeded int
		wantFailed    int
	}{
		{name: "nil records", records: nil},
		{name: "empty records", records: []Record{}},
		{
			name:          "all successful",
			records:       []Record{{Success: true}, {Success: true}},
			wantSucceeded: 2,
		},
		{
			name:       "all failed",
			records:    []Record{{}, {Error: errors.New("boom")}},
			wantFailed: 2,
		},
		{
			name:          "mixed",
			records:       []Record{{Success: true}, {}, {Success: true}},
			wantSucceeded: 2,
			wantFailed:    1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			succeeded, failed := CountOutcomes(tc.records)
			if succeeded != tc.wantSucceeded || failed != tc.wantFailed {
				t.Errorf("succeeded/failed = %d/%d, want %d/%d", succeeded, failed, tc.wantSucceeded, tc.wantFailed)
			}
			if succeeded+failed != len(tc.records) {
				t.Errorf("succeeded+failed = %d, want %d — every record must be counted once", succeeded+failed, len(tc.records))
			}
		})
	}
}

func TestGroupBySection(t *testing.T) {
	testCases := []struct {
		name         string
		records      []Record
		wantSections []string
		wantCounts   []int
	}{
		{name: "nil records", records: nil, wantSections: []string{}, wantCounts: []int{}},
		{
			name:         "single section",
			records:      []Record{{Section: "Alpha", URL: "a"}, {Section: "Alpha", URL: "b"}},
			wantSections: []string{"Alpha"},
			wantCounts:   []int{2},
		},
		{
			name: "sections are sorted regardless of arrival order",
			records: []Record{
				{Section: "Charlie", URL: "c"},
				{Section: "Alpha", URL: "a"},
				{Section: "Bravo", URL: "b"},
			},
			wantSections: []string{"Alpha", "Bravo", "Charlie"},
			wantCounts:   []int{1, 1, 1},
		},
		{
			name: "nested section names sort as strings",
			records: []Record{
				{Section: "Docs/B", URL: "b"},
				{Section: "Docs/A", URL: "a"},
				{Section: "Docs", URL: "root"},
			},
			wantSections: []string{"Docs", "Docs/A", "Docs/B"},
			wantCounts:   []int{1, 1, 1},
		},
		{
			name: "the empty section is preserved",
			records: []Record{
				{Section: "", URL: "loose"},
				{Section: "Alpha", URL: "a"},
			},
			wantSections: []string{"", "Alpha"},
			wantCounts:   []int{1, 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			groups := GroupBySection(tc.records)

			if len(groups) != len(tc.wantSections) {
				t.Fatalf("groups = %d, want %d", len(groups), len(tc.wantSections))
			}
			for i, group := range groups {
				if group.Section != tc.wantSections[i] {
					t.Errorf("groups[%d].Section = %q, want %q", i, group.Section, tc.wantSections[i])
				}
				if len(group.Records) != tc.wantCounts[i] {
					t.Errorf("groups[%d] holds %d records, want %d", i, len(group.Records), tc.wantCounts[i])
				}
			}
		})
	}
}

// Grouping must not drop or duplicate records, since reducers report counts
// from the flat slice but list entries from the groups.
func TestGroupBySectionPreservesEveryRecord(t *testing.T) {
	records := []Record{
		{Section: "Beta", URL: "b1"},
		{Section: "Alpha", URL: "a1"},
		{Section: "Beta", URL: "b2"},
		{Section: "Alpha", URL: "a2"},
	}

	seen := 0
	for _, group := range GroupBySection(records) {
		seen += len(group.Records)
	}
	if seen != len(records) {
		t.Errorf("grouped records = %d, want %d", seen, len(records))
	}
}

// Within a section, records keep the order they arrived in, so an artifact
// lists a section's URLs in the order the config declared them.
func TestGroupBySectionPreservesOrderWithinASection(t *testing.T) {
	records := []Record{
		{Section: "Alpha", URL: "first"},
		{Section: "Beta", URL: "other"},
		{Section: "Alpha", URL: "second"},
		{Section: "Alpha", URL: "third"},
	}

	groups := GroupBySection(records)
	if groups[0].Section != "Alpha" {
		t.Fatalf("groups[0].Section = %q, want Alpha", groups[0].Section)
	}

	want := []string{"first", "second", "third"}
	for i, record := range groups[0].Records {
		if record.URL != want[i] {
			t.Errorf("Alpha[%d].URL = %q, want %q", i, record.URL, want[i])
		}
	}
}
