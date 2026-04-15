package klozeo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	defaultBaseURL    = "https://api.klozeo.com/api/v1"
	defaultTimeout    = 30 * time.Second
	defaultMaxRetries = 3
)

// Client is the Klozeo API client. Create one with New.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	maxRetries int

	// Rate limit state, updated atomically after each response.
	rateMu        sync.Mutex
	rateLimitState RateLimitState
}

// Option is a functional option for configuring a Client.
type Option func(*Client)

// WithBaseURL overrides the default API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithTimeout sets the per-request HTTP timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.httpClient.Timeout = d }
}

// WithMaxRetries sets the maximum number of retries on 429 or 5xx responses.
// The default is 3.
func WithMaxRetries(n int) Option {
	return func(c *Client) { c.maxRetries = n }
}

// WithHTTPClient replaces the underlying HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// New creates a new Client authenticated with the given API key.
// Pass functional options to override defaults.
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		maxRetries: defaultMaxRetries,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// RateLimitState returns the last observed rate limit headers from the API.
func (c *Client) RateLimitState() RateLimitState {
	c.rateMu.Lock()
	defer c.rateMu.Unlock()
	return c.rateLimitState
}

// updateRateLimit stores the latest rate limit headers.
func (c *Client) updateRateLimit(resp *http.Response) {
	limit, err1 := strconv.Atoi(resp.Header.Get("X-RateLimit-Limit"))
	remaining, err2 := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining"))
	if err1 != nil || err2 != nil {
		return
	}
	c.rateMu.Lock()
	c.rateLimitState = RateLimitState{Limit: limit, Remaining: remaining}
	c.rateMu.Unlock()
}

// apiErrorResponse is the JSON shape of an error from the API.
type apiErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// do executes an HTTP request with retry logic and returns the response body.
// It handles rate limit (429) and server (5xx) retries with exponential backoff.
//
// For non-2xx responses it calls newAPIError with the decoded body.
// The caller is responsible for closing the body when a nil error is returned.
func (c *Client) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	url := c.baseURL + path

	var (
		resp    *http.Response
		attempt int
	)

	for {
		// Build request body.
		var bodyReader io.Reader
		if body != nil {
			b, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("klozeo: marshal request: %w", err)
			}
			bodyReader = bytes.NewReader(b)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("klozeo: build request: %w", err)
		}
		req.Header.Set("X-API-Key", c.apiKey)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Accept", "application/json")

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("klozeo: http: %w", err)
		}

		c.updateRateLimit(resp)

		// Success — return the response for the caller to decode.
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		// Read the error body.
		errBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Parse retry-after for 429.
		var retryAfter time.Duration
		if resp.StatusCode == 429 {
			if s := resp.Header.Get("Retry-After"); s != "" {
				if secs, err := strconv.Atoi(s); err == nil {
					retryAfter = time.Duration(secs) * time.Second
				}
			}
		}

		// Decide whether to retry.
		shouldRetry := (resp.StatusCode == 429 || resp.StatusCode >= 500) &&
			attempt < c.maxRetries

		if shouldRetry {
			attempt++
			wait := retryAfter
			if wait == 0 {
				// Exponential backoff: 1s, 2s, 4s, ...
				wait = time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
			continue
		}

		// Decode structured error.
		var apiErr apiErrorResponse
		_ = json.Unmarshal(errBody, &apiErr)
		if apiErr.Message == "" {
			apiErr.Message = string(errBody)
		}
		return nil, newAPIError(resp.StatusCode, apiErr.Message, apiErr.Code, retryAfter)
	}
}

// doJSON executes a request and decodes a JSON response into dest.
func (c *Client) doJSON(ctx context.Context, method, path string, reqBody, dest any) error {
	resp, err := c.do(ctx, method, path, reqBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if dest == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}

// doNoContent executes a request and expects a 204 No Content response.
func (c *Client) doNoContent(ctx context.Context, method, path string, reqBody any) error {
	resp, err := c.do(ctx, method, path, reqBody)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

