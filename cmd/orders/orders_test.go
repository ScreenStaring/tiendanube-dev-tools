package orders

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

func TestPrintOrders(t *testing.T) {
	orders := []api.Order{
		{
			ID:             871254203,
			Number:         101,
			Name:           "#101",
			Status:         "open",
			PaymentStatus:  "paid",
			ShippingStatus: "unpacked",
			Note:           "Gift order",
			CreatedAt:      "2024-01-01T10:00:00Z",
			UpdatedAt:      "2024-01-01T11:00:00Z",
			Products: []api.OrderProduct{
				{
					ID:        1001,
					ProductID: 5001,
					VariantID: "6001",
					Name:      "Cool Shirt",
					SKU:       api.NullableString("SHIRT-RED"),
					Quantity:  "2",
					Status:    "unpacked",
				},
			},
		},
	}

	out := captureOutput(func() {
		printOrders(orders)
	})

	if !strings.Contains(out, "871254203") || !strings.Contains(out, "#101") {
		t.Errorf("expected order ID and Name in output, got: %s", out)
	}
	if !strings.Contains(out, "Cool Shirt") || !strings.Contains(out, "SHIRT-RED") {
		t.Errorf("expected product in output, got: %s", out)
	}
	if !strings.Contains(out, `"Gift order"`) {
		t.Errorf("expected quoted note in output, got: %s", out)
	}
}

func TestPrintOrdersJSONL(t *testing.T) {
	orders := []api.Order{
		{
			ID:     871254203,
			Number: 101,
			Name:   "#101",
		},
	}

	out := captureOutput(func() {
		if err := printJSONL(orders); err != nil {
			t.Fatalf("printJSONL failed: %v", err)
		}
	})

	if !strings.Contains(out, `"id":871254203`) || !strings.Contains(out, `"#101"`) {
		t.Errorf("expected JSONL with order info, got: %s", out)
	}
}

func TestOrdersInvalidArgs(t *testing.T) {
	app := setupTestApp()

	// Invalid ID
	err := app.Run(context.Background(), []string{"tndt", "orders", "ls", "--store", "12345", "--access-token", "test-token", "invalid"})
	if err == nil {
		t.Fatalf("expected error for invalid arg, got nil")
	}

	// Missing name value
	err = app.Run(context.Background(), []string{"tndt", "orders", "ls", "--store", "12345", "--access-token", "test-token", "name:"})
	if err == nil {
		t.Fatalf("expected error for missing name, got nil")
	}
	if !strings.Contains(err.Error(), "Name value missing") {
		t.Errorf("expected 'Name value missing' error, got: %v", err)
	}
}

func TestCLIOrdersRealAPI(t *testing.T) {
	requireCLIEnvCredentials(t)

	app := setupTestApp()

	out := captureOutput(func() {
		err := app.Run(context.Background(), []string{"tndt", "orders", "ls"})
		if err != nil {
			t.Fatalf("orders ls failed: %v", err)
		}
	})
	_ = out

	outJSONL := captureOutput(func() {
		err := app.Run(context.Background(), []string{"tndt", "orders", "ls", "-j"})
		if err != nil {
			t.Fatalf("orders ls -j failed: %v", err)
		}
	})
	_ = outJSONL
}
