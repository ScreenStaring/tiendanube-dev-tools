package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// NullableString unmarshals JSON strings, numbers, or null into a string.
type NullableString string

func (ns *NullableString) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*ns = ""
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*ns = NullableString(s)
		return nil
	}
	*ns = NullableString(string(b))
	return nil
}

func (ns NullableString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(ns))
}

func (ns NullableString) String() string {
	return string(ns)
}

func (ns NullableString) Int() int {
	v, _ := strconv.Atoi(string(ns))
	return v
}

func (ns NullableString) Int64() int64 {
	v, _ := strconv.ParseInt(string(ns), 10, 64)
	return v
}

// LocalizedValue represents a translatable string in Tiendanube (e.g. {"en":"Shirt","es":"Remera"}).
type LocalizedValue map[string]string

func (l *LocalizedValue) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*l = nil
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*l = map[string]string{"default": s}
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		var raw map[string]interface{}
		if err2 := json.Unmarshal(b, &raw); err2 == nil {
			res := make(map[string]string)
			for k, v := range raw {
				if v != nil {
					res[k] = fmt.Sprint(v)
				}
			}
			*l = res
			return nil
		}
		return err
	}
	*l = m
	return nil
}

func (l LocalizedValue) String() string {
	if len(l) == 0 {
		return ""
	}
	for _, lang := range []string{"es", "pt", "en", "default"} {
		if val, ok := l[lang]; ok && val != "" {
			return val
		}
	}
	for _, prefix := range []string{"es", "pt", "en"} {
		for k, val := range l {
			if strings.HasPrefix(strings.ToLower(k), prefix) && val != "" {
				return val
			}
		}
	}
	for _, val := range l {
		if val != "" {
			return val
		}
	}
	return ""
}

// InventoryLevel represents inventory balance for a variant at a specific location.
type InventoryLevel struct {
	ID         NullableString `json:"id,omitempty"`
	VariantID  NullableString `json:"variant_id,omitempty"`
	LocationID NullableString `json:"location_id"`
	Stock      *int           `json:"stock"`
}

// ProductVariant represents a Tiendanube product variant.
type ProductVariant struct {
	ID               int64            `json:"id"`
	ProductID        int64            `json:"product_id"`
	Price            NullableString   `json:"price"`
	CompareAtPrice   NullableString   `json:"compare_at_price"`
	PromotionalPrice NullableString   `json:"promotional_price"`
	Stock            *int             `json:"stock"`
	SKU              NullableString   `json:"sku"`
	Barcode          NullableString   `json:"barcode"`
	Weight           NullableString   `json:"weight"`
	Width            NullableString   `json:"width"`
	Height           NullableString   `json:"height"`
	Depth            NullableString   `json:"depth"`
	Values           []LocalizedValue `json:"values"`
	InventoryLevels  []InventoryLevel `json:"inventory_levels"`
	CreatedAt        string           `json:"created_at"`
	UpdatedAt        string           `json:"updated_at"`
}

func (v *ProductVariant) Title() string {
	if len(v.Values) == 0 {
		return "Default"
	}
	var parts []string
	for _, val := range v.Values {
		s := val.String()
		if s != "" {
			parts = append(parts, s)
		}
	}
	if len(parts) == 0 {
		return "Default"
	}
	return strings.Join(parts, " / ")
}

func (v *ProductVariant) Inventory() string {
	if v.Stock != nil {
		return strconv.Itoa(*v.Stock)
	}
	if len(v.InventoryLevels) > 0 {
		total := 0
		hasStock := false
		for _, lvl := range v.InventoryLevels {
			if lvl.Stock != nil {
				total += *lvl.Stock
				hasStock = true
			}
		}
		if hasStock {
			return strconv.Itoa(total)
		}
	}
	return "-"
}

// InventoryDisplay returns a human-friendly representation of the variant's inventory,
// resolving location IDs to names using the provided locations map when available.
func (v *ProductVariant) InventoryDisplay(locations map[string]string) string {
	if len(v.InventoryLevels) == 0 {
		return v.Inventory()
	}

	total := 0
	hasStock := false
	type locStock struct {
		name  string
		stock int
	}
	var entries []locStock

	for _, lvl := range v.InventoryLevels {
		if lvl.Stock == nil {
			continue
		}
		hasStock = true
		total += *lvl.Stock

		name := ""
		locID := strings.TrimSpace(lvl.LocationID.String())
		if locations != nil && locID != "" {
			name = locations[locID]
		}
		if name == "" && len(locations) == 1 {
			for _, locName := range locations {
				if locName != "" {
					name = locName
					break
				}
			}
		}
		if name == "" {
			name = locID
		}

		entries = append(entries, locStock{name: name, stock: *lvl.Stock})
	}

	if !hasStock {
		return v.Inventory()
	}

	if len(entries) == 1 {
		if entries[0].name != "" {
			return fmt.Sprintf("%d (%s)", total, entries[0].name)
		}
		return strconv.Itoa(total)
	}

	var parts []string
	for _, e := range entries {
		if e.name != "" {
			parts = append(parts, fmt.Sprintf("%s: %d", e.name, e.stock))
		} else {
			parts = append(parts, strconv.Itoa(e.stock))
		}
	}

	return fmt.Sprintf("%d (%s)", total, strings.Join(parts, ", "))
}

