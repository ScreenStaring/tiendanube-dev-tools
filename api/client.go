package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	Version           = "0.0.1"
	DefaultHost       = "https://api.tiendanube.com"
	DefaultAPIVersion = "2025-03"
	initialRetryDelay = 500 * time.Millisecond
)

var (
	DefaultUserAgent = fmt.Sprintf("Tiendanube Development Tools v%s", Version)
	// ActiveAPIVersion is set process-wide by the CLI from --api-version flag.
	ActiveAPIVersion = DefaultAPIVersion
)

type Client struct {
	httpClient *http.Client
	apiVersion string
	storeID    string
	token      string
	userAgent  string
	verbose    bool
}

type ClientOption func(*Client)

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

func WithAPIVersion(version string) ClientOption {
	return func(c *Client) {
		c.apiVersion = version
	}
}

func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		if userAgent != "" {
			c.userAgent = userAgent
		}
	}
}

func WithVerbose(verbose bool) ClientOption {
	return func(c *Client) {
		c.verbose = verbose
	}
}

func NewClient(storeID, token string, opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{},
		apiVersion: ActiveAPIVersion,
		storeID:    cleanStoreID(storeID),
		token:      strings.TrimSpace(token),
		userAgent:  DefaultUserAgent,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// cleanStoreID normalizes store identifier (e.g. extracts ID if URL or domain passed).
func cleanStoreID(store string) string {
	store = strings.TrimSpace(store)
	if strings.HasPrefix(store, "http://") || strings.HasPrefix(store, "https://") {
		if u, err := url.Parse(store); err == nil {
			store = u.Host
		}
	}
	parts := strings.Split(store, ".")
	if len(parts) > 1 {
		return parts[0]
	}
	return store
}

func maxAttempts() int {
	const defaultAttempts = 10
	v := os.Getenv("TNDT_MAX_RETRY_ATTEMPTS")
	if v == "" {
		return defaultAttempts
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return defaultAttempts
	}
	return n
}

// BaseURL returns the store-scoped base URL for API calls.
func (c *Client) BaseURL() string {
	if c.apiVersion == "" {
		return fmt.Sprintf("%s/%s", DefaultHost, c.storeID)
	}
	return fmt.Sprintf("%s/%s/%s", DefaultHost, c.apiVersion, c.storeID)
}

// UserAgent returns the User-Agent header used for API requests.
func (c *Client) UserAgent() string {
	return c.userAgent
}

// APIError represents an error response from the Tiendanube API.
type APIError struct {
	StatusCode int
	Body       string
	Message    string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API request failed with HTTP %d: %s", e.StatusCode, e.Message)
	}
	if e.Body != "" {
		return fmt.Sprintf("API request failed with HTTP %d: %s", e.StatusCode, e.Body)
	}
	return fmt.Sprintf("API request failed with HTTP %d", e.StatusCode)
}

func parseAPIError(statusCode int, body []byte) *APIError {
	apiErr := &APIError{
		StatusCode: statusCode,
		Body:       string(body),
	}

	var genericMap map[string]interface{}
	if err := json.Unmarshal(body, &genericMap); err == nil {
		var messages []string

		if desc, ok := genericMap["description"].(string); ok && desc != "" {
			messages = append(messages, desc)
		}
		if msg, ok := genericMap["message"].(string); ok && msg != "" && msg != genericMap["description"] {
			messages = append(messages, msg)
		}
		if errStr, ok := genericMap["error"].(string); ok && errStr != "" {
			messages = append(messages, errStr)
		}

		var fieldErrors []string
		for k, v := range genericMap {
			if k == "description" || k == "message" || k == "error" || k == "code" {
				continue
			}
			if arr, ok := v.([]interface{}); ok {
				var items []string
				for _, item := range arr {
					items = append(items, fmt.Sprint(item))
				}
				fieldErrors = append(fieldErrors, fmt.Sprintf("%s: %s", k, strings.Join(items, ", ")))
			} else if str, ok := v.(string); ok {
				fieldErrors = append(fieldErrors, fmt.Sprintf("%s: %s", k, str))
			}
		}

		if len(fieldErrors) > 0 {
			messages = append(messages, strings.Join(fieldErrors, "; "))
		}

		if len(messages) > 0 {
			apiErr.Message = strings.Join(messages, " - ")
			return apiErr
		}
	}

	return apiErr
}

