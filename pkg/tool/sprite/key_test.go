package sprite

import (
	"testing"
)

func TestColorKey(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid single ASCII character",
			input:   "a",
			wantErr: false,
		},
		{
			name:    "valid narrow multi-byte character",
			input:   "é", // U+00E9, 2-byte UTF-8, narrow
			wantErr: false,
		},
		{
			name:    "invalid empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid multiple characters",
			input:   "ab",
			wantErr: true,
		},
		{
			name:    "invalid wide character",
			input:   "広", // U+5E83, CJK wide
			wantErr: true,
		},
		{
			name:    "invalid surrogate (malformed UTF-8)",
			input:   "\xed\xa0\x80", // U+D800 encoded as if valid UTF-8
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		_, err := fromString(tc.input)
		if (err != nil) != tc.wantErr {
			t.Errorf("%s: got error %v, wantErr %v", tc.name, err, tc.wantErr)
		}
	}
}
