package fetch

import (
	"fmt"
	"io"
)

// attempt performs one round trip, classifying the status for Get.
func (c *Client) attempt(url string) (*Result, error) {
	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnknown, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		drain(resp.Body)
		return nil, fmt.Errorf("%w: %d", ErrNonRetryable, resp.StatusCode)
	} else if resp.StatusCode >= 500 {
		drain(resp.Body)
		return nil, fmt.Errorf("%w: %d", ErrRetryable, resp.StatusCode)
	}

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	return &Result{
		Content:     payload,
		ContentType: resp.Header.Get("Content-Type"),
	}, nil
}

// drain returns the connection to the idle pool; closing a body with bytes
// still unread forces it shut instead.
func drain(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, body)
}
