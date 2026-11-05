package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ICMPMonitor represents an ICMP monitor.
type ICMPMonitor struct {
	ID                          string                     `json:"id,omitempty"`
	Name                        string                     `json:"name"`
	Description                 *string                    `json:"description"`
	Frequency                   int                        `json:"frequency"`
	FrequencyOffset             int                        `json:"frequencyOffset,omitempty"`
	Activated                   bool                       `json:"activated"`
	Muted                       bool                       `json:"muted"`
	RunParallel                 bool                       `json:"runParallel"`
	Locations                   []string                   `json:"locations"`
	DegradedPacketLossThreshold int                        `json:"degradedPacketLossThreshold,omitempty"`
	MaxPacketLossThreshold      int                        `json:"maxPacketLossThreshold,omitempty"`
	Tags                        []string                   `json:"tags,omitempty"`
	AlertSettings               *AlertSettings             `json:"alertSettings,omitempty"`
	UseGlobalAlertSettings      bool                       `json:"useGlobalAlertSettings"`
	Request                     ICMPRequest                `json:"request"`
	GroupID                     int64                      `json:"groupId,omitempty"`
	GroupOrder                  int                        `json:"groupOrder,omitempty"`
	AlertChannelSubscriptions   []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	// PrivateLocations are the slugs of the private locations the monitor runs
	// from. The Checkly agent at those locations must be version 7.0.0 or later and
	// have the NET_RAW capability.
	PrivateLocations *[]string        `json:"privateLocations"`
	RetryStrategy    *RetryStrategy   `json:"retryStrategy"`
	TriggerIncident  *IncidentTrigger `json:"triggerIncident"`
	CreatedAt        time.Time        `json:"created_at,omitempty"`
	UpdatedAt        time.Time        `json:"updated_at,omitempty"`
}

// ICMPRequest represents the parameters for the request made by an ICMP monitor.
type ICMPRequest struct {
	Hostname   string      `json:"hostname"`
	IPFamily   string      `json:"ipFamily,omitempty"`
	PingCount  int         `json:"pingCount,omitempty"`
	Assertions []Assertion `json:"assertions"`
}

type icmpMonitorPayload struct {
	ICMPMonitor
	Type        string     `json:"checkType"`
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
		Type: "ICMP",
		// Unfortunately, this will default to true if not set.
		DoubleCheck: false,
		// Unfortunately, this will default to true if not set.
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

// GetICMPMonitor takes the ID of an existing ICMP monitor, and returns the
// monitor parameters, or an error. It fails with an error matching
// ErrCheckTypeMismatch when the ID belongs to a check of another type.
func (c *client) GetICMPMonitor(
	ctx context.Context,
	ID string,
) (*ICMPMonitor, error) {
	var result ICMPMonitor
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s", ID), "ICMP", &result); err != nil {
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
