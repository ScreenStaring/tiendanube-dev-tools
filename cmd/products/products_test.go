package products

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

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

func TestDisplayFieldName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"CreatedAt", "Created At"},
		{"CanonicalURL", "Canonical URL"},
		{"FreeShipping", "Free Shipping"},
		{"RequiresShipping", "Requires Shipping"},
		{"IsKit", "Is Kit"},
		{"ID", "ID"},
		{"Name", "Name"},
	}

	for _, tt := range tests {
		if got := displayFieldName(tt.input); got != tt.want {
			t.Errorf("displayFieldName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsFieldToPrint(t *testing.T) {
	fields := []string{"name", "id"}
	if !isFieldToPrint("name", fields) {
		t.Error("expected 'name' to be field to print")
	}
	if !isFieldToPrint("id", fields) {
		t.Error("expected 'id' to be field to print")
	}
	if isFieldToPrint("price", fields) {
		t.Error("did not expect 'price' to be field to print")
	}
}

func TestPrintJSONL(t *testing.T) {
	prods := []api.Product{
		{
			ID:   123,
			Name: api.LocalizedValue{"es": "Remera"},
		},
	}

	out := captureOutput(func() {
		if err := PrintJSONL(prods); err != nil {
			t.Fatalf("PrintJSONL failed: %v", err)
		}
	})

	if !strings.Contains(out, `"id":123`) || !strings.Contains(out, `"Remera"`) {
		t.Errorf("unexpected PrintJSONL output: %s", out)
	}
}

func TestPrintFormatted(t *testing.T) {
	stock := 10
	prods := []api.Product{
		{
			ID:           456,
			Name:         api.LocalizedValue{"es": "Pantalon"},
			Published:    true,
			FreeShipping: false,
			Attributes:   []api.LocalizedValue{{"es": "Color"}},
			Variants: []api.ProductVariant{
				{
					ID:     1,
					Price:  "150.00",
					SKU:    "PAN-BLK",
					Stock:  &stock,
					Values: []api.LocalizedValue{{"es": "Negro"}},
				},
			},
		},
	}

	out := captureOutput(func() {
		PrintFormatted(prods, nil)
	})

	if !strings.Contains(out, "Pantalon") {
		t.Errorf("expected output to contain Pantalon: %s", out)
	}
	if !strings.Contains(out, "Variants") {
		t.Errorf("expected output to contain Variants section: %s", out)
	}
	if !strings.Contains(out, "PAN-BLK") {
		t.Errorf("expected output to contain SKU PAN-BLK: %s", out)
	}
	if strings.Contains(out, "Free Shipping") {
		t.Errorf("expected default output NOT to contain Free Shipping: %s", out)
	}
	if strings.Contains(out, "Requires Shipping") {
		t.Errorf("expected default output NOT to contain Requires Shipping: %s", out)
	}
	if !strings.Contains(out, "Is Kit") {
		t.Errorf("expected default output to contain Is Kit: %s", out)
	}

	// Test selecting free_shipping and requires_shipping explicitly
	selectedOut := captureOutput(func() {
		PrintFormatted(prods, []string{"free_shipping", "requires_shipping"})
	})
	if !strings.Contains(selectedOut, "Free Shipping") {
		t.Errorf("expected selected output to contain Free Shipping: %s", selectedOut)
	}
	if !strings.Contains(selectedOut, "Requires Shipping") {
		t.Errorf("expected selected output to contain Requires Shipping: %s", selectedOut)
	}
	if strings.Contains(selectedOut, "Pantalon") {
		t.Errorf("did not expect Pantalon when only shipping fields requested: %s", selectedOut)
	}

	// Test variant with inventory levels and location mapping
	stk1 := 7
	stk2 := 3
	prodsWithLocs := []api.Product{
		{
			ID:   789,
			Name: api.LocalizedValue{"es": "Campera"},
			Variants: []api.ProductVariant{
				{
					ID:    2,
					Price: "200.00",
					SKU:   "CMP-BLU",
					InventoryLevels: []api.InventoryLevel{
						{LocationID: "loc-1", Stock: &stk1},
						{LocationID: "loc-2", Stock: &stk2},
					},
				},
			},
		},
	}
	locMap := map[string]string{
		"loc-1": "Depósito",
		"loc-2": "Sucursal",
	}
	locOut := captureOutput(func() {
		PrintFormatted(prodsWithLocs, nil, locMap)
	})
	if !strings.Contains(locOut, "10 (Depósito: 7, Sucursal: 3)") {
		t.Errorf("expected output to contain location stock breakdown, got: %s", locOut)
	}

	// Test single location inventory shows location name
	stkSingle := 100
	prodSingleLoc := []api.Product{
		{
			ID:   1001,
			Name: api.LocalizedValue{"es": "Remera Classic"},
			Variants: []api.ProductVariant{
				{
					ID:    11,
					Price: "50.00",
					SKU:   "REM-CLS",
					InventoryLevels: []api.InventoryLevel{
						{LocationID: "01M33MQVR7W28EGK6TKW2RQ588", Stock: &stkSingle},
					},
				},
			},
		},
	}
	singleMap := map[string]string{
		"01M33MQVR7W28EGK6TKW2RQ588": "Depósito Central",
	}
	singleOut := captureOutput(func() {
		PrintFormatted(prodSingleLoc, nil, singleMap)
	})
	if !strings.Contains(singleOut, "100 (Depósito Central)") {
		t.Errorf("expected output to contain '100 (Depósito Central)', got: %s", singleOut)
	}
}
