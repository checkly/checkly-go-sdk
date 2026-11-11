package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ScriptDependency is a file attached to the script of a [BrowserCheck] or
// [MultiStepCheck]. The API silently drops a dependency whose Path is empty
// or longer than 1000 characters, or whose Content is empty.
type ScriptDependency struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// BrowserCheck is a browser check, which runs a script in a browser. It is
// the input and result of [Client.CreateBrowserCheck] and
// [Client.UpdateBrowserCheck], and the result of [Client.GetBrowserCheck].
type BrowserCheck struct {
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
	// ShouldFail inverts the result: a run that would fail passes, and the
	// other way around.
	ShouldFail bool `json:"shouldFail"`
	// RunParallel runs the check in all its locations in parallel.
	RunParallel bool `json:"runParallel"`
	// Locations are the public locations to run from, as region codes such
	// as "us-east-1". Nil is sent as an empty list.
	Locations []string `json:"locations"`
	// Script is the check's code. An empty Script is sent as null.
	Script string `json:"script"`
	// ScriptPath is the path of the script in the runtime.
	ScriptPath *string `json:"scriptPath"`
	// Dependencies are files attached to the script. Nil is sent as an
	// empty list.
	Dependencies         []ScriptDependency    `json:"dependencies"`
	EnvironmentVariables []EnvironmentVariable `json:"environmentVariables"`
	SSLCheckDomain       string                `json:"sslCheckDomain"`
	Tags                 []string              `json:"tags"`
	// AlertSettings is left out of the request when nil.
	AlertSettings *AlertSettings `json:"alertSettings,omitempty"`
	// UseGlobalAlertSettings applies the account's alert settings instead
	// of AlertSettings.
	UseGlobalAlertSettings bool `json:"useGlobalAlertSettings"`
	// GroupID is the ID of the check group the check belongs to; 0 means
	// none.
	GroupID int64 `json:"groupId,omitempty"`
	// GroupOrder is the check's position in its group, which sets the run
	// order when the group is triggered from the API or CI/CD. 0 is the
	// first position. Nil means no position, and is sent as null so that
	// an update clears it.
	GroupOrder *int `json:"groupOrder"`
	// AlertChannelSubscriptions are the alert channels that receive the
	// check's alerts. The SDK disables automatic assignment, so an empty
	// list leaves the check without any.
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions"`
	// PrivateLocations are the slug names of the account's private
	// locations to run from. Nil is sent as an empty list.
	PrivateLocations *[]string `json:"privateLocations"`
	// RuntimeID is the runtime version the check runs with, such as
	// "2025.04".
	RuntimeID *string `json:"runtimeId"`
	// RetryStrategy sets how failed runs are retried; nil disables retries.
	// In a group whose RetryStrategy is not [RetryStrategyTypeFallback], the
	// group's strategy applies instead.
	RetryStrategy *RetryStrategy `json:"retryStrategy"`
	// TriggerIncident makes the check open and resolve an incident on a
	// status page service, based on its alert configuration. Nil turns it
	// off.
	TriggerIncident *IncidentTrigger `json:"triggerIncident"`
	CreatedAt       time.Time        `json:"created_at,omitempty"`
	UpdatedAt       time.Time        `json:"updated_at,omitempty"`
	// Intent describes what the check protects, for automatic AI repair; nil
	// removes it.
	Intent *CheckIntent `json:"intent"`
	// AIAutoRepairEnabled overrides the account's automatic AI repair
	// setting for this check; nil uses the account setting.
	AIAutoRepairEnabled *bool `json:"aiAutoRepairEnabled"`
}

func (*BrowserCheck) isCheck() {}

type browserCheckPayload struct {
	BrowserCheck
	Type CheckType `json:"checkType"`
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
		Type: CheckTypeBrowser,
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

func (c *client) GetBrowserCheck(
	ctx context.Context,
	id string,
) (*BrowserCheck, error) {
	var result BrowserCheck
	// Without includeDependencies the response omits the script dependencies.
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s?includeDependencies=true", id), CheckTypeBrowser, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateBrowserCheck(
	ctx context.Context,
	id string,
	check BrowserCheck,
) (*BrowserCheck, error) {
	var result BrowserCheck
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/browser/%s", id)), createBrowserCheckPayload(check), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteBrowserCheck(
	ctx context.Context,
	id string,
) error {
	return c.DeleteCheck(ctx, id)
}
