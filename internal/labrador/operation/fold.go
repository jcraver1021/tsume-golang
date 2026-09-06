package operation

import "sort"

// Group is the records of one section, as laid out by GroupBySection.
type Group struct {
	Section string
	Records []Record
}

func CountOutcomes(records []Record) (succeeded, failed int) {
	for _, record := range records {
		if record.Success {
			succeeded++
		} else {
			failed++
		}
	}
	return succeeded, failed
}

// GroupBySection returns sections in sorted order so that every artifact a run
// produces lays them out identically.
func GroupBySection(records []Record) []Group {
	grouped := make(map[string][]Record)
	for _, record := range records {
		grouped[record.Section] = append(grouped[record.Section], record)
	}

	sections := make([]string, 0, len(grouped))
	for section := range grouped {
		sections = append(sections, section)
	}
	sort.Strings(sections)

	groups := make([]Group, 0, len(sections))
	for _, section := range sections {
		groups = append(groups, Group{Section: section, Records: grouped[section]})
	}
	return groups
}
