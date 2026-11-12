package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// TCPMonitor connects to a host and port over TCP, optionally sends data, and
// checks the response against its assertions. It is the input of
// [Client.CreateTCPMonitor] and [Client.UpdateTCPMonitor], and is returned by
// [Client.GetTCPMonitor].
type TCPMonitor struct {
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
	Request                TCPRequest `json:"request"`
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
	// RuntimeID is the runtime version, such as "2025.04".
	RuntimeID *string `json:"runtimeId"`
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

func (*TCPMonitor) isCheck() {}

// TCPRequest is the connection of a [TCPMonitor].
type TCPRequest struct {
	Hostname string `json:"hostname"`
	Port     uint16 `json:"port"`
	// Data is sent to the server after the connection is established.
	Data       string      `json:"data,omitempty"`
	Assertions []Assertion `json:"assertions,omitempty"`
	// IPFamily is left out of the request when empty. An update with an
	// empty IPFamily keeps the stored family, as the API rejects an empty
	// value.
	IPFamily IPFamily `json:"ipFamily,omitempty"`
}

// tcpRequestPayload wraps TCPRequest for create and update payloads. The
// endpoints are PATCH-style down to the nested request object: request keys
// omitted from the payload keep their server-side value. Assertions shadows
// the embedded field, whose omitempty would drop the key for an empty list,
// so that removing all assertions reliably clears them. Data shadows the
// embedded field to send an empty value as null, which clears the stored
// data on update; the API rejects an empty string. IPFamily is not shadowed
// because an empty string is not a valid value for it; it can only be
// changed to another valid value.
type tcpRequestPayload struct {
	TCPRequest
	Assertions []Assertion `json:"assertions"`
	Data       *string     `json:"data"`
}

type tcpMonitorPayload struct {
	TCPMonitor
	Type        CheckType `json:"checkType"`
	DoubleCheck bool      `json:"doubleCheck"`
	GroupID     *int64    `json:"groupId"`
	// The update endpoint is PATCH-style: keys omitted from the payload keep
	// their server-side value. The fields below shadow embedded fields whose
	// zero value would be omitted so that they are always sent and can
	// actually be cleared via an update.
	Tags                      []string                   `json:"tags"`
	DegradedResponseTime      int                        `json:"degradedResponseTime"`
	MaxResponseTime           int                        `json:"maxResponseTime"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions"`
	Request                   tcpRequestPayload          `json:"request"`
}

func createTCPMonitorPayload(monitor TCPMonitor) tcpMonitorPayload {
	payload := tcpMonitorPayload{
		TCPMonitor: monitor,
		// Unfortunately `checkType` is required for this endpoint.
		Type: CheckTypeTCP,
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

	payload.Request = tcpRequestPayload{
		TCPRequest: monitor.Request,
		Assertions: emptyIfNil(monitor.Request.Assertions),
		Data:       nullIfZero(monitor.Request.Data),
	}

	return payload
}

func (c *client) CreateTCPMonitor(
	ctx context.Context,
	monitor TCPMonitor,
) (*TCPMonitor, error) {
	var result TCPMonitor
	err := c.callJSON(ctx, http.MethodPost, withAutoAssignAlertsFlag("/v1/checks/tcp"), createTCPMonitorPayload(monitor), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetTCPMonitor(
	ctx context.Context,
	id string,
) (*TCPMonitor, error) {
	var result TCPMonitor
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s", id), CheckTypeTCP, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateTCPMonitor(
	ctx context.Context,
	id string,
	monitor TCPMonitor,
) (*TCPMonitor, error) {
	var result TCPMonitor
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/tcp/%s", id)), createTCPMonitorPayload(monitor), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteTCPMonitor(
	ctx context.Context,
	id string,
) error {
	return c.DeleteCheck(ctx, id)
}
