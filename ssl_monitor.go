package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// SSLMonitor connects to a host over TLS and checks its certificate and
// connection against its assertions and security baseline. It is the input of
// [Client.CreateSSLMonitor] and [Client.UpdateSSLMonitor], and is returned by
// [Client.GetSSLMonitor].
type SSLMonitor struct {
	// ID is the monitor's UUID, assigned by the API.
	ID          string  `json:"id,omitempty"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	// Frequency is the number of minutes between runs: 0, 1, 2, 5, 10, 15,
	// 30, 60, 120, 180, 360, 720 or 1440.
	Frequency int `json:"frequency"`
	// FrequencyOffset is 10, 20 or 30 when Frequency is 0, for a run every
	// that many seconds. For any other Frequency the API picks the offset
	// itself and keeps it until Frequency changes: a value sent has no
	// effect, and the value read back is the API's. The API still rejects a
	// value above Frequency * 10 for a Frequency up to 60, or above
	// Frequency / 5 rounded up for a higher one. FrequencyOffset is left out
	// of the request when zero.
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
	// response time limits in milliseconds, each at most 30000, with
	// MaxResponseTime at least DegradedResponseTime. DegradedResponseTime and
	// MaxResponseTime are left out of the request when zero.
	DegradedResponseTime int      `json:"degradedResponseTime,omitempty"`
	MaxResponseTime      int      `json:"maxResponseTime,omitempty"`
	Tags                 []string `json:"tags,omitempty"`
	// AlertSettings is left out of the request when nil.
	AlertSettings *AlertSettings `json:"alertSettings,omitempty"`
	// UseGlobalAlertSettings applies the account's alert settings instead
	// of AlertSettings.
	UseGlobalAlertSettings bool       `json:"useGlobalAlertSettings"`
	Request                SSLRequest `json:"request"`
	// GroupID is the ID of the check group the monitor belongs to; 0 means
	// none.
	GroupID int64 `json:"groupId,omitempty"`
	// GroupOrder is the monitor's position in its group, which sets the run
	// order when the group is triggered from the API or CI/CD. 0 is the
	// first position. Nil means no position, and is sent as null so that
	// an update clears it.
	GroupOrder *int `json:"groupOrder"`
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

func (*SSLMonitor) isCheck() {}

// SSLRequest is the connection of an [SSLMonitor]. The API requires
// SSLConfig, so it is a value rather than a pointer.
type SSLRequest struct {
	SSLConfig SSLConfig `json:"sslConfig"`
	// SSLClientCertificateID is the ID of a [ClientCertificate]. It is
	// required when SSLConfig.ClientCertificateMode is
	// [SSLClientCertificateModeExplicit]. Nil is sent as null, which removes
	// a stored certificate.
	SSLClientCertificateID *string `json:"sslClientCertificateId"`
	// Assertions is sent as an empty list when nil.
	Assertions []Assertion `json:"assertions,omitempty"`
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

// SSLConfig is the TLS configuration in an [SSLRequest].
type SSLConfig struct {
	Hostname string `json:"hostname"`
	// Port is 1 to 65535; the API defaults to 443. Port is left out of the
	// request when zero. An update with a zero Port keeps the stored port, as
	// the API rejects a zero value.
	Port int `json:"port,omitempty"`
	// ServerName is an optional SNI override. Nil is sent as null, which
	// removes a stored override.
	ServerName *string `json:"serverName"`
	// IPFamily is left out of the request when empty. An update with an
	// empty IPFamily keeps the stored family, as the API rejects an empty
	// value.
	IPFamily            IPFamily `json:"ipFamily,omitempty"`
	SkipChainValidation bool     `json:"skipChainValidation"`
	// HandshakeTimeoutMs is 1000 to 30000; the API defaults to 10000.
	// HandshakeTimeoutMs is left out of the request when zero. An update
	// with a zero HandshakeTimeoutMs keeps the stored timeout, as the API
	// rejects a zero value.
	HandshakeTimeoutMs int `json:"handshakeTimeoutMs,omitempty"`
	// AlertDaysBeforeExpiry is 1 to 365; the API defaults to 20.
	// AlertDaysBeforeExpiry is left out of the request when zero. An update
	// with a zero AlertDaysBeforeExpiry keeps the stored value, as the API
	// rejects a zero value.
	AlertDaysBeforeExpiry int `json:"alertDaysBeforeExpiry,omitempty"`
	// SecurityBaseline is left out of the request when nil, and the API
	// then applies its default baseline.
	SecurityBaseline *SecurityBaseline `json:"securityBaseline,omitempty"`
	// With [SSLClientCertificateModeExplicit],
	// SSLRequest.SSLClientCertificateID is required.
	// ClientCertificateMode is left out of the request when empty.
	ClientCertificateMode SSLClientCertificateMode `json:"clientCertificateMode,omitempty"`
}

// SecurityBaseline is the set of TLS rules in [SSLConfig] that an SSL monitor
// enforces or reports on, each with a severity. Set it only to override the
// API defaults, which apply on create to the rules and values left unset.
type SecurityBaseline struct {
	// Enabled defaults to true, but an update that sets a MinTLSVersion or
	// MinKeySizeBits value with a nil Enabled stores none, and the baseline
	// then runs disabled; set Enabled explicitly. Enabled is left out of the
	// request when nil.
	Enabled *bool `json:"enabled,omitempty"`
	// Enforceable rules, with the default severity
	// [SSLBaselineSeverityFail]. The defaults for MinTLSVersion and
	// MinKeySizeBits are [TLSVersion12] and 2048. MinTLSVersion,
	// MinKeySizeBits, WeakSignatureAlgorithm, WeakCipherSuite and KnownBadCA
	// are left out of the request when nil.
	MinTLSVersion          *SSLBaselineTLSRule      `json:"minTLSVersion,omitempty"`
	MinKeySizeBits         *SSLBaselineKeySizeRule  `json:"minKeySizeBits,omitempty"`
	WeakSignatureAlgorithm *SSLBaselineSeverityRule `json:"weakSignatureAlgorithm,omitempty"`
	WeakCipherSuite        *SSLBaselineSeverityRule `json:"weakCipherSuite,omitempty"`
	KnownBadCA             *SSLBaselineSeverityRule `json:"knownBadCA,omitempty"`
	// Advisory rules, with the default severity
	// [SSLBaselineSeverityIgnore]. The defaults for RecommendedTLSVersion
	// and RecommendedKeySizeBits are [TLSVersion13] and 3072.
	// RecommendedTLSVersion, RecommendedKeySizeBits, OCSPMustStapleRespected
	// and SCTPresent are left out of the request when nil.
	RecommendedTLSVersion   *SSLBaselineTLSRule      `json:"recommendedTLSVersion,omitempty"`
	RecommendedKeySizeBits  *SSLBaselineKeySizeRule  `json:"recommendedKeySizeBits,omitempty"`
	OCSPMustStapleRespected *SSLBaselineSeverityRule `json:"ocspMustStapleRespected,omitempty"`
	SCTPresent              *SSLBaselineSeverityRule `json:"sctPresent,omitempty"`
}

// SSLBaselineTLSRule is a [SecurityBaseline] rule whose value is a TLS
// version.
type SSLBaselineTLSRule struct {
	// Value is left out of the request when empty.
	Value TLSVersion `json:"value,omitempty"`
	// Severity is left out of the request when empty.
	Severity SSLBaselineSeverity `json:"severity,omitempty"`
}

// SSLBaselineKeySizeRule is a [SecurityBaseline] rule whose value is a key
// size in bits, 1024 to 16384.
type SSLBaselineKeySizeRule struct {
	// Value is left out of the request when zero.
	Value int `json:"value,omitempty"`
	// Severity is left out of the request when empty.
	Severity SSLBaselineSeverity `json:"severity,omitempty"`
}

// SSLBaselineSeverityRule is a [SecurityBaseline] rule without a value.
type SSLBaselineSeverityRule struct {
	// Severity is left out of the request when empty.
	Severity SSLBaselineSeverity `json:"severity,omitempty"`
}

type sslMonitorPayload struct {
	SSLMonitor
	Type        CheckType `json:"checkType"`
	DoubleCheck bool      `json:"doubleCheck"`
	GroupID     *int64    `json:"groupId"`
	// The update endpoint keeps the stored value of an omitted key. The
	// fields below shadow embedded fields whose zero value would be omitted,
	// so that an update resets them: lists are sent empty.
	Tags                      []string                   `json:"tags"`
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
	id string,
) (*SSLMonitor, error) {
	var result SSLMonitor
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s", id), CheckTypeSSL, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateSSLMonitor(
	ctx context.Context,
	id string,
	monitor SSLMonitor,
) (*SSLMonitor, error) {
	var result SSLMonitor
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/ssl/%s", id)), createSSLMonitorPayload(monitor), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteSSLMonitor(
	ctx context.Context,
	id string,
) error {
	return c.DeleteCheck(ctx, id)
}
