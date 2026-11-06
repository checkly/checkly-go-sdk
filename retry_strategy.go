package checkly

import (
	"encoding/json"
	"fmt"
)

// RetryStrategyType is how failed runs are retried, in [RetryStrategy].Type.
type RetryStrategyType string

// Retry strategy types, the values of [RetryStrategy].Type.
const (
	RetryStrategyTypeFixed       RetryStrategyType = "FIXED"
	RetryStrategyTypeLinear      RetryStrategyType = "LINEAR"
	RetryStrategyTypeExponential RetryStrategyType = "EXPONENTIAL"
	// RetryStrategyTypeSingleRetry retries once; MaxRetries and
	// MaxDurationSeconds are not sent.
	RetryStrategyTypeSingleRetry RetryStrategyType = "SINGLE_RETRY"
	// RetryStrategyTypeNoRetries disables retries. It is not an API value:
	// the SDK sends the retry strategy as null.
	RetryStrategyTypeNoRetries RetryStrategyType = "NO_RETRIES"
	// RetryStrategyTypeFallback lets each check in a group use its own retry
	// strategy, and only groups accept it. It is not an API type value: the
	// SDK sends the retry strategy as the string "FALLBACK".
	RetryStrategyTypeFallback RetryStrategyType = "FALLBACK"
)

// RetryOnlyOn is a failure cause that retries can be limited to, in
// [RetryStrategy].OnlyOn.
type RetryOnlyOn string

// Retry failure causes, the values of [RetryStrategy].OnlyOn.
const (
	// RetryOnlyOnNetworkError is a run that got no response.
	RetryOnlyOnNetworkError RetryOnlyOn = "NETWORK_ERROR"
)

// RetryStrategy sets how the failed runs of a check, monitor or group are
// retried.
type RetryStrategy struct {
	// Type is how failed runs are retried. MarshalJSON encodes
	// [RetryStrategyTypeNoRetries] and [RetryStrategyTypeFallback] specially.
	Type RetryStrategyType `json:"type"`
	// BaseBackoffSeconds is the number of seconds to wait before the first
	// retry.
	BaseBackoffSeconds int `json:"baseBackoffSeconds"`
	// MaxRetries is the maximum number of retries, 1 to 10.
	MaxRetries int `json:"maxRetries"`
	// MaxDurationSeconds is how long retrying may continue, 0 to 600
	// seconds.
	MaxDurationSeconds int `json:"maxDurationSeconds"`
	// SameRegion runs retries in the same region as the failed run.
	SameRegion bool `json:"sameRegion"`
	// OnlyOn limits retries to the listed failure causes.
	// OnlyOn is left out of the request when empty.
	OnlyOn []RetryOnlyOn `json:"onlyOn,omitempty"`
}

// MarshalJSON encodes Type [RetryStrategyTypeNoRetries] as null and
// [RetryStrategyTypeFallback] as the string "FALLBACK". For
// [RetryStrategyTypeSingleRetry] it leaves out MaxRetries and
// MaxDurationSeconds, which do not apply to that type.
func (s RetryStrategy) MarshalJSON() ([]byte, error) {
	type flexibleRetryStrategy struct {
		Type               RetryStrategyType `json:"type"`
		BaseBackoffSeconds *int              `json:"baseBackoffSeconds,omitempty"`
		MaxRetries         *int              `json:"maxRetries,omitempty"`
		MaxDurationSeconds *int              `json:"maxDurationSeconds,omitempty"`
		SameRegion         *bool             `json:"sameRegion,omitempty"`
		OnlyOn             []RetryOnlyOn     `json:"onlyOn,omitempty"`
	}

	switch s.Type {
	case RetryStrategyTypeSingleRetry:
		// Single retries can't have MaxRetries set to 0 or backend validation
		// will fail. We have to make sure to not send the extra properties
		// at all.
		return json.Marshal(flexibleRetryStrategy{
			Type:               s.Type,
			BaseBackoffSeconds: &s.BaseBackoffSeconds,
			SameRegion:         &s.SameRegion,
			OnlyOn:             s.OnlyOn,
		})
	case RetryStrategyTypeNoRetries:
		return json.Marshal(nil)
	case RetryStrategyTypeFallback:
		return json.Marshal(RetryStrategyTypeFallback)
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

// UnmarshalJSON is the reverse of MarshalJSON: it decodes null as Type
// [RetryStrategyTypeNoRetries] and the string "FALLBACK" as Type
// [RetryStrategyTypeFallback].
func (s *RetryStrategy) UnmarshalJSON(data []byte) error {
	// Clear all fields so previous values don't leak through.
	*s = RetryStrategy{}

	// null -> NO_RETRIES (reverse of MarshalJSON which emits null for NO_RETRIES)
	if string(data) == "null" {
		s.Type = RetryStrategyTypeNoRetries
		return nil
	}

	// Plain string -> FALLBACK
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if str != string(RetryStrategyTypeFallback) {
			return fmt.Errorf("unsupported value %q for retry strategy", str)
		}
		s.Type = RetryStrategyTypeFallback
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
