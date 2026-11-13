package checkly

import (
	"context"
	"net/http"
	"slices"
)

// AccountFeature is a feature flag of an account, in [Account].Features.
type AccountFeature string

// Account feature flags, the values of [Account].Features.
const (
	// AccountFeatureExtendedResponseTimeLimits marks accounts whose checks
	// may set response time limits above the standard runner timeouts: such
	// checks currently wait up to 45 seconds for a response, instead of 30
	// seconds for API checks and URL monitors and 5 seconds for TCP
	// monitors.
	AccountFeatureExtendedResponseTimeLimits AccountFeature = "EXTENDED_RESPONSE_TIME_LIMITS"
)

// AccountIntegrationType is the type of an account integration, in
// [AccountIntegration].Type.
type AccountIntegrationType string

// Account integration types, the values of [AccountIntegration].Type.
const (
	// AccountIntegrationSlack is the Checkly Slack app, which alert channels
	// of type [AlertChannelTypeSlackApp] need.
	AccountIntegrationSlack AccountIntegrationType = "SLACK"
)

// Account is a Checkly account, as returned by GetAccount.
type Account struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// Features lists the account's feature flags that clients can act on,
	// such as AccountFeatureExtendedResponseTimeLimits.
	Features []AccountFeature `json:"features"`

	// Integrations lists integrations of the account, such as
	// AccountIntegrationSlack, each with whether it is connected. A type
	// that is not listed is one the API does not report, which says nothing
	// about whether it is connected.
	Integrations []AccountIntegration `json:"integrations"`
}

// AccountIntegration is an integration of an [Account].
type AccountIntegration struct {
	Type      AccountIntegrationType `json:"type"`
	Connected bool                   `json:"connected"`
}

// HasFeature reports whether the account has the feature flag feature, such
// as AccountFeatureExtendedResponseTimeLimits.
func (a *Account) HasFeature(feature AccountFeature) bool {
	return slices.Contains(a.Features, feature)
}

// Integration returns the account's integration of type integrationType,
// such as AccountIntegrationSlack. It returns false if the API does not
// report that type.
func (a *Account) Integration(integrationType AccountIntegrationType) (AccountIntegration, bool) {
	i := slices.IndexFunc(a.Integrations, func(i AccountIntegration) bool {
		return i.Type == integrationType
	})
	if i < 0 {
		return AccountIntegration{}, false
	}
	return a.Integrations[i], true
}

func (c *client) GetAccount(
	ctx context.Context,
) (*Account, error) {
	result := Account{}
	err := c.callJSON(ctx, http.MethodGet, "/v1/accounts/me", nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