func retryDelay(resp *http.Response) time.Duration {
	if resp == nil {
		return 0
	}
	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		return 0
	}
	if secs, err := strconv.Atoi(retryAfter); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	return 0
}

func isRetryableStatus(code int) bool {
	return code == http.StatusTooManyRequests || code >= 500
}

// Request performs an HTTP request against the Tiendanube API with retries and logging.
func (c *Client) Request(ctx context.Context, method, path string, query url.Values, reqBody interface{}, respDest interface{}) error {
	readonly := os.Getenv("TNDT_READONLY")
	if (readonly == "1" || readonly == "true") && (method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete) {
		return fmt.Errorf("Request not allowed in read-only mode (TNDT_READONLY environment variable is set)")
	}

	endpoint := fmt.Sprintf("%s/%s", c.BaseURL(), strings.TrimPrefix(path, "/"))
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	var jsonBytes []byte
	if reqBody != nil {
		var err error
		jsonBytes, err = json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	attempts := maxAttempts()
	retryable := (method == http.MethodGet || method == http.MethodHead)

	for attempt := 0; ; attempt++ {
		var bodyReader io.Reader
		if len(jsonBytes) > 0 {
			bodyReader = bytes.NewReader(jsonBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %w", err)
		}

		req.Header.Set("Accept", "application/json")
		if c.token != "" {
			token := c.token
			if !strings.HasPrefix(strings.ToLower(token), "bearer ") {
				token = "Bearer " + token
			}
			req.Header.Set("Authorization", token)
		}
		if c.userAgent != "" {
			req.Header.Set("User-Agent", c.userAgent)
		}
		if len(jsonBytes) > 0 {
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
		}

		if c.verbose {
			fmt.Fprintf(os.Stderr, "> %s %s\n", req.Method, req.URL)
			for name, values := range req.Header {
				for _, v := range values {
					fmt.Fprintf(os.Stderr, "> %s: %s\n", name, v)
				}
			}
			if len(jsonBytes) > 0 {
				fmt.Fprintf(os.Stderr, ">\n%s\n\n", string(jsonBytes))
			} else {
				fmt.Fprintln(os.Stderr, ">")
			}
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if !retryable || attempt >= attempts-1 {
				return fmt.Errorf("API request to %s failed: %w", endpoint, err)
			}
			delay := initialRetryDelay << attempt
			fmt.Fprintf(os.Stderr, "Request to %s failed (%s); retrying in %s\n", endpoint, err, delay)
			time.Sleep(delay)
			continue
		}

		respBytes, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			if !retryable || attempt >= attempts-1 {
				return fmt.Errorf("failed to read response body: %w", readErr)
			}
			delay := initialRetryDelay << attempt
			time.Sleep(delay)
			continue
		}

		if c.verbose {
			fmt.Fprintf(os.Stderr, "< %s\n", resp.Status)
			for name, values := range resp.Header {
				for _, v := range values {
					fmt.Fprintf(os.Stderr, "< %s: %s\n", name, v)
				}
			}
			if len(respBytes) > 0 {
				fmt.Fprintf(os.Stderr, "<\n%s\n\n", string(respBytes))
			} else {
				fmt.Fprintln(os.Stderr, "<")
			}
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			apiErr := parseAPIError(resp.StatusCode, respBytes)
			if retryable && isRetryableStatus(resp.StatusCode) && attempt < attempts-1 {
				delay := retryDelay(resp)
				backoff := initialRetryDelay << attempt
				if delay < backoff {
					delay = backoff
				}
				fmt.Fprintf(os.Stderr, "Request to %s failed with status %d; retrying in %s\n", endpoint, resp.StatusCode, delay)
				time.Sleep(delay)
				continue
			}
			return apiErr
		}

		if respDest != nil && len(respBytes) > 0 {
			if err := json.Unmarshal(respBytes, respDest); err != nil {
				return fmt.Errorf("failed to unmarshal response: %w", err)
			}
		}

		return nil
	}
}
