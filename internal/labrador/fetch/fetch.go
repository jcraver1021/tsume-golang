// Package fetch retrieves URLs over HTTP with a retry policy. A Client owns one
// http.Client, so connections are pooled across a whole run.
package fetch

import (
	"errors"
	"net/http"
	"time"
)

const (
	defaultRetryCount = 3
	defaultBackoffMs  = 1000
	defaultTimeout    = 30 * time.Second
)

var (
	ErrUnknown      = errors.New("unknown error")
	ErrRetryable    = errors.New("retryable error")
	ErrNonRetryable = errors.New("non-retryable error")
)

type Result struct {
	Content     []byte
	ContentType string
}

type Client struct {
	http       *http.Client
	transport  *http.Transport
	retryCount int
	backoffMs  int
}

type Option func(*Client)

// New builds a client safe for concurrent use; a worker pool should share one.
func New(options ...Option) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	client := &Client{
		http:       &http.Client{Timeout: defaultTimeout, Transport: transport},
		transport:  transport,
		retryCount: defaultRetryCount,
		backoffMs:  defaultBackoffMs,
	}

	for _, option := range options {
		option(client)
	}

	return client
}

func WithRetryCount(count int) Option {
	return func(client *Client) {
		if count < 1 {
			count = 1
		}
		client.retryCount = count
	}
}

func WithBackoff(backoffMs int) Option {
	return func(client *Client) {
		if backoffMs < 0 {
			backoffMs = 0
		}
		client.backoffMs = backoffMs
	}
}

// WithIdleConnsPerHost should match the number of concurrent callers; the
// net/http default of 2 leaves most workers redialling.
func WithIdleConnsPerHost(count int) Option {
	return func(client *Client) {
		if count < 1 {
			count = 1
		}
		client.transport.MaxIdleConnsPerHost = count
	}
}

func (c *Client) Get(url string) (*Result, error) {
	var lastErr error

	for attempt := range c.retryCount {
		result, err := c.attempt(url)
		if err == nil {
			return result, nil
		}
		lastErr = err

		if errors.Is(err, ErrNonRetryable) {
			break
		}

		if attempt < c.retryCount-1 {
			time.Sleep(time.Duration(c.backoffMs) * time.Millisecond)
		}
	}

	return nil, lastErr
}
