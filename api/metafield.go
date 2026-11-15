package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Metafield represents a namespaced key-value store item in Tiendanube.
type Metafield struct {
	ID            int64          `json:"id"`
	Namespace     string         `json:"namespace"`
	Key           string         `json:"key"`
	Value         string         `json:"value"`
	Description   NullableString `json:"description"`
	OwnerResource string         `json:"owner_resource"`
	OwnerID       int64          `json:"owner_id"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
}

// NormalizeOwnerResource standardizes owner resource name for API endpoint paths.
func NormalizeOwnerResource(resource string) string {
	res := strings.ToLower(strings.TrimSpace(resource))
	switch res {
	case "product", "products":
		return "products"
	case "variant", "variants", "product_variant", "product_variants":
		return "product_variants"
	case "category", "categories", "collection", "collections":
		return "categories"
	case "page", "pages":
		return "pages"
	case "order", "orders":
		return "orders"
	case "customer", "customers":
		return "customers"
	default:
		if strings.HasSuffix(res, "s") {
			return res
		}
		return res + "s"
	}
}

// ListMetafieldsParams contains query parameters for GET /metafields/{owner_resource}.
type ListMetafieldsParams struct {
	OwnerID      int64
	Namespace    string
	Key          string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
	Page         int
	PerPage      int
	Fields       string
}

func (p *ListMetafieldsParams) toQuery() url.Values {
	if p == nil {
		return nil
	}
	q := url.Values{}
	if p.OwnerID > 0 {
		q.Set("owner_id", strconv.FormatInt(p.OwnerID, 10))
	}
	if p.Namespace != "" {
		q.Set("namespace", p.Namespace)
	}
	if p.Key != "" {
		q.Set("key", p.Key)
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
	return q
}

// CreateMetafieldRequest contains the fields to create a new metafield.
type CreateMetafieldRequest struct {
	OwnerResource string `json:"owner_resource"`
	OwnerID       int64  `json:"owner_id"`
	Namespace     string `json:"namespace"`
	Key           string `json:"key"`
	Value         string `json:"value"`
	Description   string `json:"description,omitempty"`
}

// UpdateMetafieldRequest contains the fields to update an existing metafield.
type UpdateMetafieldRequest struct {
	Value       string `json:"value,omitempty"`
	Description string `json:"description,omitempty"`
}

// ListMetafields retrieves metafields for the specified owner resource.
func (c *Client) ListMetafields(ctx context.Context, ownerResource string, params *ListMetafieldsParams) ([]Metafield, error) {
	norm := NormalizeOwnerResource(ownerResource)
	path := fmt.Sprintf("metafields/%s", url.PathEscape(norm))
	var query url.Values
	if params != nil {
		query = params.toQuery()
	}

	var metafields []Metafield
	err := c.Request(ctx, http.MethodGet, path, query, nil, &metafields)
	if err != nil {
		return nil, fmt.Errorf("failed to list metafields for %s: %w", ownerResource, err)
	}
	return metafields, nil
}

// GetMetafield retrieves a single metafield by ID.
func (c *Client) GetMetafield(ctx context.Context, id int64) (*Metafield, error) {
	path := fmt.Sprintf("metafields/%d", id)
	var metafield Metafield
	err := c.Request(ctx, http.MethodGet, path, nil, nil, &metafield)
	if err != nil {
		return nil, fmt.Errorf("failed to get metafield %d: %w", id, err)
	}
	return &metafield, nil
}

// CreateMetafield creates a new metafield.
func (c *Client) CreateMetafield(ctx context.Context, req *CreateMetafieldRequest) (*Metafield, error) {
	if req == nil {
		return nil, fmt.Errorf("create metafield request cannot be nil")
	}
	var metafield Metafield
	err := c.Request(ctx, http.MethodPost, "metafields", nil, req, &metafield)
	if err != nil {
		return nil, fmt.Errorf("failed to create metafield: %w", err)
	}
	return &metafield, nil
}

// UpdateMetafield updates an existing metafield.
func (c *Client) UpdateMetafield(ctx context.Context, id int64, req *UpdateMetafieldRequest) (*Metafield, error) {
	if req == nil {
		return nil, fmt.Errorf("update metafield request cannot be nil")
	}
	path := fmt.Sprintf("metafields/%d", id)
	var metafield Metafield
	err := c.Request(ctx, http.MethodPut, path, nil, req, &metafield)
	if err != nil {
		return nil, fmt.Errorf("failed to update metafield %d: %w", id, err)
	}
	return &metafield, nil
}

// DeleteMetafield deletes an existing metafield by ID.
func (c *Client) DeleteMetafield(ctx context.Context, id int64) error {
	path := fmt.Sprintf("metafields/%d", id)
	err := c.Request(ctx, http.MethodDelete, path, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete metafield %d: %w", id, err)
	}
	return nil
}
