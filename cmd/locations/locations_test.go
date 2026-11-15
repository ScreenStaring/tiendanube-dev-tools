package locations

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestFormatAddress(t *testing.T) {
	tests := []struct {
		name string
		addr *api.LocationAddress
		want string
	}{
		{
			name: "nil address",
			addr: nil,
			want: "",
		},
		{
			name: "full address",
			addr: &api.LocationAddress{
				Street:   "Av. Santa Fe",
				Number:   "1234",
				Floor:    "3A",
				Locality: "Recoleta",
				City:     "Buenos Aires",
				Zipcode:  "C1059",
			},
			want: "Av. Santa Fe 1234 3A, Recoleta, Buenos Aires, C1059",
		},
		{
			name: "address with street and city",
			addr: &api.LocationAddress{
				Street: "Corrientes",
				Number: "500",
				City:   "Buenos Aires",
			},
			want: "Corrientes 500, Buenos Aires",
		},
		{
			name: "reference and city only",
			addr: &api.LocationAddress{
				Reference: "Local Belgrano",
				City:      "Buenos Aires",
			},
			want: "Local Belgrano, Buenos Aires",
		},
		{
			name: "empty address",
			addr: &api.LocationAddress{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatAddress(tt.addr); got != tt.want {
				t.Errorf("formatAddress() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintLocations(t *testing.T) {
	locs := []api.Location{
		{
			ID:        api.NullableString("01HTMFDH09VC6E2Q8KGTGP44D3"),
			Name:      api.LocalizedValue{"es": "Depósito Central"},
			IsDefault: true,
			Priority:  1,
			Address: &api.LocationAddress{
				Street: "Av. Rivadavia",
				Number: "1000",
				City:   "CABA",
			},
			CreatedAt: "2026-01-01T00:00:00Z",
			UpdatedAt: "2026-01-02T00:00:00Z",
		},
	}

	out := captureOutput(func() {
		printLocations(locs)
	})

	if !strings.Contains(out, "01HTMFDH09VC6E2Q8KGTGP44D3") {
		t.Errorf("expected ID in output, got: %s", out)
	}
	if !strings.Contains(out, "Depósito Central") {
		t.Errorf("expected Name in output, got: %s", out)
	}
	if !strings.Contains(out, "Default") || !strings.Contains(out, "true") {
		t.Errorf("expected Default true in output, got: %s", out)
	}
	if !strings.Contains(out, "Priority") || !strings.Contains(out, "1") {
		t.Errorf("expected Priority 1 in output, got: %s", out)
	}
	if !strings.Contains(out, "Av. Rivadavia 1000, CABA") {
		t.Errorf("expected formatted address in output, got: %s", out)
	}
	if !strings.Contains(out, "--------------------") {
		t.Errorf("expected separator in output, got: %s", out)
	}
}

func TestPrintLocationsJSONL(t *testing.T) {
	locs := []api.Location{
		{
			ID:        api.NullableString("123"),
			Name:      api.LocalizedValue{"es": "Sucursal Norte"},
			IsDefault: false,
		},
	}

	out := captureOutput(func() {
		if err := printLocationsJSONL(locs); err != nil {
			t.Fatalf("printLocationsJSONL failed: %v", err)
		}
	})

	var parsed map[string]interface{}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if err := json.Unmarshal([]byte(lines[0]), &parsed); err != nil {
		t.Fatalf("invalid JSON line: %v", err)
	}
	if parsed["id"] != "123" {
		t.Errorf("expected id '123', got %v", parsed["id"])
	}
	if parsed["is_default"] != false {
		t.Errorf("expected is_default false, got %v", parsed["is_default"])
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

func TestCLILocationsRealAPI(t *testing.T) {
	requireCLIEnvCredentials(t)

	app := setupTestApp()
	out := captureOutput(func() {
		err := app.Run(context.Background(), []string{"tndt", "locations", "ls"})
		if err != nil {
			t.Fatalf("locations ls failed: %v", err)
		}
	})
	_ = out

	outJSONL := captureOutput(func() {
		err := app.Run(context.Background(), []string{"tndt", "locations", "ls", "-j"})
		if err != nil {
			t.Fatalf("locations ls -j failed: %v", err)
		}
	})
	_ = outJSONL
}
