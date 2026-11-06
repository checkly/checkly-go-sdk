package checkly

import (
	"context"
	"encoding/json"
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

	// Intent describes what the check protects, for automatic AI repair.
	// It is sent as null when nil, so an update with a nil Intent clears an
	// intent set elsewhere, for example in the web app.
	Intent *CheckIntent `json:"intent"`
	// AIAutoRepairEnabled overrides the account's automatic AI repair
	// setting for this check. Nil inherits the account setting; it is sent
	// as null, so an update with nil removes an override set elsewhere.
	AIAutoRepairEnabled *bool `json:"aiAutoRepairEnabled"`
}

// BodyType is the format of a request body, in [Request].BodyType.
type BodyType string

// Body types, the values of [Request].BodyType. The API default is
// BodyTypeNone.
const (
	BodyTypeJSON    BodyType = "JSON"
	BodyTypeForm    BodyType = "FORM"
	BodyTypeRaw     BodyType = "RAW"
	BodyTypeGraphQL BodyType = "GRAPHQL"
	BodyTypeNone    BodyType = "NONE"
)

// IPFamily is the IP version a check or monitor connects with, in the
// IPFamily field of [Request], [URLRequest], [TCPRequest], [ICMPRequest],
// [TracerouteRequest], [GRPCRequest] and [SSLConfig].
type IPFamily string

// IP families, the values of the IPFamily fields. The API default is
// IPFamilyIPv4.
const (
	IPFamilyIPv4 IPFamily = "IPv4"
	IPFamilyIPv6 IPFamily = "IPv6"
)

// Request represents the parameters for the request made by the check.
type Request struct {
	// Method is GET, POST, PUT, HEAD, DELETE or PATCH, as the net/http
	// constants such as http.MethodGet name them.
	Method          string      `json:"method"`
	URL             string      `json:"url"`
	FollowRedirects bool        `json:"followRedirects"`
	SkipSSL         bool        `json:"skipSSL"`
	Body            string      `json:"body"`
	BodyType        BodyType    `json:"bodyType,omitempty"`
	Headers         []KeyValue  `json:"headers"`
	QueryParameters []KeyValue  `json:"queryParameters"`
	Assertions      []Assertion `json:"assertions"`
	// BasicAuth must always be serialized (as null when absent): the check
	// update endpoints are PATCH-style and keep the server-side credentials
	// when the key is omitted, so omitting it would make it impossible to
	// remove previously configured credentials.
	BasicAuth *BasicAuth `json:"basicAuth"`
	IPFamily  IPFamily   `json:"ipFamily,omitempty"`
}

// MarshalJSON sends unset headers, query parameters and assertions as [],
// as the API rejects null for these lists.
func (r Request) MarshalJSON() ([]byte, error) {
	type alias Request
	r.Headers = emptyIfNil(r.Headers)
	r.QueryParameters = emptyIfNil(r.QueryParameters)
	r.Assertions = emptyIfNil(r.Assertions)
	return json.Marshal(alias(r))
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
	Type CheckType `json:"checkType"`
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
		Type: CheckTypeAPI,
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
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s", ID), CheckTypeAPI, &result); err != nil {
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
