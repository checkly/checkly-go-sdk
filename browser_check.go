package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ScriptDependency represents a dependency file attached to a browser or
// multi-step check script.
type ScriptDependency struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// BrowserCheck represents a browser check.
type BrowserCheck struct {
	ID                        string                     `json:"id,omitempty"`
	Name                      string                     `json:"name"`
	Description               *string                    `json:"description"`
	Frequency                 int                        `json:"frequency"`
	Activated                 bool                       `json:"activated"`
	Muted                     bool                       `json:"muted"`
	ShouldFail                bool                       `json:"shouldFail"`
	RunParallel               bool                       `json:"runParallel"`
	Locations                 []string                   `json:"locations"`
	Script                    string                     `json:"script"`
	ScriptPath                *string                    `json:"scriptPath"`
	Dependencies              []ScriptDependency         `json:"dependencies"`
	EnvironmentVariables      []EnvironmentVariable      `json:"environmentVariables"`
	SSLCheckDomain            string                     `json:"sslCheckDomain"`
	Tags                      []string                   `json:"tags"`
	AlertSettings             *AlertSettings             `json:"alertSettings,omitempty"`
	UseGlobalAlertSettings    bool                       `json:"useGlobalAlertSettings"`
	GroupID                   int64                      `json:"groupId,omitempty"`
	GroupOrder                int                        `json:"groupOrder"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions"`
	PrivateLocations          *[]string                  `json:"privateLocations"`
	RuntimeID                 *string                    `json:"runtimeId"`
	RetryStrategy             *RetryStrategy             `json:"retryStrategy"`
	TriggerIncident           *IncidentTrigger           `json:"triggerIncident"`
	CreatedAt                 time.Time                  `json:"created_at,omitempty"`
	UpdatedAt                 time.Time                  `json:"updated_at,omitempty"`

	// Intent describes what the check protects, for automatic AI repair.
	// It is sent as null when nil, so an update with a nil Intent clears an
	// intent set elsewhere, for example in the web app.
	Intent *CheckIntent `json:"intent"`
	// AIAutoRepairEnabled overrides the account's automatic AI repair
	// setting for this check. Nil inherits the account setting; it is sent
	// as null, so an update with nil removes an override set elsewhere.
	AIAutoRepairEnabled *bool `json:"aiAutoRepairEnabled"`
}

type browserCheckPayload struct {
	BrowserCheck
	Type string `json:"checkType"`
	// The update endpoint is PATCH-style: keys omitted from the payload keep
	// their server-side value. The fields below shadow embedded fields whose
	// Go type cannot express the value that clears them.
	DoubleCheck bool   `json:"doubleCheck"`
	GroupID     *int64 `json:"groupId"`
	// Script is null rather than empty when unset: the API rejects an empty
	// script, and a check that runs ScriptPath has no inline script.
	Script *string `json:"script"`
}

func createBrowserCheckPayload(check BrowserCheck) browserCheckPayload {
	payload := browserCheckPayload{
		BrowserCheck: check,
		// The endpoint defaults `checkType` correctly, but be explicit anyway.
		Type: "BROWSER",
		// Unfortunately, this will default to true if not set. Always sending
		// false also clears the deprecated doubleCheck flag on updates of
		// checks that still have it set; retryStrategy is its replacement.
		DoubleCheck: false,
		GroupID:     nullIfZero(check.GroupID),
		Script:      nullIfZero(check.Script),
	}

	// A nil value for a list will cause the backend to not update the value.
	// We must send empty lists instead.
	payload.Locations = emptyIfNil(check.Locations)
	if check.PrivateLocations == nil {
		payload.PrivateLocations = &[]string{}
	}
	payload.Tags = emptyIfNil(check.Tags)
	payload.AlertChannelSubscriptions = emptyIfNil(check.AlertChannelSubscriptions)
	payload.Dependencies = emptyIfNil(check.Dependencies)

	return payload
}

func (c *client) CreateBrowserCheck(
	ctx context.Context,
	check BrowserCheck,
) (*BrowserCheck, error) {
	var result BrowserCheck
	err := c.callJSON(ctx, http.MethodPost, withAutoAssignAlertsFlag("/v1/checks/browser"), createBrowserCheckPayload(check), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBrowserCheck takes the ID of an existing browser check, and returns the
// check parameters, or an error. It fails with an error matching
// ErrCheckTypeMismatch when the ID belongs to a check of another type. The
// includeDependencies query parameter is required for the response to contain
// the check's script dependencies.
func (c *client) GetBrowserCheck(
	ctx context.Context,
	ID string,
) (*BrowserCheck, error) {
	var result BrowserCheck
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s?includeDependencies=true", ID), "BROWSER", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateBrowserCheck(
	ctx context.Context,
	ID string,
	check BrowserCheck,
) (*BrowserCheck, error) {
	var result BrowserCheck
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/browser/%s", ID)), createBrowserCheckPayload(check), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteBrowserCheck(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}
