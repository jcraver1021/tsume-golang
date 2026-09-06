package fetch_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	. "tsumegolang/internal/labrador/fetch"
)

func TestAttemptClassifiesStatus(t *testing.T) {
	testCases := []struct {
		name    string
		status  int
		wantErr error
	}{
		{name: "200 succeeds", status: http.StatusOK},
		{name: "204 succeeds", status: http.StatusNoContent},
		{name: "400 is non-retryable", status: http.StatusBadRequest, wantErr: ErrNonRetryable},
		{name: "404 is non-retryable", status: http.StatusNotFound, wantErr: ErrNonRetryable},
		{name: "429 is non-retryable", status: http.StatusTooManyRequests, wantErr: ErrNonRetryable},
		{name: "500 is retryable", status: http.StatusInternalServerError, wantErr: ErrRetryable},
		{name: "503 is retryable", status: http.StatusServiceUnavailable, wantErr: ErrRetryable},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
			}))
			defer server.Close()

			_, err := New(WithRetryCount(1), WithBackoff(0)).Get(server.URL)

			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("Get() = %v, want success", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestAttemptReturnsContentAndType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{"a":1}`))
	}))
	defer server.Close()

	result, err := New(WithRetryCount(1)).Get(server.URL)
	if err != nil {
		t.Fatalf("Get() = %v", err)
	}

	if string(result.Content) != `{"a":1}` {
		t.Errorf("Content = %q, want %q", result.Content, `{"a":1}`)
	}
	if result.ContentType != "application/json; charset=utf-8" {
		t.Errorf("ContentType = %q, want application/json; charset=utf-8", result.ContentType)
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

func TestAttemptStatusErrorIncludesCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "gone", http.StatusGone)
	}))
	defer server.Close()

	_, err := New(WithRetryCount(1), WithBackoff(0)).Get(server.URL)
	if err == nil {
		t.Fatal("Get() = nil, want an error")
	}
	if got := err.Error(); got != "non-retryable error: 410" {
		t.Errorf("err = %q, want %q", got, "non-retryable error: 410")
	}
}
