package labrador

import (
	"errors"
	"time"
)

const (
	defaultRetryCount = 3
	defaultBackoffMs  = 1000
)

type DownloadHandler struct {
	retryCount int
	backoffMs  int
}

type DownloadHandlerOption func(*DownloadHandler)

func NewDownloadHandler(options ...DownloadHandlerOption) *DownloadHandler {
	handler := &DownloadHandler{
		retryCount: defaultRetryCount,
		backoffMs:  defaultBackoffMs,
	}

	for _, option := range options {
		option(handler)
	}

	return handler
}

func WithRetryCount(count int) DownloadHandlerOption {
	return func(handler *DownloadHandler) {
		if count < 1 {
			count = 1
		}
		handler.retryCount = count
	}
}

func WithBackoff(backoff int) DownloadHandlerOption {
	return func(handler *DownloadHandler) {
		if backoff < 0 {
			backoff = 0
		}
		handler.backoffMs = backoff
	}
}

func (h *DownloadHandler) Download(url string) (*DownloadResult, error) {
	var lastErr error
	for i := range h.retryCount {
		result, err := TryDownload(url)
		if err == nil {
			return result, nil
		}

		lastErr = err

		if errors.Is(err, ErrNonRetryable) {
			break
		}

		if i < h.retryCount {
			time.Sleep(time.Duration(h.backoffMs) * time.Millisecond)
		}
	}

	return nil, lastErr
}
