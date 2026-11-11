package checkly

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TracerouteMonitor traces the network path to a host and checks the response
// time, hop count and packet loss against its assertions. It is the input of
// [Client.CreateTracerouteMonitor] and [Client.UpdateTracerouteMonitor], and
// is returned by [Client.GetTracerouteMonitor].
type TracerouteMonitor struct {
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
	// response time limits in milliseconds, each at most 30000; the API
	// defaults them to 10000 and 20000. DegradedResponseTime and
	// MaxResponseTime are left out of the request when zero.
	DegradedResponseTime int      `json:"degradedResponseTime,omitempty"`
	MaxResponseTime      int      `json:"maxResponseTime,omitempty"`
	Tags                 []string `json:"tags,omitempty"`
	// AlertSettings is left out of the request when nil.
	AlertSettings *AlertSettings `json:"alertSettings,omitempty"`
	// UseGlobalAlertSettings applies the account's alert settings instead
	// of AlertSettings.
	UseGlobalAlertSettings bool              `json:"useGlobalAlertSettings"`
	Request                TracerouteRequest `json:"request"`
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
	// at those locations must be version 8.4.0 or later and have the NET_RAW
	// capability.
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

func (*TracerouteMonitor) isCheck() {}

// TracerouteProtocol is the protocol of a traceroute monitor's probes, in
// [TracerouteRequest].Protocol.
type TracerouteProtocol string

// Traceroute protocols, the values of [TracerouteRequest].Protocol. The API
// default is TracerouteProtocolTCP.
const (
	TracerouteProtocolTCP  TracerouteProtocol = "TCP"
	TracerouteProtocolICMP TracerouteProtocol = "ICMP"
	TracerouteProtocolUDP  TracerouteProtocol = "UDP"
	TracerouteProtocolSCTP TracerouteProtocol = "SCTP"
)

// TracerouteRequest is the probe of a [TracerouteMonitor].
type TracerouteRequest struct {
	URL string `json:"url"`
	// An empty Protocol means [TracerouteProtocolTCP].
	Protocol TracerouteProtocol `json:"protocol,omitempty"`
	// Port is 1 to 65535; zero means 443 for TCP and 33434 for UDP and SCTP.
	// It is ignored when Protocol is [TracerouteProtocolICMP].
	Port int `json:"port,omitempty"`
	// IPFamily is left out of the request when empty. An update with an
	// empty IPFamily keeps the stored family, as the API rejects an empty
	// value.
	IPFamily IPFamily `json:"ipFamily,omitempty"`
	// MaxHops is 1 to 64; zero means 30.
	MaxHops int `json:"maxHops,omitempty"`
	// MaxUnknownHops is 1 to 30 and at most MaxHops; zero means 15, or
	// MaxHops if that is lower.
	MaxUnknownHops int `json:"maxUnknownHops,omitempty"`
	// PtrLookup is a pointer so that an explicit false can be sent; the API
	// defaults it to true. PtrLookup is left out of the request when nil.
	PtrLookup *bool `json:"ptrLookup,omitempty"`
	// Timeout is 1 to 30; zero means 10.
	Timeout    int         `json:"timeout,omitempty"`
	Assertions []Assertion `json:"assertions,omitempty"`
}

// MarshalJSON leaves out Port when Protocol is [TracerouteProtocolICMP], as
// the API ignores the port of an ICMP traceroute.
func (r TracerouteRequest) MarshalJSON() ([]byte, error) {
	type alias TracerouteRequest
	if r.Protocol == TracerouteProtocolICMP {
		// Zero value combined with the `omitempty` tag drops `port` from the
		// marshaled body, matching the API, which ignores it for ICMP.
		r.Port = 0
	}
	return json.Marshal(alias(r))
}

type tracerouteMonitorPayload struct {
	TracerouteMonitor
	Type        CheckType `json:"checkType"`
	DoubleCheck bool      `json:"doubleCheck"`
	GroupID     *int64    `json:"groupId"`
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
	URL            string              `json:"url"`
	Protocol       *TracerouteProtocol `json:"protocol"`
	Port           *int                `json:"port"`
	IPFamily       IPFamily            `json:"ipFamily,omitempty"`
	MaxHops        *int                `json:"maxHops"`
	MaxUnknownHops *int                `json:"maxUnknownHops"`
	PtrLookup      *bool               `json:"ptrLookup,omitempty"`
	Timeout        *int                `json:"timeout"`
	Assertions     []Assertion         `json:"assertions"`
}

func createTracerouteMonitorPayload(monitor TracerouteMonitor) tracerouteMonitorPayload {
	payload := tracerouteMonitorPayload{
		TracerouteMonitor: monitor,
		// Unfortunately `checkType` is required for this endpoint.
		Type: CheckTypeTraceroute,
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
	if request.Protocol != TracerouteProtocolICMP {
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

func (c *client) GetTracerouteMonitor(
	ctx context.Context,
	id string,
) (*TracerouteMonitor, error) {
	var result TracerouteMonitor
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s", id), CheckTypeTraceroute, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateTracerouteMonitor(
	ctx context.Context,
	id string,
	monitor TracerouteMonitor,
) (*TracerouteMonitor, error) {
	var result TracerouteMonitor
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/traceroute/%s", id)), updateTracerouteMonitorPayload(monitor), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteTracerouteMonitor(
	ctx context.Context,
	id string,
) error {
	return c.DeleteCheck(ctx, id)
}
