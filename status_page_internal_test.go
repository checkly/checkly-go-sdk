package checkly

import (
	"reflect"
	"testing"
)

// TestStatusPagePayloadSendsUnsetServicesAsEmpty pins cards[].services: the
// API rejects null for it.
func TestStatusPagePayloadSendsUnsetServicesAsEmpty(t *testing.T) {
	page := StatusPage{
		Name: "page",
		URL:  "page",
		Cards: []StatusPageCard{
			{Name: "empty"},
			{Name: "set", Services: []StatusPageService{{ID: "svc-1"}}},
		},
	}
	original := []StatusPageCard{
		{Name: "empty"},
		{Name: "set", Services: []StatusPageService{{ID: "svc-1"}}},
	}

	cards, ok := marshalToMap(t, createStatusPagePayload(page))["cards"].([]any)
	if !ok || len(cards) != 2 {
		t.Fatalf("cards: want 2 cards, got %#v", cards)
	}
	if services, ok := cards[0].(map[string]any)["services"].([]any); !ok || len(services) != 0 {
		t.Errorf("cards[0].services: want [], got %#v", cards[0].(map[string]any)["services"])
	}
	services, ok := cards[1].(map[string]any)["services"].([]any)
	if !ok || len(services) != 1 || services[0].(map[string]any)["id"] != "svc-1" {
		t.Errorf("cards[1].services: want [{id: svc-1}], got %#v", cards[1].(map[string]any)["services"])
	}
	if !reflect.DeepEqual(page.Cards, original) {
		t.Errorf("caller's cards changed: got %#v", page.Cards)
	}
}
