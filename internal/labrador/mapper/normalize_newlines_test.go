package mapper_test

import (
	"testing"

	. "tsumegolang/internal/labrador/mapper"
)

func TestNormalizeNewlines(t *testing.T) {
	testCases := []struct {
		name        string
		contentType string
		content     string
		want        string
	}{
		{
			name:        "rewrites CRLF in plain text",
			contentType: "text/plain",
			content:     "a\r\nb",
			want:        "a\nb",
		},
		{
			name:        "rewrites a lone CR",
			contentType: "text/plain",
			content:     "a\rb",
			want:        "a\nb",
		},
		{
			name:        "applies to JSON",
			contentType: "application/json",
			content:     "{\r\n\"a\": 1\r\n}",
			want:        "{\n\"a\": 1\n}",
		},
		{
			name:        "applies to XML",
			contentType: "application/xml",
			content:     "<a>\r\n</a>",
			want:        "<a>\n</a>",
		},
		{
			name:        "leaves LF-only content alone",
			contentType: "text/markdown",
			content:     "a\nb",
			want:        "a\nb",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := applyChain(t, []string{"normalize-newlines"}, Payload{ContentType: tc.contentType, Content: []byte(tc.content)})
			if string(got.Content) != tc.want {
				t.Errorf("Content = %q, want %q", got.Content, tc.want)
			}
		})
	}
}

func TestNormalizeNewlinesSkipsBinary(t *testing.T) {
	payload := Payload{ContentType: "image/png", Content: []byte("a\r\nb")}

	got := applyChain(t, []string{"normalize-newlines"}, payload)
	if string(got.Content) != "a\r\nb" {
		t.Errorf("Content = %q, want it untouched", got.Content)
	}
}
