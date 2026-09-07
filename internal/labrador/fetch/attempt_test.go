package fetch_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	. "tsumegolang/internal/labrador/fetch"
)

func TestAttempt(t *testing.T) {
	testCases := []struct {
		name         string
		status       int
		body         string
		contentType  string
		suppressType bool // net/http sniffs a type unless the header is nil
		wantErr      error
		wantMessage  string
		wantContent  string
		wantType     string
	}{
		{
			name:        "200 returns the body and its type",
			status:      http.StatusOK,
			body:        `{"a":1}`,
			contentType: "application/json; charset=utf-8",
			wantContent: `{"a":1}`,
			wantType:    "application/json; charset=utf-8",
		},
		{
			name:   "204 succeeds with no body",
			status: http.StatusNoContent,
		},
		{
			name:         "an absent Content-Type comes back empty",
			status:       http.StatusOK,
			body:         "plain",
			suppressType: true,
			wantContent:  "plain",
		},
		{
			name:        "a sniffed Content-Type is passed through",
			status:      http.StatusOK,
			body:        "plain",
			wantContent: "plain",
			wantType:    "text/plain; charset=utf-8",
		},
		{name: "400 is non-retryable", status: http.StatusBadRequest, wantErr: ErrNonRetryable, wantMessage: "non-retryable error: 400"},
		{name: "404 is non-retryable", status: http.StatusNotFound, wantErr: ErrNonRetryable, wantMessage: "non-retryable error: 404"},
		{name: "410 is non-retryable", status: http.StatusGone, wantErr: ErrNonRetryable, wantMessage: "non-retryable error: 410"},
		{name: "429 is non-retryable", status: http.StatusTooManyRequests, wantErr: ErrNonRetryable, wantMessage: "non-retryable error: 429"},
		{name: "500 is retryable", status: http.StatusInternalServerError, wantErr: ErrRetryable, wantMessage: "retryable error: 500"},
		{name: "503 is retryable", status: http.StatusServiceUnavailable, wantErr: ErrRetryable, wantMessage: "retryable error: 503"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case tc.suppressType:
					w.Header()["Content-Type"] = nil
				case tc.contentType != "":
					w.Header().Set("Content-Type", tc.contentType)
				}
				w.WriteHeader(tc.status)
				w.Write([]byte(tc.body))
			}))
			defer server.Close()

			result, err := New(WithRetryCount(1), WithBackoff(0)).Get(server.URL)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
				if err.Error() != tc.wantMessage {
					t.Errorf("err = %q, want %q", err, tc.wantMessage)
				}
				return
			}

			if err != nil {
				t.Fatalf("Get() = %v", err)
			}
			if string(result.Content) != tc.wantContent {
				t.Errorf("Content = %q, want %q", result.Content, tc.wantContent)
			}
			if result.ContentType != tc.wantType {
				t.Errorf("ContentType = %q, want %q", result.ContentType, tc.wantType)
			}
		})
	}
}

func TestAttemptWrapsTransportErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	unreachable := server.URL
	server.Close()

	_, err := New(WithRetryCount(1), WithBackoff(0)).Get(unreachable)
	if !errors.Is(err, ErrUnknown) {
		t.Fatalf("err = %v, want %v", err, ErrUnknown)
	}
}
