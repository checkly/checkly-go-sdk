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

// Heartbeat represents the parameter for the heartbeat check.
type Heartbeat struct {
	Period     int    `json:"period"`
	PeriodUnit string `json:"periodUnit"`
	Grace      int    `json:"grace"`
	GraceUnit  string `json:"graceUnit"`
	PingToken  string `json:"pingToken"`
}

type heartbeatMonitorPayload struct {
	HeartbeatMonitor
	// The update endpoint is PATCH-style: keys omitted from the payload keep
	// their server-side value. The field below shadows an embedded field
	// whose zero value would be omitted so that it is always sent and can
	// actually be cleared via an update. AlertChannelSubscriptions is not
	// shadowed: this endpoint verifiably clears subscriptions when the key
	// is omitted.
	Tags []string `json:"tags"`
}

func createHeartbeatMonitorPayload(monitor HeartbeatMonitor) heartbeatMonitorPayload {
	payload := heartbeatMonitorPayload{
		HeartbeatMonitor: monitor,
	}

	payload.Tags = emptyIfNil(monitor.Tags)

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
	ID string,
) (*HeartbeatMonitor, error) {
	result := HeartbeatMonitor{}
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/checks/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
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
