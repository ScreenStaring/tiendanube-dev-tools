package cmd

import (
	"context"
	"testing"

	"github.com/urfave/cli/v3"

	"github.com/ScreenStaring/tiendanube-dev-tools/api"
)

func TestLookupAccessToken(t *testing.T) {
	// Plain token returns as-is
	plain := "my-secret-token"
	got := LookupAccessToken("123456", plain)
	if got != plain {
		t.Errorf("LookupAccessToken() = %v, want %v", got, plain)
	}

	// Dynamic token command via "< echo" passes shop argument
	cmdToken := "< echo"
	got = LookupAccessToken("123456", cmdToken)
	if got != "123456" {
		t.Errorf("LookupAccessToken() = %v, want 123456", got)
	}
}

func TestNewClientMissingFlags(t *testing.T) {
	_, err := NewClient(&cli.Command{})
	if err == nil {
		t.Fatal("expected error when store is missing, got nil")
	}

	var clientErr error
	testCmd := &cli.Command{
		Flags: Flags,
		Action: func(ctx context.Context, c *cli.Command) error {
			_, clientErr = NewClient(c)
			return nil
		},
	}
	_ = testCmd.Run(context.Background(), []string{"test", "--store", "123456"})
	if clientErr == nil {
		t.Fatal("expected error when access token is missing, got nil")
	}
}

func TestEnvVars(t *testing.T) {
	tests := []struct {
		name      string
		env       map[string]string
		wantStore string
		wantToken string
	}{
		{
			name: "TNDT env vars",
			env: map[string]string{
				"TNDT_STORE_ID":     "111",
				"TNDT_ACCESS_TOKEN": "token-tndt",
			},
			wantStore: "111",
			wantToken: "token-tndt",
		},
		{
			name: "TIENDANUBE env vars",
			env: map[string]string{
				"TIENDANUBE_STORE_ID":     "222",
				"TIENDANUBE_ACCESS_TOKEN": "token-tiendanube",
			},
			wantStore: "222",
			wantToken: "token-tiendanube",
		},
		{
			name: "TNDT precedence over TIENDANUBE",
			env: map[string]string{
				"TNDT_STORE_ID":           "111",
				"TIENDANUBE_STORE_ID":     "222",
				"TNDT_ACCESS_TOKEN":       "token-tndt",
				"TIENDANUBE_ACCESS_TOKEN": "token-tiendanube",
			},
			wantStore: "111",
			wantToken: "token-tndt",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			var gotStore, gotToken string
			testCmd := &cli.Command{
				Flags: BaseFlags(),
				Action: func(ctx context.Context, c *cli.Command) error {
					gotStore = c.String("store")
					gotToken = c.String("access-token")
					_, err := NewClient(c)
					return err
				},
			}

			if err := testCmd.Run(context.Background(), []string{"test"}); err != nil {
				t.Fatalf("unexpected command error: %v", err)
			}

			if gotStore != tc.wantStore {
				t.Errorf("store = %q, want %q", gotStore, tc.wantStore)
			}
			if gotToken != tc.wantToken {
				t.Errorf("access-token = %q, want %q", gotToken, tc.wantToken)
			}
		})
	}
}

func TestUserAgentOption(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		env       map[string]string
		wantAgent string
	}{
		{
			name:      "default user agent when unset",
			args:      []string{"test", "--store", "123", "--access-token", "token"},
			wantAgent: api.DefaultUserAgent,
		},
		{
			name:      "set via --user-agent",
			args:      []string{"test", "--store", "123", "--access-token", "token", "--user-agent", "CustomUA/1.0"},
			wantAgent: "CustomUA/1.0",
		},
		{
			name: "set via TNDT_USER_AGENT env var",
			args: []string{"test", "--store", "123", "--access-token", "token"},
			env: map[string]string{
				"TNDT_USER_AGENT": "EnvUA/1.0",
			},
			wantAgent: "EnvUA/1.0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			var client *api.Client
			testCmd := &cli.Command{
				Flags: BaseFlags(),
				Action: func(ctx context.Context, c *cli.Command) error {
					var err error
					client, err = NewClient(c)
					return err
				},
			}

			if err := testCmd.Run(context.Background(), tc.args); err != nil {
				t.Fatalf("unexpected command error: %v", err)
			}

			if client == nil {
				t.Fatal("expected client to be created, got nil")
			}

			if got := client.UserAgent(); got != tc.wantAgent {
				t.Errorf("UserAgent() = %q, want %q", got, tc.wantAgent)
			}
		})
	}
}

func TestParseIDArgs(t *testing.T) {
	ids, skus, err := ParseIDArgs([]string{"123", "sku:ABC", "456", "SKU:XYZ"}, "a product id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 || ids[0] != 123 || ids[1] != 456 {
		t.Errorf("unexpected ids: %v", ids)
	}
	if len(skus) != 2 || skus[0] != "ABC" || skus[1] != "XYZ" {
		t.Errorf("unexpected skus: %v", skus)
	}

	_, _, err = ParseIDArgs([]string{"sku:"}, "a product id")
	if err == nil {
		t.Error("expected error for empty sku, got nil")
	}

	_, _, err = ParseIDArgs([]string{"invalid"}, "a product id")
	if err == nil {
		t.Error("expected error for non-int argument, got nil")
	}
}

func TestNormalizeField(t *testing.T) {
	if got := NormalizeField(" Created At "); got != "createdat" {
		t.Errorf("NormalizeField() = %v, want createdat", got)
	}
	if got := NormalizeField("free_shipping"); got != "freeshipping" {
		t.Errorf("NormalizeField() = %v, want freeshipping", got)
	}
	if got := NormalizeField("is_kit"); got != "iskit" {
		t.Errorf("NormalizeField() = %v, want iskit", got)
	}
}
