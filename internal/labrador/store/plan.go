package store

import (
	"net/url"
	"strings"
)

const nameSeparator = "_"

// PlanFilenames gives every URL in a section a name unique within it, taking
// more of the URL path until the names stop colliding. Planning up front from
// the config keeps the result independent of the order downloads finish in.
func PlanFilenames(urls []string) map[string]string {
	// Only distinct URLs compete for a name; a URL repeated in a section still
	// names one file.
	segments := make(map[string][]string, len(urls))
	contested := make([]string, 0, len(urls))
	for _, raw := range urls {
		if _, seen := segments[raw]; seen {
			continue
		}
		segments[raw] = nameSegments(raw)
		contested = append(contested, raw)
	}

	planned := make(map[string]string, len(segments))

	for depth := 1; len(contested) > 0; depth++ {
		grouped := make(map[string][]string, len(contested))
		for _, raw := range contested {
			name := tailSegments(segments[raw], depth)
			grouped[name] = append(grouped[name], raw)
		}

		var stillContested []string
		for name, group := range grouped {
			// Once the deepest URL in a group is exhausted, a longer name
			// cannot separate them: they name the same resource.
			if len(group) == 1 || depth >= deepest(segments, group) {
				for _, raw := range group {
					planned[raw] = name
				}
				continue
			}
			stillContested = append(stillContested, group...)
		}
		contested = stillContested
	}

	return planned
}

func nameSegments(raw string) []string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return []string{raw}
	}

	path := strings.Trim(parsed.Path, "/")
	if path == "" {
		return []string{parsed.Host}
	}
	return append([]string{parsed.Host}, strings.Split(path, "/")...)
}

func tailSegments(segments []string, depth int) string {
	depth = min(depth, len(segments))
	return strings.Join(segments[len(segments)-depth:], nameSeparator)
}

func deepest(segments map[string][]string, group []string) int {
	most := 0
	for _, raw := range group {
		most = max(most, len(segments[raw]))
	}
	return most
}
