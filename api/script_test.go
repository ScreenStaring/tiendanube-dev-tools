package api

import (
	"context"
	"strings"
	"testing"
)

func TestParseScriptsJSONArray(t *testing.T) {
	raw := []byte(`[
		{
			"id": 101,
			"name": "My Script",
			"event": "onload",
			"location": "store",
			"src": "https://example.com/app.js",
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-02T00:00:00Z"
		}
	]`)

	scripts, err := parseScriptsJSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(scripts) != 1 {
		t.Fatalf("expected 1 script, got %d", len(scripts))
	}

	s := scripts[0]
	if s.ID != 101 {
		t.Errorf("expected ID 101, got %d", s.ID)
	}
	if s.GetSrc() != "https://example.com/app.js" {
		t.Errorf("expected src https://example.com/app.js, got %s", s.GetSrc())
	}
	if s.Location != "store" {
		t.Errorf("expected location store, got %s", s.Location)
	}
}

func TestParseScriptsJSONWrapped(t *testing.T) {
	raw := []byte(`{
		"result": [
			{
				"id": 202,
				"name": "Script 202",
				"event": "onfirstinteraction",
				"location": "store",
				"status": "active",
				"is_auto_install": false,
				"current_version": {
					"id": 1,
					"src": "https://apps-scripts.tiendanube.com/app/1.js",
					"version": 1
				},
				"created_at": "2024-09-11T20:19:29.000Z",
				"updated_at": "2024-09-11T20:20:19.000Z"
			}
		],
		"total": 1
	}`)

	scripts, err := parseScriptsJSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(scripts) != 1 {
		t.Fatalf("expected 1 script, got %d", len(scripts))
	}

	s := scripts[0]
	if s.ID != 202 {
		t.Errorf("expected ID 202, got %d", s.ID)
	}
	if s.GetSrc() != "https://apps-scripts.tiendanube.com/app/1.js" {
		t.Errorf("expected version src, got %s", s.GetSrc())
	}
	if s.Location != "store" {
		t.Errorf("expected location store, got %s", s.Location)
	}
}

func TestScriptHelpers(t *testing.T) {
	// Priority: Src > CurrentVersion > DraftVersion
	s1 := Script{
		Src: "https://example.com/direct.js",
		CurrentVersion: &ScriptVersion{
			Src: "https://example.com/current.js",
		},
	}
	if s1.GetSrc() != "https://example.com/direct.js" {
		t.Errorf("GetSrc() = %q, want direct.js", s1.GetSrc())
	}

	s2 := Script{
		DraftVersion: &ScriptVersion{
			Src: "https://example.com/draft.js",
		},
	}
	if s2.GetSrc() != "https://example.com/draft.js" {
		t.Errorf("GetSrc() = %q, want draft.js", s2.GetSrc())
	}
}

func TestCreateScriptValidation(t *testing.T) {
	client := NewClient("123", "token")

	_, err := client.CreateScript(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "cannot be nil") {
		t.Errorf("expected nil request error, got %v", err)
	}

	_, err = client.CreateScript(context.Background(), &CreateScriptRequest{
		ScriptID: 0,
	})
	if err == nil || !strings.Contains(err.Error(), "script_id is required") {
		t.Errorf("expected script_id required error, got %v", err)
	}
}

func TestUpdateScriptValidation(t *testing.T) {
	client := NewClient("123", "token")

	_, err := client.UpdateScript(context.Background(), 123, nil)
	if err == nil || !strings.Contains(err.Error(), "cannot be nil") {
		t.Errorf("expected nil request error, got %v", err)
	}

	_, err = client.UpdateScript(context.Background(), 0, &UpdateScriptRequest{})
	if err == nil || !strings.Contains(err.Error(), "script id is required") {
		t.Errorf("expected script id required error, got %v", err)
	}
}

func TestDeleteScriptValidation(t *testing.T) {
	client := NewClient("123", "token")

	err := client.DeleteScript(context.Background(), 0)
	if err == nil || !strings.Contains(err.Error(), "script id is required") {
		t.Errorf("expected script id required error, got %v", err)
	}
}

func TestScriptsRealAPI(t *testing.T) {
	store, token := requireAPICredentials(t)
	client := NewClient(store, token)

	scripts, err := client.ListScripts(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListScripts failed: %v", err)
	}
	t.Logf("Retrieved %d scripts", len(scripts))
}
