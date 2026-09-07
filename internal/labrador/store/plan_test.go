package store_test

import (
	"strings"
	"testing"

	. "tsumegolang/internal/labrador/store"
)

func TestPlanFilenames(t *testing.T) {
	testCases := []struct {
		name string
		urls []string
		want map[string]string
	}{
		{
			name: "distinct last segments keep their short names",
			urls: []string{"https://go.dev/a/one", "https://go.dev/b/two"},
			want: map[string]string{
				"https://go.dev/a/one": "one",
				"https://go.dev/b/two": "two",
			},
		},
		{
			name: "a collision takes one more path segment",
			urls: []string{"https://x.com/a/index.html", "https://x.com/b/index.html"},
			want: map[string]string{
				"https://x.com/a/index.html": "a_index.html",
				"https://x.com/b/index.html": "b_index.html",
			},
		},
		{
			name: "only the colliding URLs are lengthened",
			urls: []string{"https://x.com/a/index.html", "https://x.com/b/index.html", "https://x.com/c/other.html"},
			want: map[string]string{
				"https://x.com/a/index.html": "a_index.html",
				"https://x.com/b/index.html": "b_index.html",
				"https://x.com/c/other.html": "other.html",
			},
		},
		{
			name: "it keeps going until the names separate",
			urls: []string{"https://x.com/v1/docs/api", "https://x.com/v2/docs/api"},
			want: map[string]string{
				"https://x.com/v1/docs/api": "v1_docs_api",
				"https://x.com/v2/docs/api": "v2_docs_api",
			},
		},
		{
			name: "the host separates identical paths",
			urls: []string{"https://a.com/docs/api", "https://b.com/docs/api"},
			want: map[string]string{
				"https://a.com/docs/api": "a.com_docs_api",
				"https://b.com/docs/api": "b.com_docs_api",
			},
		},
		{
			name: "a shorter path is separated by the longer one growing",
			urls: []string{"https://x.com/index.html", "https://x.com/a/index.html"},
			want: map[string]string{
				"https://x.com/index.html":   "x.com_index.html",
				"https://x.com/a/index.html": "a_index.html",
			},
		},
		{
			name: "a root URL names itself after its host",
			urls: []string{"https://go.dev"},
			want: map[string]string{"https://go.dev": "go.dev"},
		},
		{
			name: "roots of different hosts do not collide",
			urls: []string{"https://a.com", "https://b.com"},
			want: map[string]string{"https://a.com": "a.com", "https://b.com": "b.com"},
		},
		{
			name: "the same URL twice cannot be separated",
			urls: []string{"https://x.com/a", "https://x.com/a"},
			want: map[string]string{"https://x.com/a": "a"},
		},
		{name: "no URLs", urls: nil, want: map[string]string{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := PlanFilenames(tc.urls)

			if len(got) != len(tc.want) {
				t.Fatalf("planned %d names, want %d: %v", len(got), len(tc.want), got)
			}
			for url, want := range tc.want {
				if got[url] != want {
					t.Errorf("%s -> %q, want %q", url, got[url], want)
				}
			}
		})
	}
}

// Names must not depend on map iteration order, or a re-run would shuffle them.
func TestPlanFilenamesIsDeterministic(t *testing.T) {
	urls := []string{
		"https://x.com/a/index.html",
		"https://x.com/b/index.html",
		"https://x.com/c/index.html",
		"https://y.com/a/index.html",
		"https://x.com/a/other.html",
	}

	first := PlanFilenames(urls)
	for range 50 {
		again := PlanFilenames(urls)
		for url, name := range first {
			if again[url] != name {
				t.Fatalf("%s -> %q then %q", url, name, again[url])
			}
		}
	}
}

// Whatever the shape of the input, no two distinct URLs may share a name.
func TestPlanFilenamesAreUnique(t *testing.T) {
	testCases := []struct {
		name string
		urls []string
	}{
		{name: "deep shared prefixes", urls: []string{"https://x.com/a/b/c/d", "https://x.com/a/b/c/e", "https://x.com/z/b/c/d"}},
		{name: "same leaf everywhere", urls: []string{"https://x.com/1/i", "https://x.com/2/i", "https://x.com/3/i", "https://y.com/1/i"}},
		{name: "mixed depths", urls: []string{"https://x.com/i", "https://x.com/a/i", "https://x.com/a/b/i", "https://x.com/a/b/c/i"}},
		{name: "roots and paths", urls: []string{"https://x.com", "https://x.com/x.com", "https://y.com/x.com"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			planned := PlanFilenames(tc.urls)

			owners := map[string]string{}
			for url, name := range planned {
				if owner, taken := owners[name]; taken {
					t.Errorf("%q is used by both %s and %s", name, owner, url)
				}
				owners[name] = url
			}
			if len(planned) != len(tc.urls) {
				t.Errorf("planned %d names for %d URLs", len(planned), len(tc.urls))
			}
			for url, name := range planned {
				if strings.TrimSpace(name) == "" {
					t.Errorf("%s got an empty name", url)
				}
			}
		})
	}
}
