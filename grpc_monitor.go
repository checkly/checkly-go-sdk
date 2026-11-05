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
	IPFamily   IPFamily    `json:"ipFamily,omitempty"`
	SkipSSL    bool        `json:"skipSSL"`
	Timeout    int         `json:"timeout,omitempty"`
	GRPCConfig GRPCConfig  `json:"grpcConfig"`
	Assertions []Assertion `json:"assertions,omitempty"`
}

// GRPCMode is what a gRPC monitor checks, in [GRPCConfig].Mode.
type GRPCMode string

// gRPC modes, the values of [GRPCConfig].Mode. An empty mode is sent as
// GRPCModeBehavior, the API default.
const (
	// GRPCModeBehavior calls the method in GRPCConfig.Method.
	GRPCModeBehavior GRPCMode = "BEHAVIOR"
	// GRPCModeHealth calls the standard gRPC health check of
	// GRPCConfig.Service.
	GRPCModeHealth GRPCMode = "HEALTH"
)

// GRPCEncoding is the message encoding of a gRPC monitor, in
// [GRPCConfig].Encoding.
type GRPCEncoding string

// gRPC encodings, the values of [GRPCConfig].Encoding. The API default is
// GRPCEncodingProtobuf.
const (
	GRPCEncodingProtobuf    GRPCEncoding = "PROTOBUF"
	GRPCEncodingFlatbuffers GRPCEncoding = "FLATBUFFERS"
)

// GRPCServiceDefinition is where a gRPC monitor gets the service definition
// from, in [GRPCConfig].ServiceDefinition.
type GRPCServiceDefinition string

// gRPC service definitions, the values of [GRPCConfig].ServiceDefinition.
// The API default is GRPCServiceDefinitionReflection.
const (
	// GRPCServiceDefinitionReflection uses server reflection.
	GRPCServiceDefinitionReflection GRPCServiceDefinition = "REFLECTION"
	// GRPCServiceDefinitionProtoFile uses GRPCConfig.ProtoContent.
	GRPCServiceDefinitionProtoFile GRPCServiceDefinition = "PROTO_FILE"
)

// GRPCConfig represents the gRPC-specific configuration nested inside a gRPC
// monitor's request. `Mode` defaults to BEHAVIOR (which requires `Method`);
// the HEALTH mode forbids the BEHAVIOR-only fields, so they are tagged
// omitempty to drop them from the marshaled body when unset.
type GRPCConfig struct {
	Mode     GRPCMode       `json:"mode,omitempty"`
	TLS      bool           `json:"tls"`
	Metadata []GRPCMetadata `json:"metadata,omitempty"`
	// BEHAVIOR-mode only (forbidden in HEALTH mode).
	// Encoding is PROTOBUF or FLATBUFFERS. An update that omits it keeps the
	// stored encoding, so send PROTOBUF explicitly to switch a FLATBUFFERS
	// monitor back; that also clears the stored BfbsContent.
	Encoding          GRPCEncoding          `json:"encoding,omitempty"`
	ServiceDefinition GRPCServiceDefinition `json:"serviceDefinition,omitempty"`
	Method            string                `json:"method,omitempty"`
	ProtoContent      string                `json:"protoContent,omitempty"`
	BfbsContent       string                `json:"bfbsContent,omitempty"`
	Message           string                `json:"message,omitempty"`
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
	Type        CheckType          `json:"checkType"`
	DoubleCheck bool               `json:"doubleCheck"`
	Request     grpcRequestPayload `json:"request"`
	GroupID     *int64             `json:"groupId"`
	// The update endpoint keeps the stored value of an omitted key. The
	// fields below shadow embedded fields whose zero value would be omitted,
	// so that an update resets them: lists are sent empty, and the group
	// order as null, which is also its default on create.
	Tags                      []string                   `json:"tags"`
	GroupOrder                *int                       `json:"groupOrder"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions"`
}

