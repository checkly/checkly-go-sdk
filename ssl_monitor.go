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
// thresholds are top-level DegradedResponseTime/MaxResponseTime fields. SSL
// monitors do support private locations.
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
	// required when SSLConfig.ClientCertificateMode is "explicit" and omitted
	// otherwise, so it is a pointer with omitempty.
	SSLClientCertificateID *string     `json:"sslClientCertificateId,omitempty"`
	Assertions             []Assertion `json:"assertions,omitempty"`
}

// SSLConfig represents the SSL-specific configuration nested inside an SSL
// monitor's request. SecurityBaseline is an optional pointer:
// leave it nil to inherit the server-side default baseline.
type SSLConfig struct {
	Hostname string `json:"hostname"`
	Port     int    `json:"port,omitempty"`
	// ServerName is an optional SNI override (nullable in the API).
	ServerName            *string           `json:"serverName,omitempty"`
	IPFamily              string            `json:"ipFamily,omitempty"`
	SkipChainValidation   bool              `json:"skipChainValidation"`
	HandshakeTimeoutMs    int               `json:"handshakeTimeoutMs,omitempty"`
	AlertDaysBeforeExpiry int               `json:"alertDaysBeforeExpiry,omitempty"`
	SecurityBaseline      *SecurityBaseline `json:"securityBaseline,omitempty"`
	// ClientCertificateMode is "auto" or "explicit" (optional).
	ClientCertificateMode string `json:"clientCertificateMode,omitempty"`
}

// SecurityBaseline represents the SSL security baseline — a set of enforceable
// and advisory rules, each with an optional value and a severity. Construct one
// only to override the server defaults; a nil *SecurityBaseline on SSLConfig
// inherits them. Enabled is a pointer so an explicit `false` can be sent while
// an unset value is omitted (the server defaults it to true).
type SecurityBaseline struct {
	Enabled *bool `json:"enabled,omitempty"`
	// Enforceable rules — server default severity "fail".
	MinTLSVersion          *SSLBaselineTLSRule      `json:"minTLSVersion,omitempty"`
	MinKeySizeBits         *SSLBaselineKeySizeRule  `json:"minKeySizeBits,omitempty"`
	WeakSignatureAlgorithm *SSLBaselineSeverityRule `json:"weakSignatureAlgorithm,omitempty"`
	WeakCipherSuite        *SSLBaselineSeverityRule `json:"weakCipherSuite,omitempty"`
	KnownBadCA             *SSLBaselineSeverityRule `json:"knownBadCA,omitempty"`
	// Advisory rules — server default severity "ignore".
	RecommendedTLSVersion   *SSLBaselineTLSRule      `json:"recommendedTLSVersion,omitempty"`
	RecommendedKeySizeBits  *SSLBaselineKeySizeRule  `json:"recommendedKeySizeBits,omitempty"`
	OCSPMustStapleRespected *SSLBaselineSeverityRule `json:"ocspMustStapleRespected,omitempty"`
	SCTPresent              *SSLBaselineSeverityRule `json:"sctPresent,omitempty"`
}

// SSLBaselineTLSRule is a baseline rule whose value is a TLS version string
// (e.g. "TLS1.2"/"TLS1.3").
type SSLBaselineTLSRule struct {
	Value    string `json:"value,omitempty"`
	Severity string `json:"severity,omitempty"`
}

// SSLBaselineKeySizeRule is a baseline rule whose value is a key size in bits.
type SSLBaselineKeySizeRule struct {
	Value    int    `json:"value,omitempty"`
	Severity string `json:"severity,omitempty"`
}

// SSLBaselineSeverityRule is a baseline rule that only carries a severity.
type SSLBaselineSeverityRule struct {
	Severity string `json:"severity,omitempty"`
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
	err := c.callJSON(ctx, http.MethodPost, withAutoAssignAlertsFlag("/v1/checks/ssl"), createSSLMonitorPayload(monitor), &result, http.StatusCreated)
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
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/checks/%s", ID), nil, &result, http.StatusOK)
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
