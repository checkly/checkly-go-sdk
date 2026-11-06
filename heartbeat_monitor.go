package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HeartbeatMonitor expects pings at its ping URL instead of running checks
// itself: [Heartbeat] sets how often a ping is expected. It is the input of
// [Client.CreateHeartbeatMonitor] and [Client.UpdateHeartbeatMonitor], and is
// returned by [Client.GetHeartbeatMonitor].
type HeartbeatMonitor struct {
	// ID is the monitor's UUID, assigned by the API.
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Activated   bool    `json:"activated"`
	// Muted stops notifications when the monitor fails or recovers.
	Muted bool     `json:"muted"`
	Tags  []string `json:"tags,omitempty"`
	// AlertSettings with the [EscalationTypeRunBased] escalation type
	// requires RunBasedEscalation.FailedRunThreshold.
	AlertSettings AlertSettings `json:"alertSettings,omitempty"`
	// UseGlobalAlertSettings applies the account's alert settings instead
	// of AlertSettings.
	UseGlobalAlertSettings bool `json:"useGlobalAlertSettings"`
	// AlertChannelSubscriptions are the alert channels that receive the
	// monitor's alerts. The SDK disables automatic assignment, so an
	// empty list leaves the monitor without any.
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	Heartbeat                 Heartbeat                  `json:"heartbeat"`
	// TriggerIncident makes the monitor open and resolve an incident on a
	// status page service, based on its alert configuration. Nil turns it
	// off.
	TriggerIncident *IncidentTrigger `json:"triggerIncident"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

func (*HeartbeatMonitor) isCheck() {}

// HeartbeatUnit is the time unit of a heartbeat monitor's period and grace,
// in [Heartbeat].PeriodUnit and [Heartbeat].GraceUnit.
type HeartbeatUnit string

// Heartbeat units, the values of [Heartbeat].PeriodUnit and
// [Heartbeat].GraceUnit.
const (
	HeartbeatUnitSeconds HeartbeatUnit = "seconds"
	HeartbeatUnitMinutes HeartbeatUnit = "minutes"
	HeartbeatUnitHours   HeartbeatUnit = "hours"
	HeartbeatUnitDays    HeartbeatUnit = "days"
)

// Heartbeat is the ping schedule of a [HeartbeatMonitor].
type Heartbeat struct {
	// Period is the interval expected between pings, from 30 seconds to 365
	// days.
	Period     int           `json:"period"`
	PeriodUnit HeartbeatUnit `json:"periodUnit"`
	// Grace is added to the period, and must be less than 365 days.
	Grace     int           `json:"grace"`
	GraceUnit HeartbeatUnit `json:"graceUnit"`
	// PingToken is the UUID that the ping URL is built from. An update with
	// an empty PingToken keeps the stored token.
	PingToken string `json:"pingToken"`
}

type heartbeatMonitorPayload struct {
	HeartbeatMonitor
	// The update endpoint is PATCH-style: keys omitted from the payload keep
	// their server-side value. The fields below shadow embedded fields whose
	// zero value would be omitted so that they are always sent and can
	// actually be cleared via an update.
	Tags                      []string                   `json:"tags"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions"`
}

func createHeartbeatMonitorPayload(monitor HeartbeatMonitor) heartbeatMonitorPayload {
	payload := heartbeatMonitorPayload{
		HeartbeatMonitor: monitor,
	}

	payload.Tags = emptyIfNil(monitor.Tags)
	payload.AlertChannelSubscriptions = emptyIfNil(monitor.AlertChannelSubscriptions)

	return payload
}

func (c *client) CreateHeartbeatMonitor(
	ctx context.Context,
	monitor HeartbeatMonitor,
) (*HeartbeatMonitor, error) {
	var result HeartbeatMonitor
	err := c.callJSON(ctx, http.MethodPost, withAutoAssignAlertsFlag("/v1/checks/heartbeat"), createHeartbeatMonitorPayload(monitor), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetHeartbeatMonitor(
	ctx context.Context,
	id string,
) (*HeartbeatMonitor, error) {
	var result HeartbeatMonitor
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s", id), CheckTypeHeartbeat, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateHeartbeatMonitor(
	ctx context.Context,
	id string,
	monitor HeartbeatMonitor,
) (*HeartbeatMonitor, error) {
	var result HeartbeatMonitor
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/heartbeat/%s", id)), createHeartbeatMonitorPayload(monitor), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteHeartbeatMonitor(
	ctx context.Context,
	id string,
) error {
	return c.DeleteCheck(ctx, id)
}
