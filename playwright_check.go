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

	CacheHash         string   `json:"cacheHash"`
	PlaywrightVersion *string  `json:"playwrightVersion"`
	TestCommand       *string  `json:"testCommand"`
	InstallCommand    *string  `json:"installCommand"`
	Browsers          []string `json:"browsers"`
	CodeBundlePath    string   `json:"codeBundlePath"`
	WorkingDir        *string  `json:"workingDir"`
	Engine            *string  `json:"engine"`
	EngineVersion     *string  `json:"engineVersion"`

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
	payload := playwrightCheckPayload{
		PlaywrightCheck: check,
		// Unfortunately `checkType` is required for this endpoint.
		Type: "PLAYWRIGHT",
	}

	if payload.Browsers == nil {
		payload.Browsers = []string{}
	}

	return payload
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

// GetPlaywrightCheck retrieves an existing Playwright check.
func (c *client) GetPlaywrightCheck(
	ctx context.Context,
	ID string,
) (*PlaywrightCheck, error) {
	var result PlaywrightCheck
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/checks/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
