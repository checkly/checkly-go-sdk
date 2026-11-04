package checkly

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/netip"
	"net/url"
	"os"
	"strings"
	"time"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// NewClient constructs a Checkly API client.
func NewClient(
	//checkly API's base url
	baseURL,
	//checkly's api key
	apiKey string,
	//optional, defaults to http.DefaultClient
	httpClient *http.Client,
	debug io.Writer,
) Client {
	c := &client{
		apiKey:     apiKey,
		url:        baseURL,
		httpClient: httpClient,
		debug:      debug,
		sleep:      sleepContext,
	}
	if httpClient != nil {
		c.httpClient = httpClient
	} else {
		c.httpClient = http.DefaultClient
	}
	return c
}

// SetAccountId sets ID on a client which is required when using User API keys.
func (c *client) SetAccountId(ID string) {
	c.accountId = ID
}

// SetChecklySource sets the source on a client which is required for analytics.
func (c *client) SetChecklySource(source string) {
	c.source = source
}

// Create creates a new check with the specified details. It returns the
// newly-created check, or an error.
//
// Deprecated: this method would be removed in future versions,
// use CreateCheck instead.
func (c *client) Create(
	ctx context.Context,
	check Check,
) (*Check, error) {
	// There are differences between /v1/checks and /v1/checks/<type>. Keep
	// using /v1/checks here for backwards compatibility reasons.
	return c.createCheck(ctx, check, "checks")
}

// Update updates an existing check with the specified details. It returns the
// updated check, or an error.
//
// Deprecated: this method would be removed in future versions,
// use UpdateCheck instead.
func (c *client) Update(
	ctx context.Context,
	ID string, check Check,
) (*Check, error) {
	return c.UpdateCheck(ctx, ID, check)
}

// Delete deletes the check with the specified ID.
//
// Deprecated: this method would be removed in future versions,
// use DeleteCheck instead.
func (c *client) Delete(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}

