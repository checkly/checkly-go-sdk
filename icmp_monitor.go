package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ICMPMonitor pings a host and checks the packet loss and its assertions. It
// is the input of [Client.CreateICMPMonitor] and [Client.UpdateICMPMonitor],
// and is returned by [Client.GetICMPMonitor].
type ICMPMonitor struct {
	// ID is the monitor's UUID, assigned by the API.
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
	// Muted stops notifications when the monitor fails or recovers.
	Muted bool `json:"muted"`
	// RunParallel runs the monitor in all its locations in parallel.
	RunParallel bool `json:"runParallel"`
	// Locations are the public locations to run from, as region codes such
	// as "us-east-1". Nil is sent as an empty list.
	Locations []string `json:"locations"`
	// DegradedPacketLossThreshold and MaxPacketLossThreshold are the
	// percentages of lost packets at which a run counts as degraded and as
	// failed: greater than 0 and at most 100, with MaxPacketLossThreshold at
	// least DegradedPacketLossThreshold. The SDK always sends both, so both
	// must be set.
	DegradedPacketLossThreshold int      `json:"degradedPacketLossThreshold,omitempty"`
	MaxPacketLossThreshold      int      `json:"maxPacketLossThreshold,omitempty"`
	Tags                        []string `json:"tags,omitempty"`
	// AlertSettings is left out of the request when nil.
	AlertSettings *AlertSettings `json:"alertSettings,omitempty"`
	// UseGlobalAlertSettings applies the account's alert settings instead
	// of AlertSettings.
	UseGlobalAlertSettings bool        `json:"useGlobalAlertSettings"`
	Request                ICMPRequest `json:"request"`
	// GroupID is the ID of the check group the monitor belongs to; 0 means
	// none.
	GroupID int64 `json:"groupId,omitempty"`
	// GroupOrder is the monitor's position in its group, which sets the run
	// order when the group is triggered from the API or CI/CD.
	GroupOrder int `json:"groupOrder,omitempty"`
	// AlertChannelSubscriptions are the alert channels that receive the
	// monitor's alerts. The SDK disables automatic assignment, so an
	// empty list leaves the monitor without any.
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	// PrivateLocations are the slug names of the account's private
	// locations to run from. Nil is sent as an empty list. The Checkly agent
	// at those locations must be version 7.0.0 or later and have the NET_RAW
	// capability.
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

// ICMPRequest is the ping of an [ICMPMonitor].
type ICMPRequest struct {
	Hostname string `json:"hostname"`
	// IPFamily is left out of the request when empty.
	IPFamily IPFamily `json:"ipFamily,omitempty"`
	// PingCount is 1 to 50; the API defaults to 10.
	// PingCount is left out of the request when zero.
	PingCount  int         `json:"pingCount,omitempty"`
	Assertions []Assertion `json:"assertions"`
}

type icmpMonitorPayload struct {
	ICMPMonitor
	Type        CheckType  `json:"checkType"`
	DoubleCheck bool       `json:"doubleCheck"`
	ShouldFail  bool       `json:"shouldFail"`
	GroupID     *int64     `json:"groupId"`
	ID          *string    `json:"id,omitempty"`         // Skip, can't be changed.
	CreatedAt   *time.Time `json:"created_at,omitempty"` // Skip, can't be changed.
	UpdatedAt   *time.Time `json:"updated_at,omitempty"` // Skip, can't be changed.
	// The update endpoint is PATCH-style: keys omitted from the payload keep
	// their server-side value. The fields below shadow embedded fields whose
	// zero value would be omitted so that they are always sent and can
	// actually be cleared via an update.
	Tags                        []string                   `json:"tags"`
	AlertChannelSubscriptions   []AlertChannelSubscription `json:"alertChannelSubscriptions"`
	DegradedPacketLossThreshold int                        `json:"degradedPacketLossThreshold"`
	MaxPacketLossThreshold      int                        `json:"maxPacketLossThreshold"`
	GroupOrder                  *int                       `json:"groupOrder"`
}

func createICMPMonitorPayload(monitor ICMPMonitor) icmpMonitorPayload {
	payload := icmpMonitorPayload{
		ICMPMonitor: monitor,
		// Unfortunately `checkType` is required for this endpoint.
		Type: CheckTypeICMP,
		// Unfortunately, this will default to true if not set.
		DoubleCheck: false,
		// The monitor type has no ShouldFail field, so it is always sent as
		// false, which is also the API default.
		ShouldFail: false,
	}

	// The API rejects null for the assertion list.
	payload.Request.Assertions = emptyIfNil(monitor.Request.Assertions)

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

	// The packet loss thresholds must always be sent: omitting the keys
	// would make the backend default them on create and keep the previous
	// values on update. The backend rejects an explicit 0, and sending it
	// surfaces that as a validation error rather than silently storing a
	// different value.
	payload.DegradedPacketLossThreshold = monitor.DegradedPacketLossThreshold
	payload.MaxPacketLossThreshold = monitor.MaxPacketLossThreshold

	// GroupOrder must be null if empty or the order will not get unset on
	// update. Note that the check and TCP monitor endpoints differ: they
	// accept and reset on a literal 0.
	payload.GroupOrder = nullIfZero(monitor.GroupOrder)

	return payload
}

func (c *client) CreateICMPMonitor(
	ctx context.Context,
	monitor ICMPMonitor,
) (*ICMPMonitor, error) {
	var result ICMPMonitor
	err := c.callJSON(ctx, http.MethodPost, withAutoAssignAlertsFlag("/v1/checks/icmp"), createICMPMonitorPayload(monitor), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetICMPMonitor(
	ctx context.Context,
	ID string,
) (*ICMPMonitor, error) {
	var result ICMPMonitor
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s", ID), CheckTypeICMP, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateICMPMonitor(
	ctx context.Context,
	ID string,
	monitor ICMPMonitor,
) (*ICMPMonitor, error) {
	var result ICMPMonitor
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/icmp/%s", ID)), createICMPMonitorPayload(monitor), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteICMPMonitor(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}
