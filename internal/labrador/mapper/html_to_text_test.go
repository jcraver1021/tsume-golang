package mapper_test

import (
	"testing"

	. "tsumegolang/internal/labrador/mapper"
)

func TestHTMLToText(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "unescapes entities",
			content: `<p>Tom &amp; Jerry</p>`,
			want:    "Tom & Jerry\n",
		},
		{
			name:    "breaks on block-level closing tags",
			content: `<h1>Title</h1><p>one</p><p>two</p>`,
			want:    "Title\none\ntwo\n",
		},
		{
			name:    "breaks on br",
			content: `a<br/>b`,
			want:    "a\nb\n",
		},
		{
			name:    "drops script bodies",
			content: `<script>var secret = 1;</script><p>visible</p>`,
			want:    "visible\n",
		},
		{
			name:    "drops style bodies",
			content: `<style>p{color:red}</style><p>visible</p>`,
			want:    "visible\n",
		},
		{
			name:    "drops comments",
			content: `<!-- hidden --><p>visible</p>`,
			want:    "visible\n",
		},
		{
			name:    "collapses runs of blank lines",
			content: `<p>a</p><div></div><div></div><div></div><p>b</p>`,
			want:    "a\n\nb\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := applyOne(t, "html-to-text", Payload{ContentType: "text/html", Content: []byte(tc.content)})
			if string(got.Content) != tc.want {
				t.Errorf("Content = %q, want %q", got.Content, tc.want)
			}
		})
	}
}

func TestHTMLToTextRetypesPayload(t *testing.T) {
	got := applyOne(t, "html-to-text", Payload{ContentType: "text/html", Content: []byte(`<p>hi</p>`)})

	if got.ContentType != "text/plain; charset=utf-8" {
		t.Errorf("ContentType = %q, want text/plain; charset=utf-8", got.ContentType)
	}
	if got.Extension != "txt" {
		t.Errorf("Extension = %q, want txt", got.Extension)
	}
	if KindOf(got.ContentType) != KindText {
		t.Errorf("KindOf(%q) = %q, want %q", got.ContentType, KindOf(got.ContentType), KindText)
	}
}

func TestHTMLToTextSkipsNonHTML(t *testing.T) {
	payload := Payload{ContentType: "application/pdf", Content: []byte("%PDF-1.4")}

	got := applyChain(t, []string{"html-to-text"}, payload)
	if string(got.Content) != "%PDF-1.4" || got.Extension != "" {
		t.Errorf("payload = %q/%q, want it untouched", got.Content, got.Extension)
	}
}
