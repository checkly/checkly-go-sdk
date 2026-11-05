package checkly

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// CheckResult represents a Check result
type CheckResult struct {
	ID                  string              `json:"id"`
	Name                string              `json:"name"`
	CheckID             string              `json:"checkId"`
	HasFailures         bool                `json:"hasFailures"`
	HasErrors           bool                `json:"hasErrors"`
	IsDegraded          bool                `json:"isDegraded"`
	OverMaxResponseTime bool                `json:"overMaxResponseTime"`
	RunLocation         string              `json:"runLocation"`
	ResponseTime        int64               `json:"responseTime"`
	APICheckResult      *APICheckResult     `json:"apiCheckResult"`
	BrowserCheckResult  *BrowserCheckResult `json:"browserCheckResult"`
	// Typed failure-debug diagnostics for the uptime-monitor types. Each is nil
	// for the other check types, mirroring the apiCheckResult sibling.
	TracerouteCheckResult *TracerouteCheckResult `json:"tracerouteCheckResult"`
	GRPCCheckResult       *GRPCCheckResult       `json:"grpcCheckResult"`
	SSLCheckResult        *SSLCheckResult        `json:"sslCheckResult"`
	CheckRunID            int64                  `json:"checkRunId"`
	Attempts              int64                  `json:"attempts"`
	StartedAt             time.Time              `json:"startedAt"`
	StoppedAt             time.Time              `json:"stoppedAt"`
	CreatedAt             time.Time              `json:"created_at"`
}

// APICheckResult represents an API Check result
type APICheckResult map[string]interface{}

// BrowserCheckResult represents a Browser Check result
type BrowserCheckResult map[string]interface{}

// The TRACEROUTE / GRPC / SSL result types below mirror the additive optional
// fields the public check-results response carries alongside apiCheckResult (see
// the backend public-api check-results schemas). They are the read-path types an
// SDK consumer decodes to debug a failing uptime-monitor result. Documented
// scalars are typed; the open runner sub-objects (timingPhases / request /
// assertions / certificate / securityBaseline) stay map/slice of interface{} so
// the full runner artifact survives JSON decoding without dropping fields.

// TracerouteCheckResult holds the failure-debug diagnostics for a traceroute
// check result.
type TracerouteCheckResult struct {
	TotalHops          *int                     `json:"totalHops,omitempty"`
	DestinationReached *bool                    `json:"destinationReached,omitempty"`
	FinalHopLatency    map[string]interface{}   `json:"finalHopLatency,omitempty"`
	TimingPhases       map[string]interface{}   `json:"timingPhases,omitempty"`
	RequestError       *string                  `json:"requestError,omitempty"`
	Request            map[string]interface{}   `json:"request,omitempty"`
	Assertions         []map[string]interface{} `json:"assertions,omitempty"`
	Response           *TracerouteCheckResponse `json:"response,omitempty"`
}

// TracerouteCheckResponse is the detailed traceroute response artifact.
type TracerouteCheckResponse struct {
	Hostname           string                   `json:"hostname,omitempty"`
	ResolvedIP         string                   `json:"resolvedIp,omitempty"`
	TotalHops          int                      `json:"totalHops,omitempty"`
	DestinationReached bool                     `json:"destinationReached,omitempty"`
	TruncationReason   string                   `json:"truncationReason,omitempty"`
	FinalHopLatency    map[string]interface{}   `json:"finalHopLatency,omitempty"`
	Hops               []map[string]interface{} `json:"hops,omitempty"`
	Protocol           string                   `json:"protocol,omitempty"`
	ProbeProtocol      string                   `json:"probeProtocol,omitempty"`
}

// GRPCCheckResult holds the failure-debug diagnostics for a gRPC check result.
type GRPCCheckResult struct {
	GRPCStatusCode *int                     `json:"grpcStatusCode,omitempty"`
	HealthStatus   *int                     `json:"healthStatus,omitempty"`
	TimingPhases   map[string]interface{}   `json:"timingPhases,omitempty"`
	RequestError   *string                  `json:"requestError,omitempty"`
	Request        map[string]interface{}   `json:"request,omitempty"`
	Assertions     []map[string]interface{} `json:"assertions,omitempty"`
	Response       *GRPCCheckResponse       `json:"response,omitempty"`
}

