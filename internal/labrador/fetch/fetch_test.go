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

func TestGetRetriesRetryableStatus(t *testing.T) {
	server := newCountingServer(t, func(w http.ResponseWriter, r *http.Request, call int) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := New(WithRetryCount(3), WithBackoff(0)).Get(server.URL)
	if !errors.Is(err, ErrRetryable) {
		t.Fatalf("err = %v, want %v", err, ErrRetryable)
	}
	if got := server.callCount(); got != 3 {
		t.Errorf("requests = %d, want 3", got)
	}
}

func TestGetStopsOnNonRetryableStatus(t *testing.T) {
	server := newCountingServer(t, func(w http.ResponseWriter, r *http.Request, call int) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := New(WithRetryCount(5), WithBackoff(0)).Get(server.URL)
	if !errors.Is(err, ErrNonRetryable) {
		t.Fatalf("err = %v, want %v", err, ErrNonRetryable)
	}
	if got := server.callCount(); got != 1 {
		t.Errorf("requests = %d, want 1 — a 4xx must not be retried", got)
	}
}

func TestGetSucceedsAfterTransientFailure(t *testing.T) {
	server := newCountingServer(t, func(w http.ResponseWriter, r *http.Request, call int) {
		if call < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("recovered"))
	})

	result, err := New(WithRetryCount(3), WithBackoff(0)).Get(server.URL)
	if err != nil {
		t.Fatalf("Get() = %v", err)
	}
	if string(result.Content) != "recovered" {
		t.Errorf("Content = %q, want %q", result.Content, "recovered")
	}
	if got := server.callCount(); got != 3 {
		t.Errorf("requests = %d, want 3", got)
	}
}

// The backoff belongs between attempts, so N attempts sleep N-1 times. Sleeping
// after the final attempt would only delay reporting a failure that has already
// been decided.
func TestGetDoesNotBackOffAfterTheFinalAttempt(t *testing.T) {
	server := newCountingServer(t, func(w http.ResponseWriter, r *http.Request, call int) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	const backoffMs = 60
	start := time.Now()
	if _, err := New(WithRetryCount(3), WithBackoff(backoffMs)).Get(server.URL); err == nil {
		t.Fatal("Get() = nil, want an error")
	}
	elapsed := time.Since(start)

	minimum := 2 * backoffMs * time.Millisecond
	maximum := 3 * backoffMs * time.Millisecond
	if elapsed < minimum {
		t.Errorf("elapsed = %v, want at least %v for two gaps between three attempts", elapsed, minimum)
	}
	if elapsed >= maximum {
		t.Errorf("elapsed = %v, want under %v — it should not sleep after the last attempt", elapsed, maximum)
	}
}

func TestGetHonoursOptionFloors(t *testing.T) {
	server := newCountingServer(t, func(w http.ResponseWriter, r *http.Request, call int) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	// A retry count below one would otherwise mean never attempting at all.
	if _, err := New(WithRetryCount(0), WithBackoff(-5)).Get(server.URL); err == nil {
		t.Fatal("Get() = nil, want an error")
	}
	if got := server.callCount(); got != 1 {
		t.Errorf("requests = %d, want 1", got)
	}
}

func TestClientReusesConnections(t *testing.T) {
	server := newCountingServer(t, func(w http.ResponseWriter, r *http.Request, call int) {
		w.Write([]byte("ok"))
	})

	client := New(WithRetryCount(1))
	for range 5 {
		if _, err := client.Get(server.URL); err != nil {
			t.Fatalf("Get() = %v", err)
		}
	}

	if got := server.connectionCount(); got != 1 {
		t.Errorf("connections = %d, want 1 — a shared client should pool its connections", got)
	}
}

// A failed status still has to leave the connection reusable, which only holds
// if the body is drained before it is closed.
func TestClientReusesConnectionsAcrossFailedStatuses(t *testing.T) {
	server := newCountingServer(t, func(w http.ResponseWriter, r *http.Request, call int) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})

	if _, err := New(WithRetryCount(4), WithBackoff(0)).Get(server.URL); err == nil {
		t.Fatal("Get() = nil, want an error")
	}

	if got := server.connectionCount(); got != 1 {
		t.Errorf("connections = %d, want 1 across 4 attempts", got)
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
