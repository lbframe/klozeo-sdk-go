package klozeo

import (
	"errors"
	"fmt"
	"time"
)

// Sentinel errors returned by the client. Use errors.Is to test for these.
var (
	// ErrNotFound is returned when the API responds with HTTP 404.
	ErrNotFound = errors.New("klozeo: resource not found")
	// ErrUnauthorized is returned when the API responds with HTTP 401.
	ErrUnauthorized = errors.New("klozeo: unauthorized — check your API key")
	// ErrForbidden is returned when the API responds with HTTP 403 (leads limit reached).
	ErrForbidden = errors.New("klozeo: forbidden — leads limit reached")
	// ErrRateLimited is returned when the API responds with HTTP 429.
	ErrRateLimited = errors.New("klozeo: rate limit exceeded")
	// ErrBadRequest is returned when the API responds with HTTP 400.
	ErrBadRequest = errors.New("klozeo: bad request")
)

// APIError represents a structured error response from the Klozeo API.
// It wraps the appropriate sentinel error so errors.Is works correctly.
type APIError struct {
	// StatusCode is the HTTP status code.
	StatusCode int
	// Message is the detailed error description from the API.
	Message string
	// Code is the machine-readable error code from the API.
	Code string

	sentinel error
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("klozeo: HTTP %d — %s (code: %s)", e.StatusCode, e.Message, e.Code)
}

// Is enables errors.Is to match against sentinel errors.
func (e *APIError) Is(target error) bool {
	return errors.Is(e.sentinel, target)
}

// Unwrap returns the sentinel error for errors.Is/As chaining.
func (e *APIError) Unwrap() error {
	return e.sentinel
}

// RateLimitError is an *APIError subtype that also exposes the Retry-After duration.
type RateLimitError struct {
	APIError
	// RetryAfter is the duration to wait before retrying, parsed from the
	// Retry-After response header. May be zero if the header was absent.
	RetryAfter time.Duration
}

// Error implements the error interface.
func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("klozeo: rate limit exceeded — retry after %s", e.RetryAfter)
	}
	return e.APIError.Error()
}

// newAPIError constructs the appropriate error type for the given HTTP status.
func newAPIError(statusCode int, message, code string, retryAfter time.Duration) error {
	var sentinel error
	switch statusCode {
	case 400:
		sentinel = ErrBadRequest
	case 401:
		sentinel = ErrUnauthorized
	case 403:
		sentinel = ErrForbidden
	case 404:
		sentinel = ErrNotFound
	case 429:
		sentinel = ErrRateLimited
	default:
		sentinel = fmt.Errorf("klozeo: unexpected HTTP %d", statusCode)
	}

	base := APIError{
		StatusCode: statusCode,
		Message:    message,
		Code:       code,
		sentinel:   sentinel,
	}

	if statusCode == 429 {
		return &RateLimitError{APIError: base, RetryAfter: retryAfter}
	}
	return &base
}
