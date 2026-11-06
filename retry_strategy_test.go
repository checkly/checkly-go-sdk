package checkly_test

import (
	"encoding/json"
	"reflect"
	"testing"

	checkly "github.com/checkly/checkly-go-sdk"
)

// TestRetryStrategyRoundTrip asserts the JSON each retry strategy type is
// sent as, and that it decodes back to the same value.
func TestRetryStrategyRoundTrip(t *testing.T) {
	networkError := []checkly.RetryOnlyOn{checkly.RetryOnlyOnNetworkError}
	tests := []struct {
		name     string
		strategy checkly.RetryStrategy
		wantJSON string
	}{
		{
			name:     "FIXED sends every field",
			strategy: checkly.RetryStrategy{Type: checkly.RetryStrategyTypeFixed, MaxRetries: 2, MaxDurationSeconds: 600, BaseBackoffSeconds: 60, SameRegion: true},
			wantJSON: `{"type":"FIXED","baseBackoffSeconds":60,"maxRetries":2,"maxDurationSeconds":600,"sameRegion":true}`,
		},
		{
			name:     "FIXED with OnlyOn",
			strategy: checkly.RetryStrategy{Type: checkly.RetryStrategyTypeFixed, MaxRetries: 2, MaxDurationSeconds: 600, BaseBackoffSeconds: 60, SameRegion: true, OnlyOn: networkError},
			wantJSON: `{"type":"FIXED","baseBackoffSeconds":60,"maxRetries":2,"maxDurationSeconds":600,"sameRegion":true,"onlyOn":["NETWORK_ERROR"]}`,
		},
		{
			name:     "LINEAR sends every field",
			strategy: checkly.RetryStrategy{Type: checkly.RetryStrategyTypeLinear, MaxRetries: 3, MaxDurationSeconds: 300, BaseBackoffSeconds: 30},
			wantJSON: `{"type":"LINEAR","baseBackoffSeconds":30,"maxRetries":3,"maxDurationSeconds":300,"sameRegion":false}`,
		},
		{
			name:     "SINGLE_RETRY omits MaxRetries and MaxDurationSeconds",
			strategy: checkly.RetryStrategy{Type: checkly.RetryStrategyTypeSingleRetry, BaseBackoffSeconds: 10, SameRegion: true, OnlyOn: networkError},
			wantJSON: `{"type":"SINGLE_RETRY","baseBackoffSeconds":10,"sameRegion":true,"onlyOn":["NETWORK_ERROR"]}`,
		},
		{
			name:     "NO_RETRIES marshals as null",
			strategy: checkly.RetryStrategy{Type: checkly.RetryStrategyTypeNoRetries},
			wantJSON: `null`,
		},
		{
			name:     "FALLBACK marshals as a string",
			strategy: checkly.RetryStrategy{Type: checkly.RetryStrategyTypeFallback},
			wantJSON: `"FALLBACK"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.strategy)
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}
			if string(data) != tt.wantJSON {
				t.Errorf("got JSON %s, want %s", data, tt.wantJSON)
			}
			var got checkly.RetryStrategy
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.strategy) {
				t.Errorf("got %+v, want %+v (JSON was: %s)", got, tt.strategy, data)
			}
		})
	}
}

func TestRetryStrategyUnmarshalClearsFields(t *testing.T) {
	// Unmarshal into a struct that already has values to verify fields are cleared.
	s := checkly.RetryStrategy{
		Type:               checkly.RetryStrategyTypeFixed,
		MaxRetries:         5,
		MaxDurationSeconds: 600,
		BaseBackoffSeconds: 60,
		SameRegion:         true,
	}
	if err := json.Unmarshal([]byte("null"), &s); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	want := checkly.RetryStrategy{Type: checkly.RetryStrategyTypeNoRetries}
	if !reflect.DeepEqual(s, want) {
		t.Errorf("got %+v, want %+v", s, want)
	}
}

func TestRetryStrategyUnmarshalRejectsUnknownString(t *testing.T) {
	var s checkly.RetryStrategy
	err := json.Unmarshal([]byte(`"UNKNOWN"`), &s)
	if err == nil {
		t.Fatal("expected error for unknown string value, got nil")
	}
}
