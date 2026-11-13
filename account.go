package checkly

import (
	"context"
	"net/http"
	"slices"
)

// AccountFeatureExtendedResponseTimeLimits is the [Account].Features flag of
// accounts whose checks may set response time limits above the standard
// runner timeouts: such checks currently wait up to 45 seconds for a
// response, instead of 30 seconds for API checks and URL monitors and 5
// seconds for TCP monitors.
const AccountFeatureExtendedResponseTimeLimits = "EXTENDED_RESPONSE_TIME_LIMITS"

// Account is a Checkly account, as returned by GetAccount.
type Account struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// Features lists the account's feature flags that clients can act on,
	// such as AccountFeatureExtendedResponseTimeLimits.
	Features []string `json:"features"`
}

// HasFeature reports whether the account has the feature flag feature, such
// as AccountFeatureExtendedResponseTimeLimits.
func (a *Account) HasFeature(feature string) bool {
	return slices.Contains(a.Features, feature)
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
