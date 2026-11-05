package checkly

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TracerouteMonitor represents a traceroute monitor.
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
	PrivateLocations          *[]string                  `json:"privateLocations"`
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
	// The update endpoint keeps the stored value of an omitted key. The
	// fields below shadow embedded fields whose zero value would be omitted,
	// so that an update resets them: lists are sent empty, and the group
	// order as null, which is also its default on create.
	Tags                      []string                   `json:"tags"`
	GroupOrder                *int                       `json:"groupOrder"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions"`
}

// tracerouteMonitorUpdatePayload is the update request. It differs from the
// create request only in its request object, see
// tracerouteUpdateRequestPayload.
type tracerouteMonitorUpdatePayload struct {
	tracerouteMonitorPayload
	Request tracerouteUpdateRequestPayload `json:"request"`
}

// tracerouteUpdateRequestPayload is the request object of an update. The
// create request leaves unset fields out so that the API stores its
// defaults, but the update endpoint keeps the stored value of an omitted
// key, so here unset fields are sent as null, which the runner treats as its
// default, and an empty assertion list is sent as []. IPFamily and PtrLookup
// stay omitted when unset: the API cannot store a null IP family, and a null
// PTR lookup would run as false rather than the default true. The port of an
// ICMP probe is sent as null, as the API discards it anyway.
type tracerouteUpdateRequestPayload struct {
	URL            string      `json:"url"`
	Protocol       *string     `json:"protocol"`
	Port           *int        `json:"port"`
	IPFamily       string      `json:"ipFamily,omitempty"`
	MaxHops        *int        `json:"maxHops"`
	MaxUnknownHops *int        `json:"maxUnknownHops"`
	PtrLookup      *bool       `json:"ptrLookup,omitempty"`
	Timeout        *int        `json:"timeout"`
	Assertions     []Assertion `json:"assertions"`
}

func createTracerouteMonitorPayload(monitor TracerouteMonitor) tracerouteMonitorPayload {
	payload := tracerouteMonitorPayload{
		TracerouteMonitor: monitor,
		// Unfortunately `checkType` is required for this endpoint.
		Type: "TRACEROUTE",
		// Unfortunately, this will default to true if not set.
		DoubleCheck:               false,
		Tags:                      emptyIfNil(monitor.Tags),
		GroupOrder:                nullIfZero(monitor.GroupOrder),
		AlertChannelSubscriptions: emptyIfNil(monitor.AlertChannelSubscriptions),
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

	return payload
}

func updateTracerouteMonitorPayload(monitor TracerouteMonitor) tracerouteMonitorUpdatePayload {
	request := monitor.Request
	payload := tracerouteMonitorUpdatePayload{
		tracerouteMonitorPayload: createTracerouteMonitorPayload(monitor),
		Request: tracerouteUpdateRequestPayload{
			URL:            request.URL,
			Protocol:       nullIfZero(request.Protocol),
			IPFamily:       request.IPFamily,
			MaxHops:        nullIfZero(request.MaxHops),
			MaxUnknownHops: nullIfZero(request.MaxUnknownHops),
			PtrLookup:      request.PtrLookup,
			Timeout:        nullIfZero(request.Timeout),
			Assertions:     emptyIfNil(request.Assertions),
		},
	}
	if request.Protocol != "ICMP" {
		payload.Request.Port = nullIfZero(request.Port)
	}
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
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/traceroute/%s", ID)), updateTracerouteMonitorPayload(monitor), &result, http.StatusOK)
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
