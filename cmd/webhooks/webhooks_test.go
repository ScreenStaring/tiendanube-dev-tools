package webhooks

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/urfave/cli/v3"
)

func setupTestApp() *cli.Command {
	return &cli.Command{
		Name:                   "tndt",
		UseShortOptionHandling: true,
		Commands: []*cli.Command{
			&Cmd,
		},
	}
}

func captureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
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

func TestParseWebhookIDArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    []int64
		wantErr bool
	}{
		{
			name: "single id",
			args: []string{"123"},
			want: []int64{123},
		},
		{
			name: "multiple ids deduplicated",
			args: []string{"123", "456", "123"},
			want: []int64{123, 456},
		},
		{
			name: "empty args",
			args: []string{},
			want: nil,
		},
		{
			name:    "invalid id",
			args:    []string{"abc"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseWebhookIDArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseWebhookIDArgs(%v) expected error, got nil", tt.args)
				}
				return
			}
			if err != nil {
				t.Errorf("parseWebhookIDArgs(%v) unexpected error: %v", tt.args, err)
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("parseWebhookIDArgs(%v) returned %d items, want %d", tt.args, len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("got[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestListWebhooksInvalidArg(t *testing.T) {
	app := setupTestApp()
	err := app.Run(context.Background(), []string{"tndt", "webhooks", "ls", "--store", "123", "abc"})
	if err == nil {
		t.Fatal("expected error for non-numeric webhook id, got nil")
	}
	expected := "Argument 'abc' invalid: must be a webhook id"
	if err.Error() != expected {
		t.Errorf("got error %q, want %q", err.Error(), expected)
	}
}

func TestCLIWebhooksRealAPI(t *testing.T) {
	requireCLIEnvCredentials(t)

	app := setupTestApp()
	testURL := fmt.Sprintf("https://example.com/tndt-cli-%d", time.Now().UnixNano())
	testEvent := "app/uninstalled"

	// 1. Create Webhook (using webhooks command and c alias)
	createArgs := []string{
		"tndt", "webhooks", "create",
		"-e", testEvent,
		"-a", testURL,
	}

	var createdID string
	out := captureOutput(func() {
		if err := app.Run(context.Background(), createArgs); err != nil {
			t.Fatalf("webhook create failed: %v", err)
		}
	})

	if !strings.Contains(out, "Webhook created:") {
		t.Fatalf("expected 'Webhook created:', got: %s", out)
	}

	for _, field := range strings.Fields(out) {
		if field != "Webhook" && field != "created:" {
			createdID = field
			break
		}
	}
	if createdID == "" {
		t.Fatalf("could not extract created webhook ID from: %s", out)
	}
	t.Logf("Created webhook %s", createdID)

	defer func() {
		_ = app.Run(context.Background(), []string{"tndt", "webhooks", "delete", createdID})
	}()

	// 2. Duplicate Check (--one)
	oneArgs := []string{
		"tndt", "webhooks", "create", "--one",
		"-e", testEvent,
		"-a", testURL,
	}
	if err := app.Run(context.Background(), oneArgs); err == nil {
		t.Errorf("expected error with --one for duplicate webhook, got nil")
	}

	// 3. List Webhooks using webhooks ls
	lsArgs := []string{
		"tndt", "webhooks", "ls",
		"-e", testEvent,
	}
	lsOut := captureOutput(func() {
		if err := app.Run(context.Background(), lsArgs); err != nil {
			t.Fatalf("webhook ls failed: %v", err)
		}
	})
	if !strings.Contains(lsOut, createdID) {
		t.Errorf("expected ls output to contain %s, got:\n%s", createdID, lsOut)
	}

	// 4. List Webhooks (JSONL) using webhooks ls
	jsonlArgs := []string{
		"tndt", "webhooks", "ls", "-j",
		"-e", testEvent,
	}
	jsonlOut := captureOutput(func() {
		if err := app.Run(context.Background(), jsonlArgs); err != nil {
			t.Fatalf("webhook ls -j failed: %v", err)
		}
	})
	if !strings.Contains(jsonlOut, createdID) {
		t.Errorf("expected JSONL output to contain %s, got:\n%s", createdID, jsonlOut)
	}

	// 4b. List Webhooks by ID using positional argument
	lsByIDArgs := []string{
		"tndt", "webhooks", "ls",
		createdID,
	}
	lsByIDOut := captureOutput(func() {
		if err := app.Run(context.Background(), lsByIDArgs); err != nil {
			t.Fatalf("webhook ls with ID failed: %v", err)
		}
	})
	if !strings.Contains(lsByIDOut, createdID) {
		t.Errorf("expected ls with ID output to contain %s, got:\n%s", createdID, lsByIDOut)
	}

	// 4c. List Webhooks by ID (JSONL)
	jsonlByIDArgs := []string{
		"tndt", "webhooks", "ls", "-j",
		createdID,
	}
	jsonlByIDOut := captureOutput(func() {
		if err := app.Run(context.Background(), jsonlByIDArgs); err != nil {
			t.Fatalf("webhook ls -j with ID failed: %v", err)
		}
	})
	if !strings.Contains(jsonlByIDOut, createdID) {
		t.Errorf("expected JSONL with ID output to contain %s, got:\n%s", createdID, jsonlByIDOut)
	}

	// 5. Update Webhook
	newURL := fmt.Sprintf("https://example.com/tndt-cli-updated-%d", time.Now().UnixNano())
	updateArgs := []string{
		"tndt", "webhooks", "update",
		"-a", newURL,
		createdID,
	}
	updateOut := captureOutput(func() {
		if err := app.Run(context.Background(), updateArgs); err != nil {
			t.Fatalf("webhook update failed: %v", err)
		}
	})
	if !strings.Contains(updateOut, "Webhook updated") {
		t.Errorf("expected 'Webhook updated', got: %s", updateOut)
	}

	// 6. Delete Webhook
	delArgs := []string{
		"tndt", "webhooks", "delete",
		createdID,
	}
	delOut := captureOutput(func() {
		if err := app.Run(context.Background(), delArgs); err != nil {
			t.Fatalf("webhook delete failed: %v", err)
		}
	})
	if !strings.Contains(delOut, "1 webhook(s) deleted") {
		t.Errorf("expected '1 webhook(s) deleted', got: %s", delOut)
	}
}
