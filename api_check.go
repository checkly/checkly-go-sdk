package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// APICheck represents an API check.
//
// The per-type check update endpoints are PATCH-style: keys omitted from the
// payload keep their server-side value. Clearable fields of APICheck,
// BrowserCheck and MultiStepCheck are therefore serialized even when empty,
// and the client sends nil slices as empty lists.
type APICheck struct {
	ID              string   `json:"id,omitempty"`
	Name            string   `json:"name"`
	Description     *string  `json:"description"`
	Frequency       int      `json:"frequency"`
	FrequencyOffset int      `json:"frequencyOffset,omitempty"`
	Activated       bool     `json:"activated"`
	Muted           bool     `json:"muted"`
	ShouldFail      bool     `json:"shouldFail"`
	RunParallel     bool     `json:"runParallel"`
	Locations       []string `json:"locations"`
	// DegradedResponseTime is the response time in milliseconds where the
	// check should be considered degraded. The API accepts values between 0
	// and 30000. It is always sent, so leaving it at 0 sets a limit of 0
	// rather than the API's default of 5000.
	DegradedResponseTime int `json:"degradedResponseTime"`
	// MaxResponseTime is the response time in milliseconds where the check
	// should be considered failing. The API accepts values between 0 and
	// 30000. It is always sent, so leaving it at 0 sets a limit of 0 rather
	// than the API's default of 20000.
	MaxResponseTime           int                        `json:"maxResponseTime"`
	Tags                      []string                   `json:"tags"`
	AlertSettings             *AlertSettings             `json:"alertSettings,omitempty"`
	UseGlobalAlertSettings    bool                       `json:"useGlobalAlertSettings"`
	Request                   Request                    `json:"request"`
	SetupSnippetID            int64                      `json:"setupSnippetId,omitempty"`
	TearDownSnippetID         int64                      `json:"tearDownSnippetId,omitempty"`
	LocalSetupScript          *string                    `json:"localSetupScript"`
	LocalTearDownScript       *string                    `json:"localTearDownScript"`
	GroupID                   int64                      `json:"groupId,omitempty"`
	GroupOrder                int                        `json:"groupOrder"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions"`
	PrivateLocations          *[]string                  `json:"privateLocations"`
	RuntimeID                 *string                    `json:"runtimeId"`
	RetryStrategy             *RetryStrategy             `json:"retryStrategy"`
	TriggerIncident           *IncidentTrigger           `json:"triggerIncident"`
	CreatedAt                 time.Time                  `json:"created_at,omitempty"`
	UpdatedAt                 time.Time                  `json:"updated_at,omitempty"`
}

// Request represents the parameters for the request made by the check.
type Request struct {
	Method          string      `json:"method"`
	URL             string      `json:"url"`
	FollowRedirects bool        `json:"followRedirects"`
	SkipSSL         bool        `json:"skipSSL"`
	Body            string      `json:"body"`
	BodyType        string      `json:"bodyType,omitempty"`
	Headers         []KeyValue  `json:"headers"`
	QueryParameters []KeyValue  `json:"queryParameters"`
	Assertions      []Assertion `json:"assertions"`
	// BasicAuth must always be serialized (as null when absent): the check
	// update endpoints are PATCH-style and keep the server-side credentials
	// when the key is omitted, so omitting it would make it impossible to
	// remove previously configured credentials.
	BasicAuth *BasicAuth `json:"basicAuth"`
	IPFamily  string     `json:"ipFamily,omitempty"`
}

// BasicAuth represents the HTTP basic authentication credentials for a request.
type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// KeyValue represents a key-value pair, for example a request header setting,
// or a query parameter.
type KeyValue struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Locked bool   `json:"locked"`
}

type apiCheckPayload struct {
	APICheck
	Type string `json:"checkType"`
	// The update endpoint is PATCH-style: keys omitted from the payload keep
	// their server-side value. The fields below shadow embedded fields whose
	// Go type cannot express the value that clears them.
	DoubleCheck       bool   `json:"doubleCheck"`
	GroupID           *int64 `json:"groupId"`
	SetupSnippetID    *int64 `json:"setupSnippetId"`
	TearDownSnippetID *int64 `json:"tearDownSnippetId"`
}

func createAPICheckPayload(check APICheck) apiCheckPayload {
	payload := apiCheckPayload{
		APICheck: check,
		// The endpoint defaults `checkType` correctly, but be explicit anyway.
		Type: "API",
		// Unfortunately, this will default to true if not set. Always sending
		// false also clears the deprecated doubleCheck flag on updates of
		// checks that still have it set; retryStrategy is its replacement.
		DoubleCheck:       false,
		GroupID:           nullIfZero(check.GroupID),
		SetupSnippetID:    nullIfZero(check.SetupSnippetID),
		TearDownSnippetID: nullIfZero(check.TearDownSnippetID),
	}

	// A nil value for a list will cause the backend to not update the value.
	// We must send empty lists instead.
	payload.Locations = emptyIfNil(check.Locations)
	if check.PrivateLocations == nil {
		payload.PrivateLocations = &[]string{}
	}
	payload.Tags = emptyIfNil(check.Tags)
	payload.AlertChannelSubscriptions = emptyIfNil(check.AlertChannelSubscriptions)

	return payload
}

func (c *client) CreateAPICheck(
	ctx context.Context,
	check APICheck,
) (*APICheck, error) {
	var result APICheck
	err := c.callJSON(ctx, http.MethodPost, withAutoAssignAlertsFlag("/v1/checks/api"), createAPICheckPayload(check), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAPICheck takes the ID of an existing API check, and returns the check
// parameters, or an error. It fails with an error matching ErrCheckTypeMismatch
// when the ID belongs to a check of another type.
func (c *client) GetAPICheck(
	ctx context.Context,
	ID string,
) (*APICheck, error) {
	var result APICheck
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s", ID), "API", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateAPICheck(
	ctx context.Context,
	ID string,
	check APICheck,
) (*APICheck, error) {
	var result APICheck
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/api/%s", ID)), createAPICheckPayload(check), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteAPICheck(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}
