package api

import (
	"encoding/json"
	"testing"
)

func TestLocalizedValueUnmarshal(t *testing.T) {
	// JSON object with multiple languages
	rawObj := []byte(`{"en":"Shirt","es":"Remera","pt":"Camiseta"}`)
	var lv LocalizedValue
	if err := json.Unmarshal(rawObj, &lv); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if lv.String() != "Remera" { // es is prioritized for Tiendanube
		t.Errorf("lv.String() = %q, want Remera", lv.String())
	}

	// JSON string
	rawStr := []byte(`"Plain text"`)
	var lvStr LocalizedValue
	if err := json.Unmarshal(rawStr, &lvStr); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if lvStr.String() != "Plain text" {
		t.Errorf("lvStr.String() = %q, want 'Plain text'", lvStr.String())
	}

	// Null
	rawNull := []byte(`null`)
	var lvNull LocalizedValue
	if err := json.Unmarshal(rawNull, &lvNull); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if lvNull.String() != "" {
		t.Errorf("lvNull.String() = %q, want ''", lvNull.String())
	}
}

func TestNullableString(t *testing.T) {
	var ns NullableString
	if err := json.Unmarshal([]byte(`null`), &ns); err != nil {
		t.Fatalf("unmarshal null failed: %v", err)
	}
	if ns.String() != "" {
		t.Errorf("expected empty string for null, got %q", ns.String())
	}

	if err := json.Unmarshal([]byte(`"val"`), &ns); err != nil {
		t.Fatalf("unmarshal string failed: %v", err)
	}
	if ns.String() != "val" {
		t.Errorf("expected val, got %q", ns.String())
	}

	if err := json.Unmarshal([]byte(`12345`), &ns); err != nil {
		t.Fatalf("unmarshal number failed: %v", err)
	}
	if ns.String() != "12345" {
		t.Errorf("expected 12345, got %q", ns.String())
	}
}

func TestProductVariantAndOptions(t *testing.T) {
	productJSON := `{
		"id": 123,
		"name": {"es": "Remera"},
		"attributes": [{"es": "Color"}, {"es": "Talle"}],
		"variants": [
			{
				"id": 1,
				"price": "100.00",
				"sku": "REM-RED-S",
				"values": [{"es": "Rojo"}, {"es": "S"}]
			},
			{
				"id": 2,
				"price": "100.00",
				"sku": "REM-BLU-M",
				"values": [{"es": "Azul"}, {"es": "M"}]
			}
		]
	}`

	var p Product
	if err := json.Unmarshal([]byte(productJSON), &p); err != nil {
		t.Fatalf("unmarshal product failed: %v", err)
	}

	if p.Name.String() != "Remera" {
		t.Errorf("product name = %q, want Remera", p.Name.String())
	}

	options := p.GetOptions()
	if len(options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(options))
	}
	if options[0].Name != "Color" || len(options[0].Values) != 2 {
		t.Errorf("unexpected option 0: %+v", options[0])
	}
	if options[1].Name != "Talle" || len(options[1].Values) != 2 {
		t.Errorf("unexpected option 1: %+v", options[1])
	}

	if p.Variants[0].Title() != "Rojo / S" {
		t.Errorf("variant 0 title = %q, want 'Rojo / S'", p.Variants[0].Title())
	}
	if p.Variants[1].Title() != "Azul / M" {
		t.Errorf("variant 1 title = %q, want 'Azul / M'", p.Variants[1].Title())
	}
}

func TestInventoryLevelsAndStock(t *testing.T) {
	variantJSON := `{
		"id": 101,
		"stock": null,
		"inventory_levels": [
			{
				"id": "1",
				"variant_id": "101",
				"location_id": "01HTMFDH09VC6E2Q8KGTGP44D3",
				"stock": 5
			},
			{
				"id": "2",
				"variant_id": "101",
				"location_id": "01HTMFDH09VC6E2Q8KGTGP44D4",
				"stock": 10
			}
		]
	}`

	var v ProductVariant
	if err := json.Unmarshal([]byte(variantJSON), &v); err != nil {
		t.Fatalf("unmarshal variant failed: %v", err)
	}

	if len(v.InventoryLevels) != 2 {
		t.Fatalf("expected 2 inventory levels, got %d", len(v.InventoryLevels))
	}
	if v.InventoryLevels[0].LocationID.String() != "01HTMFDH09VC6E2Q8KGTGP44D3" {
		t.Errorf("expected location_id '01HTMFDH09VC6E2Q8KGTGP44D3', got %q", v.InventoryLevels[0].LocationID.String())
	}
	if v.Inventory() != "15" {
		t.Errorf("v.Inventory() = %q, want 15", v.Inventory())
	}

	locations := map[string]string{
		"01HTMFDH09VC6E2Q8KGTGP44D3": "Depósito Central",
		"01HTMFDH09VC6E2Q8KGTGP44D4": "Sucursal Palermo",
	}
	wantDisplay := "15 (Depósito Central: 5, Sucursal Palermo: 10)"
	if got := v.InventoryDisplay(locations); got != wantDisplay {
		t.Errorf("v.InventoryDisplay() = %q, want %q", got, wantDisplay)
	}

	// Single location
	vSingle := ProductVariant{
		InventoryLevels: []InventoryLevel{
			{LocationID: "01HTMFDH09VC6E2Q8KGTGP44D3", Stock: v.InventoryLevels[0].Stock},
		},
	}
	if got := vSingle.InventoryDisplay(locations); got != "5 (Depósito Central)" {
		t.Errorf("vSingle.InventoryDisplay() = %q, want '5 (Depósito Central)'", got)
	}

	// Fallback to location ID when name unknown
	if got := vSingle.InventoryDisplay(nil); got != "5 (01HTMFDH09VC6E2Q8KGTGP44D3)" {
		t.Errorf("vSingle.InventoryDisplay(nil) = %q, want '5 (01HTMFDH09VC6E2Q8KGTGP44D3)'", got)
	}

	// Single location in map resolves name even if ID key has whitespace
	vSingleTrim := ProductVariant{
		InventoryLevels: []InventoryLevel{
			{LocationID: " 01M33MQVR7W28EGK6TKW2RQ588 ", Stock: v.InventoryLevels[0].Stock},
		},
	}
	singleLocMap := map[string]string{
		"01M33MQVR7W28EGK6TKW2RQ588": "Depósito Central",
	}
	if got := vSingleTrim.InventoryDisplay(singleLocMap); got != "5 (Depósito Central)" {
		t.Errorf("vSingleTrim.InventoryDisplay(singleLocMap) = %q, want '5 (Depósito Central)'", got)
	}

	// Variant with direct stock
	variantStock := `{"id": 102, "stock": 42}`
	var v2 ProductVariant
	if err := json.Unmarshal([]byte(variantStock), &v2); err != nil {
		t.Fatalf("unmarshal variant with stock failed: %v", err)
	}
	if v2.Inventory() != "42" {
		t.Errorf("v2.Inventory() = %q, want 42", v2.Inventory())
	}

	// Variant with null stock and no inventory levels
	variantNull := `{"id": 103, "stock": null}`
	var v3 ProductVariant
	if err := json.Unmarshal([]byte(variantNull), &v3); err != nil {
		t.Fatalf("unmarshal variant with null stock failed: %v", err)
	}
	if v3.Inventory() != "-" {
		t.Errorf("v3.Inventory() = %q, want '-'", v3.Inventory())
	}
}
