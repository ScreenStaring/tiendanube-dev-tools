package scripts

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"

	"github.com/ScreenStaring/tiendanube-dev-tools/api"
)

func captureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func setupTestApp() *cli.Command {
	return &cli.Command{
		Name:                   "tndt",
		UseShortOptionHandling: true,
		Commands: []*cli.Command{
			&Cmd,
		},
	}
}

func requireCLIEnvCredentials(t *testing.T) {
	t.Helper()
	store := os.Getenv("TNDT_STORE_ID")
	if store == "" {
		store = os.Getenv("TIENDANUBE_STORE_ID")
	}

	token := os.Getenv("TNDT_ACCESS_TOKEN")
	if token == "" {
		token = os.Getenv("TIENDANUBE_ACCESS_TOKEN")
	}

	if store == "" || token == "" {
		t.Skip("Skipping CLI integration test: set TNDT_STORE_ID / TIENDANUBE_STORE_ID and TNDT_ACCESS_TOKEN / TIENDANUBE_ACCESS_TOKEN to run against real API")
	}
}

func TestScriptURLRegex(t *testing.T) {
	tests := []struct {
		url   string
		match bool
	}{
		{"https://example.com/app.js", true},
		{"http://example.com/app.js", false}, // scriptURL pattern is (?i)\A(?:https:)?//[\da-z]
		{"//example.com/app.js", true},
		{"12345", false},
		{"not-a-url", false},
	}

	for _, tt := range tests {
		if got := scriptURL.MatchString(tt.url); got != tt.match {
			t.Errorf("scriptURL.MatchString(%q) = %v, want %v", tt.url, got, tt.match)
		}
	}
}

func TestPrintFormatted(t *testing.T) {
	scripts := []api.Script{
		{
			ID:        1001,
			Name:      "Analytics",
			Event:     "onload",
			Location:  "store",
			Status:    "active",
			Src:       "https://example.com/analytics.js",
			CreatedAt: "2024-01-01T00:00:00Z",
			UpdatedAt: "2024-01-02T00:00:00Z",
		},
	}

	out := captureOutput(func() {
		printFormatted(scripts)
	})

	if !strings.Contains(out, "Analytics") || !strings.Contains(out, "1001") {
		t.Errorf("expected script info in output, got: %s", out)
	}
	if !strings.Contains(out, "https://example.com/analytics.js") {
		t.Errorf("expected script URL in output, got: %s", out)
	}
	if !strings.Contains(out, "--------------------") {
		t.Errorf("expected separator in output, got: %s", out)
	}
}

func TestPrintJSONL(t *testing.T) {
	scripts := []api.Script{
		{
			ID:   1001,
			Name: "Analytics",
			Src:  "https://example.com/analytics.js",
		},
	}

	out := captureOutput(func() {
		if err := printJSONL(scripts); err != nil {
			t.Fatalf("printJSONL failed: %v", err)
		}
	})

	if !strings.Contains(out, `"id":1001`) || !strings.Contains(out, `"Analytics"`) {
		t.Errorf("expected script in JSONL, got: %s", out)
	}
}

