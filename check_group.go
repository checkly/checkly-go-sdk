package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Group represents a check group.
type Group struct {
	ID                        int64                      `json:"id,omitempty"`
	Name                      string                     `json:"name"`
	Activated                 bool                       `json:"activated"`
	Muted                     bool                       `json:"muted"`
	RunParallel               bool                       `json:"runParallel"`
	Tags                      []string                   `json:"tags"`
	Locations                 []string                   `json:"locations"`
	Concurrency               int                        `json:"concurrency"`
	APICheckDefaults          APICheckDefaults           `json:"apiCheckDefaults"`
	EnvironmentVariables      []EnvironmentVariable      `json:"environmentVariables"`
	UseGlobalAlertSettings    bool                       `json:"useGlobalAlertSettings"`
	AlertSettings             AlertSettings              `json:"alertSettings,omitempty"`
	SetupSnippetID            int64                      `json:"setupSnippetId,omitempty"`
	TearDownSnippetID         int64                      `json:"tearDownSnippetId,omitempty"`
	LocalSetupScript          string                     `json:"localSetupScript,omitempty"`
	LocalTearDownScript       string                     `json:"localTearDownScript,omitempty"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	CreatedAt                 time.Time                  `json:"created_at"`
	UpdatedAt                 time.Time                  `json:"updated_at"`

	// Pointers
	RuntimeID        *string        `json:"runtimeId"`
	PrivateLocations *[]string      `json:"privateLocations"`
	RetryStrategy    *RetryStrategy `json:"retryStrategy"`

	// Deprecated: this property will be removed in future versions. Please use RetryStrategy instead.
	DoubleCheck bool `json:"doubleCheck"`
}

// Group represents a check group v2.
type GroupV2 struct {
	ID                        int64                      `json:"id,omitempty"`
	Name                      string                     `json:"name"`
	Activated                 bool                       `json:"activated"`
	Muted                     bool                       `json:"muted"`
	RunParallel               *bool                      `json:"runParallel"`
	Tags                      []string                   `json:"tags"`
	Locations                 []string                   `json:"locations"`
	Concurrency               int                        `json:"concurrency"`
	APICheckDefaults          APICheckDefaults           `json:"apiCheckDefaults"`
	EnvironmentVariables      []EnvironmentVariable      `json:"environmentVariables"`
	UseGlobalAlertSettings    *bool                      `json:"useGlobalAlertSettings"`
	AlertSettings             *AlertSettings             `json:"alertSettings"`
	SetupSnippetID            *int64                     `json:"setupSnippetId"`
	TearDownSnippetID         *int64                     `json:"tearDownSnippetId"`
	LocalSetupScript          *string                    `json:"localSetupScript"`
	LocalTearDownScript       *string                    `json:"localTearDownScript"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	RuntimeID                 *string                    `json:"runtimeId"`
	PrivateLocations          *[]string                  `json:"privateLocations"`
	RetryStrategy             *RetryStrategy             `json:"retryStrategy"`
	CreatedAt                 time.Time                  `json:"created_at"`
	UpdatedAt                 time.Time                  `json:"updated_at"`
}

// APICheckDefaults represents the default settings for API checks within a
// given group.
type APICheckDefaults struct {
	BaseURL         string      `json:"url"`
	Headers         []KeyValue  `json:"headers,omitempty"`
	QueryParameters []KeyValue  `json:"queryParameters,omitempty"`
	Assertions      []Assertion `json:"assertions,omitempty"`
	BasicAuth       BasicAuth   `json:"basicAuth,omitempty"`
}

type groupPayload struct {
	Group
	// The update endpoint is PATCH-style: keys omitted from the payload keep
	// their server-side value. The fields below shadow embedded fields whose
	// zero value would be omitted so that they are always sent and can
	// actually be cleared via an update.
	SetupSnippetID            *int64                     `json:"setupSnippetId"`
	TearDownSnippetID         *int64                     `json:"tearDownSnippetId"`
	LocalSetupScript          *string                    `json:"localSetupScript"`
	LocalTearDownScript       *string                    `json:"localTearDownScript"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions"`
}

func createGroupPayload(group Group) groupPayload {
	payload := groupPayload{
		Group: group,
	}

	// A nil value for a list will cause the backend to not update the value.
	// We must send empty lists instead.
	if group.Locations == nil {
		payload.Locations = []string{}
	}

	if group.PrivateLocations == nil {
		payload.PrivateLocations = &[]string{}
	}

	// The snippet IDs must be null if empty or they will not get unset on
	// update.
	payload.SetupSnippetID = nullIfZero(group.SetupSnippetID)

	payload.TearDownSnippetID = nullIfZero(group.TearDownSnippetID)

	// The local scripts must be null if empty or they will not get unset on
	// update.
	payload.LocalSetupScript = nullIfZero(group.LocalSetupScript)

	payload.LocalTearDownScript = nullIfZero(group.LocalTearDownScript)

	payload.AlertChannelSubscriptions = emptyIfNil(group.AlertChannelSubscriptions)

	return payload
}

// CreateGroup creates a new check group with the specified details. It returns
// the newly-created group, or an error.
func (c *client) CreateGroup(
	ctx context.Context,
	group Group,
) (*Group, error) {
	var result Group
	err := c.callJSON(ctx, http.MethodPost, withAutoAssignAlertsFlag("/v1/check-groups"), createGroupPayload(group), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
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

	return payload
}

// GetGroup takes the ID of an existing check group, and returns the
// corresponding group, or an error.
func (c *client) GetGroup(
	ctx context.Context,
	ID int64,
) (*Group, error) {
	result := Group{}
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/check-groups/%d", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateGroup takes the ID of an existing check group, and updates the
// corresponding check group to match the supplied group. It returns the updated
// group, or an error.
func (c *client) UpdateGroup(
	ctx context.Context,
	ID int64,
	group Group,
) (*Group, error) {
	var result Group
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/check-groups/%d", ID)), createGroupPayload(group), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteGroup deletes the check group with the specified ID.
func (c *client) DeleteGroup(
	ctx context.Context,
	ID int64,
) error {
	_, err := c.call(ctx, http.MethodDelete, fmt.Sprintf("/v1/check-groups/%d", ID), nil, http.StatusNoContent)
	return err
}

// CreateGroupV2 creates a new check group with the specified details. It returns
// the newly-created group, or an error.
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

// GetGroupV2 takes the ID of an existing check group, and returns the
// corresponding group, or an error.
func (c *client) GetGroupV2(
	ctx context.Context,
	ID int64,
) (*GroupV2, error) {
	var result GroupV2
	// The API has no v2 route for reading a group. The v1 read response has
	// the same shape as the v2 create and update responses, plus the
	// read-only projectBindings, which GroupV2 does not decode.
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/check-groups/%d", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateGroupV2 takes the ID of an existing check group, and updates the
// corresponding check group to match the supplied group. It returns the updated
// group, or an error.
func (c *client) UpdateGroupV2(
	ctx context.Context,
	ID int64,
	group GroupV2,
) (*GroupV2, error) {
	var result GroupV2
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v2/check-groups/%d", ID)), createGroupV2Payload(group), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteGroupV2 deletes the check group with the specified ID.
func (c *client) DeleteGroupV2(
	ctx context.Context,
	ID int64,
) error {
	// The API has no v2 route for deleting a group; the v1 route deletes
	// groups created through either version.
	return c.DeleteGroup(ctx, ID)
}
