package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// OrderProduct represents a line item in an Order.
// Note: VariantID and Quantity use NullableString because Tiendanube returns them
// as integers in order listings (GET /orders), but as strings in single order
// retrievals (GET /orders/{id}).
type OrderProduct struct {
	ID           int64          `json:"id"`
	ProductID    int64          `json:"product_id"`
	VariantID    NullableString `json:"variant_id"` // int in GET /orders, string in GET /orders/{id}
	Name         string         `json:"name"`
	Price        string         `json:"price"`
	Quantity     NullableString `json:"quantity"` // int in GET /orders, string in GET /orders/{id}
	Weight       string         `json:"weight,omitempty"`
	Width        string         `json:"width,omitempty"`
	Height       string         `json:"height,omitempty"`
	Depth        string         `json:"depth,omitempty"`
	FreeShipping bool           `json:"free_shipping,omitempty"`
	SKU          NullableString `json:"sku,omitempty"`
	Barcode      NullableString `json:"barcode,omitempty"`
	Status       string         `json:"status,omitempty"`
}

// OrderCustomer represents customer info attached to an Order.
type OrderCustomer struct {
	ID             int64          `json:"id"`
	Name           string         `json:"name"`
	Email          string         `json:"email"`
	Identification NullableString `json:"identification,omitempty"`
	Phone          NullableString `json:"phone,omitempty"`
}

// Order represents a Tiendanube/Nuvemshop order.
type Order struct {
	ID             int64          `json:"id"`
	Token          string         `json:"token,omitempty"`
	StoreID        NullableString `json:"store_id,omitempty"`
	Number         int64          `json:"number,omitempty"`
	Name           string         `json:"name,omitempty"`
	ContactEmail   string         `json:"contact_email,omitempty"`
	ContactName    string         `json:"contact_name,omitempty"`
	ContactPhone   string         `json:"contact_phone,omitempty"`
	Status         string         `json:"status,omitempty"`
	PaymentStatus  string         `json:"payment_status,omitempty"`
	ShippingStatus string         `json:"shipping_status,omitempty"`
	Subtotal       string         `json:"subtotal,omitempty"`
	Discount       string         `json:"discount,omitempty"`
	Total          string         `json:"total,omitempty"`
	TotalUSD       string         `json:"total_usd,omitempty"`
	Currency       string         `json:"currency,omitempty"`
	Language       string         `json:"language,omitempty"`
	Gateway        string         `json:"gateway,omitempty"`
	GatewayName    string         `json:"gateway_name,omitempty"`
	Note           string         `json:"note,omitempty"`
	OwnerNote      string         `json:"owner_note,omitempty"`
	CreatedAt      string         `json:"created_at,omitempty"`
	UpdatedAt      string         `json:"updated_at,omitempty"`
	CancelledAt    NullableString `json:"cancelled_at,omitempty"`
	ClosedAt       NullableString `json:"closed_at,omitempty"`
	PaidAt         NullableString `json:"paid_at,omitempty"`
	Products       []OrderProduct `json:"products,omitempty"`
	Customer       *OrderCustomer `json:"customer,omitempty"`
}

// DisplayName returns order name (or #<number> / #<id>).
func (o *Order) DisplayName() string {
	if o.Name != "" {
		return o.Name
	}
	if o.Number != 0 {
		return fmt.Sprintf("#%d", o.Number)
	}
	return fmt.Sprintf("#%d", o.ID)
}

// DisplayFinancialStatus returns order payment/financial status.
func (o *Order) DisplayFinancialStatus() string {
	return o.PaymentStatus
}

// DisplayFulfillmentStatus returns order shipping/fulfillment status.
func (o *Order) DisplayFulfillmentStatus() string {
	return o.ShippingStatus
}

// ListOrdersParams contains query parameters for GET /orders.
type ListOrdersParams struct {
	SinceID         int64
	Status          string
	PaymentStatus   string
	ShippingStatus  string
	Channels        string
	CreatedAtMin    string
	CreatedAtMax    string
	UpdatedAtMin    string
	UpdatedAtMax    string
	TotalMin        string
	TotalMax        string
	CustomerIDs     string
	Page            int
	PerPage         int
	Fields          string
	Q               string
	AppID           string
	PaymentMethods  string
	PaymentProvider string
	Aggregates      string
}

func (p *ListOrdersParams) toQuery() url.Values {
	if p == nil {
		return nil
	}
	q := url.Values{}
	if p.SinceID > 0 {
		q.Set("since_id", strconv.FormatInt(p.SinceID, 10))
	}
	if p.Status != "" {
		q.Set("status", p.Status)
	}
	if p.PaymentStatus != "" {
		q.Set("payment_status", p.PaymentStatus)
	}
	if p.ShippingStatus != "" {
		q.Set("shipping_status", p.ShippingStatus)
	}
	if p.Channels != "" {
		q.Set("channels", p.Channels)
	}
	if p.CreatedAtMin != "" {
		q.Set("created_at_min", p.CreatedAtMin)
	}
	if p.CreatedAtMax != "" {
		q.Set("created_at_max", p.CreatedAtMax)
	}
	if p.UpdatedAtMin != "" {
		q.Set("updated_at_min", p.UpdatedAtMin)
	}
	if p.UpdatedAtMax != "" {
		q.Set("updated_at_max", p.UpdatedAtMax)
	}
	if p.TotalMin != "" {
		q.Set("total_min", p.TotalMin)
	}
	if p.TotalMax != "" {
		q.Set("total_max", p.TotalMax)
	}
	if p.CustomerIDs != "" {
		q.Set("customer_ids", p.CustomerIDs)
	}
	if p.Page > 0 {
		q.Set("page", strconv.Itoa(p.Page))
	}
	if p.PerPage > 0 {
		q.Set("per_page", strconv.Itoa(p.PerPage))
	}
	if p.Fields != "" {
		q.Set("fields", p.Fields)
	}
	if p.Q != "" {
		q.Set("q", p.Q)
	}
	if p.AppID != "" {
		q.Set("app_id", p.AppID)
	}
	if p.PaymentMethods != "" {
		q.Set("payment_methods", p.PaymentMethods)
	}
	if p.PaymentProvider != "" {
		q.Set("payment_provider", p.PaymentProvider)
	}
	if p.Aggregates != "" {
		q.Set("aggregates", p.Aggregates)
	}
	return q
}

// ListOrders retrieves orders matching the query parameters.
// Endpoint: GET /{api_version}/{store_id}/orders
func (c *Client) ListOrders(ctx context.Context, params *ListOrdersParams) ([]Order, error) {
	var orders []Order
	var query url.Values
	if params != nil {
		query = params.toQuery()
	}

	err := c.Request(ctx, http.MethodGet, "orders", query, nil, &orders)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	return orders, nil
}

// GetOrder retrieves a single order by ID.
// Endpoint: GET /{api_version}/{store_id}/orders/{id}
func (c *Client) GetOrder(ctx context.Context, id int64) (*Order, error) {
	var order Order
	err := c.Request(ctx, http.MethodGet, fmt.Sprintf("orders/%d", id), nil, nil, &order)
	if err != nil {
		return nil, fmt.Errorf("failed to get order %d: %w", id, err)
	}
	return &order, nil
}
