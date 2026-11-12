package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// GroupV2 is a check group: settings shared by the checks in it, as used by
// [Client.CreateGroupV2], [Client.GetGroupV2] and [Client.UpdateGroupV2].
type GroupV2 struct {
	// ID is the group's ID, assigned by the API.
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name"`
	// Activated, when false, deactivates every check in the group.
	Activated bool `json:"activated"`
	// Muted mutes every check in the group.
	Muted bool `json:"muted"`
	// RunParallel runs the group's checks in all their locations in
	// parallel; nil leaves each check its own setting.
	RunParallel *bool `json:"runParallel"`
	// Tags are added to the tags of every check in the group.
	Tags []string `json:"tags"`
	// Locations are the public locations to run from, as region codes such
	// as "us-east-1". When the group has any
	// public or private locations, they replace those of its checks.
	Locations []string `json:"locations"`
	// Concurrency is how many of the group's checks run at once when the
	// group is triggered from the API or CI/CD; at least 1.
	Concurrency      int              `json:"concurrency"`
	APICheckDefaults APICheckDefaults `json:"apiCheckDefaults"`
	// EnvironmentVariables are merged into those of every check in the
	// group.
	EnvironmentVariables []EnvironmentVariable `json:"environmentVariables"`
	// UseGlobalAlertSettings applies the account's alert settings instead
	// of AlertSettings, for every check in the group; nil leaves each check
	// its own setting.
	UseGlobalAlertSettings *bool `json:"useGlobalAlertSettings"`
	// AlertSettings replaces the alert settings of the checks in the group;
	// nil leaves each check its own. A check that uses the account's alert
	// settings keeps them unless UseGlobalAlertSettings is false.
	AlertSettings *AlertSettings `json:"alertSettings"`
	// SetupSnippetID and TearDownSnippetID are the IDs of the [Snippet] run
	// in the setup and teardown phase of the group's API checks; nil means
	// none.
	SetupSnippetID    *int64 `json:"setupSnippetId"`
	TearDownSnippetID *int64 `json:"tearDownSnippetId"`
	// LocalSetupScript and LocalTearDownScript are Node.js code run in the
	// setup and teardown phase of the group's API checks; nil means none.
	LocalSetupScript    *string `json:"localSetupScript"`
	LocalTearDownScript *string `json:"localTearDownScript"`
	// AlertChannelSubscriptions are the alert channels that receive alerts
	// for the group's checks. They apply only when UseGlobalAlertSettings is
	// non-nil or AlertSettings is set; otherwise each check uses its own
	// subscriptions. The SDK disables automatic assignment, so an empty list
	// leaves the group without any.
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	// RuntimeID is the runtime version, such as "2025.04", for the checks
	// in the group that have none of their own.
	RuntimeID *string `json:"runtimeId"`
	// PrivateLocations are the slug names of the account's private
	// locations to run from.
	PrivateLocations *[]string `json:"privateLocations"`
	// RetryStrategy applies to every check in the group instead of the
	// check's own, unless its Type is [RetryStrategyTypeFallback], which lets
	// each check use its own. Nil disables retries.
	RetryStrategy *RetryStrategy `json:"retryStrategy"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// APICheckDefaults are request settings of a [GroupV2] for the API checks in
// it.
type APICheckDefaults struct {
	// BaseURL is the group's base URL, which the checks' request URLs can
	// reference as {{GROUP_BASE_URL}}.
	BaseURL         string      `json:"url"`
	Headers         []KeyValue  `json:"headers,omitempty"`
	QueryParameters []KeyValue  `json:"queryParameters,omitempty"`
	Assertions      []Assertion `json:"assertions,omitempty"`
	BasicAuth       BasicAuth   `json:"basicAuth,omitempty"`
}

type groupV2Payload struct {
	GroupV2
	// The embedded field's omitempty would drop the key for nil and empty
	// slices, leaving it up to the backend whether the server-side value is
	// kept or cleared. Shadow the field so an explicit empty list is always
	// sent and removing all subscriptions reliably clears them.
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions"`
}

func createGroupV2Payload(group GroupV2) groupV2Payload {
	payload := groupV2Payload{
		GroupV2: group,
	}

	// A nil value for a list will cause the backend to not update the value.
	// We must send empty lists instead.
	if group.Locations == nil {
		payload.Locations = []string{}
	}

	if group.PrivateLocations == nil {
		payload.PrivateLocations = &[]string{}
	}

	payload.AlertChannelSubscriptions = emptyIfNil(group.AlertChannelSubscriptions)

	// The API rejects null tags.
	payload.Tags = emptyIfNil(group.Tags)

	return payload
}

func (c *client) CreateGroupV2(
	ctx context.Context,
	group GroupV2,
) (*GroupV2, error) {
	var result GroupV2
	err := c.callJSON(ctx, http.MethodPost, withAutoAssignAlertsFlag("/v2/check-groups"), createGroupV2Payload(group), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetGroupV2(
	ctx context.Context,
	id int64,
) (*GroupV2, error) {
	var result GroupV2
	// The API has no v2 route for reading a group. The v1 read response has
	// the same shape as the v2 create and update responses, plus the
	// read-only projectBindings, which GroupV2 does not decode.
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/check-groups/%d", id), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateGroupV2(
	ctx context.Context,
	id int64,
	group GroupV2,
) (*GroupV2, error) {
	var result GroupV2
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v2/check-groups/%d", id)), createGroupV2Payload(group), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteGroupV2(
	ctx context.Context,
	id int64,
) error {
	// The API has no v2 route for deleting a group; the v1 route deletes
	// groups created through either version.
	_, err := c.call(ctx, http.MethodDelete, fmt.Sprintf("/v1/check-groups/%d", id), nil, http.StatusNoContent)
	return err
}
