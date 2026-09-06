package labrador_test

import (
	"os"
	"testing"

	. "tsumegolang/internal/labrador"
)

func TestParseSectionsFromYAML(t *testing.T) {
	testCases := []struct {
		name        string
		yamlContent string
		want        []Section
		wantErr     bool
	}{
		{
			name: "valid sections",
			yamlContent: `"Chapter 1":
  - https://example.com
  - https://go.dev
"Chapter 2":
  - http://test.org`,
			want: []Section{
				{
					Name: "Chapter 1",
					URLs: []string{"https://example.com", "https://go.dev"},
				},
				{
					Name: "Chapter 2",
					URLs: []string{"http://test.org"},
				},
			},
			wantErr: false,
		},
		{
			name: "sections with comments",
			yamlContent: `# This is a comment
"Chapter 1":
  # Another comment
  - https://example.com
  - https://go.dev`,
			want: []Section{
				{
					Name: "Chapter 1",
					URLs: []string{"https://example.com", "https://go.dev"},
				},
			},
			wantErr: false,
		},
		{
			name: "sections with invalid URLs filtered out",
			yamlContent: `"Chapter 1":
  - https://example.com
  - not a url
  - https://go.dev
  - ftp://invalid.com`,
			want: []Section{
				{
					Name: "Chapter 1",
					URLs: []string{"https://example.com", "https://go.dev"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty sections ignored",
			yamlContent: `"Chapter 1":
  - https://example.com
"Chapter 2":
  - not a url
  - also not a url
"Chapter 3":
  - https://go.dev`,
			want: []Section{
				{
					Name: "Chapter 1",
					URLs: []string{"https://example.com"},
				},
				{
					Name: "Chapter 3",
					URLs: []string{"https://go.dev"},
				},
			},
			wantErr: false,
		},
		{
			name:        "empty YAML",
			yamlContent: "{}",
			want:        []Section{},
			wantErr:     false,
		},
		{
			name:        "invalid YAML syntax",
			yamlContent: "invalid: yaml: syntax: here:",
			want:        nil,
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "test-sections-*.yaml")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tc.yamlContent); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tmpFile.Close()

			got, err := ParseSectionsFromYAML(tmpFile.Name())
			if (err != nil) != tc.wantErr {
				t.Errorf("ParseSectionsFromYAML() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if tc.wantErr {
				return
			}

			if len(got) != len(tc.want) {
				t.Errorf("ParseSectionsFromYAML() got %d sections, want %d sections", len(got), len(tc.want))
				return
			}

			gotMap := make(map[string][]string)
			for _, section := range got {
				gotMap[section.Name] = section.URLs
			}

			for _, wantSection := range tc.want {
				gotURLs, exists := gotMap[wantSection.Name]
				if !exists {
					t.Errorf("ParseSectionsFromYAML() missing section %q", wantSection.Name)
					continue
				}

				if len(gotURLs) != len(wantSection.URLs) {
					t.Errorf("ParseSectionsFromYAML() section %q has %d URLs, want %d URLs",
						wantSection.Name, len(gotURLs), len(wantSection.URLs))
					continue
				}

				for i, wantURL := range wantSection.URLs {
					if gotURLs[i] != wantURL {
						t.Errorf("ParseSectionsFromYAML() section %q URL[%d] = %q; want %q",
							wantSection.Name, i, gotURLs[i], wantURL)
					}
				}
			}
		})
	}
}

func TestParseSectionsFromYAML_FileNotFound(t *testing.T) {
	_, err := ParseSectionsFromYAML("/nonexistent/file.yaml")
	if err == nil {
		t.Error("ParseSectionsFromYAML() expected error for nonexistent file, got nil")
	}
}
