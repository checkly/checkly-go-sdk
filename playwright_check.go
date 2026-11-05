package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// PlaywrightCheck represents a Playwright check.
type PlaywrightCheck struct {
	ID                        string                     `json:"id,omitempty"`
	Name                      string                     `json:"name"`
	Description               *string                    `json:"description"`
	Frequency                 int                        `json:"frequency"`
	Activated                 bool                       `json:"activated"`
	Muted                     bool                       `json:"muted"`
	RunParallel               bool                       `json:"runParallel"`
	Locations                 []string                   `json:"locations"`
	PrivateLocations          *[]string                  `json:"privateLocations"`
	EnvironmentVariables      []EnvironmentVariable      `json:"environmentVariables"`
	Tags                      []string                   `json:"tags,omitempty"`
	AlertSettings             *AlertSettings             `json:"alertSettings,omitempty"`
	UseGlobalAlertSettings    bool                       `json:"useGlobalAlertSettings"`
	GroupID                   int64                      `json:"groupId,omitempty"`
	GroupOrder                int                        `json:"groupOrder,omitempty"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	TriggerIncident           *IncidentTrigger           `json:"triggerIncident"`
	// Intent describes what the check protects, for automatic AI repair.
	// It is sent as null when nil, so an update with a nil Intent clears an
	// intent set elsewhere, for example in the web app.
	Intent *CheckIntent `json:"intent"`
	// AIAutoRepairEnabled overrides the account's automatic AI repair
	// setting for this check. Nil inherits the account setting; it is sent
	// as null, so an update with nil removes an override set elsewhere.
	AIAutoRepairEnabled *bool `json:"aiAutoRepairEnabled"`

	CacheHash         string  `json:"cacheHash"`
	PlaywrightVersion *string `json:"playwrightVersion"`
	TestCommand       *string `json:"testCommand"`
	InstallCommand    *string `json:"installCommand"`
	// Browsers lists the browsers the tests run in. The API requires at least
	// one, so a nil or empty list is left out of create and update requests
	// and the API's default applies (on update too, because the update
	// endpoint applies create defaults to omitted keys).
	Browsers       []string `json:"browsers,omitempty"`
	CodeBundlePath string   `json:"codeBundlePath"`
	WorkingDir     *string  `json:"workingDir"`
	Engine         *string  `json:"engine"`
	EngineVersion  *string  `json:"engineVersion"`

	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type playwrightCheckPayload struct {
	PlaywrightCheck
	Type      string     `json:"checkType"`
	ID        *string    `json:"id,omitempty"`         // Skip, can't be changed.
	CreatedAt *time.Time `json:"created_at,omitempty"` // Skip, can't be changed.
	UpdatedAt *time.Time `json:"updated_at,omitempty"` // Skip, can't be changed.
}

func createPlaywrightCheckPayload(check PlaywrightCheck) playwrightCheckPayload {
	return playwrightCheckPayload{
		PlaywrightCheck: check,
		// Unfortunately `checkType` is required for this endpoint.
		Type: "PLAYWRIGHT",
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

// GetPlaywrightCheck takes the ID of an existing Playwright check, and returns
// the check parameters, or an error. It fails with an error matching
// ErrCheckTypeMismatch when the ID belongs to a check of another type.
func (c *client) GetPlaywrightCheck(
	ctx context.Context,
	ID string,
) (*PlaywrightCheck, error) {
	var result PlaywrightCheck
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s", ID), "PLAYWRIGHT", &result); err != nil {
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
