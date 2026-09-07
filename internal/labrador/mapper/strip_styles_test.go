package mapper_test

import (
	"cmp"
	"testing"

	. "tsumegolang/internal/labrador/mapper"
)

func TestStripStyles(t *testing.T) {
	testCases := []struct {
		name        string
		contentType string
		content     string
		want        string
	}{
		{
			name:    "removes a style block",
			content: `<style>p{color:red}</style><p>hi</p>`,
			want:    `<p>hi</p>`,
		},
		{
			name:    "removes a style with a media attribute",
			content: `<style media="print">p{color:red}</style>keep`,
			want:    `keep`,
		},
		{
			name:    "leaves scripts in place",
			content: `<style>a{}</style><script>b</script>`,
			want:    `<script>b</script>`,
		},
		{
			name:        "skips a payload that is not HTML",
			contentType: "text/plain",
			content:     `<style>p{}</style>`,
			want:        `<style>p{}</style>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := applyChain(t, []string{"strip-styles"}, Payload{ContentType: cmp.Or(tc.contentType, "text/html"), Content: []byte(tc.content)})
			if string(got.Content) != tc.want {
				t.Errorf("Content = %q, want %q", got.Content, tc.want)
			}
		})
	}
}
