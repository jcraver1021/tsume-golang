package mapper_test

import (
	"testing"

	. "tsumegolang/internal/labrador/mapper"
)

func TestStripScripts(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "removes a script with attributes",
			content: `<p>hi</p><script src="x.js" defer>alert(1)</script><p>bye</p>`,
			want:    `<p>hi</p><p>bye</p>`,
		},
		{
			name:    "removes multiple scripts",
			content: `<script>a</script>keep<script>b</script>`,
			want:    `keep`,
		},
		{
			name:    "spans newlines",
			content: "<script>\nvar x = 1;\n</script><p>hi</p>",
			want:    `<p>hi</p>`,
		},
		{
			name:    "tolerates whitespace in the closing tag",
			content: `<script>a</script >keep`,
			want:    `keep`,
		},
		{
			name:    "leaves markup without scripts alone",
			content: `<p>plain</p>`,
			want:    `<p>plain</p>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := applyOne(t, "strip-scripts", Payload{ContentType: "text/html", Content: []byte(tc.content)})
			if string(got.Content) != tc.want {
				t.Errorf("Content = %q, want %q", got.Content, tc.want)
			}
			if got.ContentType != "text/html" {
				t.Errorf("ContentType = %q, want it unchanged", got.ContentType)
			}
			if got.Extension != "" {
				t.Errorf("Extension = %q, want it unset", got.Extension)
			}
		})
	}
}

func TestStripScriptsSkipsNonHTML(t *testing.T) {
	payload := Payload{ContentType: "application/json", Content: []byte(`{"a":"<script>x</script>"}`)}

	got := applyChain(t, []string{"strip-scripts"}, payload)
	if string(got.Content) != string(payload.Content) {
		t.Errorf("Content = %q, want it untouched", got.Content)
	}
}
