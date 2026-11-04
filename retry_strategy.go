package checkly

import (
	"encoding/json"
	"fmt"
)

type RetryStrategy struct {
	Type               string   `json:"type"`
	BaseBackoffSeconds int      `json:"baseBackoffSeconds"`
	MaxRetries         int      `json:"maxRetries"`
	MaxDurationSeconds int      `json:"maxDurationSeconds"`
	SameRegion         bool     `json:"sameRegion"`
	OnlyOn             []string `json:"onlyOn,omitempty"`
}

func (s RetryStrategy) MarshalJSON() ([]byte, error) {
	type flexibleRetryStrategy struct {
		Type               string   `json:"type"`
		BaseBackoffSeconds *int     `json:"baseBackoffSeconds,omitempty"`
		MaxRetries         *int     `json:"maxRetries,omitempty"`
		MaxDurationSeconds *int     `json:"maxDurationSeconds,omitempty"`
		SameRegion         *bool    `json:"sameRegion,omitempty"`
		OnlyOn             []string `json:"onlyOn,omitempty"`
	}

	switch s.Type {
	case "SINGLE_RETRY":
		// Single retries can't have MaxRetries set to 0 or backend validation
		// will fail. We have to make sure to not send the extra properties
		// at all.
		return json.Marshal(flexibleRetryStrategy{
			Type:               s.Type,
			BaseBackoffSeconds: &s.BaseBackoffSeconds,
			SameRegion:         &s.SameRegion,
			OnlyOn:             s.OnlyOn,
		})
	case "NO_RETRIES":
		return json.Marshal(nil)
	case "FALLBACK":
		return json.Marshal("FALLBACK")
	default:
		return json.Marshal(flexibleRetryStrategy{
			Type:               s.Type,
			BaseBackoffSeconds: &s.BaseBackoffSeconds,
			MaxRetries:         &s.MaxRetries,
			MaxDurationSeconds: &s.MaxDurationSeconds,
			SameRegion:         &s.SameRegion,
			OnlyOn:             s.OnlyOn,
		})
	}
}

func (s *RetryStrategy) UnmarshalJSON(data []byte) error {
	// Clear all fields so previous values don't leak through.
	*s = RetryStrategy{}

	// null -> NO_RETRIES (reverse of MarshalJSON which emits null for NO_RETRIES)
	if string(data) == "null" {
		s.Type = "NO_RETRIES"
		return nil
	}

	// Plain string -> FALLBACK
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if str != "FALLBACK" {
			return fmt.Errorf("unsupported value %q for retry strategy", str)
		}
		s.Type = str
		return nil
	}

	// Object -> standard unmarshal. Use an alias type to avoid infinite
	// recursion: the alias has the same fields but no UnmarshalJSON method,
	// so json.Unmarshal uses the default struct decoder.
	type retryStrategyAlias RetryStrategy
	var alias retryStrategyAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*s = RetryStrategy(alias)
	return nil
}
