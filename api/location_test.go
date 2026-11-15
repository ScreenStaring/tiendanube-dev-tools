package api

import (
	"context"
	"encoding/json"
	"testing"
)

func TestLocationUnmarshal(t *testing.T) {
	raw := []byte(`[
		{
			"id": "01HTMFDH09VC6E2Q8KGTGP44D3",
			"name": {"es": "Depósito Central", "en": "Main Warehouse"},
			"is_default": true,
			"priority": 1,
			"created_at": "2026-01-01T00:00:00Z",
			"updated_at": "2026-01-02T00:00:00Z"
		},
		{
			"id": 12345,
			"name": "Sucursal Palermo",
			"is_default": false,
			"priority": 2
		}
	]`)

	var locs []Location
	if err := json.Unmarshal(raw, &locs); err != nil {
		t.Fatalf("unmarshal locations failed: %v", err)
	}

	if len(locs) != 2 {
		t.Fatalf("expected 2 locations, got %d", len(locs))
	}

	if locs[0].ID.String() != "01HTMFDH09VC6E2Q8KGTGP44D3" {
		t.Errorf("expected id 01HTMFDH09VC6E2Q8KGTGP44D3, got %q", locs[0].ID.String())
	}
	if locs[0].Name.String() != "Depósito Central" {
		t.Errorf("expected Depósito Central, got %q", locs[0].Name.String())
	}
	if !locs[0].IsDefault {
		t.Errorf("expected is_default=true")
	}

	if locs[1].ID.String() != "12345" {
		t.Errorf("expected id 12345, got %q", locs[1].ID.String())
	}
	if locs[1].Name.String() != "Sucursal Palermo" {
		t.Errorf("expected Sucursal Palermo, got %q", locs[1].Name.String())
	}

	// Test regional name
	flexibleRaw := []byte(`[
		{
			"id": "01M33MQVR7W28EGK6TKW2RQ588",
			"name": {"es_AR": "Depósito Norte"},
			"is_default": true,
			"priority": 0
		},
		{
			"id": "01M33MQVR7W28EGK6TKW2RQ589",
			"name": null,
			"address": {"reference": "Local Belgrano", "city": "Buenos Aires"},
			"is_default": false,
			"priority": 0
		}
	]`)

	var flexLocs []Location
	if err := json.Unmarshal(flexibleRaw, &flexLocs); err != nil {
		t.Fatalf("unmarshal flexible locations failed: %v", err)
	}

	if len(flexLocs) != 2 {
		t.Fatalf("expected 2 locations, got %d", len(flexLocs))
	}
	if !flexLocs[0].IsDefault {
		t.Errorf("expected is_default=true")
	}
	if flexLocs[0].Priority != 0 {
		t.Errorf("expected priority=0, got %d", flexLocs[0].Priority)
	}
	if flexLocs[0].DisplayName() != "Depósito Norte" {
		t.Errorf("expected DisplayName 'Depósito Norte', got %q", flexLocs[0].DisplayName())
	}

	if flexLocs[1].IsDefault {
		t.Errorf("expected is_default=false for string 'false'")
	}
	if flexLocs[1].DisplayName() != "Local Belgrano" {
		t.Errorf("expected DisplayName fallback to reference 'Local Belgrano', got %q", flexLocs[1].DisplayName())
	}

	// Test parseLocationsJSON with object wrapper
	wrapperRaw := []byte(`{
		"results": [
			{
				"id": "01M33MQVR7W28EGK6TKW2RQ588",
				"name": "Depósito Central"
			}
		]
	}`)
	wrappedLocs, err := parseLocationsJSON(wrapperRaw)
	if err != nil {
		t.Fatalf("parseLocationsJSON wrapper failed: %v", err)
	}
	if len(wrappedLocs) != 1 || wrappedLocs[0].DisplayName() != "Depósito Central" {
		t.Errorf("unexpected wrapped locations: %+v", wrappedLocs)
	}

	// Test parseLocationsJSON with single object
	singleRaw := []byte(`{
		"id": "01M33MQVR7W28EGK6TKW2RQ588",
		"name": "Depósito Central"
	}`)
	singleLocs, err := parseLocationsJSON(singleRaw)
	if err != nil {
		t.Fatalf("parseLocationsJSON single failed: %v", err)
	}
	if len(singleLocs) != 1 || singleLocs[0].DisplayName() != "Depósito Central" {
		t.Errorf("unexpected single location: %+v", singleLocs)
	}
}

func TestLocationRealAPI(t *testing.T) {
	store, token := requireAPICredentials(t)
	client := NewClient(store, token)

	locs, err := client.ListLocations(context.Background())
	if err != nil {
		t.Fatalf("ListLocations failed: %v", err)
	}
	t.Logf("Retrieved %d locations", len(locs))

	if len(locs) > 0 {
		loc, err := client.GetLocation(context.Background(), locs[0].ID.String())
		if err != nil {
			t.Fatalf("GetLocation failed: %v", err)
		}
		if loc.ID != locs[0].ID {
			t.Errorf("got location ID %v, want %v", loc.ID, locs[0].ID)
		}
	}
}
