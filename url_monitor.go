package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// URLMonitor represents a URL monitor.
type URLMonitor struct {
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
	Request                   URLRequest                 `json:"request"`
	GroupID                   int64                      `json:"groupId,omitempty"`
	GroupOrder                int                        `json:"groupOrder,omitempty"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	PrivateLocations          *[]string                  `json:"privateLocations"`
	RetryStrategy             *RetryStrategy             `json:"retryStrategy"`
	TriggerIncident           *IncidentTrigger           `json:"triggerIncident"`
	CreatedAt                 time.Time                  `json:"created_at,omitempty"`
	UpdatedAt                 time.Time                  `json:"updated_at,omitempty"`
}

// URLRequest represents the parameters for the request made by a URL monitor.
type URLRequest struct {
	URL             string      `json:"url"`
	FollowRedirects bool        `json:"followRedirects"`
	SkipSSL         bool        `json:"skipSSL"`
	Assertions      []Assertion `json:"assertions"`
	IPFamily        string      `json:"ipFamily,omitempty"`
}

func (r *URLRequest) toRequest() Request {
	return Request{
		Method:          "GET",
		URL:             r.URL,
		FollowRedirects: r.FollowRedirects,
		SkipSSL:         r.SkipSSL,
		Assertions:      r.Assertions,
		IPFamily:        r.IPFamily,
	}
}

type urlMonitorPayload struct {
	URLMonitor
	Type        string  `json:"checkType"`
	Request     Request `json:"request"`
	DoubleCheck bool    `json:"doubleCheck"`
	GroupID     *int64  `json:"groupId"`
	// The update endpoint is PATCH-style: keys omitted from the payload keep
	// their server-side value. The fields below shadow embedded fields whose
	// zero value would be omitted so that they are always sent and can
	// actually be cleared via an update.
	Tags                      []string                   `json:"tags"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions"`
	DegradedResponseTime      int                        `json:"degradedResponseTime"`
	MaxResponseTime           int                        `json:"maxResponseTime"`
	GroupOrder                *int                       `json:"groupOrder"`
}

func createURLMonitorPayload(monitor URLMonitor) urlMonitorPayload {
	payload := urlMonitorPayload{
		URLMonitor: monitor,
		// Unfortunately `checkType` is required for this endpoint.
		Type:    "URL",
		Request: monitor.Request.toRequest(),
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

	// GroupOrder must be null if empty or the order will not get unset on
	// update. Note that the check and TCP monitor endpoints differ: they
	// accept and reset on a literal 0.
	payload.GroupOrder = nullIfZero(monitor.GroupOrder)

	return payload
}

func (c *client) CreateURLMonitor(
	ctx context.Context,
	monitor URLMonitor,
) (*URLMonitor, error) {
	var result URLMonitor
	err := c.callJSON(ctx, http.MethodPost, withAutoAssignAlertsFlag("/v1/checks/url"), createURLMonitorPayload(monitor), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateURLMonitor(
	ctx context.Context,
	ID string,
	monitor URLMonitor,
) (*URLMonitor, error) {
	var result URLMonitor
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/url/%s", ID)), createURLMonitorPayload(monitor), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteURLMonitor(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}

// GetURLMonitor takes the ID of an existing URL monitor, and returns the
// monitor parameters, or an error.
func (c *client) GetURLMonitor(
	ctx context.Context,
	ID string,
) (*URLMonitor, error) {
	var result URLMonitor
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/checks/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
