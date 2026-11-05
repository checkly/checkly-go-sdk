package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// TCPMonitor represents a TCP monitor.
type TCPMonitor struct {
	ID                        string                     `json:"id,omitempty"`
	Name                      string                     `json:"name"`
	Description               *string                    `json:"description"`
	Frequency                 int                        `json:"frequency"`
	FrequencyOffset           int                        `json:"frequencyOffset,omitempty"`
	Activated                 bool                       `json:"activated"`
	Muted                     bool                       `json:"muted"`
	ShouldFail                bool                       `json:"shouldFail"`
	RunParallel               bool                       `json:"runParallel"`
	Locations                 []string                   `json:"locations"`
	DegradedResponseTime      int                        `json:"degradedResponseTime,omitempty"`
	MaxResponseTime           int                        `json:"maxResponseTime,omitempty"`
	Tags                      []string                   `json:"tags,omitempty"`
	AlertSettings             *AlertSettings             `json:"alertSettings,omitempty"`
	UseGlobalAlertSettings    bool                       `json:"useGlobalAlertSettings"`
	Request                   TCPRequest                 `json:"request"`
	GroupID                   int64                      `json:"groupId,omitempty"`
	GroupOrder                int                        `json:"groupOrder,omitempty"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	PrivateLocations          *[]string                  `json:"privateLocations"`
	RuntimeID                 *string                    `json:"runtimeId"`
	RetryStrategy             *RetryStrategy             `json:"retryStrategy"`
	TriggerIncident           *IncidentTrigger           `json:"triggerIncident"`
	CreatedAt                 time.Time                  `json:"created_at,omitempty"`
	UpdatedAt                 time.Time                  `json:"updated_at,omitempty"`
}

// TCPRequest represents the parameters for a TCP check's connection.
type TCPRequest struct {
	Hostname   string      `json:"hostname"`
	Port       uint16      `json:"port"`
	Data       string      `json:"data,omitempty"`
	Assertions []Assertion `json:"assertions,omitempty"`
	IPFamily   string      `json:"ipFamily,omitempty"`
}

// tcpRequestPayload wraps TCPRequest for create and update payloads. The
// endpoints are PATCH-style down to the nested request object: request keys
// omitted from the payload keep their server-side value. Assertions shadows
// the embedded field, whose omitempty would drop the key for an empty list,
// so that removing all assertions reliably clears them. Data is not shadowed
// because the backend rejects an empty string ("request.data is not allowed
// to be empty") and silently keeps the previous value on null, so a data
// value can never be cleared, only changed. IPFamily is not shadowed because
// an empty string is not a valid value for it either; it can only be changed
// to another valid value.
type tcpRequestPayload struct {
	TCPRequest
	Assertions []Assertion `json:"assertions"`
}

type tcpMonitorPayload struct {
	TCPMonitor
	Type        string `json:"checkType"`
	DoubleCheck bool   `json:"doubleCheck"`
	GroupID     *int64 `json:"groupId"`
	// The update endpoint is PATCH-style: keys omitted from the payload keep
	// their server-side value. The fields below shadow embedded fields whose
	// zero value would be omitted so that they are always sent and can
	// actually be cleared via an update. AlertChannelSubscriptions is not
	// shadowed: this endpoint verifiably clears subscriptions when the key
	// is omitted.
	Tags                 []string          `json:"tags"`
	DegradedResponseTime int               `json:"degradedResponseTime"`
	MaxResponseTime      int               `json:"maxResponseTime"`
	GroupOrder           int               `json:"groupOrder"`
	Request              tcpRequestPayload `json:"request"`
}

func createTCPMonitorPayload(monitor TCPMonitor) tcpMonitorPayload {
	payload := tcpMonitorPayload{
		TCPMonitor: monitor,
		// Unfortunately `checkType` is required for this endpoint.
		Type: "TCP",
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

	// The response time limits must always be sent: 0 is a valid value, and
	// omitting the keys would make the backend default them on create and
	// keep the previous values on update.
	payload.DegradedResponseTime = monitor.DegradedResponseTime
	payload.MaxResponseTime = monitor.MaxResponseTime

	// The group order must always be sent: 0 is the value it resets to, and
	// omitting the key would make the backend keep the previous value on
	// update. Note that the URL, DNS and ICMP monitor endpoints differ:
	// they only reset the order on an explicit null, not 0.
	payload.GroupOrder = monitor.GroupOrder

	payload.Request = tcpRequestPayload{
		TCPRequest: monitor.Request,
		Assertions: emptyIfNil(monitor.Request.Assertions),
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

// GetTCPMonitor takes the ID of an existing TCP monitor, and returns the
// monitor parameters, or an error. It fails with an error matching
// ErrCheckTypeMismatch when the ID belongs to a check of another type.
func (c *client) GetTCPMonitor(
	ctx context.Context,
	ID string,
) (*TCPMonitor, error) {
	var result TCPMonitor
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s", ID), "TCP", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateTCPMonitor(
	ctx context.Context,
	ID string,
	monitor TCPMonitor,
) (*TCPMonitor, error) {
	var result TCPMonitor
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/tcp/%s", ID)), createTCPMonitorPayload(monitor), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteTCPMonitor(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}
