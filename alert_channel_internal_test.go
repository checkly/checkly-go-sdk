package checkly

import "testing"

// TestAlertChannelPayloadOmitsUnsetSettings asserts the alert channel payload
// omits unset settings, which the API rejects as null, sends the set ones,
// and never sends the ID, which the API strips from the body.
func TestAlertChannelPayloadOmitsUnsetSettings(t *testing.T) {
	t.Parallel()
	unset := marshalToMap(t, createAlertChannelPayload(AlertChannel{
		ID:        7,
		Type:      AlertTypeEmail,
		Email:     &AlertChannelEmail{Address: "a@example.com"},
		CreatedAt: "2026-01-02T03:04:05Z",
		UpdatedAt: "2026-01-02T03:04:05Z",
	}))
	if len(unset) != 2 || unset["type"] != AlertTypeEmail || unset["config"] == nil {
		t.Errorf("want only type and config, got %v", unset)
	}

	f, threshold := false, 30
	set := marshalToMap(t, createAlertChannelPayload(AlertChannel{
		ID:                 7,
		Type:               AlertTypeEmail,
		SendRecovery:       &f,
		SendFailure:        &f,
		SendDegraded:       &f,
		SSLExpiry:          &f,
		SSLExpiryThreshold: &threshold,
	}))
	for _, key := range []string{"sendRecovery", "sendFailure", "sendDegraded", "sslExpiry"} {
		if v, ok := set[key]; !ok || v != false {
			t.Errorf("want %s false, got %#v", key, v)
		}
	}
	if set["sslExpiryThreshold"] != float64(30) {
		t.Errorf("want sslExpiryThreshold 30, got %#v", set["sslExpiryThreshold"])
	}
	if _, ok := set["id"]; ok {
		t.Errorf("want no id, got %v", set)
	}
}
