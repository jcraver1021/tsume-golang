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
			name:    "valid single character",
			input:   "a",
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
	}

	for _, tc := range testCases {
		_, err := fromString(tc.input)
		if (err != nil) != tc.wantErr {
			t.Errorf("%s: got error %v, wantErr %v", tc.name, err, tc.wantErr)
		}
	}
}
