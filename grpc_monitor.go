package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// GRPCMonitor represents a gRPC monitor.
type GRPCMonitor struct {
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
	Request                   GRPCRequest                `json:"request"`
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

// GRPCRequest represents the parameters for a gRPC monitor's connection.
//
// GRPCConfig is required by the public API (the wire schema is `.required()`),
// so it is sent as a value rather than an optional pointer.
type GRPCRequest struct {
	URL string `json:"url"`
	// Port accepts a number or a `{{template}}` string, mirroring the public
	// API's `Joi.alternatives`, so its type is intentionally open.
	Port       interface{} `json:"port"`
	IPFamily   string      `json:"ipFamily,omitempty"`
	SkipSSL    bool        `json:"skipSSL"`
	Timeout    int         `json:"timeout,omitempty"`
	GRPCConfig GRPCConfig  `json:"grpcConfig"`
	Assertions []Assertion `json:"assertions,omitempty"`
}

// GRPCConfig represents the gRPC-specific configuration nested inside a gRPC
// monitor's request. `Mode` defaults to "BEHAVIOR" (which requires `Method`);
// the "HEALTH" mode forbids the BEHAVIOR-only fields, so they are tagged
// omitempty to drop them from the marshaled body when unset.
type GRPCConfig struct {
	Mode     string         `json:"mode,omitempty"`
	TLS      bool           `json:"tls"`
	Metadata []GRPCMetadata `json:"metadata,omitempty"`
	// BEHAVIOR-mode only (forbidden in HEALTH mode).
	// Encoding is PROTOBUF or FLATBUFFERS. An update that omits it keeps the
	// stored encoding, so send PROTOBUF explicitly to switch a FLATBUFFERS
	// monitor back; that also clears the stored BfbsContent.
	Encoding          string `json:"encoding,omitempty"`
	ServiceDefinition string `json:"serviceDefinition,omitempty"`
	Method            string `json:"method,omitempty"`
	ProtoContent      string `json:"protoContent,omitempty"`
	BfbsContent       string `json:"bfbsContent,omitempty"`
	Message           string `json:"message,omitempty"`
	// HEALTH-mode only (forbidden in BEHAVIOR mode).
	Service string `json:"service,omitempty"`
}

// GRPCMetadata represents a single gRPC metadata key/value pair.
type GRPCMetadata struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type grpcMonitorPayload struct {
	GRPCMonitor
	Type        string             `json:"checkType"`
	DoubleCheck bool               `json:"doubleCheck"`
	Request     grpcRequestPayload `json:"request"`
	GroupID     *int64             `json:"groupId"`
}

// grpcRequestPayload sends empty lists rather than omitting them, as the
// update endpoint keeps the stored value of an omitted field.
type grpcRequestPayload struct {
	GRPCRequest
	Assertions []Assertion       `json:"assertions"`
	GRPCConfig grpcConfigPayload `json:"grpcConfig"`
}

// grpcConfigPayload sends empty metadata as an empty list, and always sends
// the free-text field of the configured mode, even when it is empty: the
// update endpoint keeps the stored value of an omitted field, so an empty
// value would otherwise never clear the old one. The free-text field of the
// other mode stays omitted, as the API discards it anyway.
type grpcConfigPayload struct {
	GRPCConfig
	Metadata []GRPCMetadata `json:"metadata"`
	Message  *string        `json:"message,omitempty"`
	Service  *string        `json:"service,omitempty"`
}

func createGRPCConfigPayload(config GRPCConfig) grpcConfigPayload {
	payload := grpcConfigPayload{
		GRPCConfig: config,
		Metadata:   emptyIfNil(config.Metadata),
	}
	// An empty mode means BEHAVIOR, the API's default.
	if config.Mode == "HEALTH" {
		payload.Service = &config.Service
	} else {
		payload.Message = &config.Message
	}
	return payload
}

func createGRPCMonitorPayload(monitor GRPCMonitor) grpcMonitorPayload {
	payload := grpcMonitorPayload{
		GRPCMonitor: monitor,
		// Unfortunately `checkType` is required for this endpoint.
		Type: "GRPC",
		// Unfortunately, this will default to true if not set.
		DoubleCheck: false,
		Request: grpcRequestPayload{
			GRPCRequest: monitor.Request,
			Assertions:  emptyIfNil(monitor.Request.Assertions),
			GRPCConfig:  createGRPCConfigPayload(monitor.Request.GRPCConfig),
		},
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

func (c *client) CreateGRPCMonitor(
	ctx context.Context,
	monitor GRPCMonitor,
) (*GRPCMonitor, error) {
	var result GRPCMonitor
	err := c.callJSON(ctx, http.MethodPost, withAutoAssignAlertsFlag("/v1/checks/grpc"), createGRPCMonitorPayload(monitor), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetGRPCMonitor takes the ID of an existing gRPC monitor, and returns the
// monitor parameters, or an error.
func (c *client) GetGRPCMonitor(
	ctx context.Context,
	ID string,
) (*GRPCMonitor, error) {
	var result GRPCMonitor
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/checks/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateGRPCMonitor(
	ctx context.Context,
	ID string,
	monitor GRPCMonitor,
) (*GRPCMonitor, error) {
	var result GRPCMonitor
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/grpc/%s", ID)), createGRPCMonitorPayload(monitor), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteGRPCMonitor(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}
