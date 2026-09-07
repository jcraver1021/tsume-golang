package fetch_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	. "tsumegolang/internal/labrador/fetch"
)

// countingServer reports how many requests it received and which client
// connections they arrived on.
type countingServer struct {
	*httptest.Server
	mu      sync.Mutex
	calls   int
	remotes map[string]int
}

func newCountingServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request, call int)) *countingServer {
	t.Helper()

	cs := &countingServer{remotes: map[string]int{}}
	cs.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cs.mu.Lock()
		cs.calls++
		call := cs.calls
		cs.remotes[r.RemoteAddr]++
		cs.mu.Unlock()

		handler(w, r, call)
	}))
	t.Cleanup(cs.Close)

	return cs
}

func (cs *countingServer) callCount() int {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.calls
}

func (cs *countingServer) connectionCount() int {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return len(cs.remotes)
}

func TestGetRetryPolicy(t *testing.T) {
	testCases := []struct {
		name        string
		failUntil   int // fail every call before this one; 0 fails them all
		status      int
		retryCount  int
		backoffMs   int
		wantCalls   int
		wantErr     error
		wantContent string
	}{
		{
			name:       "a retryable status is retried up to the limit",
			status:     http.StatusInternalServerError,
			retryCount: 3,
			wantCalls:  3,
			wantErr:    ErrRetryable,
		},
		{
			name:       "a non-retryable status stops immediately",
			status:     http.StatusNotFound,
			retryCount: 5,
			wantCalls:  1,
			wantErr:    ErrNonRetryable,
		},
		{
			name:        "a transient failure recovers",
			failUntil:   3,
			status:      http.StatusServiceUnavailable,
			retryCount:  3,
			wantCalls:   3,
			wantContent: "recovered",
		},
		{
			name:        "the first attempt can simply succeed",
			failUntil:   1,
			retryCount:  3,
			wantCalls:   1,
			wantContent: "recovered",
		},
		{
			name:       "a retry count below one still attempts once",
			status:     http.StatusInternalServerError,
			retryCount: 0,
			backoffMs:  -5,
			wantCalls:  1,
			wantErr:    ErrRetryable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := newCountingServer(t, func(w http.ResponseWriter, r *http.Request, call int) {
				if tc.failUntil > 0 && call >= tc.failUntil {
					w.Write([]byte("recovered"))
					return
				}
				w.WriteHeader(tc.status)
			})

			result, err := New(WithRetryCount(tc.retryCount), WithBackoff(tc.backoffMs)).Get(server.URL)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
			} else {
				if err != nil {
					t.Fatalf("Get() = %v", err)
				}
				if string(result.Content) != tc.wantContent {
					t.Errorf("Content = %q, want %q", result.Content, tc.wantContent)
				}
			}

			if got := server.callCount(); got != tc.wantCalls {
				t.Errorf("requests = %d, want %d", got, tc.wantCalls)
			}
		})
	}
}

func TestClientPoolsConnections(t *testing.T) {
	testCases := []struct {
		name       string
		status     int
		retryCount int
		requests   int
	}{
		{name: "across successful requests", status: http.StatusOK, retryCount: 1, requests: 5},
		// Only holds if a failed response body is drained before being closed.
		{name: "across retried failures", status: http.StatusInternalServerError, retryCount: 4, requests: 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := newCountingServer(t, func(w http.ResponseWriter, r *http.Request, call int) {
				http.Error(w, "body", tc.status)
			})

			client := New(WithRetryCount(tc.retryCount), WithBackoff(0))
			for range tc.requests {
				client.Get(server.URL)
			}

			if got := server.connectionCount(); got != 1 {
				t.Errorf("connections = %d, want 1", got)
			}
		})
	}
}

// The backoff belongs between attempts, so N attempts sleep N-1 times.
func TestGetBacksOffBetweenAttemptsOnly(t *testing.T) {
	const backoffMs = 60

	testCases := []struct {
		name       string
		retryCount int
		wantGaps   int
	}{
		{name: "a single attempt never sleeps", retryCount: 1, wantGaps: 0},
		{name: "two attempts sleep once", retryCount: 2, wantGaps: 1},
		{name: "three attempts sleep twice", retryCount: 3, wantGaps: 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := newCountingServer(t, func(w http.ResponseWriter, r *http.Request, call int) {
				w.WriteHeader(http.StatusInternalServerError)
			})

			start := time.Now()
			if _, err := New(WithRetryCount(tc.retryCount), WithBackoff(backoffMs)).Get(server.URL); err == nil {
				t.Fatal("Get() = nil, want an error")
			}
			elapsed := time.Since(start)

			if minimum := time.Duration(tc.wantGaps) * backoffMs * time.Millisecond; elapsed < minimum {
				t.Errorf("elapsed = %v, want at least %v for %d gaps", elapsed, minimum, tc.wantGaps)
			}
			if maximum := time.Duration(tc.wantGaps+1) * backoffMs * time.Millisecond; elapsed >= maximum {
				t.Errorf("elapsed = %v, want under %v — it should not sleep after the last attempt", elapsed, maximum)
			}
		})
	}
}

func TestClientIsSafeForConcurrentUse(t *testing.T) {
	server := newCountingServer(t, func(w http.ResponseWriter, r *http.Request, call int) {
		w.Write([]byte("ok"))
	})

	client := New(WithRetryCount(1), WithIdleConnsPerHost(8))

	var wg sync.WaitGroup
	errs := make(chan error, 24)
	for range 24 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := client.Get(server.URL); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent Get() = %v", err)
	}
	if got := server.callCount(); got != 24 {
		t.Errorf("requests = %d, want 24", got)
	}
}
