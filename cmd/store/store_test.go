package store

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
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

func TestFormatField(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"ID", "ID"},
		{"Name", "Name"},
		{"Description", "Description"},
		{"Email", "Email"},
		{"CustomerEmail", "Customer Email"},
		{"CreatedAt", "Created"},
		{"UpdatedAt", "Updated"},
		{"CountryCode", "Country Code"},
		{"OriginalDomain", "Original Domain"},
		{"WhatsappPhoneNumber", "Whatsapp Phone Number"},
		{"BusinessID", "Business ID"},
		{"BusinessName", "Business Name"},
		{"CustomerAccounts", "Customer Accounts"},
		{"PlanName", "Plan Name"},
		{"CurrentTheme", "Current Theme"},
		{"MainLanguage", "Main Language"},
		{"IanaTimezone", "Iana Timezone"},
	}

	for _, tt := range tests {
		if got := formatField(tt.input); got != tt.want {
			t.Errorf("formatField(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestStoreInfoMissingStore(t *testing.T) {
	app := setupTestApp()
	args := []string{
		"tndt", "store", "info",
		"--access-token", "some-token",
	}

	err := app.Run(context.Background(), args)
	if err == nil {
		t.Fatalf("expected error for missing store flag, got nil")
	}
	if !strings.Contains(err.Error(), "Required flag") && !strings.Contains(err.Error(), "Store ID is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCLIStoreInfoRealAPI(t *testing.T) {
	requireCLIEnvCredentials(t)

	app := setupTestApp()
	out := captureOutput(func() {
		err := app.Run(context.Background(), []string{"tndt", "store", "info"})
		if err != nil {
			t.Fatalf("store info failed: %v", err)
		}
	})

	if !strings.Contains(out, "Name") {
		t.Errorf("expected store info to contain Name, got:\n%s", out)
	}
}
