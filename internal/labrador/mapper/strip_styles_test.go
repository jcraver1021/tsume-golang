package mapper_test

import (
	"testing"

	. "tsumegolang/internal/labrador/mapper"
)

func TestStripStyles(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		want    string
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := applyOne(t, "strip-styles", Payload{ContentType: "text/html", Content: []byte(tc.content)})
			if string(got.Content) != tc.want {
				t.Errorf("Content = %q, want %q", got.Content, tc.want)
			}
		})
	}
}
