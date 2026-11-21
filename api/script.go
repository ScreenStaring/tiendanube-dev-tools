package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ScriptVersion represents a version of a script in Tiendanube/Nuvemshop.
type ScriptVersion struct {
	ID      int64  `json:"id"`
	Src     string `json:"src"`
	Version int    `json:"version"`
}

// Script represents an injected JavaScript tag/resource on the store storefront.
type Script struct {
	ID             int64          `json:"id"`
	Name           string         `json:"name,omitempty"`
	Event          string         `json:"event,omitempty"`
	Location       string         `json:"location,omitempty"`
	Status         string         `json:"status,omitempty"`
	IsAutoInstall  bool           `json:"is_auto_install,omitempty"`
	Src            string         `json:"src,omitempty"`
	Params         interface{}    `json:"params,omitempty"`
	CurrentVersion *ScriptVersion `json:"current_version,omitempty"`
	DraftVersion   *ScriptVersion `json:"draft_version,omitempty"`
	CreatedAt      string         `json:"created_at,omitempty"`
	UpdatedAt      string         `json:"updated_at,omitempty"`
}

// GetSrc returns the script URL, normalizing differences between unversioned and versioned API responses.
//
// Priority order:
//  1. Src: Direct, canonical script URL returned in flat/unversioned responses.
//  2. CurrentVersion.Src: The active, published version running on the live storefront for managed/versioned scripts.
//  3. DraftVersion.Src: Fallback to the staged, unpublished version when a script is in development and has no live version yet.
func (s *Script) GetSrc() string {
	if s.Src != "" {
		return s.Src
	}
	if s.CurrentVersion != nil && s.CurrentVersion.Src != "" {
		return s.CurrentVersion.Src
	}
	if s.DraftVersion != nil && s.DraftVersion.Src != "" {
		return s.DraftVersion.Src
	}
	return ""
}

// ListScriptsParams contains optional filters for listing scripts.
type ListScriptsParams struct {
	Page     int
	PerPage  int
	Event    string
	Location string
	Fields   string
}

func (p *ListScriptsParams) toQuery() url.Values {
	if p == nil {
		return nil
	}
	q := url.Values{}
	if p.Page > 0 {
		q.Set("page", strconv.Itoa(p.Page))
	}
	if p.PerPage > 0 {
		q.Set("per_page", strconv.Itoa(p.PerPage))
	}
	if p.Event != "" {
		q.Set("event", p.Event)
	}
	if p.Location != "" {
		q.Set("location", p.Location)
	}
	if p.Fields != "" {
		q.Set("fields", p.Fields)
	}
	return q
}

// ListScripts retrieves scripts from the store.
// Endpoint: GET /{api_version}/{store_id}/scripts
func (c *Client) ListScripts(ctx context.Context, params *ListScriptsParams) ([]Script, error) {
	var query url.Values
	if params != nil {
		query = params.toQuery()
	}

	var raw json.RawMessage
	err := c.Request(ctx, http.MethodGet, "scripts", query, nil, &raw)
	if err != nil {
		return nil, fmt.Errorf("failed to list scripts: %w", err)
	}

	return parseScriptsJSON(raw)
}

func parseScriptsJSON(raw []byte) ([]Script, error) {
	// Response might be a direct array: []Script
	var scripts []Script
	if err := json.Unmarshal(raw, &scripts); err == nil {
		return scripts, nil
	}

	// Or it might be wrapped in an object: {"result": []Script, "total": 1}
	var wrapper struct {
		Result []Script `json:"result"`
	}
	if err := json.Unmarshal(raw, &wrapper); err == nil {
		return wrapper.Result, nil
	}

	return nil, fmt.Errorf("failed to parse scripts response: %s", string(raw))
}

// CreateScriptRequest represents the payload to associate a script with a store.
type CreateScriptRequest struct {
	ScriptID    int64  `json:"script_id"`
	QueryParams string `json:"query_params,omitempty"`
}

// CreateScript associates a script with the store.
// Endpoint: POST /{api_version}/{store_id}/scripts
func (c *Client) CreateScript(ctx context.Context, req *CreateScriptRequest) (*Script, error) {
	if req == nil {
		return nil, fmt.Errorf("create script request cannot be nil")
	}
	if req.ScriptID == 0 {
		return nil, fmt.Errorf("script_id is required")
	}

	var raw json.RawMessage
	err := c.Request(ctx, http.MethodPost, "scripts", nil, req, &raw)
	if err != nil {
		return nil, fmt.Errorf("failed to create script: %w", err)
	}

	return parseScriptJSON(raw)
}

func parseScriptJSON(raw []byte) (*Script, error) {
	// 1. Try envelope: {"message": "...", "content": {...}}
	var wrapper struct {
		Message string  `json:"message"`
		Content *Script `json:"content"`
	}
	if err := json.Unmarshal(raw, &wrapper); err == nil && wrapper.Content != nil {
		return wrapper.Content, nil
	}

	// 2. Try direct Script: {"id": 12345, ...}
	var script Script
	if err := json.Unmarshal(raw, &script); err == nil && script.ID != 0 {
		return &script, nil
	}

	return nil, fmt.Errorf("failed to parse script response: %s", string(raw))
}

// GetScript retrieves a single script by ID.
// Endpoint: GET /{api_version}/{store_id}/scripts/{id}
func (c *Client) GetScript(ctx context.Context, id int64) (*Script, error) {
	var script Script
	err := c.Request(ctx, http.MethodGet, fmt.Sprintf("scripts/%d", id), nil, nil, &script)
	if err != nil {
		return nil, fmt.Errorf("failed to get script %d: %w", id, err)
	}
	return &script, nil
}

// UpdateScriptRequest represents the payload to update a script's association with a store.
type UpdateScriptRequest struct {
	ScriptID    int64  `json:"script_id,omitempty"`
	QueryParams string `json:"query_params,omitempty"`
}

// UpdateScript updates an existing script association by ID.
// Endpoint: PUT /{api_version}/{store_id}/scripts/{id}
func (c *Client) UpdateScript(ctx context.Context, id int64, req *UpdateScriptRequest) (*Script, error) {
	if req == nil {
		return nil, fmt.Errorf("update script request cannot be nil")
	}
	if id == 0 {
		return nil, fmt.Errorf("script id is required")
	}

	payload := *req
	if payload.ScriptID == 0 {
		payload.ScriptID = id
	}

	var raw json.RawMessage
	err := c.Request(ctx, http.MethodPut, fmt.Sprintf("scripts/%d", id), nil, &payload, &raw)
	if err != nil {
		return nil, fmt.Errorf("failed to update script %d: %w", id, err)
	}

	if len(raw) == 0 || string(raw) == "{}" {
		return &Script{ID: id}, nil
	}

	return parseScriptJSON(raw)
}

// DeleteScript deletes a script by ID.
// Endpoint: DELETE /{api_version}/{store_id}/scripts/{id}
func (c *Client) DeleteScript(ctx context.Context, id int64) error {
	if id == 0 {
		return fmt.Errorf("script id is required")
	}
	err := c.Request(ctx, http.MethodDelete, fmt.Sprintf("scripts/%d", id), nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete script %d: %w", id, err)
	}
	return nil
}