// grpcMonitorUpdatePayload is the update request. It differs from the create
// request only in its request object, see grpcUpdateRequestPayload.
type grpcMonitorUpdatePayload struct {
	grpcMonitorPayload
	Request grpcUpdateRequestPayload `json:"request"`
}

// grpcRequestPayload sends empty lists rather than omitting them, as the
// update endpoint keeps the stored value of an omitted field.
type grpcRequestPayload struct {
	GRPCRequest
	Assertions []Assertion       `json:"assertions"`
	GRPCConfig grpcConfigPayload `json:"grpcConfig"`
}

// grpcUpdateRequestPayload sends an unset timeout as null, which the update
// endpoint accepts and the runner treats as its default timeout. The create
// request leaves the key out instead, so that the API stores its default.
type grpcUpdateRequestPayload struct {
	grpcRequestPayload
	Timeout *int `json:"timeout"`
}

// grpcConfigPayload sends empty metadata as an empty list, and always sends
// the free-text field of the configured mode, even when it is empty: the
// update endpoint keeps the stored value of an omitted field, so an empty
// value would otherwise never clear the old one. The free-text field of the
// other mode stays omitted, as the API discards it anyway.
type grpcConfigPayload struct {
	GRPCConfig
	Metadata     []GRPCMetadata `json:"metadata"`
	Message      *string        `json:"message,omitempty"`
	Service      *string        `json:"service,omitempty"`
	ProtoContent *string        `json:"protoContent,omitempty"`
}

func createGRPCConfigPayload(config GRPCConfig) grpcConfigPayload {
	payload := grpcConfigPayload{
		GRPCConfig:   config,
		Metadata:     emptyIfNil(config.Metadata),
		ProtoContent: nullIfZero(config.ProtoContent),
	}
	// An empty mode means BEHAVIOR, the API's default. The mode is always
	// sent: an update without it makes the API discard every BEHAVIOR-mode
	// field of the request.
	if config.Mode == "" {
		payload.Mode = GRPCModeBehavior
	}
	if payload.Mode == GRPCModeHealth {
		payload.Service = &config.Service
	} else {
		payload.Message = &config.Message
		// An empty proto content is sent only with an explicit REFLECTION
		// service definition. The API validates the content against the
		// definition in the request alone, treating an omitted one as
		// REFLECTION, so an empty content without it would leave a stored
		// PROTO_FILE definition with an empty proto file.
		if config.ServiceDefinition == GRPCServiceDefinitionReflection {
			payload.ProtoContent = &config.ProtoContent
		}
	}
	return payload
}

func createGRPCMonitorPayload(monitor GRPCMonitor) grpcMonitorPayload {
	payload := grpcMonitorPayload{
		GRPCMonitor: monitor,
		// Unfortunately `checkType` is required for this endpoint.
		Type: CheckTypeGRPC,
		// Unfortunately, this will default to true if not set.
		DoubleCheck: false,
		Request: grpcRequestPayload{
			GRPCRequest: monitor.Request,
			Assertions:  emptyIfNil(monitor.Request.Assertions),
			GRPCConfig:  createGRPCConfigPayload(monitor.Request.GRPCConfig),
		},
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

func updateGRPCMonitorPayload(monitor GRPCMonitor) grpcMonitorUpdatePayload {
	payload := createGRPCMonitorPayload(monitor)
	return grpcMonitorUpdatePayload{
		grpcMonitorPayload: payload,
		Request: grpcUpdateRequestPayload{
			grpcRequestPayload: payload.Request,
			Timeout:            nullIfZero(monitor.Request.Timeout),
		},
	}
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

func (c *client) GetGRPCMonitor(
	ctx context.Context,
	ID string,
) (*GRPCMonitor, error) {
	var result GRPCMonitor
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s", ID), CheckTypeGRPC, &result); err != nil {
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
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/grpc/%s", ID)), updateGRPCMonitorPayload(monitor), &result, http.StatusOK)
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
