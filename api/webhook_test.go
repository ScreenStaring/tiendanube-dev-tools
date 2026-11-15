package api

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestWebhookCRUDRealAPI(t *testing.T) {
	store, token := requireAPICredentials(t)

	client := NewClient(store, token)

	testURL := fmt.Sprintf("https://example.com/tndt-test-%d", time.Now().UnixNano())
	testEvent := "app/uninstalled"

	// 1. Create Webhook
	created, err := client.CreateWebhook(context.Background(), &CreateWebhookRequest{
		Event: testEvent,
		URL:   testURL,
	})
	if err != nil {
		t.Fatalf("CreateWebhook failed: %v", err)
	}
	t.Logf("Created webhook ID: %d", created.ID)

	defer func() {
		_ = client.DeleteWebhook(context.Background(), created.ID)
	}()

	// 2. Get Webhook
	got, err := client.GetWebhook(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetWebhook failed: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("got ID %d, want %d", got.ID, created.ID)
	}
	if got.Event != testEvent {
		t.Errorf("got Event %q, want %q", got.Event, testEvent)
	}
	if got.URL != testURL {
		t.Errorf("got URL %q, want %q", got.URL, testURL)
	}

	// 3. Update Webhook
	updatedURL := fmt.Sprintf("https://example.com/tndt-test-updated-%d", time.Now().UnixNano())
	updated, err := client.UpdateWebhook(context.Background(), created.ID, &UpdateWebhookRequest{
		URL: updatedURL,
	})
	if err != nil {
		t.Fatalf("UpdateWebhook failed: %v", err)
	}
	if updated.URL != updatedURL {
		t.Errorf("updated URL %q, want %q", updated.URL, updatedURL)
	}

	// 4. List Webhooks
	list, err := client.ListWebhooks(context.Background(), &ListWebhooksParams{
		Event: testEvent,
	})
	if err != nil {
		t.Fatalf("ListWebhooks failed: %v", err)
	}
	var found bool
	for _, w := range list {
		if w.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("webhook %d not found in list", created.ID)
	}

	// 5. Delete Webhook
	err = client.DeleteWebhook(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("DeleteWebhook failed: %v", err)
	}

	// Verify deletion returns an error
	_, err = client.GetWebhook(context.Background(), created.ID)
	if err == nil {
		t.Errorf("expected error getting deleted webhook %d, got nil", created.ID)
	}
}
