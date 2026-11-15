package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// StoreLanguage represents a language supported by the store.
type StoreLanguage struct {
	Currency string `json:"currency"`
	Active   bool   `json:"active"`
}

// Store represents a Tiendanube store and its configuration.
type Store struct {
	ID                  NullableString           `json:"id"`
	Name                LocalizedValue           `json:"name"`
	Description         LocalizedValue           `json:"description,omitempty"`
	Type                NullableString           `json:"type,omitempty"`
	Email               string                   `json:"email"`
	Logo                NullableString           `json:"logo,omitempty"`
	ContactEmail        string                   `json:"contact_email"`
	Country             string                   `json:"country"`
	Facebook            NullableString           `json:"facebook,omitempty"`
	Youtube             NullableString           `json:"youtube,omitempty"`
	Tiktok              NullableString           `json:"tiktok,omitempty"`
	Twitter             NullableString           `json:"twitter,omitempty"`
	Instagram           NullableString           `json:"instagram,omitempty"`
	Pinterest           NullableString           `json:"pinterest,omitempty"`
	Blog                NullableString           `json:"blog,omitempty"`
	BusinessID          NullableString           `json:"business_id,omitempty"`
	BusinessName        NullableString           `json:"business_name,omitempty"`
	BusinessAddress     NullableString           `json:"business_address,omitempty"`
	Address             NullableString           `json:"address,omitempty"`
	Phone               NullableString           `json:"phone,omitempty"`
	WhatsappPhoneNumber NullableString           `json:"whatsapp_phone_number,omitempty"`
	CustomerAccounts    NullableString           `json:"customer_accounts,omitempty"`
	PlanName            string                   `json:"plan_name"`
	Domains             []string                 `json:"domains,omitempty"`
	Languages           map[string]StoreLanguage `json:"languages,omitempty"`
	OriginalDomain      string                   `json:"original_domain"`
	URLWithProtocol     string                   `json:"url_with_protocol"`
	MainCurrency        string                   `json:"main_currency"`
	CurrentTheme        NullableString           `json:"current_theme,omitempty"`
	MainLanguage        string                   `json:"main_language"`
	AdminLanguage       string                   `json:"admin_language"`
	CreatedAt           string                   `json:"created_at"`
	UpdatedAt           string                   `json:"updated_at"`
	TimezoneID          string                   `json:"timezone_id"`
}

func parseStoreJSON(raw []byte) (*Store, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, fmt.Errorf("empty store response")
	}

	var store Store
	if err := json.Unmarshal(raw, &store); err == nil && store.ID != "" {
		return &store, nil
	}

	var wrapper struct {
		Store Store `json:"store"`
	}
	if err := json.Unmarshal(raw, &wrapper); err == nil && wrapper.Store.ID != "" {
		return &wrapper.Store, nil
	}

	if err := json.Unmarshal(raw, &store); err != nil {
		return nil, fmt.Errorf("failed to parse store: %w", err)
	}
	return &store, nil
}

// GetStore retrieves the authenticated store's settings and information.
func (c *Client) GetStore(ctx context.Context) (*Store, error) {
	var raw json.RawMessage
	err := c.Request(ctx, http.MethodGet, "store", nil, nil, &raw)
	if err != nil {
		return nil, err
	}

	return parseStoreJSON(raw)
}
