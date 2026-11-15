package api

import (
	"context"
	"encoding/json"
	"testing"
)

func TestOrderUnmarshalRealOrder(t *testing.T) {
	raw := []byte(`{
		"id": 2120301338,
		"store_id": "8275148",
		"number": 100,
		"status": "open",
		"payment_status": "pending",
		"shipping_status": "unpacked",
		"note": null,
		"owner_note": "",
		"subtotal": "55.00",
		"total": "55.00",
		"currency": "BRL",
		"products": [
			{
				"id": 3687628509,
				"product_id": 368965690,
				"variant_id": "1602347859",
				"name": "Jar of Food",
				"price": "55.00",
				"quantity": "1",
				"sku": null,
				"barcode": null
			}
		]
	}`)

	var order Order
	if err := json.Unmarshal(raw, &order); err != nil {
		t.Fatalf("failed to unmarshal order 2120301338: %v", err)
	}

	if order.ID != 2120301338 {
		t.Errorf("expected ID 2120301338, got %d", order.ID)
	}
	if len(order.Products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(order.Products))
	}
	p := order.Products[0]
	if p.VariantID != "1602347859" {
		t.Errorf("expected VariantID '1602347859', got %q", p.VariantID)
	}
	if p.Quantity != "1" {
		t.Errorf("expected Quantity '1', got %q", p.Quantity)
	}
	if p.ID != 3687628509 {
		t.Errorf("expected product ID 3687628509, got %d", p.ID)
	}
	if p.ProductID != 368965690 {
		t.Errorf("expected product_id 368965690, got %d", p.ProductID)
	}
}

func TestOrderUnmarshalListAndDetailVariants(t *testing.T) {
	// GET /orders list endpoint returns integer quantity & variant_id
	listRaw := []byte(`[
		{
			"id": 100,
			"products": [
				{
					"id": 1,
					"product_id": 10,
					"variant_id": 20,
					"quantity": 3
				}
			]
		}
	]`)

	var orders []Order
	if err := json.Unmarshal(listRaw, &orders); err != nil {
		t.Fatalf("failed to unmarshal list order: %v", err)
	}
	if len(orders) != 1 || len(orders[0].Products) != 1 {
		t.Fatalf("unexpected order/product count")
	}
	p := orders[0].Products[0]
	if p.Quantity.String() != "3" || p.Quantity.Int() != 3 {
		t.Errorf("expected Quantity 3, got string=%q int=%d", p.Quantity.String(), p.Quantity.Int())
	}
	if p.VariantID.String() != "20" || p.VariantID.Int64() != 20 {
		t.Errorf("expected VariantID 20, got string=%q int64=%d", p.VariantID.String(), p.VariantID.Int64())
	}

	// GET /orders/{id} detail endpoint returns string quantity & variant_id
	detailRaw := []byte(`{
		"id": 100,
		"products": [
			{
				"id": 1,
				"product_id": 10,
				"variant_id": "20",
				"quantity": "3"
			}
		]
	}`)

	var order Order
	if err := json.Unmarshal(detailRaw, &order); err != nil {
		t.Fatalf("failed to unmarshal detail order: %v", err)
	}
	if len(order.Products) != 1 {
		t.Fatalf("unexpected product count")
	}
	dp := order.Products[0]
	if dp.Quantity.String() != "3" || dp.Quantity.Int() != 3 {
		t.Errorf("expected Quantity 3, got string=%q int=%d", dp.Quantity.String(), dp.Quantity.Int())
	}
	if dp.VariantID.String() != "20" || dp.VariantID.Int64() != 20 {
		t.Errorf("expected VariantID 20, got string=%q int64=%d", dp.VariantID.String(), dp.VariantID.Int64())
	}
}

func TestGetOrderUnmarshal(t *testing.T) {
	raw := []byte(`{
		"id": 999,
		"number": 500,
		"status": "closed",
		"payment_status": "paid",
		"shipping_status": "shipped"
	}`)

	var order Order
	if err := json.Unmarshal(raw, &order); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if order.ID != 999 {
		t.Errorf("expected ID 999, got %d", order.ID)
	}
	if order.DisplayName() != "#500" {
		t.Errorf("expected DisplayName #500, got %s", order.DisplayName())
	}
}

func TestOrdersRealAPI(t *testing.T) {
	store, token := requireAPICredentials(t)
	client := NewClient(store, token)

	orders, err := client.ListOrders(context.Background(), &ListOrdersParams{PerPage: 5})
	if err != nil {
		t.Fatalf("ListOrders failed: %v", err)
	}
	t.Logf("Retrieved %d orders", len(orders))

	if len(orders) > 0 {
		order, err := client.GetOrder(context.Background(), orders[0].ID)
		if err != nil {
			t.Fatalf("GetOrder failed: %v", err)
		}
		if order.ID != orders[0].ID {
			t.Errorf("got order ID %d, want %d", order.ID, orders[0].ID)
		}
	}
}
