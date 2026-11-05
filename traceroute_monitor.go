package checkly

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TracerouteMonitor represents a traceroute monitor.
//
// Unlike the other monitor types it has no PrivateLocations field: the public
// create schema marks `privateLocations` as forbidden ("Traceroute monitors do
// not support private locations"), so the field must never reach the wire.
type TracerouteMonitor struct {
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
	Request                   TracerouteRequest          `json:"request"`
	GroupID                   int64                      `json:"groupId,omitempty"`
	GroupOrder                int                        `json:"groupOrder,omitempty"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	RuntimeID                 *string                    `json:"runtimeId"`
	RetryStrategy             *RetryStrategy             `json:"retryStrategy"`
	TriggerIncident           *IncidentTrigger           `json:"triggerIncident"`
	CreatedAt                 time.Time                  `json:"created_at,omitempty"`
	UpdatedAt                 time.Time                  `json:"updated_at,omitempty"`
}

// TracerouteRequest represents the parameters for a traceroute monitor's probe.
//
// Port is stripped from the marshaled body when Protocol is "ICMP", mirroring
// the public API's `Joi.any().strip()` for that case (see MarshalJSON).
type TracerouteRequest struct {
	URL            string `json:"url"`
	Protocol       string `json:"protocol,omitempty"`
	Port           int    `json:"port,omitempty"`
	IPFamily       string `json:"ipFamily,omitempty"`
	MaxHops        int    `json:"maxHops,omitempty"`
	MaxUnknownHops int    `json:"maxUnknownHops,omitempty"`
	// PtrLookup is a pointer so that an explicit `false` can be sent; the API
	// defaults it to true when the field is omitted.
	PtrLookup  *bool       `json:"ptrLookup,omitempty"`
	Timeout    int         `json:"timeout,omitempty"`
	Assertions []Assertion `json:"assertions,omitempty"`
}

// MarshalJSON drops the `port` field when the probe protocol is "ICMP",
// mirroring the public API schema which strips `port` for ICMP traceroutes.
func (r TracerouteRequest) MarshalJSON() ([]byte, error) {
	type alias TracerouteRequest
	if r.Protocol == "ICMP" {
		// Zero value combined with the `omitempty` tag drops `port` from the
		// marshaled body, matching the Joi `.strip()` behavior.
		r.Port = 0
	}
	return json.Marshal(alias(r))
}

type tracerouteMonitorPayload struct {
	TracerouteMonitor
	Type        string `json:"checkType"`
	DoubleCheck bool   `json:"doubleCheck"`
	GroupID     *int64 `json:"groupId"`
}

func createTracerouteMonitorPayload(monitor TracerouteMonitor) tracerouteMonitorPayload {
	payload := tracerouteMonitorPayload{
		TracerouteMonitor: monitor,
		// Unfortunately `checkType` is required for this endpoint.
		Type: "TRACEROUTE",
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

	// NOTE: traceroute monitors forbid private locations (the create schema
	// marks `privateLocations` as forbidden), so — unlike the other monitor
	// payloads — we must not emit that field at all. TracerouteMonitor has no
	// PrivateLocations field, so nothing to set here.

	return payload
}

func (c *client) CreateTracerouteMonitor(
	ctx context.Context,
	monitor TracerouteMonitor,
) (*TracerouteMonitor, error) {
	var result TracerouteMonitor
	err := c.callJSON(ctx, http.MethodPost, withAutoAssignAlertsFlag("/v1/checks/traceroute"), createTracerouteMonitorPayload(monitor), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTracerouteMonitor takes the ID of an existing traceroute monitor, and
// returns the monitor parameters, or an error. It fails with an error matching
// ErrCheckTypeMismatch when the ID belongs to a check of another type.
func (c *client) GetTracerouteMonitor(
	ctx context.Context,
	ID string,
) (*TracerouteMonitor, error) {
	var result TracerouteMonitor
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s", ID), "TRACEROUTE", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateTracerouteMonitor(
	ctx context.Context,
	ID string,
	monitor TracerouteMonitor,
) (*TracerouteMonitor, error) {
	var result TracerouteMonitor
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/traceroute/%s", ID)), createTracerouteMonitorPayload(monitor), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteTracerouteMonitor(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}
