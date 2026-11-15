package api

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestClientReadOnlyMode(t *testing.T) {
	os.Setenv("TNDT_READONLY", "1")
	defer os.Unsetenv("TNDT_READONLY")

	client := NewClient("123456", "token")

	err := client.Request(context.Background(), http.MethodPost, "webhooks", nil, map[string]string{"a": "b"}, nil)
	if err == nil {
		t.Fatal("expected error in read-only mode for POST, got nil")
	}

	err = client.Request(context.Background(), http.MethodPut, "webhooks/1", nil, map[string]string{"a": "b"}, nil)
	if err == nil {
		t.Fatal("expected error in read-only mode for PUT, got nil")
	}

	err = client.Request(context.Background(), http.MethodDelete, "webhooks/1", nil, nil, nil)
	if err == nil {
		t.Fatal("expected error in read-only mode for DELETE, got nil")
	}
}

func TestClientParseAPIError(t *testing.T) {
	body := []byte(`{
		"event": ["Invalid event specified."],
		"url": ["invalid url specified"]
	}`)

	apiErr := parseAPIError(http.StatusUnprocessableEntity, body)
	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnprocessableEntity)
	}
	if apiErr.Message == "" {
		t.Errorf("expected parsed error message, got empty")
	}
}

func TestRetryDelayHeader(t *testing.T) {
	resp := &http.Response{
		Header: http.Header{
			"Retry-After": []string{"2"},
		},
	}
	d := retryDelay(resp)
	if d != 2*time.Second {
		t.Errorf("retryDelay = %v, want 2s", d)
	}
}

func TestDefaultUserAgent(t *testing.T) {
	want := "Tiendanube Development Tools v" + Version
	if DefaultUserAgent != want {
		t.Errorf("DefaultUserAgent = %q, want %q", DefaultUserAgent, want)
	}
}

func TestBaseURL(t *testing.T) {
	client := NewClient("123456", "token")
	expected := "https://api.tiendanube.com/2025-03/123456"
	if client.BaseURL() != expected {
		t.Errorf("BaseURL() = %q, want %q", client.BaseURL(), expected)
	}

	versionless := NewClient("123456", "token", WithAPIVersion(""))
	if versionless.BaseURL() != "https://api.tiendanube.com/123456" {
		t.Errorf("BaseURL() = %q, want https://api.tiendanube.com/123456", versionless.BaseURL())
	}
}

func TestWithUserAgent(t *testing.T) {
	cDefault := NewClient("123456", "token")
	if cDefault.UserAgent() != DefaultUserAgent {
		t.Errorf("UserAgent() = %q, want %q", cDefault.UserAgent(), DefaultUserAgent)
	}

	customUA := "CustomAgent/1.0.0"
	cCustom := NewClient("123456", "token", WithUserAgent(customUA))
	if cCustom.UserAgent() != customUA {
		t.Errorf("UserAgent() = %q, want %q", cCustom.UserAgent(), customUA)
	}

	cEmpty := NewClient("123456", "token", WithUserAgent(""))
	if cEmpty.UserAgent() != DefaultUserAgent {
		t.Errorf("UserAgent() = %q, want %q", cEmpty.UserAgent(), DefaultUserAgent)
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestClientUserAgentHeader(t *testing.T) {
	var gotUA string
	customClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotUA = req.Header.Get("User-Agent")
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{}`)),
			}, nil
		}),
	}

	customUA := "MyTestAgent/2.0"
	client := NewClient("123456", "token", WithUserAgent(customUA), WithHTTPClient(customClient))
	err := client.Request(context.Background(), http.MethodGet, "test", nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected Request error: %v", err)
	}

	if gotUA != customUA {
		t.Errorf("User-Agent header = %q, want %q", gotUA, customUA)
	}
}
