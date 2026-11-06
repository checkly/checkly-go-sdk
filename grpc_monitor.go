package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// GRPCMonitor calls a gRPC method, or queries a server's health status, and
// checks the result against its assertions. It is the input of
// [Client.CreateGRPCMonitor] and [Client.UpdateGRPCMonitor], and is returned
// by [Client.GetGRPCMonitor].
type GRPCMonitor struct {
	// ID is the monitor's UUID, assigned by the API.
	ID          string  `json:"id,omitempty"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	// Frequency is the number of minutes between runs: 0, 1, 2, 5, 10, 15,
	// 30, 60, 120, 180, 360, 720 or 1440.
	Frequency int `json:"frequency"`
	// FrequencyOffset is 10, 20 or 30 when Frequency is 0, for a run every
	// that many seconds. For a higher Frequency the API assigns the offset,
	// but rejects a value sent above Frequency * 10 for a Frequency up to 60,
	// or above Frequency / 5 rounded up for a higher one. FrequencyOffset is
	// left out of the request when zero.
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
	// response time limits in milliseconds, each at most 180000.
	// DegradedResponseTime and MaxResponseTime are left out of the request when
	// zero.
	DegradedResponseTime int      `json:"degradedResponseTime,omitempty"`
	MaxResponseTime      int      `json:"maxResponseTime,omitempty"`
	Tags                 []string `json:"tags,omitempty"`
	// AlertSettings is left out of the request when nil.
	AlertSettings *AlertSettings `json:"alertSettings,omitempty"`
	// UseGlobalAlertSettings applies the account's alert settings instead
	// of AlertSettings.
	UseGlobalAlertSettings bool        `json:"useGlobalAlertSettings"`
	Request                GRPCRequest `json:"request"`
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
	// locations to run from. Nil is sent as an empty list.
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

// GRPCRequest is the connection of a [GRPCMonitor]. The API requires
// GRPCConfig, so it is a value rather than a pointer.
type GRPCRequest struct {
	URL string `json:"url"`
	// Port is 1 to 65535. Its type is open because the API may return it as
	// a string; send a number or a numeric string.
	Port interface{} `json:"port"`
	// IPFamily is left out of the request when empty. An update with an
	// empty IPFamily keeps the stored family, as the API rejects an empty
	// value.
	IPFamily IPFamily `json:"ipFamily,omitempty"`
	SkipSSL  bool     `json:"skipSSL"`
	// Timeout is 1 to 180; zero means 60.
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

// GRPCConfig is the gRPC configuration in a [GRPCRequest]. Mode selects which
// of the other fields apply: the API drops the fields of the other mode.
type GRPCConfig struct {
	// An empty Mode is sent as [GRPCModeBehavior], also on update, so set
	// [GRPCModeHealth] on every write of a health monitor.
	Mode GRPCMode `json:"mode,omitempty"`
	TLS  bool     `json:"tls"`
	// Metadata holds at most 50 entries, with keys unique regardless of
	// case.
	Metadata []GRPCMetadata `json:"metadata,omitempty"`
	// The fields from Encoding to Message are for [GRPCModeBehavior] only.
	//
	// An update that omits Encoding keeps the stored encoding, so send
	// [GRPCEncodingProtobuf] explicitly to switch a [GRPCEncodingFlatbuffers]
	// monitor back; that also clears the stored BfbsContent. Encoding is left
	// out of the request when empty.
	Encoding GRPCEncoding `json:"encoding,omitempty"`
	// ServiceDefinition applies to [GRPCEncodingProtobuf] only, and
	// [GRPCServiceDefinitionProtoFile] requires ProtoContent.
	// ServiceDefinition is left out of the request when empty, so send
	// [GRPCServiceDefinitionReflection] explicitly to switch a
	// [GRPCServiceDefinitionProtoFile] monitor back.
	ServiceDefinition GRPCServiceDefinition `json:"serviceDefinition,omitempty"`
	// Method is required with [GRPCModeBehavior]. Method is left out of the
	// request when empty.
	Method string `json:"method,omitempty"`
	// ProtoContent is the proto file of [GRPCServiceDefinitionProtoFile].
	// With [GRPCModeBehavior], an empty ProtoContent is sent only with an
	// explicit [GRPCServiceDefinitionReflection] ServiceDefinition, which
	// removes the stored file; otherwise it is left out of the request.
	ProtoContent string `json:"protoContent,omitempty"`
	// BfbsContent is the base64-encoded binary FlatBuffers schema, required
	// for [GRPCEncodingFlatbuffers]. BfbsContent is left out of the request
	// when empty.
	BfbsContent string `json:"bfbsContent,omitempty"`
	Message     string `json:"message,omitempty"`
	// Service is for [GRPCModeHealth] only.
	Service string `json:"service,omitempty"`
}

// GRPCMetadata is a metadata entry in [GRPCConfig].
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