func TestParseQueryParams(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{
			name: "empty args",
			args: nil,
			want: "",
		},
		{
			name: "single key-val",
			args: []string{"key1=val1"},
			want: `{"key1":"val1"}`,
		},
		{
			name: "multiple flags",
			args: []string{"key1=val1", "key2=val2"},
			want: `{"key1":"val1","key2":"val2"}`,
		},
		{
			name: "values with comma",
			args: []string{"key1=val1,val2", "key2=val3"},
			want: `{"key1":"val1,val2","key2":"val3"}`,
		},
		{
			name: "multiple flags with various values",
			args: []string{"paramA=2", "showDialog=true", "disabled=false"},
			want: `{"disabled":"false","paramA":"2","showDialog":"true"}`,
		},
		{
			name:    "invalid no equal sign",
			args:    []string{"invalid_no_equal"},
			wantErr: true,
		},
		{
			name:    "invalid empty key",
			args:    []string{"=val"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseQueryParams(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseQueryParams() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseQueryParams() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestScriptsCreateArgsValidation(t *testing.T) {
	app := setupTestApp()

	// Missing script id
	err := app.Run(context.Background(), []string{"tndt", "scripts", "create", "--store", "123", "--access-token", "token"})
	if err == nil || !strings.Contains(err.Error(), "Script id is required") {
		t.Errorf("expected 'Script id is required', got: %v", err)
	}

	// Too many arguments
	err = app.Run(context.Background(), []string{"tndt", "scripts", "create", "--store", "123", "--access-token", "token", "123", "456"})
	if err == nil || !strings.Contains(err.Error(), "Too many arguments") {
		t.Errorf("expected 'Too many arguments', got: %v", err)
	}

	// Non-numeric script id
	err = app.Run(context.Background(), []string{"tndt", "scripts", "create", "--store", "123", "--access-token", "token", "abc"})
	if err == nil || !strings.Contains(err.Error(), "invalid: must be an int") {
		t.Errorf("expected 'must be an int', got: %v", err)
	}

	// Invalid query format
	err = app.Run(context.Background(), []string{"tndt", "scripts", "create", "--store", "123", "--access-token", "token", "-q", "invalid_no_equal", "123"})
	if err == nil || !strings.Contains(err.Error(), "key=value") {
		t.Errorf("expected 'key=value' error, got: %v", err)
	}
}

func TestScriptsUpdateArgsValidation(t *testing.T) {
	app := setupTestApp()

	// Missing script id
	err := app.Run(context.Background(), []string{"tndt", "scripts", "update", "--store", "123", "--access-token", "token"})
	if err == nil || !strings.Contains(err.Error(), "Script id is required") {
		t.Errorf("expected 'Script id is required', got: %v", err)
	}

	// Too many arguments
	err = app.Run(context.Background(), []string{"tndt", "scripts", "update", "--store", "123", "--access-token", "token", "123", "456"})
	if err == nil || !strings.Contains(err.Error(), "Too many arguments") {
		t.Errorf("expected 'Too many arguments', got: %v", err)
	}

	// Non-numeric script id
	err = app.Run(context.Background(), []string{"tndt", "scripts", "update", "--store", "123", "--access-token", "token", "abc"})
	if err == nil || !strings.Contains(err.Error(), "invalid: must be an int") {
		t.Errorf("expected 'must be an int', got: %v", err)
	}

	// Invalid query format
	err = app.Run(context.Background(), []string{"tndt", "scripts", "update", "--store", "123", "--access-token", "token", "-q", "invalid_no_equal", "123"})
	if err == nil || !strings.Contains(err.Error(), "key=value") {
		t.Errorf("expected 'key=value' error, got: %v", err)
	}
}

func TestScriptsDeleteArgsValidation(t *testing.T) {
	app := setupTestApp()

	// Missing script id
	err := app.Run(context.Background(), []string{"tndt", "scripts", "delete", "--store", "123", "--access-token", "token"})
	if err == nil || !strings.Contains(err.Error(), "Script id is required") {
		t.Errorf("expected 'Script id is required', got: %v", err)
	}

	// Too many arguments
	err = app.Run(context.Background(), []string{"tndt", "scripts", "delete", "--store", "123", "--access-token", "token", "123", "456"})
	if err == nil || !strings.Contains(err.Error(), "Too many arguments") {
		t.Errorf("expected 'Too many arguments', got: %v", err)
	}

	// Non-numeric script id
	err = app.Run(context.Background(), []string{"tndt", "scripts", "delete", "--store", "123", "--access-token", "token", "abc"})
	if err == nil || !strings.Contains(err.Error(), "invalid: must be an int") {
		t.Errorf("expected 'must be an int', got: %v", err)
	}
}

func TestScriptsInvalidArg(t *testing.T) {
	app := setupTestApp()
	err := app.Run(context.Background(), []string{"tndt", "scripts", "ls", "--store", "12345", "--access-token", "test-token", "not-a-number"})
	if err == nil {
		t.Fatalf("expected error for invalid arg, got nil")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected invalid arg error, got: %v", err)
	}
}

func TestCLIScriptsRealAPI(t *testing.T) {
	requireCLIEnvCredentials(t)

	app := setupTestApp()

	out := captureOutput(func() {
		err := app.Run(context.Background(), []string{"tndt", "scripts", "ls"})
		if err != nil {
			t.Fatalf("scripts ls failed: %v", err)
		}
	})
	_ = out

	outJSONL := captureOutput(func() {
		err := app.Run(context.Background(), []string{"tndt", "scripts", "ls", "-j"})
		if err != nil {
			t.Fatalf("scripts ls -j failed: %v", err)
		}
	})
	_ = outJSONL
}
