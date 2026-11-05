package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// SSLMonitor represents an SSL/TLS certificate monitor.
//
// Like the other connection monitors (gRPC, traceroute), the response-time
// thresholds are top-level DegradedResponseTime/MaxResponseTime fields.
type SSLMonitor struct {
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
	Request                   SSLRequest                 `json:"request"`
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

// SSLRequest represents the parameters for an SSL monitor's connection.
//
// SSLConfig is required by the public API (the wire schema is `.required()`),
// so it is sent as a value rather than an optional pointer.
type SSLRequest struct {
	SSLConfig SSLConfig `json:"sslConfig"`
	// SSLClientCertificateID is the FK to a stored client certificate. It is
	// required when SSLConfig.ClientCertificateMode is explicit. Nil is sent
	// as null, which removes a stored certificate.
	SSLClientCertificateID *string     `json:"sslClientCertificateId"`
	Assertions             []Assertion `json:"assertions,omitempty"`
}

// SSLClientCertificateMode is how an SSL monitor picks the client
// certificate it presents, in [SSLConfig].ClientCertificateMode.
type SSLClientCertificateMode string

// SSL client certificate modes, the values of
// [SSLConfig].ClientCertificateMode.
const (
	SSLClientCertificateModeAuto SSLClientCertificateMode = "auto"
	// SSLClientCertificateModeExplicit presents the certificate in
	// SSLRequest.SSLClientCertificateID.
	SSLClientCertificateModeExplicit SSLClientCertificateMode = "explicit"
)

// TLSVersion is a TLS protocol version, in the Value of [SSLBaselineTLSRule].
type TLSVersion string

// TLS versions, the values of [SSLBaselineTLSRule].Value.
const (
	TLSVersion12 TLSVersion = "TLS1.2"
	TLSVersion13 TLSVersion = "TLS1.3"
)

// SSLBaselineSeverity is what an SSL monitor does when a security baseline
// rule is violated, in the Severity of [SSLBaselineTLSRule],
// [SSLBaselineKeySizeRule] and [SSLBaselineSeverityRule].
type SSLBaselineSeverity string

// SSL baseline severities, the values of the Severity fields of the security
// baseline rules.
const (
	SSLBaselineSeverityFail    SSLBaselineSeverity = "fail"
	SSLBaselineSeverityDegrade SSLBaselineSeverity = "degrade"
	SSLBaselineSeverityIgnore  SSLBaselineSeverity = "ignore"
)

// SSLConfig represents the SSL-specific configuration nested inside an SSL
// monitor's request. SecurityBaseline is an optional pointer:
// leave it nil to inherit the server-side default baseline.
type SSLConfig struct {
	Hostname string `json:"hostname"`
	Port     int    `json:"port,omitempty"`
	// ServerName is an optional SNI override. Nil is sent as null, which
	// removes a stored override.
	ServerName            *string           `json:"serverName"`
	IPFamily              IPFamily          `json:"ipFamily,omitempty"`
	SkipChainValidation   bool              `json:"skipChainValidation"`
	HandshakeTimeoutMs    int               `json:"handshakeTimeoutMs,omitempty"`
	AlertDaysBeforeExpiry int               `json:"alertDaysBeforeExpiry,omitempty"`
	SecurityBaseline      *SecurityBaseline `json:"securityBaseline,omitempty"`
	// ClientCertificateMode is optional.
	ClientCertificateMode SSLClientCertificateMode `json:"clientCertificateMode,omitempty"`
}

// SecurityBaseline represents the SSL security baseline — a set of enforceable
// and advisory rules, each with an optional value and a severity. Construct one
// only to override the server defaults; a nil *SecurityBaseline on SSLConfig
// inherits them. Enabled is a pointer so an explicit `false` can be sent while
// an unset value is omitted (the server defaults it to true).
type SecurityBaseline struct {
	Enabled *bool `json:"enabled,omitempty"`
	// Enforceable rules — server default severity fail.
	MinTLSVersion          *SSLBaselineTLSRule      `json:"minTLSVersion,omitempty"`
	MinKeySizeBits         *SSLBaselineKeySizeRule  `json:"minKeySizeBits,omitempty"`
	WeakSignatureAlgorithm *SSLBaselineSeverityRule `json:"weakSignatureAlgorithm,omitempty"`
	WeakCipherSuite        *SSLBaselineSeverityRule `json:"weakCipherSuite,omitempty"`
	KnownBadCA             *SSLBaselineSeverityRule `json:"knownBadCA,omitempty"`
	// Advisory rules — server default severity ignore.
	RecommendedTLSVersion   *SSLBaselineTLSRule      `json:"recommendedTLSVersion,omitempty"`
	RecommendedKeySizeBits  *SSLBaselineKeySizeRule  `json:"recommendedKeySizeBits,omitempty"`
	OCSPMustStapleRespected *SSLBaselineSeverityRule `json:"ocspMustStapleRespected,omitempty"`
	SCTPresent              *SSLBaselineSeverityRule `json:"sctPresent,omitempty"`
}

// SSLBaselineTLSRule is a baseline rule whose value is a TLS version.
type SSLBaselineTLSRule struct {
	Value    TLSVersion          `json:"value,omitempty"`
	Severity SSLBaselineSeverity `json:"severity,omitempty"`
}

// SSLBaselineKeySizeRule is a baseline rule whose value is a key size in bits.
type SSLBaselineKeySizeRule struct {
	Value    int                 `json:"value,omitempty"`
	Severity SSLBaselineSeverity `json:"severity,omitempty"`
}

// SSLBaselineSeverityRule is a baseline rule that only carries a severity.
type SSLBaselineSeverityRule struct {
	Severity SSLBaselineSeverity `json:"severity,omitempty"`
}

type sslMonitorPayload struct {
	SSLMonitor
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
	Request                   sslRequestPayload          `json:"request"`
}

// sslRequestPayload sends an unset assertion list as [] rather than omitting
// it, as the update endpoint keeps the stored value of an omitted key.
type sslRequestPayload struct {
	SSLRequest
	Assertions []Assertion `json:"assertions"`
}

func createSSLMonitorPayload(monitor SSLMonitor) sslMonitorPayload {
	payload := sslMonitorPayload{
		SSLMonitor: monitor,
		// Unfortunately `checkType` is required for this endpoint.
		Type: CheckTypeSSL,
		// Unfortunately, this will default to true if not set.
		DoubleCheck:               false,
		Tags:                      emptyIfNil(monitor.Tags),
		GroupOrder:                nullIfZero(monitor.GroupOrder),
		AlertChannelSubscriptions: emptyIfNil(monitor.AlertChannelSubscriptions),
		Request: sslRequestPayload{
			SSLRequest: monitor.Request,
			Assertions: emptyIfNil(monitor.Request.Assertions),
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

func (c *client) CreateSSLMonitor(
	ctx context.Context,
	monitor SSLMonitor,
) (*SSLMonitor, error) {
	var result SSLMonitor
	err := c.callJSON(ctx, http.MethodPost, withAutoAssignAlertsFlag("/v1/checks/ssl"), createSSLMonitorPayload(monitor), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetSSLMonitor(
	ctx context.Context,
	ID string,
) (*SSLMonitor, error) {
	var result SSLMonitor
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s", ID), CheckTypeSSL, &result); err != nil {
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
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/ssl/%s", ID)), createSSLMonitorPayload(monitor), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteSSLMonitor(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}