// ProductImage represents a product image.
type ProductImage struct {
	ID        int64  `json:"id"`
	ProductID int64  `json:"product_id"`
	Src       string `json:"src"`
	Position  int    `json:"position"`
}

// ProductOption represents an option name and its distinct values.
type ProductOption struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// Product represents a Tiendanube product.
type Product struct {
	ID               int64            `json:"id"`
	Name             LocalizedValue   `json:"name"`
	Description      LocalizedValue   `json:"description"`
	Handle           LocalizedValue   `json:"handle"`
	Published        bool             `json:"published"`
	FreeShipping     bool             `json:"free_shipping"`
	RequiresShipping bool             `json:"requires_shipping"`
	IsKit            bool             `json:"is_kit"`
	CanonicalURL     NullableString   `json:"canonical_url"`
	Brand            NullableString   `json:"brand"`
	Tags             NullableString   `json:"tags"`
	CreatedAt        string           `json:"created_at"`
	UpdatedAt        string           `json:"updated_at"`
	Attributes       []LocalizedValue `json:"attributes,omitempty"`
	Variants         []ProductVariant `json:"variants,omitempty"`
	Images           []ProductImage   `json:"images,omitempty"`
}

func (p *Product) GetOptions() []ProductOption {
	if len(p.Attributes) == 0 {
		return nil
	}
	options := make([]ProductOption, len(p.Attributes))
	for i, attr := range p.Attributes {
		options[i].Name = attr.String()
		seen := make(map[string]bool)
		for _, v := range p.Variants {
			if i < len(v.Values) {
				valStr := v.Values[i].String()
				if valStr != "" && !seen[valStr] {
					seen[valStr] = true
					options[i].Values = append(options[i].Values, valStr)
				}
			}
		}
	}
	return options
}

func (p *Product) HasOnlyDefaultVariant() bool {
	if len(p.Variants) > 1 {
		return false
	}
	if len(p.Attributes) > 0 {
		return false
	}
	if len(p.Variants) == 1 && len(p.Variants[0].Values) > 0 {
		return false
	}
	return true
}

// ListProductsParams contains query parameters for GET /products.
type ListProductsParams struct {
	SinceID      int64
	Q            string
	Handle       string
	CategoryID   int64
	Published    *bool
	FreeShipping *bool
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
	Page         int
	PerPage      int
	Fields       string
	Language     string
}

func (p *ListProductsParams) toQuery() url.Values {
	if p == nil {
		return nil
	}
	q := url.Values{}
	if p.SinceID > 0 {
		q.Set("since_id", strconv.FormatInt(p.SinceID, 10))
	}
	if p.Q != "" {
		q.Set("q", p.Q)
	}
	if p.Handle != "" {
		q.Set("handle", p.Handle)
	}
	if p.CategoryID > 0 {
		q.Set("category_id", strconv.FormatInt(p.CategoryID, 10))
	}
	if p.Published != nil {
		q.Set("published", strconv.FormatBool(*p.Published))
	}
	if p.FreeShipping != nil {
		q.Set("free_shipping", strconv.FormatBool(*p.FreeShipping))
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
	if p.Page > 0 {
		q.Set("page", strconv.Itoa(p.Page))
	}
	if p.PerPage > 0 {
		q.Set("per_page", strconv.Itoa(p.PerPage))
	}
	if p.Fields != "" {
		q.Set("fields", p.Fields)
	}
	if p.Language != "" {
		q.Set("language", p.Language)
	}
	return q
}

// ListProducts retrieves products matching the given parameters.
func (c *Client) ListProducts(ctx context.Context, params *ListProductsParams) ([]Product, error) {
	var products []Product
	var query url.Values
	if params != nil {
		query = params.toQuery()
	}

	err := c.Request(ctx, http.MethodGet, "products", query, nil, &products)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	return products, nil
}

// GetProduct retrieves a single product by ID.
func (c *Client) GetProduct(ctx context.Context, id int64, fields ...string) (*Product, error) {
	path := fmt.Sprintf("products/%d", id)
	var query url.Values
	if len(fields) > 0 && fields[0] != "" {
		query = url.Values{"fields": []string{fields[0]}}
	}

	var product Product
	err := c.Request(ctx, http.MethodGet, path, query, nil, &product)
	if err != nil {
		return nil, fmt.Errorf("failed to get product %d: %w", id, err)
	}
	return &product, nil
}

// GetProductBySKU retrieves the product that owns the variant with the given SKU.
func (c *Client) GetProductBySKU(ctx context.Context, sku string) (*Product, error) {
	path := fmt.Sprintf("products/sku/%s", url.PathEscape(sku))
	var product Product
	err := c.Request(ctx, http.MethodGet, path, nil, nil, &product)
	if err != nil {
		return nil, fmt.Errorf("failed to get product by sku %s: %w", sku, err)
	}
	return &product, nil
}

// DeleteProduct deletes a product by ID.
func (c *Client) DeleteProduct(ctx context.Context, id int64) error {
	path := fmt.Sprintf("products/%d", id)
	err := c.Request(ctx, http.MethodDelete, path, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete product %d: %w", id, err)
	}
	return nil
}
