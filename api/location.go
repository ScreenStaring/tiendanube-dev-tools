package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// LocationAddress represents the address of a physical store or warehouse location.
type LocationAddress struct {
	Zipcode        string `json:"zipcode"`
	Street         string `json:"street"`
	Number         string `json:"number"`
	Floor          string `json:"floor"`
	Locality       string `json:"locality"`
	City           string `json:"city"`
	Reference      string `json:"reference"`
	BetweenStreets string `json:"between_streets"`
}

// Location represents a physical store or warehouse location in Tiendanube.
type Location struct {
	ID          NullableString   `json:"id"`
	Name        LocalizedValue   `json:"name"`
	Title       LocalizedValue   `json:"title,omitempty"`
	Label       NullableString   `json:"label,omitempty"`
	Description LocalizedValue   `json:"description,omitempty"`
	IsDefault   bool             `json:"is_default"`
	Priority    int              `json:"priority"`
	Address     *LocationAddress `json:"address,omitempty"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
}

// DisplayName returns the best human-friendly name for this location.
func (l *Location) DisplayName() string {
	if s := l.Name.String(); s != "" {
		return s
	}
	if s := l.Title.String(); s != "" {
		return s
	}
	if s := l.Label.String(); s != "" {
		return s
	}
	if s := l.Description.String(); s != "" {
		return s
	}
	if l.Address != nil {
		if l.Address.Reference != "" {
			return l.Address.Reference
		}
		if l.Address.City != "" {
			return l.Address.City
		}
	}
	if l.IsDefault {
		return "Default"
	}
	return ""
}

func parseLocationsJSON(raw []byte) ([]Location, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	// 1. Direct array of locations
	var locs []Location
	if err := json.Unmarshal(raw, &locs); err == nil {
		return locs, nil
	}

	// 2. Object wrapper: {"results": [...]} or {"locations": [...]} or {"data": [...]}
	var wrapper struct {
		Results   []Location `json:"results"`
		Locations []Location `json:"locations"`
		Data      []Location `json:"data"`
	}
	if err := json.Unmarshal(raw, &wrapper); err == nil {
		if len(wrapper.Results) > 0 {
			return wrapper.Results, nil
		}
		if len(wrapper.Locations) > 0 {
			return wrapper.Locations, nil
		}
		if len(wrapper.Data) > 0 {
			return wrapper.Data, nil
		}
	}

	// 3. Single location object: {"id": "...", "name": ...}
	var single Location
	if err := json.Unmarshal(raw, &single); err == nil && single.ID != "" {
		return []Location{single}, nil
	}

	return nil, fmt.Errorf("failed to parse locations JSON: %s", string(raw))
}

// ListLocations retrieves the store's physical locations.
func (c *Client) ListLocations(ctx context.Context) ([]Location, error) {
	var raw json.RawMessage
	err := c.Request(ctx, http.MethodGet, "locations", nil, nil, &raw)
	if err != nil {
		return nil, err
	}
	return parseLocationsJSON(raw)
}

// GetLocation retrieves a single location by ID.
func (c *Client) GetLocation(ctx context.Context, id string) (*Location, error) {
	path := fmt.Sprintf("locations/%s", url.PathEscape(id))
	var raw json.RawMessage
	err := c.Request(ctx, http.MethodGet, path, nil, nil, &raw)
	if err != nil {
		return nil, err
	}
	locs, err := parseLocationsJSON(raw)
	if err == nil && len(locs) > 0 {
		return &locs[0], nil
	}
	var single Location
	if unmarshalErr := json.Unmarshal(raw, &single); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to parse location: %w", unmarshalErr)
	}
	return &single, nil
}
