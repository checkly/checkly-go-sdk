package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// MultiStepCheck is a multi-step check, which runs a script. It is the
// input and result of [Client.CreateMultiStepCheck] and
// [Client.UpdateMultiStepCheck], and the result of
// [Client.GetMultiStepCheck].
type MultiStepCheck struct {
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
	// order when the group is triggered from the API or CI/CD.
	GroupOrder int `json:"groupOrder"`
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
	// Intent describes what the check protects, for automatic AI repair.
	// It is sent as null when nil, so an update with a nil Intent clears an
	// intent set elsewhere, for example in the web app.
	Intent *CheckIntent `json:"intent"`
	// AIAutoRepairEnabled overrides the account's automatic AI repair
	// setting for this check. Nil inherits the account setting; it is sent
	// as null, so an update with nil removes an override set elsewhere.
	AIAutoRepairEnabled *bool `json:"aiAutoRepairEnabled"`
}

type multiStepCheckPayload struct {
	MultiStepCheck
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

func createMultiStepCheckPayload(check MultiStepCheck) multiStepCheckPayload {
	payload := multiStepCheckPayload{
		MultiStepCheck: check,
		// The endpoint defaults `checkType` correctly, but be explicit anyway.
		Type: CheckTypeMultiStep,
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

func (c *client) CreateMultiStepCheck(
	ctx context.Context,
	check MultiStepCheck,
) (*MultiStepCheck, error) {
	var result MultiStepCheck
	err := c.callJSON(ctx, http.MethodPost, withAutoAssignAlertsFlag("/v1/checks/multistep"), createMultiStepCheckPayload(check), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetMultiStepCheck(
	ctx context.Context,
	ID string,
) (*MultiStepCheck, error) {
	var result MultiStepCheck
	// Without includeDependencies the response omits the script dependencies.
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s?includeDependencies=true", ID), CheckTypeMultiStep, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateMultiStepCheck(
	ctx context.Context,
	ID string,
	check MultiStepCheck,
) (*MultiStepCheck, error) {
	var result MultiStepCheck
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/multistep/%s", ID)), createMultiStepCheckPayload(check), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteMultiStepCheck(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}
