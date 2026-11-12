package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// URLMonitor requests a URL with GET and checks the response status code
// against its assertions. It is the input of [Client.CreateURLMonitor] and
// [Client.UpdateURLMonitor], and is returned by [Client.GetURLMonitor].
type URLMonitor struct {
	// ID is the monitor's UUID, assigned by the API.
	ID          string  `json:"id,omitempty"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	// Frequency is the number of minutes between runs: 0, 1, 2, 5, 10, 15,
	// 30, 60, 120, 180, 360, 720 or 1440.
	Frequency int `json:"frequency"`
	// FrequencyOffset is 10, 20 or 30 when Frequency is 0, for a run every
	// that many seconds. For any other Frequency the API picks the offset
	// itself and keeps it until Frequency changes: a value sent has no
	// effect, and the value read back is the API's. The API still rejects a
	// value above Frequency * 10 for a Frequency up to 60, or above
	// Frequency / 5 rounded up for a higher one. FrequencyOffset is left out
	// of the request when zero.
	FrequencyOffset int  `json:"frequencyOffset,omitempty"`
	Activated       bool `json:"activated"`
	// Muted stops notifications when the monitor fails or recovers.
	Muted bool `json:"muted"`
	// ShouldFail inverts the result: a run that would fail passes, and the
	// other way around.
	ShouldFail bool `json:"shouldFail"`
	// RunParallel runs the monitor in all its locations in parallel.
	RunParallel bool `json:"runParallel"`
	// Locations are the public locations to run from, as region codes such
	// as "us-east-1". Nil is sent as an empty list.
	Locations []string `json:"locations"`
	// DegradedResponseTime and MaxResponseTime are the degraded and failing
	// response time limits in milliseconds, each at most 45000.
	DegradedResponseTime int      `json:"degradedResponseTime,omitempty"`
	MaxResponseTime      int      `json:"maxResponseTime,omitempty"`
	Tags                 []string `json:"tags,omitempty"`
	// AlertSettings is left out of the request when nil.
	AlertSettings *AlertSettings `json:"alertSettings,omitempty"`
	// UseGlobalAlertSettings applies the account's alert settings instead
	// of AlertSettings.
	UseGlobalAlertSettings bool       `json:"useGlobalAlertSettings"`
	Request                URLRequest `json:"request"`
	// GroupID is the ID of the check group the monitor belongs to; 0 means
	// none.
	GroupID int64 `json:"groupId,omitempty"`
	// GroupOrder is the monitor's position in its group, which sets the run
	// order when the group is triggered from the API or CI/CD. 0 is the
	// first position. Nil means no position, and is sent as null so that
	// an update clears it.
	GroupOrder *int `json:"groupOrder"`
	// AlertChannelSubscriptions are the alert channels that receive the
	// monitor's alerts. The SDK disables automatic assignment, so an
	// empty list leaves the monitor without any.
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	// PrivateLocations are the slug names of the account's private
	// locations to run from. Nil is sent as an empty list.
	PrivateLocations *[]string `json:"privateLocations"`
	// RetryStrategy sets how failed runs are retried; nil disables retries.
	// In a group whose RetryStrategy is not [RetryStrategyTypeFallback], the
	// group's strategy applies instead.
	RetryStrategy *RetryStrategy `json:"retryStrategy"`
	// TriggerIncident makes the monitor open and resolve an incident on a
	// status page service, based on its alert configuration. Nil turns it
	// off.
	TriggerIncident *IncidentTrigger `json:"triggerIncident"`
	CreatedAt       time.Time        `json:"created_at,omitempty"`
	UpdatedAt       time.Time        `json:"updated_at,omitempty"`
}

func (*URLMonitor) isCheck() {}

// URLRequest is the request of a [URLMonitor]. The SDK always sends it with
// the method GET. It has no headers: the SDK neither reads nor sends them,
// so an update keeps headers set on the monitor in the web app or through
// the API.
type URLRequest struct {
	URL             string      `json:"url"`
	FollowRedirects bool        `json:"followRedirects"`
	SkipSSL         bool        `json:"skipSSL"`
	Assertions      []Assertion `json:"assertions"`
	// IPFamily is left out of the request when empty. An update with an
	// empty IPFamily keeps the stored family, as the API rejects an empty
	// value.
	IPFamily IPFamily `json:"ipFamily,omitempty"`
}

// urlRequestPayload is the request of a URL monitor as the SDK sends it: only
// the fields a URL monitor has, not those of an API check's [Request]. On
// accounts that can set URL monitor headers, the API stores the headers of
// every write that sends the key, so sending none keeps the headers set
// outside the SDK.
type urlRequestPayload struct {
	Method string `json:"method"`
	URLRequest
}

func (r URLRequest) toPayload() urlRequestPayload {
	// The API rejects null assertions.
	r.Assertions = emptyIfNil(r.Assertions)
	return urlRequestPayload{Method: http.MethodGet, URLRequest: r}
}

type urlMonitorPayload struct {
	URLMonitor
	Type        CheckType         `json:"checkType"`
	Request     urlRequestPayload `json:"request"`
	DoubleCheck bool              `json:"doubleCheck"`
	GroupID     *int64            `json:"groupId"`
	// The update endpoint is PATCH-style: keys omitted from the payload keep
	// their server-side value. The fields below shadow embedded fields whose
	// zero value would be omitted so that they are always sent and can
	// actually be cleared via an update.
	Tags                      []string                   `json:"tags"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions"`
	DegradedResponseTime      int                        `json:"degradedResponseTime"`
	MaxResponseTime           int                        `json:"maxResponseTime"`
}

func createURLMonitorPayload(monitor URLMonitor) urlMonitorPayload {
	payload := urlMonitorPayload{
		URLMonitor: monitor,
		// Unfortunately `checkType` is required for this endpoint.
		Type:    CheckTypeURL,
		Request: monitor.Request.toPayload(),
		// Unfortunately, this will default to true if not set.
		DoubleCheck: false,
	}

	// GroupID must be null if empty or the group will not get unset on update.
	if monitor.GroupID != 0 {
		payload.GroupID = &monitor.GroupID
	}

	// A nil value for a list will cause the backend to not update the value.
	// We must send empty lists instead.
	if monitor.Locations == nil {
		payload.Locations = []string{}
	}

	if monitor.PrivateLocations == nil {
		payload.PrivateLocations = &[]string{}
	}

	payload.Tags = emptyIfNil(monitor.Tags)

	payload.AlertChannelSubscriptions = emptyIfNil(monitor.AlertChannelSubscriptions)

	// The response time limits must always be sent: 0 is a valid value, and
	// omitting the keys would make the backend default them on create and
	// keep the previous values on update.
	payload.DegradedResponseTime = monitor.DegradedResponseTime
	payload.MaxResponseTime = monitor.MaxResponseTime

	return payload
}

func (c *client) CreateURLMonitor(
	ctx context.Context,
	monitor URLMonitor,
) (*URLMonitor, error) {
	var result URLMonitor
	err := c.callJSON(ctx, http.MethodPost, withAutoAssignAlertsFlag("/v1/checks/url"), createURLMonitorPayload(monitor), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetURLMonitor(
	ctx context.Context,
	id string,
) (*URLMonitor, error) {
	var result URLMonitor
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s", id), CheckTypeURL, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateURLMonitor(
	ctx context.Context,
	id string,
	monitor URLMonitor,
) (*URLMonitor, error) {
	var result URLMonitor
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/url/%s", id)), createURLMonitorPayload(monitor), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteURLMonitor(
	ctx context.Context,
	id string,
) error {
	return c.DeleteCheck(ctx, id)
}
