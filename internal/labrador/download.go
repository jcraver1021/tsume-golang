package labrador

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	ErrUnknown      = errors.New("unknown error")
	ErrRetryable    = fmt.Errorf("retryable error")
	ErrNonRetryable = fmt.Errorf("non-retryable error")
)

type DownloadResult struct {
	Content     []byte
	ContentType string
}

func TryDownload(url string) (*DownloadResult, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnknown, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return nil, fmt.Errorf("%w: %d", ErrNonRetryable, resp.StatusCode)
	} else if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("%w: %d", ErrRetryable, resp.StatusCode)
	}

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnknown, err)
	}

	contentType := resp.Header.Get("Content-Type")

	return &DownloadResult{
		Content:     payload,
		ContentType: contentType,
	}, nil
}
