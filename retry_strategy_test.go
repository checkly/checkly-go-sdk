package checkly_test

import (
	"encoding/json"
	"reflect"
	"testing"

	checkly "github.com/checkly/checkly-go-sdk"
)

func TestRetryStrategyRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input checkly.RetryStrategy
		want  checkly.RetryStrategy
	}{
		{
			name:  "FIXED strategy round-trips",
			input: checkly.RetryStrategy{Type: "FIXED", MaxRetries: 2, MaxDurationSeconds: 600, BaseBackoffSeconds: 60, SameRegion: true},
			want:  checkly.RetryStrategy{Type: "FIXED", MaxRetries: 2, MaxDurationSeconds: 600, BaseBackoffSeconds: 60, SameRegion: true},
		},
		{
			name:  "LINEAR strategy round-trips",
			input: checkly.RetryStrategy{Type: "LINEAR", MaxRetries: 3, MaxDurationSeconds: 300, BaseBackoffSeconds: 30, SameRegion: false},
			want:  checkly.RetryStrategy{Type: "LINEAR", MaxRetries: 3, MaxDurationSeconds: 300, BaseBackoffSeconds: 30, SameRegion: false},
		},
		{
			name:  "SINGLE_RETRY omits MaxRetries and MaxDurationSeconds",
			input: checkly.RetryStrategy{Type: "SINGLE_RETRY", BaseBackoffSeconds: 10, SameRegion: true},
			want:  checkly.RetryStrategy{Type: "SINGLE_RETRY", BaseBackoffSeconds: 10, SameRegion: true},
		},
		{
			name:  "NO_RETRIES marshals as null and back",
			input: checkly.RetryStrategy{Type: "NO_RETRIES"},
			want:  checkly.RetryStrategy{Type: "NO_RETRIES"},
		},
		{
			name:  "FALLBACK marshals as string and back",
			input: checkly.RetryStrategy{Type: "FALLBACK"},
			want:  checkly.RetryStrategy{Type: "FALLBACK"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}
			var got checkly.RetryStrategy
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v (JSON was: %s)", got, tt.want, data)
			}
		})
	}
}

func TestRetryStrategyUnmarshalClearsFields(t *testing.T) {
	// Unmarshal into a struct that already has values to verify fields are cleared.
	s := checkly.RetryStrategy{
		Type:               "FIXED",
		MaxRetries:         5,
		MaxDurationSeconds: 600,
		BaseBackoffSeconds: 60,
		SameRegion:         true,
	}
	if err := json.Unmarshal([]byte("null"), &s); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	want := checkly.RetryStrategy{Type: "NO_RETRIES"}
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

func TestRetryStrategyMarshalNoRetries(t *testing.T) {
	s := checkly.RetryStrategy{Type: "NO_RETRIES"}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	if string(data) != "null" {
		t.Errorf("got %s, want null", data)
	}
}

func TestRetryStrategyMarshalFallback(t *testing.T) {
	s := checkly.RetryStrategy{Type: "FALLBACK"}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	if string(data) != `"FALLBACK"` {
		t.Errorf("got %s, want %q", data, "FALLBACK")
	}
}
