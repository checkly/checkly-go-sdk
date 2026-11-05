package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// PlaywrightEngine is the JavaScript runtime a Playwright check suite runs
// in, in [PlaywrightCheck].Engine.
type PlaywrightEngine string

// Playwright engines, the values of [PlaywrightCheck].Engine.
const (
	PlaywrightEngineNode PlaywrightEngine = "node"
	PlaywrightEngineBun  PlaywrightEngine = "bun"
)

// PlaywrightCheck is a Playwright check, which runs a Playwright test suite
// from a code bundle. It is the input and result of
// [Client.CreatePlaywrightCheck] and [Client.UpdatePlaywrightCheck], and the
// result of [Client.GetPlaywrightCheck]. Unlike the other check types it has
// no ShouldFail, RuntimeID or RetryStrategy, which the API does not accept
// for Playwright checks.
type PlaywrightCheck struct {
	// ID is the check's UUID, assigned by the API.
	ID          string  `json:"id,omitempty"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	// Frequency is the number of minutes between runs: 1, 2, 5, 10, 15, 30,
	// 60, 120, 180, 360, 720 or 1440.
	Frequency int  `json:"frequency"`
	Activated bool `json:"activated"`
	// Muted stops notifications when the check fails or recovers.
	Muted bool `json:"muted"`
	// RunParallel runs the check in all its locations in parallel.
	RunParallel bool `json:"runParallel"`
	// Locations are the public locations to run from, as region codes such
	// as "us-east-1".
	Locations []string `json:"locations"`
	// PrivateLocations are the slug names of the account's private
	// locations to run from.
	PrivateLocations     *[]string             `json:"privateLocations"`
	EnvironmentVariables []EnvironmentVariable `json:"environmentVariables"`
	// Tags is left out of the request when empty.
	Tags []string `json:"tags,omitempty"`
	// AlertSettings is left out of the request when nil.
	AlertSettings *AlertSettings `json:"alertSettings,omitempty"`
	// UseGlobalAlertSettings applies the account's alert settings instead
	// of AlertSettings.
	UseGlobalAlertSettings bool `json:"useGlobalAlertSettings"`
	// GroupID is the ID of the check group the check belongs to; 0 means
	// none. GroupID is left out of the request when zero.
	GroupID int64 `json:"groupId,omitempty"`
	// GroupOrder is the check's position in its group, which sets the run
	// order when the group is triggered from the API or CI/CD.
	// GroupOrder is left out of the request when zero.
	GroupOrder int `json:"groupOrder,omitempty"`
	// AlertChannelSubscriptions are the alert channels that receive the
	// check's alerts. The SDK disables automatic assignment, so an empty
	// list leaves the check without any.
	// AlertChannelSubscriptions is left out of the request when empty.
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	// TriggerIncident makes the check open and resolve an incident on a
	// status page service, based on its alert configuration. Nil turns it
	// off.
	TriggerIncident *IncidentTrigger `json:"triggerIncident"`
	// Intent describes what the check protects, for automatic AI repair.
	// It is sent as null when nil, so an update with a nil Intent clears an
	// intent set elsewhere, for example in the web app.
	Intent *CheckIntent `json:"intent"`
	// AIAutoRepairEnabled overrides the account's automatic AI repair
	// setting for this check. Nil inherits the account setting; it is sent
	// as null, so an update with nil removes an override set elsewhere.
	AIAutoRepairEnabled *bool `json:"aiAutoRepairEnabled"`

	// CacheHash is a hash of the lock file and browsers, which the API
	// requires.
	CacheHash         string  `json:"cacheHash"`
	PlaywrightVersion *string `json:"playwrightVersion"`
	// TestCommand is the command that runs the tests. Nil applies the API
	// default, "npx playwright test".
	TestCommand *string `json:"testCommand"`
	// InstallCommand defaults to an empty command when nil.
	InstallCommand *string `json:"installCommand"`
	// Browsers are the browsers to run the tests in, such as "chromium",
	// without duplicates. Browsers is left out of the request when empty, and
	// the API then uses ["chromium"].
	Browsers []string `json:"browsers,omitempty"`
	// CodeBundlePath is the path of the check's code bundle, which the API
	// requires.
	CodeBundlePath string  `json:"codeBundlePath"`
	WorkingDir     *string `json:"workingDir"`
	// Engine is the JavaScript runtime that runs the tests, or nil.
	Engine *PlaywrightEngine `json:"engine"`
	// EngineVersion is the version of Engine, such as "24" for
	// [PlaywrightEngineNode]. It must be nil when Engine is nil.
	EngineVersion *string `json:"engineVersion"`

	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type playwrightCheckPayload struct {
	PlaywrightCheck
	Type      CheckType  `json:"checkType"`
	ID        *string    `json:"id,omitempty"`         // Skip, can't be changed.
	CreatedAt *time.Time `json:"created_at,omitempty"` // Skip, can't be changed.
	UpdatedAt *time.Time `json:"updated_at,omitempty"` // Skip, can't be changed.
}

func createPlaywrightCheckPayload(check PlaywrightCheck) playwrightCheckPayload {
	return playwrightCheckPayload{
		PlaywrightCheck: check,
		// Unfortunately `checkType` is required for this endpoint.
		Type: CheckTypePlaywright,
	}
}

func (c *client) CreatePlaywrightCheck(
	ctx context.Context,
	check PlaywrightCheck,
) (*PlaywrightCheck, error) {
	var result PlaywrightCheck
	err := c.callJSON(ctx, http.MethodPost, withAutoAssignAlertsFlag("/v1/checks/playwright"), createPlaywrightCheckPayload(check), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetPlaywrightCheck(
	ctx context.Context,
	ID string,
) (*PlaywrightCheck, error) {
	var result PlaywrightCheck
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s", ID), CheckTypePlaywright, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdatePlaywrightCheck(
	ctx context.Context,
	ID string,
	check PlaywrightCheck,
) (*PlaywrightCheck, error) {
	var result PlaywrightCheck
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/playwright/%s", ID)), createPlaywrightCheckPayload(check), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeletePlaywrightCheck(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}
