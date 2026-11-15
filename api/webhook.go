package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Webhook represents a Tiendanube/Nuvemshop webhook subscription.
type Webhook struct {
	ID        int64  `json:"id"`
	Event     string `json:"event"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ListWebhooksParams contains optional filters for listing webhooks.
type ListWebhooksParams struct {
	SinceID      int64
	URL          string
	Event        string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
	Page         int
	PerPage      int
	Fields       string
}

func (p *ListWebhooksParams) toQuery() url.Values {
	if p == nil {
		return nil
	}
	q := url.Values{}
	if p.SinceID > 0 {
		q.Set("since_id", strconv.FormatInt(p.SinceID, 10))
	}
	if p.URL != "" {
		q.Set("url", p.URL)
	}
	if p.Event != "" {
		q.Set("event", p.Event)
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

// CreateWebhookRequest represents the payload to create a webhook.
type CreateWebhookRequest struct {
	Event string `json:"event"`
	URL   string `json:"url"`
}

// UpdateWebhookRequest represents the payload to update an existing webhook.
type UpdateWebhookRequest struct {
	Event string `json:"event,omitempty"`
	URL   string `json:"url,omitempty"`
}

// ListWebhooks retrieves webhooks matching the optional parameters.
// Endpoint: GET /{api_version}/{store_id}/webhooks
func (c *Client) ListWebhooks(ctx context.Context, params *ListWebhooksParams) ([]Webhook, error) {
	var webhooks []Webhook
	var query url.Values
	if params != nil {
		query = params.toQuery()
	}

	err := c.Request(ctx, http.MethodGet, "webhooks", query, nil, &webhooks)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}
	return webhooks, nil
}

// GetWebhook retrieves a single webhook by ID.
// Endpoint: GET /{api_version}/{store_id}/webhooks/{id}
func (c *Client) GetWebhook(ctx context.Context, id int64, fields ...string) (*Webhook, error) {
	path := fmt.Sprintf("webhooks/%d", id)
	var query url.Values
	if len(fields) > 0 && fields[0] != "" {
		query = url.Values{"fields": []string{fields[0]}}
	}

	var webhook Webhook
	err := c.Request(ctx, http.MethodGet, path, query, nil, &webhook)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook %d: %w", id, err)
	}
	return &webhook, nil
}

// CreateWebhook registers a new webhook.
// Endpoint: POST /{api_version}/{store_id}/webhooks
func (c *Client) CreateWebhook(ctx context.Context, req *CreateWebhookRequest) (*Webhook, error) {
	if req == nil {
		return nil, fmt.Errorf("create webhook request cannot be nil")
	}
	if req.Event == "" {
		return nil, fmt.Errorf("event is required")
	}
	if req.URL == "" {
		return nil, fmt.Errorf("url is required")
	}

	var webhook Webhook
	err := c.Request(ctx, http.MethodPost, "webhooks", nil, req, &webhook)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}
	return &webhook, nil
}

// UpdateWebhook updates an existing webhook.
// Endpoint: PUT /{api_version}/{store_id}/webhooks/{id}
func (c *Client) UpdateWebhook(ctx context.Context, id int64, req *UpdateWebhookRequest) (*Webhook, error) {
	if req == nil {
		return nil, fmt.Errorf("update webhook request cannot be nil")
	}
	if req.Event == "" && req.URL == "" {
		return nil, fmt.Errorf("at least one of event or url must be provided")
	}

	path := fmt.Sprintf("webhooks/%d", id)
	var webhook Webhook
	err := c.Request(ctx, http.MethodPut, path, nil, req, &webhook)
	if err != nil {
		return nil, fmt.Errorf("failed to update webhook %d: %w", id, err)
	}
	return &webhook, nil
}

// DeleteWebhook removes a webhook.
// Endpoint: DELETE /{api_version}/{store_id}/webhooks/{id}
func (c *Client) DeleteWebhook(ctx context.Context, id int64) error {
	path := fmt.Sprintf("webhooks/%d", id)
	err := c.Request(ctx, http.MethodDelete, path, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete webhook %d: %w", id, err)
	}
	return nil
}