// Get takes the ID of an existing check, and returns the check parameters, or
// an error.
//
// Deprecated: this method would be removed in future versions,
// use GetCheck instead.
func (c *client) Get(
	ctx context.Context,
	ID string,
) (*Check, error) {
	var result Check
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("checks/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new check with the specified details. It returns the
// newly-created check, or an error.
func (c *client) CreateCheck(
	ctx context.Context,
	check Check,
) (*Check, error) {
	var endpoint string
	switch check.Type {
	case "BROWSER":
		endpoint = "checks/browser"
	case "API":
		endpoint = "checks/api"
	case "HEARTBEAT":
		endpoint = "checks/heartbeat"
	case "MULTI_STEP":
		endpoint = "checks/multistep"
	case "TCP":
		return nil, fmt.Errorf("user error: use CreateTCPMonitor to create TCP monitors")
	case "URL":
		return nil, fmt.Errorf("user error: use CreateURLMonitor to create URL monitors")
	case "DNS":
		return nil, fmt.Errorf("user error: use CreateDNSMonitor to create DNS monitors")
	case "ICMP":
		return nil, fmt.Errorf("user error: use CreateICMPMonitor to create ICMP monitors")
	case "GRPC":
		return nil, fmt.Errorf("user error: use CreateGRPCMonitor to create GRPC monitors")
	case "TRACEROUTE":
		return nil, fmt.Errorf("user error: use CreateTracerouteMonitor to create TRACEROUTE monitors")
	case "SSL":
		return nil, fmt.Errorf("user error: use CreateSSLMonitor to create SSL monitors")
	default:
		return nil, fmt.Errorf("unknown check type: %s", check.Type)
	}
	return c.createCheck(ctx, check, endpoint)
}

type checkPayload struct {
	Check
	GroupID *int64 `json:"groupId"`
}

func createCheckPayload(check Check) checkPayload {
	payload := checkPayload{
		Check: check,
	}

	// GroupID must be null if empty or the group will not get unset on update.
	if check.GroupID != 0 {
		payload.GroupID = &check.GroupID
	}

	// A nil value for a list will cause the backend to not update the value.
	// We must send empty lists instead.
	if check.Locations == nil {
		payload.Locations = []string{}
	}

	if check.PrivateLocations == nil {
		payload.PrivateLocations = &[]string{}
	}

	return payload
}

func (c *client) createCheck(
	ctx context.Context,
	check Check,
	endpoint string,
) (*Check, error) {
	var result Check
	err := c.callJSON(ctx, 1, http.MethodPost, withAutoAssignAlertsFlag(endpoint), createCheckPayload(check), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Deprecated: Use CreateHeartbeatMonitor instead.
func (c *client) CreateHeartbeat(
	ctx context.Context,
	check HeartbeatCheck,
) (*HeartbeatCheck, error) {
	return c.CreateHeartbeatMonitor(ctx, check)
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
	err := c.callJSON(ctx, 1, http.MethodPost, withAutoAssignAlertsFlag("checks/heartbeat"), createHeartbeatMonitorPayload(monitor), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Deprecated: Use CreateTCPMonitor instead.
func (c *client) CreateTCPCheck(
	ctx context.Context,
	check TCPCheck,
) (*TCPCheck, error) {
	return c.CreateTCPMonitor(ctx, check)
}

// tcpRequestPayload wraps TCPRequest for create and update payloads. The
// endpoints are PATCH-style down to the nested request object: request keys
// omitted from the payload keep their server-side value. Assertions shadows
// the embedded field, whose omitempty would drop the key for an empty list,
// so that removing all assertions reliably clears them. Data is not shadowed
// because the backend rejects an empty string ("request.data is not allowed
// to be empty") and silently keeps the previous value on null, so a data
// value can never be cleared, only changed. IPFamily is not shadowed because
// an empty string is not a valid value for it either; it can only be changed
// to another valid value.
type tcpRequestPayload struct {
	TCPRequest
	Assertions []Assertion `json:"assertions"`
}

type tcpMonitorPayload struct {
	TCPMonitor
	Type        string `json:"checkType"`
	DoubleCheck bool   `json:"doubleCheck"`
	GroupID     *int64 `json:"groupId"`
	// The update endpoint is PATCH-style: keys omitted from the payload keep
	// their server-side value. The fields below shadow embedded fields whose
	// zero value would be omitted so that they are always sent and can
	// actually be cleared via an update. AlertChannelSubscriptions is not
	// shadowed: this endpoint verifiably clears subscriptions when the key
	// is omitted.
	Tags                 []string          `json:"tags"`
	DegradedResponseTime int               `json:"degradedResponseTime"`
	MaxResponseTime      int               `json:"maxResponseTime"`
	GroupOrder           int               `json:"groupOrder"`
	Request              tcpRequestPayload `json:"request"`
}

func createTCPMonitorPayload(monitor TCPMonitor) tcpMonitorPayload {
	payload := tcpMonitorPayload{
		TCPMonitor: monitor,
		// Unfortunately `checkType` is required for this endpoint.
		Type: "TCP",
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

	// The response time limits must always be sent: 0 is a valid value, and
	// omitting the keys would make the backend default them on create and
	// keep the previous values on update.
	payload.DegradedResponseTime = monitor.DegradedResponseTime
	payload.MaxResponseTime = monitor.MaxResponseTime

	// The group order must always be sent: 0 is the value it resets to, and
	// omitting the key would make the backend keep the previous value on
	// update. Note that the URL, DNS and ICMP monitor endpoints differ:
	// they only reset the order on an explicit null, not 0.
	payload.GroupOrder = monitor.GroupOrder

	payload.Request = tcpRequestPayload{
		TCPRequest: monitor.Request,
		Assertions: emptyIfNil(monitor.Request.Assertions),
	}

	return payload
}

func (c *client) CreateTCPMonitor(
	ctx context.Context,
	monitor TCPMonitor,
) (*TCPMonitor, error) {
	var result TCPMonitor
	err := c.callJSON(ctx, 1, http.MethodPost, withAutoAssignAlertsFlag("checks/tcp"), createTCPMonitorPayload(monitor), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
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
	err := c.callJSON(ctx, 1, http.MethodPost, withAutoAssignAlertsFlag("checks/grpc"), createGRPCMonitorPayload(monitor), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
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
	err := c.callJSON(ctx, 1, http.MethodPost, withAutoAssignAlertsFlag("checks/traceroute"), createTracerouteMonitorPayload(monitor), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type sslMonitorPayload struct {
	SSLMonitor
	Type        string `json:"checkType"`
	DoubleCheck bool   `json:"doubleCheck"`
	GroupID     *int64 `json:"groupId"`
}

func createSSLMonitorPayload(monitor SSLMonitor) sslMonitorPayload {
	payload := sslMonitorPayload{
		SSLMonitor: monitor,
		// Unfortunately `checkType` is required for this endpoint.
		Type: "SSL",
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

	return payload
}

func (c *client) CreateSSLMonitor(
	ctx context.Context,
	monitor SSLMonitor,
) (*SSLMonitor, error) {
	var result SSLMonitor
	err := c.callJSON(ctx, 1, http.MethodPost, withAutoAssignAlertsFlag("checks/ssl"), createSSLMonitorPayload(monitor), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
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
	err := c.callJSON(ctx, 1, http.MethodPost, withAutoAssignAlertsFlag("checks/url"), createURLMonitorPayload(monitor), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type dnsMonitorPayload struct {
	DNSMonitor
	Type        string     `json:"checkType"`
	DoubleCheck bool       `json:"doubleCheck"`
	ShouldFail  bool       `json:"shouldFail"`
	GroupID     *int64     `json:"groupId"`
	ID          *string    `json:"id,omitempty"`         // Skip, can't be changed.
	CreatedAt   *time.Time `json:"created_at,omitempty"` // Skip, can't be changed.
	UpdatedAt   *time.Time `json:"updated_at,omitempty"` // Skip, can't be changed.
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

func createDNSMonitorPayload(monitor DNSMonitor) dnsMonitorPayload {
	payload := dnsMonitorPayload{
		DNSMonitor: monitor,
		// Unfortunately `checkType` is required for this endpoint.
		Type: "DNS",
		// Unfortunately, this will default to true if not set.
		DoubleCheck: false,
		// Unfortunately, this will default to true if not set.
		ShouldFail: false,
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

func (c *client) CreateDNSMonitor(
	ctx context.Context,
	monitor DNSMonitor,
) (*DNSMonitor, error) {
	var result DNSMonitor
	err := c.callJSON(ctx, 1, http.MethodPost, withAutoAssignAlertsFlag("checks/dns"), createDNSMonitorPayload(monitor), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type icmpMonitorPayload struct {
	ICMPMonitor
	Type        string     `json:"checkType"`
	DoubleCheck bool       `json:"doubleCheck"`
	ShouldFail  bool       `json:"shouldFail"`
	GroupID     *int64     `json:"groupId"`
	ID          *string    `json:"id,omitempty"`         // Skip, can't be changed.
	CreatedAt   *time.Time `json:"created_at,omitempty"` // Skip, can't be changed.
	UpdatedAt   *time.Time `json:"updated_at,omitempty"` // Skip, can't be changed.
	// The update endpoint is PATCH-style: keys omitted from the payload keep
	// their server-side value. The fields below shadow embedded fields whose
	// zero value would be omitted so that they are always sent and can
	// actually be cleared via an update.
	Tags                        []string                   `json:"tags"`
	AlertChannelSubscriptions   []AlertChannelSubscription `json:"alertChannelSubscriptions"`
	DegradedPacketLossThreshold int                        `json:"degradedPacketLossThreshold"`
	MaxPacketLossThreshold      int                        `json:"maxPacketLossThreshold"`
	GroupOrder                  *int                       `json:"groupOrder"`
}

func createICMPMonitorPayload(monitor ICMPMonitor) icmpMonitorPayload {
	payload := icmpMonitorPayload{
		ICMPMonitor: monitor,
		// Unfortunately `checkType` is required for this endpoint.
		Type: "ICMP",
		// Unfortunately, this will default to true if not set.
		DoubleCheck: false,
		// Unfortunately, this will default to true if not set.
		ShouldFail: false,
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

	payload.Tags = emptyIfNil(monitor.Tags)

	payload.AlertChannelSubscriptions = emptyIfNil(monitor.AlertChannelSubscriptions)

	// The packet loss thresholds must always be sent: omitting the keys
	// would make the backend default them on create and keep the previous
	// values on update. The backend rejects an explicit 0, and sending it
	// surfaces that as a validation error rather than silently storing a
	// different value.
	payload.DegradedPacketLossThreshold = monitor.DegradedPacketLossThreshold
	payload.MaxPacketLossThreshold = monitor.MaxPacketLossThreshold

	// GroupOrder must be null if empty or the order will not get unset on
	// update. Note that the check and TCP monitor endpoints differ: they
	// accept and reset on a literal 0.
	payload.GroupOrder = nullIfZero(monitor.GroupOrder)

	return payload
}

func (c *client) CreateICMPMonitor(
	ctx context.Context,
	monitor ICMPMonitor,
) (*ICMPMonitor, error) {
	var result ICMPMonitor
	err := c.callJSON(ctx, 1, http.MethodPost, withAutoAssignAlertsFlag("checks/icmp"), createICMPMonitorPayload(monitor), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type playwrightCheckPayload struct {
	PlaywrightCheck
	Type      string     `json:"checkType"`
	ID        *string    `json:"id,omitempty"`         // Skip, can't be changed.
	CreatedAt *time.Time `json:"created_at,omitempty"` // Skip, can't be changed.
	UpdatedAt *time.Time `json:"updated_at,omitempty"` // Skip, can't be changed.
}

func createPlaywrightCheckPayload(check PlaywrightCheck) playwrightCheckPayload {
	payload := playwrightCheckPayload{
		PlaywrightCheck: check,
		// Unfortunately `checkType` is required for this endpoint.
		Type: "PLAYWRIGHT",
	}

	if payload.Browsers == nil {
		payload.Browsers = []string{}
	}

	return payload
}

func (c *client) CreatePlaywrightCheck(
	ctx context.Context,
	check PlaywrightCheck,
) (*PlaywrightCheck, error) {
	var result PlaywrightCheck
	err := c.callJSON(ctx, 1, http.MethodPost, withAutoAssignAlertsFlag("checks/playwright"), createPlaywrightCheckPayload(check), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type apiCheckPayload struct {
	APICheck
	Type string `json:"checkType"`
	// The update endpoint is PATCH-style: keys omitted from the payload keep
	// their server-side value. The fields below shadow embedded fields whose
	// Go type cannot express the value that clears them.
	DoubleCheck       bool   `json:"doubleCheck"`
	GroupID           *int64 `json:"groupId"`
	SetupSnippetID    *int64 `json:"setupSnippetId"`
	TearDownSnippetID *int64 `json:"tearDownSnippetId"`
}

func createAPICheckPayload(check APICheck) apiCheckPayload {
	payload := apiCheckPayload{
		APICheck: check,
		// The endpoint defaults `checkType` correctly, but be explicit anyway.
		Type: "API",
		// Unfortunately, this will default to true if not set. Always sending
		// false also clears the deprecated doubleCheck flag on updates of
		// checks that still have it set; retryStrategy is its replacement.
		DoubleCheck:       false,
		GroupID:           nullIfZero(check.GroupID),
		SetupSnippetID:    nullIfZero(check.SetupSnippetID),
		TearDownSnippetID: nullIfZero(check.TearDownSnippetID),
	}

	// A nil value for a list will cause the backend to not update the value.
	// We must send empty lists instead.
	payload.Locations = emptyIfNil(check.Locations)
	if check.PrivateLocations == nil {
		payload.PrivateLocations = &[]string{}
	}
	payload.Tags = emptyIfNil(check.Tags)
	payload.AlertChannelSubscriptions = emptyIfNil(check.AlertChannelSubscriptions)

	return payload
}

func (c *client) CreateAPICheck(
	ctx context.Context,
	check APICheck,
) (*APICheck, error) {
	var result APICheck
	err := c.callJSON(ctx, 1, http.MethodPost, withAutoAssignAlertsFlag("checks/api"), createAPICheckPayload(check), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type browserCheckPayload struct {
	BrowserCheck
	Type string `json:"checkType"`
	// The update endpoint is PATCH-style: keys omitted from the payload keep
	// their server-side value. The fields below shadow embedded fields whose
	// Go type cannot express the value that clears them.
	DoubleCheck bool   `json:"doubleCheck"`
	GroupID     *int64 `json:"groupId"`
	// Script is null rather than empty when unset: the API rejects an empty
	// script, and a check that runs ScriptPath has no inline script.
	Script *string `json:"script"`
}

func createBrowserCheckPayload(check BrowserCheck) browserCheckPayload {
	payload := browserCheckPayload{
		BrowserCheck: check,
		// The endpoint defaults `checkType` correctly, but be explicit anyway.
		Type: "BROWSER",
		// Unfortunately, this will default to true if not set. Always sending
		// false also clears the deprecated doubleCheck flag on updates of
		// checks that still have it set; retryStrategy is its replacement.
		DoubleCheck: false,
		GroupID:     nullIfZero(check.GroupID),
		Script:      nullIfZero(check.Script),
	}

	// A nil value for a list will cause the backend to not update the value.
	// We must send empty lists instead.
	payload.Locations = emptyIfNil(check.Locations)
	if check.PrivateLocations == nil {
		payload.PrivateLocations = &[]string{}
	}
	payload.Tags = emptyIfNil(check.Tags)
	payload.AlertChannelSubscriptions = emptyIfNil(check.AlertChannelSubscriptions)
	payload.Dependencies = emptyIfNil(check.Dependencies)

	return payload
}

func (c *client) CreateBrowserCheck(
	ctx context.Context,
	check BrowserCheck,
) (*BrowserCheck, error) {
	var result BrowserCheck
	err := c.callJSON(ctx, 1, http.MethodPost, withAutoAssignAlertsFlag("checks/browser"), createBrowserCheckPayload(check), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type multiStepCheckPayload struct {
	MultiStepCheck
	Type string `json:"checkType"`
	// The update endpoint is PATCH-style: keys omitted from the payload keep
	// their server-side value. The fields below shadow embedded fields whose
	// Go type cannot express the value that clears them.
	DoubleCheck bool   `json:"doubleCheck"`
	GroupID     *int64 `json:"groupId"`
	// Script is null rather than empty when unset: the API rejects an empty
	// script, and a check that runs ScriptPath has no inline script.
	Script *string `json:"script"`
}

func createMultiStepCheckPayload(check MultiStepCheck) multiStepCheckPayload {
	payload := multiStepCheckPayload{
		MultiStepCheck: check,
		// The endpoint defaults `checkType` correctly, but be explicit anyway.
		Type: "MULTI_STEP",
		// Unfortunately, this will default to true if not set. Always sending
		// false also clears the deprecated doubleCheck flag on updates of
		// checks that still have it set; retryStrategy is its replacement.
		DoubleCheck: false,
		GroupID:     nullIfZero(check.GroupID),
		Script:      nullIfZero(check.Script),
	}

	// A nil value for a list will cause the backend to not update the value.
	// We must send empty lists instead.
	payload.Locations = emptyIfNil(check.Locations)
	if check.PrivateLocations == nil {
		payload.PrivateLocations = &[]string{}
	}
	payload.Tags = emptyIfNil(check.Tags)
	payload.AlertChannelSubscriptions = emptyIfNil(check.AlertChannelSubscriptions)
	payload.Dependencies = emptyIfNil(check.Dependencies)

	return payload
}

func (c *client) CreateMultiStepCheck(
	ctx context.Context,
	check MultiStepCheck,
) (*MultiStepCheck, error) {
	var result MultiStepCheck
	err := c.callJSON(ctx, 1, http.MethodPost, withAutoAssignAlertsFlag("checks/multistep"), createMultiStepCheckPayload(check), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an existing check with the specified details. It returns the
// updated check, or an error.
func (c *client) UpdateCheck(
	ctx context.Context,
	ID string,
	check Check,
) (*Check, error) {
	var result Check
	err := c.callJSON(ctx, 1, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("checks/%s", ID)), createCheckPayload(check), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Deprecated: Use UpdateHeartbeatMonitor instead.
func (c *client) UpdateHeartbeat(
	ctx context.Context,
	ID string, check HeartbeatCheck,
) (*HeartbeatCheck, error) {
	return c.UpdateHeartbeatMonitor(ctx, ID, check)
}

func (c *client) UpdateHeartbeatMonitor(
	ctx context.Context,
	ID string,
	monitor HeartbeatMonitor,
) (*HeartbeatMonitor, error) {
	var result HeartbeatMonitor
	err := c.callJSON(ctx, 1, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("checks/heartbeat/%s", ID)), createHeartbeatMonitorPayload(monitor), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Deprecated: Use UpdateTCPMonitor instead.
func (c *client) UpdateTCPCheck(
	ctx context.Context,
	ID string,
	check TCPCheck,
) (*TCPCheck, error) {
	return c.UpdateTCPMonitor(ctx, ID, check)
}

func (c *client) UpdateTCPMonitor(
	ctx context.Context,
	ID string,
	monitor TCPMonitor,
) (*TCPMonitor, error) {
	var result TCPMonitor
	err := c.callJSON(ctx, 1, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("checks/tcp/%s", ID)), createTCPMonitorPayload(monitor), &result, http.StatusOK)
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
	err := c.callJSON(ctx, 1, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("checks/grpc/%s", ID)), createGRPCMonitorPayload(monitor), &result, http.StatusOK)
	if err != nil {
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
	err := c.callJSON(ctx, 1, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("checks/traceroute/%s", ID)), createTracerouteMonitorPayload(monitor), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateSSLMonitor(
	ctx context.Context,
	ID string,
	monitor SSLMonitor,
) (*SSLMonitor, error) {
	var result SSLMonitor
	err := c.callJSON(ctx, 1, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("checks/ssl/%s", ID)), createSSLMonitorPayload(monitor), &result, http.StatusOK)
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
	err := c.callJSON(ctx, 1, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("checks/url/%s", ID)), createURLMonitorPayload(monitor), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateDNSMonitor(
	ctx context.Context,
	ID string,
	monitor DNSMonitor,
) (*DNSMonitor, error) {
	var result DNSMonitor
	err := c.callJSON(ctx, 1, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("checks/dns/%s", ID)), createDNSMonitorPayload(monitor), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateICMPMonitor(
	ctx context.Context,
	ID string,
	monitor ICMPMonitor,
) (*ICMPMonitor, error) {
	var result ICMPMonitor
	err := c.callJSON(ctx, 1, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("checks/icmp/%s", ID)), createICMPMonitorPayload(monitor), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdatePlaywrightCheck(
	ctx context.Context,
	ID string,
	check PlaywrightCheck,
) (*PlaywrightCheck, error) {
	var result PlaywrightCheck
	err := c.callJSON(ctx, 1, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("checks/playwright/%s", ID)), createPlaywrightCheckPayload(check), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateAPICheck(
	ctx context.Context,
	ID string,
	check APICheck,
) (*APICheck, error) {
	var result APICheck
	err := c.callJSON(ctx, 1, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("checks/api/%s", ID)), createAPICheckPayload(check), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateBrowserCheck(
	ctx context.Context,
	ID string,
	check BrowserCheck,
) (*BrowserCheck, error) {
	var result BrowserCheck
	err := c.callJSON(ctx, 1, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("checks/browser/%s", ID)), createBrowserCheckPayload(check), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateMultiStepCheck(
	ctx context.Context,
	ID string,
	check MultiStepCheck,
) (*MultiStepCheck, error) {
	var result MultiStepCheck
	err := c.callJSON(ctx, 1, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("checks/multistep/%s", ID)), createMultiStepCheckPayload(check), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes the check with the specified ID.
func (c *client) DeleteCheck(
	ctx context.Context,
	ID string,
) error {
	_, err := c.call(ctx, 1, http.MethodDelete, fmt.Sprintf("checks/%s", ID), nil, http.StatusNoContent)
	return err
}

func (c *client) DeleteHeartbeatMonitor(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}

func (c *client) DeleteTCPMonitor(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}

func (c *client) DeleteGRPCMonitor(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}

func (c *client) DeleteTracerouteMonitor(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}

func (c *client) DeleteSSLMonitor(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}

func (c *client) DeleteURLMonitor(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}

func (c *client) DeleteDNSMonitor(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}

func (c *client) DeleteICMPMonitor(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}

func (c *client) DeletePlaywrightCheck(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}

func (c *client) DeleteAPICheck(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}

func (c *client) DeleteBrowserCheck(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}

func (c *client) DeleteMultiStepCheck(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}

// Get takes the ID of an existing check, and returns the check parameters, or
// an error.
func (c *client) GetCheck(
	ctx context.Context,
	ID string,
) (*Check, error) {
	result := Check{}
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("checks/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Deprecated: Use GetHeartbeatMonitor instead.
func (c *client) GetHeartbeatCheck(
	ctx context.Context,
	ID string,
) (*HeartbeatCheck, error) {
	return c.GetHeartbeatMonitor(ctx, ID)
}

func (c *client) GetHeartbeatMonitor(
	ctx context.Context,
	ID string,
) (*HeartbeatMonitor, error) {
	result := HeartbeatMonitor{}
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("checks/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTCPCheck takes the ID of an existing TCP check, and returns the check
// parameters, or an error.
//
// Deprecated: Use GetTCPMonitor instead.
func (c *client) GetTCPCheck(
	ctx context.Context,
	ID string,
) (*TCPCheck, error) {
	return c.GetTCPMonitor(ctx, ID)
}

// GetTCPMonitor takes the ID of an existing TCP monitor, and returns the
// monitor parameters, or an error.
func (c *client) GetTCPMonitor(
	ctx context.Context,
	ID string,
) (*TCPMonitor, error) {
	var result TCPMonitor
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("checks/%s", ID), nil, &result, http.StatusOK)
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
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("checks/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTracerouteMonitor takes the ID of an existing traceroute monitor, and
// returns the monitor parameters, or an error.
func (c *client) GetTracerouteMonitor(
	ctx context.Context,
	ID string,
) (*TracerouteMonitor, error) {
	var result TracerouteMonitor
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("checks/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetSSLMonitor takes the ID of an existing SSL monitor, and returns the
// monitor parameters, or an error.
func (c *client) GetSSLMonitor(
	ctx context.Context,
	ID string,
) (*SSLMonitor, error) {
	var result SSLMonitor
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("checks/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetURLMonitor takes the ID of an existing URL monitor, and returns the
// monitor parameters, or an error.
func (c *client) GetURLMonitor(
	ctx context.Context,
	ID string,
) (*URLMonitor, error) {
	var result URLMonitor
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("checks/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetDNSMonitor takes the ID of an existing DNS monitor, and returns the
// monitor parameters, or an error.
func (c *client) GetDNSMonitor(
	ctx context.Context,
	ID string,
) (*DNSMonitor, error) {
	var result DNSMonitor
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("checks/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetICMPMonitor takes the ID of an existing ICMP monitor, and returns the
// monitor parameters, or an error.
func (c *client) GetICMPMonitor(
	ctx context.Context,
	ID string,
) (*ICMPMonitor, error) {
	var result ICMPMonitor
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("checks/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPlaywrightCheck retrieves an existing Playwright check.
func (c *client) GetPlaywrightCheck(
	ctx context.Context,
	ID string,
) (*PlaywrightCheck, error) {
	var result PlaywrightCheck
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("checks/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// getCheckOfType fetches the check at path and decodes it into result. The
// generic check endpoint returns checks of every type, so the response's
// checkType is compared with the expected one: decoding a check of another
// type would silently drop its type-specific fields.
func (c *client) getCheckOfType(
	ctx context.Context,
	path string,
	checkType string,
	result any,
) error {
	res, err := c.call(ctx, 1, http.MethodGet, path, nil, http.StatusOK)
	if err != nil {
		return err
	}
	var header struct {
		CheckType string `json:"checkType"`
	}
	if err := decode(res, &header); err != nil {
		return err
	}
	if header.CheckType != checkType {
		return fmt.Errorf("%s %s: check is of type %q, not %q", res.method, res.path, header.CheckType, checkType)
	}
	return decode(res, result)
}

// GetAPICheck takes the ID of an existing API check, and returns the check
// parameters, or an error. It returns an error when the ID belongs to a check
// of another type.
func (c *client) GetAPICheck(
	ctx context.Context,
	ID string,
) (*APICheck, error) {
	var result APICheck
	if err := c.getCheckOfType(ctx, fmt.Sprintf("checks/%s", ID), "API", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBrowserCheck takes the ID of an existing browser check, and returns the
// check parameters, or an error. It returns an error when the ID belongs to a
// check of another type. The includeDependencies query parameter is required
// for the response to contain the check's script dependencies.
func (c *client) GetBrowserCheck(
	ctx context.Context,
	ID string,
) (*BrowserCheck, error) {
	var result BrowserCheck
	if err := c.getCheckOfType(ctx, fmt.Sprintf("checks/%s?includeDependencies=true", ID), "BROWSER", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetMultiStepCheck takes the ID of an existing multi-step check, and returns
// the check parameters, or an error. It returns an error when the ID belongs
// to a check of another type. The includeDependencies query parameter is
// required for the response to contain the check's script dependencies.
func (c *client) GetMultiStepCheck(
	ctx context.Context,
	ID string,
) (*MultiStepCheck, error) {
	var result MultiStepCheck
	if err := c.getCheckOfType(ctx, fmt.Sprintf("checks/%s?includeDependencies=true", ID), "MULTI_STEP", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

type groupPayload struct {
	Group
	// The update endpoint is PATCH-style: keys omitted from the payload keep
	// their server-side value. The fields below shadow embedded fields whose
	// zero value would be omitted so that they are always sent and can
	// actually be cleared via an update.
	SetupSnippetID            *int64                     `json:"setupSnippetId"`
	TearDownSnippetID         *int64                     `json:"tearDownSnippetId"`
	LocalSetupScript          *string                    `json:"localSetupScript"`
	LocalTearDownScript       *string                    `json:"localTearDownScript"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions"`
}

func createGroupPayload(group Group) groupPayload {
	payload := groupPayload{
		Group: group,
	}

	// A nil value for a list will cause the backend to not update the value.
	// We must send empty lists instead.
	if group.Locations == nil {
		payload.Locations = []string{}
	}

	if group.PrivateLocations == nil {
		payload.PrivateLocations = &[]string{}
	}

	// The snippet IDs must be null if empty or they will not get unset on
	// update.
	payload.SetupSnippetID = nullIfZero(group.SetupSnippetID)

	payload.TearDownSnippetID = nullIfZero(group.TearDownSnippetID)

	// The local scripts must be null if empty or they will not get unset on
	// update.
	payload.LocalSetupScript = nullIfZero(group.LocalSetupScript)

	payload.LocalTearDownScript = nullIfZero(group.LocalTearDownScript)

	payload.AlertChannelSubscriptions = emptyIfNil(group.AlertChannelSubscriptions)

	return payload
}

// CreateGroup creates a new check group with the specified details. It returns
// the newly-created group, or an error.
func (c *client) CreateGroup(
	ctx context.Context,
	group Group,
) (*Group, error) {
	var result Group
	err := c.callJSON(ctx, 1, http.MethodPost, withAutoAssignAlertsFlag("check-groups"), createGroupPayload(group), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type groupV2Payload struct {
	GroupV2
	// The embedded field's omitempty would drop the key for nil and empty
	// slices, leaving it up to the backend whether the server-side value is
	// kept or cleared. Shadow the field so an explicit empty list is always
	// sent and removing all subscriptions reliably clears them.
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions"`
}

func createGroupV2Payload(group GroupV2) groupV2Payload {
	payload := groupV2Payload{
		GroupV2: group,
	}

	// A nil value for a list will cause the backend to not update the value.
	// We must send empty lists instead.
	if group.Locations == nil {
		payload.Locations = []string{}
	}

	if group.PrivateLocations == nil {
		payload.PrivateLocations = &[]string{}
	}

	payload.AlertChannelSubscriptions = emptyIfNil(group.AlertChannelSubscriptions)

	return payload
}

// CreateGroupV2 creates a new check group with the specified details. It returns
// the newly-created group, or an error.
func (c *client) CreateGroupV2(
	ctx context.Context,
	group GroupV2,
) (*GroupV2, error) {
	var result GroupV2
	err := c.callJSON(ctx, 2, http.MethodPost, withAutoAssignAlertsFlag("check-groups"), createGroupV2Payload(group), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetGroup takes the ID of an existing check group, and returns the
// corresponding group, or an error.
func (c *client) GetGroup(
	ctx context.Context,
	ID int64,
) (*Group, error) {
	result := Group{}
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("check-groups/%d", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetGroupV2 takes the ID of an existing check group, and returns the
// corresponding group, or an error.
func (c *client) GetGroupV2(
	ctx context.Context,
	ID int64,
) (*GroupV2, error) {
	var result GroupV2
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("check-groups/%d", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateGroup takes the ID of an existing check group, and updates the
// corresponding check group to match the supplied group. It returns the updated
// group, or an error.
func (c *client) UpdateGroup(
	ctx context.Context,
	ID int64,
	group Group,
) (*Group, error) {
	var result Group
	err := c.callJSON(ctx, 1, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("check-groups/%d", ID)), createGroupPayload(group), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateGroupV2 takes the ID of an existing check group, and updates the
// corresponding check group to match the supplied group. It returns the updated
// group, or an error.
func (c *client) UpdateGroupV2(
	ctx context.Context,
	ID int64,
	group GroupV2,
) (*GroupV2, error) {
	var result GroupV2
	err := c.callJSON(ctx, 2, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("check-groups/%d", ID)), createGroupV2Payload(group), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteGroup deletes the check group with the specified ID.
func (c *client) DeleteGroup(
	ctx context.Context,
	ID int64,
) error {
	_, err := c.call(ctx, 1, http.MethodDelete, fmt.Sprintf("check-groups/%d", ID), nil, http.StatusNoContent)
	return err
}

// DeleteGroupV2 deletes the check group with the specified ID.
func (c *client) DeleteGroupV2(
	ctx context.Context,
	ID int64,
) error {
	return c.DeleteGroup(ctx, ID)
}

// GetCheckResult gets a specific Check result
func (c *client) GetCheckResult(
	ctx context.Context,
	checkID,
	checkResultID string,
) (*CheckResult, error) {
	result := CheckResult{}
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("check-results/%s/%s", checkID, checkResultID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetCheckResults gets the results of the given Check
func (c *client) GetCheckResults(
	ctx context.Context,
	checkID string,
	filters *CheckResultsFilter,
) ([]CheckResult, error) {
	uri := fmt.Sprintf("check-results/%s", checkID)
	if filters != nil {
		q := url.Values{}
		if filters.Page > 0 {
			q.Add("page", fmt.Sprintf("%d", filters.Page))
		}
		if filters.Limit > 0 {
			q.Add("limit", fmt.Sprintf("%d", filters.Limit))
		}
		if filters.From > 0 {
			q.Add("from", fmt.Sprintf("%d", filters.From))
		}
		if filters.To > 0 {
			q.Add("to", fmt.Sprintf("%d", filters.To))
		}
		if filters.CheckType == TypeBrowser || filters.CheckType == TypeAPI {
			q.Add("checkType", string(filters.CheckType))
		}
		if filters.HasFailures {
			q.Add("hasFailures", "1")
		}
		if len(filters.Location) > 0 {
			q.Add("location", filters.Location)
		}
		uri = uri + "?" + q.Encode()
	}

	result := []CheckResult{}
	err := c.callJSON(ctx, 1, http.MethodGet, uri, nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CreateSnippet creates a new snippet with the specified details. It returns
// the newly-created snippet, or an error.
func (c *client) CreateSnippet(
	ctx context.Context,
	snippet Snippet,
) (*Snippet, error) {
	var result Snippet
	err := c.callJSON(ctx, 1, http.MethodPost, "snippets", snippet, &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetSnippet takes the ID of an existing snippet, and returns the
// corresponding snippet, or an error.
func (c *client) GetSnippet(
	ctx context.Context,
	ID int64,
) (*Snippet, error) {
	result := Snippet{}
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("snippets/%d", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateSnippet takes the ID of an existing snippet, and updates the
// corresponding snippet to match the supplied snippet. It returns the updated
// snippet, or an error.
func (c *client) UpdateSnippet(
	ctx context.Context,
	ID int64,
	snippet Snippet,
) (*Snippet, error) {
	var result Snippet
	err := c.callJSON(ctx, 1, http.MethodPut, fmt.Sprintf("snippets/%d", ID), snippet, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteSnippet deletes the snippet with the specified ID. It returns a
func (c *client) DeleteSnippet(
	ctx context.Context,
	ID int64,
) error {
	_, err := c.call(ctx, 1, http.MethodDelete, fmt.Sprintf("snippets/%d", ID), nil, http.StatusNoContent)
	return err
}

// CreateEnvironmentVariable creates a new environment variable with the
// specified details.  It returns the newly-created environment variable,
// or an error.
func (c *client) CreateEnvironmentVariable(
	ctx context.Context,
	envVar EnvironmentVariable,
) (*EnvironmentVariable, error) {
	var result EnvironmentVariable
	err := c.callJSON(ctx, 1, http.MethodPost, "variables", envVar, &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetEnvironmentVariable takes the ID of an existing environment variable, and returns the
// corresponding environment variable, or an error.
func (c *client) GetEnvironmentVariable(
	ctx context.Context,
	key string,
) (*EnvironmentVariable, error) {
	result := EnvironmentVariable{}
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("variables/%s", key), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateEnvironmentVariable takes the ID of an existing environment variable, and updates the
// corresponding environment variable to match the supplied environment variable. It returns the updated
// environment variable, or an error.
func (c *client) UpdateEnvironmentVariable(
	ctx context.Context,
	key string,
	envVar EnvironmentVariable,
) (*EnvironmentVariable, error) {
	var result EnvironmentVariable
	err := c.callJSON(ctx, 1, http.MethodPut, fmt.Sprintf("variables/%s", key), envVar, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteEnvironmentVariable deletes the environment variable with the specified ID.
func (c *client) DeleteEnvironmentVariable(
	ctx context.Context,
	key string,
) error {
	_, err := c.call(ctx, 1, http.MethodDelete, fmt.Sprintf("variables/%s", key), nil, http.StatusNoContent)
	return err
}

// CreateAlertChannel creates a new alert channel with the specified details. It returns
// the newly-created alert channel, or an error.
func (c *client) CreateAlertChannel(
	ctx context.Context,
	ac AlertChannel,
) (*AlertChannel, error) {
	res, err := c.call(ctx, 1, http.MethodPost, "alert-channels", payloadFromAlertChannel(ac), http.StatusOK, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return alertChannelFromJSON(res)
}

// GetAlertChannel takes the ID of an existing alert channel, and returns the
// corresponding alert channel, or an error.
func (c *client) GetAlertChannel(
	ctx context.Context,
	ID int64,
) (*AlertChannel, error) {
	res, err := c.call(ctx, 1, http.MethodGet, fmt.Sprintf("alert-channels/%d", ID), nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return alertChannelFromJSON(res)
}

// UpdateAlertChannel takes the ID of an existing alert channel, and updates the
// corresponding alert channel to match the supplied alert channel. It returns the updated
// alert channel, or an error.
func (c *client) UpdateAlertChannel(
	ctx context.Context,
	ID int64,
	ac AlertChannel,
) (*AlertChannel, error) {
	res, err := c.call(ctx, 1, http.MethodPut, fmt.Sprintf("alert-channels/%d", ID), payloadFromAlertChannel(ac), http.StatusOK)
	if err != nil {
		return nil, err
	}
	return alertChannelFromJSON(res)
}

// DeleteAlertChannel deletes the alert channel with the specified ID. It returns a
func (c *client) DeleteAlertChannel(
	ctx context.Context,
	ID int64,
) error {
	_, err := c.call(ctx, 1, http.MethodDelete, fmt.Sprintf("alert-channels/%d", ID), nil, http.StatusNoContent)
	return err
}

// CreateDashboard creates a new dashboard with the specified details. It returns
// the newly-created dashboard, or an error.
func (c *client) CreateDashboard(
	ctx context.Context,
	dashboard Dashboard,
) (*Dashboard, error) {
	var result Dashboard
	err := c.callJSON(ctx, 1, http.MethodPost, "dashboards", dashboard, &result, http.StatusOK, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetDashboard takes the ID of an existing dashboard, and returns the
// corresponding dashboard, or an error.
func (c *client) GetDashboard(
	ctx context.Context,
	ID string,
) (*Dashboard, error) {
	result := Dashboard{}
	err := c.callJSON(ctx, 1, http.MethodGet, "dashboards/"+ID, nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteDashboard deletes the dashboard with the specified ID.
func (c *client) DeleteDashboard(
	ctx context.Context,
	ID string,
) error {
	_, err := c.call(ctx, 1, http.MethodDelete, "dashboards/"+ID, nil, http.StatusNoContent)
	return err
}

// UpdateDashboard takes the ID of an existing dashboard, and updates the
// corresponding dashboard to match the supplied dashboard. It returns the updated
// dashboard, or an error.
func (c *client) UpdateDashboard(
	ctx context.Context,
	ID string,
	dashboard Dashboard,
) (*Dashboard, error) {
	var result Dashboard
	err := c.callJSON(ctx, 1, http.MethodPut, "dashboards/"+ID, dashboard, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateMaintenanceWindow creates a new window with the specified details.
func (c *client) CreateMaintenanceWindow(
	ctx context.Context,
	mw MaintenanceWindow,
) (*MaintenanceWindow, error) {
	var result MaintenanceWindow
	err := c.callJSON(ctx, 1, http.MethodPost, "maintenance-windows", mw, &result, http.StatusOK, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetMaintenanceWindow takes the ID of an existing window, and returns the
// corresponding window.
func (c *client) GetMaintenanceWindow(
	ctx context.Context,
	ID int64,
) (*MaintenanceWindow, error) {
	result := MaintenanceWindow{}
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("maintenance-windows/%d", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteMaintenanceWindow deletes the window with the specified ID.
func (c *client) DeleteMaintenanceWindow(
	ctx context.Context,
	ID int64,
) error {
	_, err := c.call(ctx, 1, http.MethodDelete, fmt.Sprintf("maintenance-windows/%d", ID), nil, http.StatusNoContent)
	return err
}

// UpdateMaintenanceWindow takes the ID of an existing window, and updates the
// corresponding window to match the supplied one.
func (c *client) UpdateMaintenanceWindow(
	ctx context.Context,
	ID int64,
	mw MaintenanceWindow,
) (*MaintenanceWindow, error) {
	var result MaintenanceWindow
	err := c.callJSON(ctx, 1, http.MethodPut, fmt.Sprintf("maintenance-windows/%d", ID), mw, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CreatePrivateLocation creates a new private location with the specified details.
func (c *client) CreatePrivateLocation(
	ctx context.Context,
	pl PrivateLocation,
) (*PrivateLocation, error) {
	var result PrivateLocation
	err := c.callJSON(ctx, 1, http.MethodPost, "private-locations", pl, &result, http.StatusOK, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPrivateLocation takes the ID of an existing location, and returns the
// corresponding one.
func (c *client) GetPrivateLocation(
	ctx context.Context,
	ID string,
) (*PrivateLocation, error) {
	result := PrivateLocation{}
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("private-locations/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeletePrivateLocation deletes the private location with the specified ID.
func (c *client) DeletePrivateLocation(
	ctx context.Context,
	ID string,
) error {
	_, err := c.call(ctx, 1, http.MethodDelete, fmt.Sprintf("private-locations/%s", ID), nil, http.StatusNoContent)
	return err
}

// UpdatePrivateLocation takes the ID of an existing private location and updates it
// to match the new one.
func (c *client) UpdatePrivateLocation(
	ctx context.Context,
	ID string,
	pl PrivateLocation,
) (*PrivateLocation, error) {
	var result PrivateLocation
	err := c.callJSON(ctx, 1, http.MethodPut, fmt.Sprintf("private-locations/%s", ID), pl, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateTriggerCheck creates a new trigger with the specified details.
func (c *client) CreateTriggerCheck(
	ctx context.Context,
	checkID string,
) (*TriggerCheck, error) {
	var result TriggerCheck
	err := c.callJSON(ctx, 1, http.MethodPost, fmt.Sprintf("triggers/checks/%s", checkID), nil, &result, http.StatusOK, http.StatusCreated)
	if err != nil {
		return nil, err
	}

	result.URL = fmt.Sprintf("%s/checks/%s/trigger/%s", c.url, checkID, result.Token)

	return &result, nil
}

// GetTriggerCheck takes the ID of an existing trigger, and returns the
// corresponding trigger.
func (c *client) GetTriggerCheck(
	ctx context.Context,
	checkID string,
) (*TriggerCheck, error) {
	result := TriggerCheck{}
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("triggers/checks/%s", checkID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	result.URL = fmt.Sprintf("%s/checks/%s/trigger/%s", c.url, checkID, result.Token)
	return &result, nil
}

// DeleteTriggerCheck deletes the window with the specified ID.
func (c *client) DeleteTriggerCheck(
	ctx context.Context,
	checkID string,
) error {
	_, err := c.call(ctx, 1, http.MethodDelete, fmt.Sprintf("triggers/checks/%s", checkID), nil, http.StatusNoContent)
	return err
}

// CreateTriggerGroup creates a new trigger with the specified details.
func (c *client) CreateTriggerGroup(
	ctx context.Context,
	groupID int64,
) (*TriggerGroup, error) {
	var result TriggerGroup
	err := c.callJSON(ctx, 1, http.MethodPost, fmt.Sprintf("triggers/check-groups/%d", groupID), nil, &result, http.StatusOK, http.StatusCreated)
	if err != nil {
		return nil, err
	}

	result.URL = fmt.Sprintf("%s/check-groups/%d/trigger/%s", c.url, groupID, result.Token)

	return &result, nil
}

// GetTriggerGroup takes the ID of an existing trigger, and returns the
// corresponding trigger.
func (c *client) GetTriggerGroup(
	ctx context.Context,
	groupID int64,
) (*TriggerGroup, error) {
	result := TriggerGroup{}
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("triggers/check-groups/%d", groupID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}

	result.URL = fmt.Sprintf("%s/check-groups/%d/trigger/%s", c.url, groupID, result.Token)
	return &result, nil
}

// DeleteTriggerGroup deletes the window with the specified ID.
func (c *client) DeleteTriggerGroup(
	ctx context.Context,
	groupID int64,
) error {
	_, err := c.call(ctx, 1, http.MethodDelete, fmt.Sprintf("triggers/check-groups/%d", groupID), nil, http.StatusNoContent)
	return err
}

func (c *client) CreateClientCertificate(
	ctx context.Context,
	cs ClientCertificate,
) (*ClientCertificate, error) {
	var result ClientCertificate
	err := c.callJSON(ctx, 1, http.MethodPost, "client-certificates", cs, &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetClientCertificate(
	ctx context.Context,
	ID string,
) (*ClientCertificate, error) {
	var result ClientCertificate
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("client-certificates/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteClientCertificate(
	ctx context.Context,
	ID string,
) error {
	_, err := c.call(ctx, 1, http.MethodDelete, fmt.Sprintf("client-certificates/%s", ID), nil, http.StatusNoContent)
	return err
}

func (c *client) CreateStatusPage(
	ctx context.Context,
	page StatusPage,
) (*StatusPage, error) {
	var result StatusPage
	err := c.callJSON(ctx, 1, http.MethodPost, "status-pages", page, &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetStatusPage(
	ctx context.Context,
	ID string,
) (*StatusPage, error) {
	var result StatusPage
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("status-pages/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateStatusPage(
	ctx context.Context,
	ID string,
	page StatusPage,
) (*StatusPage, error) {
	var result StatusPage
	err := c.callJSON(ctx, 1, http.MethodPut, fmt.Sprintf("status-pages/%s", ID), page, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteStatusPage(
	ctx context.Context,
	ID string,
) error {
	_, err := c.call(ctx, 1, http.MethodDelete, fmt.Sprintf("status-pages/%s", ID), nil, http.StatusNoContent)
	return err
}

func (c *client) CreateStatusPageService(
	ctx context.Context,
	service StatusPageService,
) (*StatusPageService, error) {
	var result StatusPageService
	err := c.callJSON(ctx, 1, http.MethodPost, "status-pages/services", service, &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetStatusPageService(
	ctx context.Context,
	ID string,
) (*StatusPageService, error) {
	var result StatusPageService
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("status-pages/services/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateStatusPageService(
	ctx context.Context,
	ID string,
	service StatusPageService,
) (*StatusPageService, error) {
	var result StatusPageService
	err := c.callJSON(ctx, 1, http.MethodPut, fmt.Sprintf("status-pages/services/%s", ID), service, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteStatusPageService(
	ctx context.Context,
	ID string,
) error {
	_, err := c.call(ctx, 1, http.MethodDelete, fmt.Sprintf("status-pages/services/%s", ID), nil, http.StatusNoContent)
	return err
}

func (c *client) CreateStatusPageV3(
	ctx context.Context,
	page StatusPageV3,
) (*StatusPageV3, error) {
	var result StatusPageV3
	err := c.callJSON(ctx, 3, http.MethodPost, "status-pages", page, &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetStatusPageV3(
	ctx context.Context,
	ID string,
) (*StatusPageV3, error) {
	var result StatusPageV3
	err := c.callJSON(ctx, 3, http.MethodGet, fmt.Sprintf("status-pages/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateStatusPageV3(
	ctx context.Context,
	ID string,
	page StatusPageV3,
) (*StatusPageV3, error) {
	var result StatusPageV3
	err := c.callJSON(ctx, 3, http.MethodPut, fmt.Sprintf("status-pages/%s", ID), page, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteStatusPageV3(
	ctx context.Context,
	ID string,
) error {
	_, err := c.call(ctx, 3, http.MethodDelete, fmt.Sprintf("status-pages/%s", ID), nil, http.StatusNoContent)
	return err
}

func (c *client) CreateStatusPageComponentV3(
	ctx context.Context,
	statusPageID string,
	component StatusPageComponentV3,
) (*StatusPageComponentV3, error) {
	var result StatusPageComponentV3
	err := c.callJSON(ctx, 3, http.MethodPost, fmt.Sprintf("status-pages/%s/components", statusPageID), component, &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetStatusPageComponentV3(
	ctx context.Context,
	statusPageID string,
	ID string,
) (*StatusPageComponentV3, error) {
	var result StatusPageComponentV3
	err := c.callJSON(ctx, 3, http.MethodGet, fmt.Sprintf("status-pages/%s/components/%s", statusPageID, ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateStatusPageComponentV3(
	ctx context.Context,
	statusPageID string,
	ID string,
	component StatusPageComponentV3,
) (*StatusPageComponentV3, error) {
	var result StatusPageComponentV3
	err := c.callJSON(ctx, 3, http.MethodPut, fmt.Sprintf("status-pages/%s/components/%s", statusPageID, ID), component, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteStatusPageComponentV3(
	ctx context.Context,
	statusPageID string,
	ID string,
) error {
	_, err := c.call(ctx, 3, http.MethodDelete, fmt.Sprintf("status-pages/%s/components/%s", statusPageID, ID), nil, http.StatusNoContent)
	return err
}

func (c *client) CreateStatusPageAutomationRuleV3(
	ctx context.Context,
	statusPageID string,
	rule StatusPageAutomationRuleV3,
) (*StatusPageAutomationRuleV3, error) {
	var result StatusPageAutomationRuleV3
	err := c.callJSON(ctx, 3, http.MethodPost, fmt.Sprintf("status-pages/%s/automation-rules", statusPageID), rule, &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetStatusPageAutomationRuleV3(
	ctx context.Context,
	statusPageID string,
	ID string,
) (*StatusPageAutomationRuleV3, error) {
	var result StatusPageAutomationRuleV3
	err := c.callJSON(ctx, 3, http.MethodGet, fmt.Sprintf("status-pages/%s/automation-rules/%s", statusPageID, ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateStatusPageAutomationRuleV3(
	ctx context.Context,
	statusPageID string,
	ID string,
	rule StatusPageAutomationRuleV3,
) (*StatusPageAutomationRuleV3, error) {
	var result StatusPageAutomationRuleV3
	err := c.callJSON(ctx, 3, http.MethodPut, fmt.Sprintf("status-pages/%s/automation-rules/%s", statusPageID, ID), rule, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteStatusPageAutomationRuleV3(
	ctx context.Context,
	statusPageID string,
	ID string,
) error {
	_, err := c.call(ctx, 3, http.MethodDelete, fmt.Sprintf("status-pages/%s/automation-rules/%s", statusPageID, ID), nil, http.StatusNoContent)
	return err
}

func payloadFromAlertChannel(ac AlertChannel) map[string]interface{} {
	payload := map[string]interface{}{
		"id":     ac.ID,
		"type":   ac.Type,
		"config": ac.GetConfig(),
	}
	if ac.SendRecovery != nil {
		payload["sendRecovery"] = *ac.SendRecovery
	}
	if ac.SendDegraded != nil {
		payload["sendDegraded"] = *ac.SendDegraded
	}
	if ac.SendFailure != nil {
		payload["sendFailure"] = *ac.SendFailure
	}
	if ac.SSLExpiry != nil {
		payload["sslExpiry"] = *ac.SSLExpiry
	}
	if ac.SSLExpiryThreshold != nil {
		payload["sslExpiryThreshold"] = *ac.SSLExpiryThreshold
	}
	return payload
}

// alertChannelFromJSON decodes res, a response holding an alert channel. A
// scalar field (id, type, the send flags, sslExpiry and sslExpiryThreshold)
// that is JSON null is left unset; one of another type than expected is an
// error that names the field. The config is decoded by
// AlertChannelConfigFromJSON.
func alertChannelFromJSON(res rawResponse) (*AlertChannel, error) {
	result := map[string]any{}
	if err := decode(res, &result); err != nil {
		return nil, err
	}
	// fieldError reports the type, never the value, of a field: alert
	// channel configs hold secrets.
	fieldError := func(field string, v any, want string) error {
		return fmt.Errorf("%s %s: decoding %d response: field %q is %T, want %s", res.method, res.path, res.status, field, v, want)
	}
	optionalBool := func(field string) (*bool, error) {
		switch v := result[field].(type) {
		case nil:
			return nil, nil
		case bool:
			return &v, nil
		default:
			return nil, fieldError(field, v, "a boolean")
		}
	}

	resultAc := &AlertChannel{}
	switch v := result["id"].(type) {
	case nil:
	case float64:
		resultAc.ID = int64(v)
	default:
		return nil, fieldError("id", v, "a number")
	}
	switch v := result["type"].(type) {
	case nil:
	case string:
		resultAc.Type = v
	default:
		return nil, fieldError("type", v, "a string")
	}
	var err error
	if resultAc.SendRecovery, err = optionalBool("sendRecovery"); err != nil {
		return nil, err
	}
	if resultAc.SendFailure, err = optionalBool("sendFailure"); err != nil {
		return nil, err
	}
	if resultAc.SendDegraded, err = optionalBool("sendDegraded"); err != nil {
		return nil, err
	}
	if resultAc.SSLExpiry, err = optionalBool("sslExpiry"); err != nil {
		return nil, err
	}
	switch v := result["sslExpiryThreshold"].(type) {
	case nil:
	case float64:
		t := int(v)
		resultAc.SSLExpiryThreshold = &t
	default:
		return nil, fieldError("sslExpiryThreshold", v, "a number")
	}
	if cfg, ok := result["config"]; ok {
		cfgJSON, err := json.Marshal(cfg)
		if err != nil {
			return nil, fmt.Errorf("%s %s: decoding %d response: config: %w", res.method, res.path, res.status, err)
		}
		c, err := AlertChannelConfigFromJSON(resultAc.Type, cfgJSON)
		if err != nil {
			return nil, fmt.Errorf("%s %s: decoding %d response: config: %w", res.method, res.path, res.status, err)
		}
		resultAc.SetConfig(c)
	}
	return resultAc, nil
}

// Get a specific runtime specs
func (c *client) GetRuntime(
	ctx context.Context,
	ID string,
) (*Runtime, error) {
	result := Runtime{}
	err := c.callJSON(ctx, 1, http.MethodGet, fmt.Sprintf("runtimes/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get static IP lists
func (c *client) GetStaticIPs(
	ctx context.Context,
) ([]StaticIP, error) {
	var IPs []StaticIP

	// getting IPv6 first
	var datav6 map[string][]string
	err := c.callJSON(ctx, 1, http.MethodGet, "static-ipv6s-by-region", nil, &datav6, http.StatusOK)
	if err != nil {
		return nil, err
	}

	for region, ips := range datav6 {
		for _, ip := range ips {
			addr, err := netip.ParsePrefix(ip)
			if err != nil {
				return nil, fmt.Errorf("could not parse CIDR from %s: %w", ip, err)
			}

			IPs = append(IPs, StaticIP{Region: region, Address: addr})
		}
	}

	// and then IPv4
	var datav4 map[string][]string
	err = c.callJSON(ctx, 1, http.MethodGet, "static-ips-by-region", nil, &datav4, http.StatusOK)
	if err != nil {
		return nil, err
	}

	for region, ips := range datav4 {
		for _, ip := range ips {
			addr, err := netip.ParseAddr(ip)
			if err != nil {
				return nil, fmt.Errorf("could not parse IP address from %s: %w", ip, err)
			}
			IPs = append(IPs, StaticIP{Region: region, Address: netip.PrefixFrom(addr, 32)})
		}
	}

	return IPs, nil
}

// dumpResponse writes the raw response data to the debug output, if set, or
// standard error otherwise.
func (c *client) dumpResponse(resp *http.Response) {
	// ignore errors dumping response - no recovery from this
	responseDump, _ := httputil.DumpResponse(resp, true)
	fmt.Fprintln(c.debug, string(responseDump))
	fmt.Fprintln(c.debug)
}

func (c *client) addAuthHeaders(req *http.Request) error {
	req.Header.Add("Authorization", "Bearer "+c.apiKey)

	if strings.HasPrefix(c.apiKey, "cu") && c.accountId == "" {
		return errors.New("missing Checkly Account ID (required when using User API Keys)")
	}

	if c.accountId != "" {
		req.Header.Add("x-checkly-account", c.accountId)
	}
	if c.source != "" {
		req.Header.Add("x-checkly-source", c.source)
	} else {
		req.Header.Add("x-checkly-source", "go-sdk")
	}

	return nil
}

// rawResponse is a response as send returns it, before its status is
// checked, together with the request it answers.
type rawResponse struct {
	// method and path name the request in error messages; path is as
	// apiPath returns it.
	method string
	path   string
	status int
	body   []byte
	header http.Header
	// retries is how many times the request was re-sent after a 429.
	retries int
}

// apiPath returns the path of an API request for error messages: the
// versioned path without the query string, such as "/v1/checks/abc".
func apiPath(v int, path string) string {
	path, _, _ = strings.Cut(path, "?")
	return fmt.Sprintf("/v%d/%s", v, path)
}

// send sends a request to version v of the API and returns the response,
// whatever its status. path is relative to the version, such as
// "checks/abc?x=y". A 429 response is retried as described in retry.go; once
// the retries are spent, or the caller's deadline does not leave time for
// another one, the 429 is returned like any other response. When the context
// ends during a wait, the error matches both ErrRateLimited and the context
// error.
func (c *client) send(
	ctx context.Context,
	v int,
	method string,
	path string,
	data []byte,
) (rawResponse, error) {
	requestURL := fmt.Sprintf("%s/v%d/%s", c.url, v, path)
	for retry := 0; ; retry++ {
		res, err := c.doRequest(ctx, method, requestURL, apiPath(v, path), data)
		res.retries = retry
		if err != nil || res.status != http.StatusTooManyRequests || retry >= rateLimitMaxRetries {
			return res, err
		}

		// Rate limited: wait and re-send, unless the wait would leave too
		// little of the caller's deadline for the re-sent request, in which
		// case the caller sees the 429 itself.
		now := time.Now()
		wait := rateLimitWait(res.header, retry, now)
		if deadline, ok := ctx.Deadline(); ok && now.Add(wait+rateLimitRequestMargin).After(deadline) {
			return res, nil
		}
		if c.debug != nil {
			fmt.Fprintf(c.debug, "rate limited, retrying in %v (retry %d of %d)\n\n", wait, retry+1, rateLimitMaxRetries)
		}
		if err := c.sleep(ctx, wait); err != nil {
			return res, fmt.Errorf("%w; gave up waiting to retry: %w", newAPIError(res, nil, c.accountId), err)
		}
	}
}

// call marshals in as the request body, unless it is nil, sends the request
// to version v of the API and returns the response. A response status not
// among want is returned as an *APIError.
func (c *client) call(
	ctx context.Context,
	v int,
	method string,
	path string,
	in any,
	want ...int,
) (rawResponse, error) {
	var data []byte
	if in != nil {
		var err error
		if data, err = json.Marshal(in); err != nil {
			return rawResponse{}, fmt.Errorf("%s %s: encoding request: %w", method, apiPath(v, path), err)
		}
	}
	res, err := c.send(ctx, v, method, path, data)
	if err != nil {
		return rawResponse{}, err
	}
	for _, status := range want {
		if res.status == status {
			return res, nil
		}
	}
	return rawResponse{}, newAPIError(res, want, c.accountId)
}

// callJSON is call followed by decoding the response body into out.
func (c *client) callJSON(
	ctx context.Context,
	v int,
	method string,
	path string,
	in any,
	out any,
	want ...int,
) error {
	res, err := c.call(ctx, v, method, path, in, want...)
	if err != nil {
		return err
	}
	return decode(res, out)
}

// decode decodes the body of res, a successful response, into out. The error
// names the request, the status and the body's size, but never contains the
// body, which may hold secrets. Like the decoding it replaces, it reads the
// first JSON value and ignores the rest.
func decode(res rawResponse, out any) error {
	if err := json.NewDecoder(bytes.NewReader(res.body)).Decode(out); err != nil {
		return fmt.Errorf("%s %s: decoding %d response (%d bytes): %w", res.method, res.path, res.status, len(res.body), err)
	}
	return nil
}

// doRequest sends a single API request. The request is rebuilt from data on
// every call so that it can be re-sent after a rate-limited response. path is
// the request path for error messages, as apiPath returns it.
func (c *client) doRequest(
	ctx context.Context,
	method string,
	requestURL string,
	path string,
	data []byte,
) (rawResponse, error) {
	// The context is attached after the debug dump: DumpRequestOut fails on
	// a cancelled context, which would hide the transport error that Do
	// reports for it.
	req, err := http.NewRequest(method, requestURL, bytes.NewReader(data))
	if err != nil {
		return rawResponse{}, fmt.Errorf("%s %s: building request: %w", method, path, err)
	}

	err = c.addAuthHeaders(req)
	if err != nil {
		return rawResponse{}, err
	}

	req.Header.Add("content-type", "application/json")

	if c.debug != nil {
		requestDump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return rawResponse{}, fmt.Errorf("%s %s: dumping request for debug output: %w", method, path, err)
		}
		fmt.Fprintln(c.debug, string(requestDump))
		fmt.Fprintln(c.debug)
	}

	req = req.WithContext(ctx)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			return rawResponse{}, &transportError{method: method, path: path, host: req.URL.Host, err: urlErr}
		}
		return rawResponse{}, fmt.Errorf("%s %s: request to %s failed: %w", method, path, req.URL.Host, err)
	}

	defer resp.Body.Close()
	if c.debug != nil {
		c.dumpResponse(resp)
	}

	res := rawResponse{method: method, path: path, status: resp.StatusCode, header: resp.Header}
	res.body, err = io.ReadAll(resp.Body)
	if err != nil {
		return res, fmt.Errorf("%s %s: reading %d response: %w", method, path, resp.StatusCode, err)
	}
	return res, nil
}

func withAutoAssignAlertsFlag(url string) string {
	flag := "autoAssignAlerts=false"
	if strings.Contains(url, "?") {
		return url + "&" + flag
	}
	return url + "?" + flag
}

func (c *client) UploadCodeBundle(
	ctx context.Context,
	data io.Reader,
	size int64,
	options UploadCodeBundleOptions,
) (*CodeBundle, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/next/checkly-storage/upload-code-bundle", c.url),
		io.LimitReader(data, size),
	)
	if err != nil {
		return nil, err
	}

	err = c.addAuthHeaders(req)
	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/octet-stream")
	req.ContentLength = size

	if options.ChecksumSha256 != "" {
		req.Header.Add("x-bundle-checksum-sha256", options.ChecksumSha256)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == 200:
		decoder := json.NewDecoder(resp.Body)

		var output CodeBundle
		err = decoder.Decode(&output)
		if err != nil {
			return nil, err
		}

		return &output, nil
	default:
		return nil, fmt.Errorf("unexpected HTTP %d response code: %v", resp.StatusCode, resp.Status)
	}
}

type peekCodeBundlePayload struct {
	Key string `json:"key"`
}

var ErrCodeBundleNotFound = errors.New("code bundle not found")

func (c *client) PeekCodeBundle(
	ctx context.Context,
	key string,
) (*CodeBundle, error) {
	payload := peekCodeBundlePayload{
		Key: key,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/next/checkly-storage/peek-code-bundle", c.url),
		bytes.NewReader(data),
	)
	if err != nil {
		return nil, err
	}

	err = c.addAuthHeaders(req)
	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == 404, resp.StatusCode == 403:
		return nil, ErrCodeBundleNotFound
	case resp.StatusCode == 200:
		decoder := json.NewDecoder(resp.Body)

		var output CodeBundle
		err = decoder.Decode(&output)
		if err != nil {
			return nil, err
		}

		return &output, nil
	default:
		return nil, fmt.Errorf("unexpected HTTP %d response code: %v", resp.StatusCode, resp.Status)
	}
}
