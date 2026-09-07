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
		wantOrder    map[string][]string // URLs a section must list, in order
	}{
		{name: "nil records", records: nil, wantSections: []string{}, wantCounts: []int{}},
		{
			name:         "single section",
			records:      []Record{{Section: "Alpha", URL: "a"}, {Section: "Alpha", URL: "b"}},
			wantSections: []string{"Alpha"},
			wantCounts:   []int{2},
		},
		{
			name: "records keep their arrival order within a section",
			records: []Record{
				{Section: "Alpha", URL: "first"},
				{Section: "Beta", URL: "other"},
				{Section: "Alpha", URL: "second"},
				{Section: "Alpha", URL: "third"},
			},
			wantSections: []string{"Alpha", "Beta"},
			wantCounts:   []int{3, 1},
			wantOrder:    map[string][]string{"Alpha": {"first", "second", "third"}},
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

			grouped := 0
			for i, group := range groups {
				grouped += len(group.Records)

				if group.Section != tc.wantSections[i] {
					t.Errorf("groups[%d].Section = %q, want %q", i, group.Section, tc.wantSections[i])
				}
				if len(group.Records) != tc.wantCounts[i] {
					t.Errorf("groups[%d] holds %d records, want %d", i, len(group.Records), tc.wantCounts[i])
				}
				for j, record := range group.Records {
					if want := tc.wantOrder[group.Section]; want != nil && record.URL != want[j] {
						t.Errorf("%s[%d].URL = %q, want %q", group.Section, j, record.URL, want[j])
					}
				}
			}
			if grouped != len(tc.records) {
				t.Errorf("grouped %d records, want all %d", grouped, len(tc.records))
			}
		})
	}
}
