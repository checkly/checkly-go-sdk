package checkly

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// APICheck is an API check: it sends an HTTP request and runs assertions on
// the response. It is the input and result of [Client.CreateAPICheck] and
// [Client.UpdateAPICheck], and the result of [Client.GetAPICheck].
//
// Fields whose zero value would be left out of the request are sent anyway
// where the zero value is how they are cleared: nil Locations,
// PrivateLocations, Tags and AlertChannelSubscriptions are sent as empty
// lists, and a GroupID, SetupSnippetID or TearDownSnippetID of 0 as null.
type APICheck struct {
	// ID is the check's UUID, assigned by the API.
	ID          string  `json:"id,omitempty"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	// Frequency is the number of minutes between runs: 0, 1, 2, 5, 10, 15,
	// 30, 60, 120, 180, 360, 720 or 1440.
	Frequency int `json:"frequency"`
	// FrequencyOffset must be 10, 20 or 30 when Frequency is 0, at most ten
	// times Frequency when Frequency is 1 to 60, and at most Frequency / 5
	// (rounded up) above 60. FrequencyOffset is left out of the request
	// when zero.
	FrequencyOffset int  `json:"frequencyOffset,omitempty"`
	Activated       bool `json:"activated"`
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
	// DegradedResponseTime and MaxResponseTime are the degraded and failing
	// response time limits in milliseconds, each at most 45000.
	DegradedResponseTime int      `json:"degradedResponseTime"`
	MaxResponseTime      int      `json:"maxResponseTime"`
	Tags                 []string `json:"tags"`
	// AlertSettings is left out of the request when nil.
	AlertSettings *AlertSettings `json:"alertSettings,omitempty"`
	// UseGlobalAlertSettings applies the account's alert settings instead
	// of AlertSettings.
	UseGlobalAlertSettings bool    `json:"useGlobalAlertSettings"`
	Request                Request `json:"request"`
	// SetupSnippetID and TearDownSnippetID are the IDs of the [Snippet] run
	// in the setup and teardown phase; 0 means none.
	SetupSnippetID    int64 `json:"setupSnippetId,omitempty"`
	TearDownSnippetID int64 `json:"tearDownSnippetId,omitempty"`
	// LocalSetupScript and LocalTearDownScript are Node.js code run in the
	// setup and teardown phase.
	LocalSetupScript    *string `json:"localSetupScript"`
	LocalTearDownScript *string `json:"localTearDownScript"`
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

// Request is the HTTP request of an API check, with the assertions run on
// its response.
type Request struct {
	// Method is GET, POST, PUT, HEAD, DELETE or PATCH, as the net/http
	// constants such as http.MethodGet name them.
	Method          string `json:"method"`
	URL             string `json:"url"`
	FollowRedirects bool   `json:"followRedirects"`
	SkipSSL         bool   `json:"skipSSL"`
	Body            string `json:"body"`
	// BodyType is left out of the request when empty.
	BodyType        BodyType    `json:"bodyType,omitempty"`
	Headers         []KeyValue  `json:"headers"`
	QueryParameters []KeyValue  `json:"queryParameters"`
	Assertions      []Assertion `json:"assertions"`
	// BasicAuth is always sent, as null when nil: the update endpoints keep
	// the stored credentials when the key is left out, so sending null is how
	// previously configured credentials are removed.
	BasicAuth *BasicAuth `json:"basicAuth"`
	// IPFamily is left out of the request when empty.
	IPFamily IPFamily `json:"ipFamily,omitempty"`
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

// BasicAuth holds the HTTP basic authentication credentials of a [Request]
// or of [APICheckDefaults].
type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// KeyValue is a key-value pair, such as a request header or query parameter.
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
