package checkly

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// CheckResult is the result of one check run, as returned by
// [Client.GetCheckResult] and, in a [CheckResultsPage],
// [Client.GetCheckResults].
type CheckResult struct {
	ID string `json:"id"`
	// Name is the name of the check.
	Name    string `json:"name"`
	CheckID string `json:"checkId"`
	// HasFailures reports whether the run failed, for example through a
	// failing assertion, a timeout or a failing script.
	HasFailures bool `json:"hasFailures"`
	// HasErrors reports an internal error in Checkly rather than in the
	// checked service.
	HasErrors bool `json:"hasErrors"`
	// IsDegraded and OverMaxResponseTime report whether the run exceeded the
	// degraded and failing limits of the check or monitor, such as its
	// DegradedResponseTime and MaxResponseTime.
	IsDegraded          bool `json:"isDegraded"`
	OverMaxResponseTime bool `json:"overMaxResponseTime"`
	// RunLocation is the location the check ran from.
	RunLocation string `json:"runLocation"`
	// ResponseTime is the time in milliseconds the check spent producing its
	// result: the measured operation, such as the request or the TLS
	// handshake, for API checks and protocol monitors, and the whole run for
	// browser, multi-step and Playwright checks. ICMP and traceroute monitors
	// report their average latency, which has a fractional part.
	ResponseTime       float64             `json:"responseTime"`
	APICheckResult     *APICheckResult     `json:"apiCheckResult"`
	BrowserCheckResult *BrowserCheckResult `json:"browserCheckResult"`
	// TracerouteCheckResult, GRPCCheckResult and SSLCheckResult hold the
	// diagnostics of a traceroute, gRPC or SSL monitor run, and are nil for
	// the other check types.
	TracerouteCheckResult *TracerouteCheckResult `json:"tracerouteCheckResult"`
	GRPCCheckResult       *GRPCCheckResult       `json:"grpcCheckResult"`
	SSLCheckResult        *SSLCheckResult        `json:"sslCheckResult"`
	// CheckRunID is the ID of the check run that produced this result.
	CheckRunID int64 `json:"checkRunId"`
	// Attempts is how often the run was retried.
	Attempts  int64     `json:"attempts"`
	StartedAt time.Time `json:"startedAt"`
	StoppedAt time.Time `json:"stoppedAt"`
	CreatedAt time.Time `json:"created_at"`
}

// APICheckResult is the request and response detail of an API check run,
// decoded as an untyped JSON object.
type APICheckResult map[string]interface{}

// BrowserCheckResult is the detail of a browser check run, decoded as an
// untyped JSON object.
type BrowserCheckResult map[string]interface{}

// The traceroute, gRPC and SSL result types below type the fields the API
// documents, and keep its open-ended objects (timing phases, the request,
// assertions, certificate and security baseline) as maps, so that decoding
// drops nothing the API reports.

// TracerouteCheckResult holds the diagnostics of a traceroute monitor run.
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

// TracerouteCheckResponse is the detailed response of a traceroute monitor
// run.
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

// GRPCCheckResult holds the diagnostics of a gRPC monitor run.
type GRPCCheckResult struct {
	GRPCStatusCode *int                     `json:"grpcStatusCode,omitempty"`
	HealthStatus   *int                     `json:"healthStatus,omitempty"`
	TimingPhases   map[string]interface{}   `json:"timingPhases,omitempty"`
	RequestError   *string                  `json:"requestError,omitempty"`
	Request        map[string]interface{}   `json:"request,omitempty"`
	Assertions     []map[string]interface{} `json:"assertions,omitempty"`
	Response       *GRPCCheckResponse       `json:"response,omitempty"`
}

// GRPCCheckResponse is the detailed response of a gRPC monitor run.
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

// SSLCheckResult holds the diagnostics of an SSL monitor run.
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

// SSLCheckResponse is the detailed response of an SSL monitor run.
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

// CheckResultsPage is one page of the results of a check, as returned by
// [Client.GetCheckResults].
type CheckResultsPage struct {
	// Entries holds the results of this page.
	Entries []CheckResult `json:"entries"`
	// NextID is the cursor of the next page, to set as
	// [CheckResultsFilter].NextID. It is empty on the last page.
	NextID string `json:"nextId"`
}

// CheckResultsFilter narrows the results returned by
// [Client.GetCheckResults]. Zero-valued fields are not sent.
//
// The API allows 60 requests per 60 seconds to this endpoint, and a request
// that sets HasFailures counts as five.
type CheckResultsFilter struct {
	// Limit is the number of results per page, at most 100; the API
	// defaults to 10.
	Limit int64
	// NextID is the cursor of the page to read, from
	// [CheckResultsPage].NextID; empty reads the first page. The cursor does
	// not carry the other fields, so send the same ones on every page.
	NextID string
	// Location limits the results to one location, such as "eu-west-1".
	Location string
	// From and To limit the results to runs from From (inclusive) to To
	// (exclusive), as Unix timestamps in seconds. There is no default window:
	// an unset From or To leaves that end open, so a nil filter spans all the
	// raw results the account retains, which depends on its plan. The API
	// answers 408 when a wide query takes too long.
	To   int64
	From int64
	// CheckType is the type of the check, such as [CheckTypeMultiStep] or
	// [CheckTypeURL]. The API answers 404 when the check is of another type,
	// and 400 for a type it does not know.
	CheckType CheckType
	// HasFailures limits the results to failed runs.
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
) (*CheckResultsPage, error) {
	uri := fmt.Sprintf("/v2/check-results/%s", checkID)
	if filters != nil {
		q := url.Values{}
		if filters.NextID != "" {
			q.Add("nextId", filters.NextID)
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
			q.Add("hasFailures", "true")
		}
		if len(filters.Location) > 0 {
			q.Add("location", filters.Location)
		}
		if len(q) > 0 {
			uri += "?" + q.Encode()
		}
	}

	result := CheckResultsPage{}
	err := c.callJSON(ctx, http.MethodGet, uri, nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
