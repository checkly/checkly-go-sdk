package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type HeartbeatMonitor struct {
	ID                        string                     `json:"id"`
	Name                      string                     `json:"name"`
	Description               *string                    `json:"description"`
	Activated                 bool                       `json:"activated"`
	Muted                     bool                       `json:"muted"`
	Tags                      []string                   `json:"tags,omitempty"`
	AlertSettings             AlertSettings              `json:"alertSettings,omitempty"`
	UseGlobalAlertSettings    bool                       `json:"useGlobalAlertSettings"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	Heartbeat                 Heartbeat                  `json:"heartbeat"`
	TriggerIncident           *IncidentTrigger           `json:"triggerIncident"`
	CreatedAt                 time.Time                  `json:"created_at"`
	UpdatedAt                 time.Time                  `json:"updated_at"`
}

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

// Heartbeat represents the parameter for the heartbeat check.
type Heartbeat struct {
	Period     int           `json:"period"`
	PeriodUnit HeartbeatUnit `json:"periodUnit"`
	Grace      int           `json:"grace"`
	GraceUnit  HeartbeatUnit `json:"graceUnit"`
	PingToken  string        `json:"pingToken"`
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

// GetHeartbeatMonitor takes the ID of an existing heartbeat monitor, and
// returns the monitor parameters, or an error. It fails with an error matching
// ErrCheckTypeMismatch when the ID belongs to a check of another type.
func (c *client) GetHeartbeatMonitor(
	ctx context.Context,
	ID string,
) (*HeartbeatMonitor, error) {
	var result HeartbeatMonitor
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s", ID), CheckTypeHeartbeat, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateHeartbeatMonitor(
	ctx context.Context,
	ID string,
	monitor HeartbeatMonitor,
) (*HeartbeatMonitor, error) {
	var result HeartbeatMonitor
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/heartbeat/%s", ID)), createHeartbeatMonitorPayload(monitor), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteHeartbeatMonitor(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}
