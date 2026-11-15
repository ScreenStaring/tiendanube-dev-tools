package api

import (
	"context"
	"testing"
)

func TestNormalizeOwnerResource(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"product", "products"},
		{"products", "products"},
		{"variant", "product_variants"},
		{"product_variants", "product_variants"},
		{"category", "categories"},
		{"collection", "categories"},
		{"collections", "categories"},
		{"page", "pages"},
		{"order", "orders"},
		{"customer", "customers"},
	}

	for _, tt := range tests {
		if got := NormalizeOwnerResource(tt.input); got != tt.want {
			t.Errorf("NormalizeOwnerResource(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestListMetafieldsParamsToQuery(t *testing.T) {
	params := &ListMetafieldsParams{
		OwnerID:   12345,
		Namespace: "custom",
		Key:       "color",
		Page:      2,
		PerPage:   50,
	}

	q := params.toQuery()
	if q.Get("owner_id") != "12345" {
		t.Errorf("owner_id = %v, want 12345", q.Get("owner_id"))
	}
	if q.Get("namespace") != "custom" {
		t.Errorf("namespace = %v, want custom", q.Get("namespace"))
	}
	if q.Get("key") != "color" {
		t.Errorf("key = %v, want color", q.Get("key"))
	}
	if q.Get("page") != "2" {
		t.Errorf("page = %v, want 2", q.Get("page"))
	}
	if q.Get("per_page") != "50" {
		t.Errorf("per_page = %v, want 50", q.Get("per_page"))
	}
}

func TestMetafieldsRealAPI(t *testing.T) {
	store, token := requireAPICredentials(t)
	client := NewClient(store, token)

	mfs, err := client.ListMetafields(context.Background(), "products", &ListMetafieldsParams{PerPage: 5})
	if err != nil {
		t.Fatalf("ListMetafields failed: %v", err)
	}
	t.Logf("Retrieved %d metafields", len(mfs))

	if len(mfs) > 0 {
		mf, err := client.GetMetafield(context.Background(), mfs[0].ID)
		if err != nil {
			t.Fatalf("GetMetafield failed: %v", err)
		}
		if mf.ID != mfs[0].ID {
			t.Errorf("got metafield ID %d, want %d", mf.ID, mfs[0].ID)
		}
	}
}
