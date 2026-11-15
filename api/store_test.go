package api

import (
	"context"
	"testing"
)

func TestStoreUnmarshal(t *testing.T) {
	sampleJSON := []byte(`{
		"id": 1234,
		"name": {
			"en": "Poké Store",
			"pt": "Poké Loja"
		},
		"description": {
			"en": "Best poke store",
			"pt": "Melhor loja poke"
		},
		"type": "gaming",
		"email": "owner@pokestore.com",
		"logo": "//d26lpennugtm8s.cloudfront.net/stores/046/logo.png",
		"contact_email": "contact@pokestore.com",
		"country": "BR",
		"facebook": "https://www.facebook.com/pokestore",
		"youtube": "https://www.youtube.com/@pokestore",
		"tiktok": "https://www.tiktok.com/@pokestore",
		"twitter": "https://twitter.com/pokestore",
		"instagram": "https://instagram.com/pokestore",
		"pinterest": null,
		"blog": null,
		"business_id": "12.345.678/0001-90",
		"business_name": "Poke Store LTDA",
		"business_address": "Rua Augusta 100",
		"address": "Av. Paulista 1000",
		"phone": "+55 11 912345678",
		"whatsapp_phone_number": "+55 11 999998888",
		"customer_accounts": "optional",
		"plan_name": "BR_2022_05_next_0000_000",
		"domains": [
			"http://www.pokestore.com",
			"http://pokestore.com"
		],
		"languages": {
			"en": {
				"currency": "USD",
				"active": true
			},
			"pt": {
				"currency": "BRL",
				"active": true
			}
		},
		"original_domain": "pokeloja.lojavirtualnuvem.com.br",
		"url_with_protocol": "https://www.pokestore.com",
		"main_currency": "BRL",
		"current_theme": "ipanema",
		"main_language": "pt",
		"admin_language": "pt",
		"created_at": "2013-01-01T08:12:51+0000",
		"updated_at": "2025-08-26T20:20:55+0000",
		"timezone_id": "America/Sao_Paulo"
	}`)

	store, err := parseStoreJSON(sampleJSON)
	if err != nil {
		t.Fatalf("unexpected error parsing store JSON: %v", err)
	}

	if store.ID.String() != "1234" {
		t.Errorf("expected ID 1234, got %q", store.ID.String())
	}
	if store.Name.String() != "Poké Loja" && store.Name.String() != "Poké Store" {
		t.Errorf("unexpected name: %q", store.Name.String())
	}
	if store.Email != "owner@pokestore.com" {
		t.Errorf("expected email owner@pokestore.com, got %q", store.Email)
	}
	if store.ContactEmail != "contact@pokestore.com" {
		t.Errorf("expected contact_email contact@pokestore.com, got %q", store.ContactEmail)
	}
	if store.Country != "BR" {
		t.Errorf("expected country BR, got %q", store.Country)
	}
	if store.MainCurrency != "BRL" {
		t.Errorf("expected main_currency BRL, got %q", store.MainCurrency)
	}
	if store.OriginalDomain != "pokeloja.lojavirtualnuvem.com.br" {
		t.Errorf("expected original_domain, got %q", store.OriginalDomain)
	}
	if store.URLWithProtocol != "https://www.pokestore.com" {
		t.Errorf("expected url_with_protocol, got %q", store.URLWithProtocol)
	}
	if store.TimezoneID != "America/Sao_Paulo" {
		t.Errorf("expected timezone_id America/Sao_Paulo, got %q", store.TimezoneID)
	}
	if len(store.Domains) != 2 {
		t.Errorf("expected 2 domains, got %d", len(store.Domains))
	}
}

func TestStoreUnmarshalWrapped(t *testing.T) {
	wrappedJSON := []byte(`{
		"store": {
			"id": "5678",
			"name": "Tienda Test",
			"email": "test@tienda.com",
			"main_currency": "ARS",
			"country": "AR"
		}
	}`)

	store, err := parseStoreJSON(wrappedJSON)
	if err != nil {
		t.Fatalf("unexpected error parsing wrapped store JSON: %v", err)
	}
	if store.ID.String() != "5678" {
		t.Errorf("expected ID 5678, got %q", store.ID.String())
	}
	if store.Name.String() != "Tienda Test" {
		t.Errorf("expected Name Tienda Test, got %q", store.Name.String())
	}
	if store.MainCurrency != "ARS" {
		t.Errorf("expected Currency ARS, got %q", store.MainCurrency)
	}
}

func TestGetStoreRealAPI(t *testing.T) {
	store, token := requireAPICredentials(t)
	client := NewClient(store, token)

	s, err := client.GetStore(context.Background())
	if err != nil {
		t.Fatalf("GetStore failed: %v", err)
	}
	if s.ID.String() == "" {
		t.Errorf("expected non-empty store ID")
	}
}