// GRPCCheckResponse is the detailed gRPC response artifact.
type GRPCCheckResponse struct {
	GRPCMode          string                 `json:"grpcMode,omitempty"`
	Host              string                 `json:"host,omitempty"`
	ResolvedIP        string                 `json:"resolvedIp,omitempty"`
	Port              int                    `json:"port,omitempty"`
	GRPCMethod        string                 `json:"grpcMethod,omitempty"`
	ResponseMessage   string                 `json:"responseMessage,omitempty"`
	GRPCStatusCode    int                    `json:"grpcStatusCode,omitempty"`
	GRPCStatusMessage string                 `json:"grpcStatusMessage,omitempty"`
	HealthStatus      *int                   `json:"healthStatus,omitempty"`
	HealthStatusLabel string                 `json:"healthStatusLabel,omitempty"`
	Metadata          []GRPCMetadata         `json:"metadata,omitempty"`
	DiscoveredMethods []string               `json:"discoveredMethods,omitempty"`
	RequestError      string                 `json:"requestError,omitempty"`
	TimingPhases      map[string]interface{} `json:"timingPhases,omitempty"`
}

// SSLCheckResult holds the failure-debug diagnostics for an SSL check result.
type SSLCheckResult struct {
	TLSVersion       string                   `json:"tlsVersion,omitempty"`
	CipherSuite      string                   `json:"cipherSuite,omitempty"`
	DaysUntilExpiry  *int                     `json:"daysUntilExpiry,omitempty"`
	HandshakeTimeMs  *float64                 `json:"handshakeTimeMs,omitempty"`
	ChainTrusted     *bool                    `json:"chainTrusted,omitempty"`
	HostnameVerified *bool                    `json:"hostnameVerified,omitempty"`
	BaselineVerdict  string                   `json:"baselineVerdict,omitempty"`
	BaselineGrade    string                   `json:"baselineGrade,omitempty"`
	FailureCategory  string                   `json:"failureCategory,omitempty"`
	RequestError     *string                  `json:"requestError,omitempty"`
	Request          map[string]interface{}   `json:"request,omitempty"`
	Assertions       []map[string]interface{} `json:"assertions,omitempty"`
	Response         *SSLCheckResponse        `json:"response,omitempty"`
}

// SSLCheckResponse is the detailed SSL/TLS response artifact.
type SSLCheckResponse struct {
	ResolvedIP       string                   `json:"resolvedIp,omitempty"`
	Protocol         string                   `json:"protocol,omitempty"`
	CipherSuite      string                   `json:"cipherSuite,omitempty"`
	HandshakeTimeMs  float64                  `json:"handshakeTimeMs,omitempty"`
	HostnameVerified bool                     `json:"hostnameVerified,omitempty"`
	ChainTrusted     bool                     `json:"chainTrusted,omitempty"`
	DaysUntilExpiry  int                      `json:"daysUntilExpiry,omitempty"`
	OCSPStapled      bool                     `json:"ocspStapled,omitempty"`
	SecurityBaseline map[string]interface{}   `json:"securityBaseline,omitempty"`
	Certificate      map[string]interface{}   `json:"certificate,omitempty"`
	Chain            []map[string]interface{} `json:"chain,omitempty"`
}

// CheckResultsFilter represents the parameters that can be passed while
// getting Check Results. Zero-valued fields are not sent. A non-empty
// CheckType, such as "MULTI_STEP" or "URL", is sent as is: the API answers
// 404 when the check is of another type, and 400 for a type it does not know.
type CheckResultsFilter struct {
	Limit       int64
	Page        int64
	Location    string
	To          int64
	From        int64
	CheckType   CheckType
	HasFailures bool
}

func (c *client) GetCheckResult(
	ctx context.Context,
	checkID,
	checkResultID string,
) (*CheckResult, error) {
	result := CheckResult{}
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/check-results/%s/%s", checkID, checkResultID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetCheckResults(
	ctx context.Context,
	checkID string,
	filters *CheckResultsFilter,
) ([]CheckResult, error) {
	uri := fmt.Sprintf("/v1/check-results/%s", checkID)
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
		if filters.CheckType != "" {
			q.Add("checkType", string(filters.CheckType))
		}
		if filters.HasFailures {
			q.Add("hasFailures", "1")
		}
		if len(filters.Location) > 0 {
			q.Add("location", filters.Location)
		}
		if len(q) > 0 {
			uri += "?" + q.Encode()
		}
	}

	result := []CheckResult{}
	err := c.callJSON(ctx, http.MethodGet, uri, nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return result, nil
}
