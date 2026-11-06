package checkly

// AssertionSource is the value an [Assertion] checks, in Assertion.Source.
// The sources a check type accepts:
//
//   - API checks and the API check defaults of groups: STATUS_CODE,
//     JSON_BODY, HEADERS, TEXT_BODY, RESPONSE_TIME.
//   - URL monitors: STATUS_CODE.
//   - TCP monitors: RESPONSE_DATA, RESPONSE_TIME.
//   - DNS monitors: RESPONSE_TIME, RESPONSE_CODE, TEXT_ANSWER, JSON_ANSWER.
//   - ICMP monitors: LATENCY, JSON_RESPONSE.
//   - Traceroute monitors: RESPONSE_TIME, HOP_COUNT, PACKET_LOSS.
//   - gRPC monitors: RESPONSE_TIME, GRPC_STATUS_CODE,
//     GRPC_HEALTHCHECK_STATUS, GRPC_RESPONSE, TEXT_BODY, GRPC_METADATA.
//   - SSL monitors: CERTIFICATE, CONNECTION, RESPONSE_TIME, JSON_RESPONSE,
//     TEXT_RESPONSE.
//
// Browser and multistep checks, Playwright check suites and heartbeat
// monitors take no assertions.
type AssertionSource string

// Assertion sources, the values of [Assertion].Source.
const (
	AssertionSourceStatusCode   AssertionSource = "STATUS_CODE"
	AssertionSourceJSONBody     AssertionSource = "JSON_BODY"
	AssertionSourceTextBody     AssertionSource = "TEXT_BODY"
	AssertionSourceHeaders      AssertionSource = "HEADERS"
	AssertionSourceResponseTime AssertionSource = "RESPONSE_TIME"
	AssertionSourceResponseData AssertionSource = "RESPONSE_DATA"
	// AssertionSourceCertificate is a certificate field of an SSL monitor,
	// named by the assertion's Property.
	AssertionSourceCertificate AssertionSource = "CERTIFICATE"
	// AssertionSourceConnection is a connection field of an SSL monitor, such
	// as tlsVersion, named by the assertion's Property.
	AssertionSourceConnection            AssertionSource = "CONNECTION"
	AssertionSourceJSONResponse          AssertionSource = "JSON_RESPONSE"
	AssertionSourceTextResponse          AssertionSource = "TEXT_RESPONSE"
	AssertionSourceResponseCode          AssertionSource = "RESPONSE_CODE"
	AssertionSourceTextAnswer            AssertionSource = "TEXT_ANSWER"
	AssertionSourceJSONAnswer            AssertionSource = "JSON_ANSWER"
	AssertionSourceLatency               AssertionSource = "LATENCY"
	AssertionSourceHopCount              AssertionSource = "HOP_COUNT"
	AssertionSourcePacketLoss            AssertionSource = "PACKET_LOSS"
	AssertionSourceGRPCResponse          AssertionSource = "GRPC_RESPONSE"
	AssertionSourceGRPCMetadata          AssertionSource = "GRPC_METADATA"
	AssertionSourceGRPCHealthcheckStatus AssertionSource = "GRPC_HEALTHCHECK_STATUS"
	AssertionSourceGRPCStatusCode        AssertionSource = "GRPC_STATUS_CODE"
)

// AssertionComparison is how an [Assertion] compares its source with its
// target, in Assertion.Comparison. Below, "numeric" means EQUALS, NOT_EQUALS,
// GREATER_THAN and LESS_THAN, and "text" means those plus IS_EMPTY,
// NOT_EMPTY, CONTAINS and NOT_CONTAINS. The comparisons a check type accepts:
//
//   - API checks, the API check defaults of groups, and DNS monitors: all.
//   - URL monitors: numeric.
//   - TCP monitors: text for RESPONSE_DATA, numeric for RESPONSE_TIME.
//   - ICMP monitors: numeric for LATENCY; text, IS_NULL and NOT_NULL for
//     JSON_RESPONSE.
//   - Traceroute monitors: numeric for RESPONSE_TIME; EQUALS, GREATER_THAN
//     and LESS_THAN for HOP_COUNT and PACKET_LOSS.
//   - gRPC monitors: numeric for RESPONSE_TIME and GRPC_STATUS_CODE; EQUALS
//     and NOT_EQUALS for GRPC_HEALTHCHECK_STATUS; all for GRPC_RESPONSE,
//     TEXT_BODY and GRPC_METADATA.
//   - SSL monitors: numeric for RESPONSE_TIME; text for TEXT_RESPONSE; text,
//     IS_NULL and NOT_NULL for JSON_RESPONSE; for CERTIFICATE and CONNECTION,
//     a set that depends on the Property.
type AssertionComparison string

// Assertion comparisons, the values of [Assertion].Comparison.
const (
	AssertionComparisonEquals      AssertionComparison = "EQUALS"
	AssertionComparisonNotEquals   AssertionComparison = "NOT_EQUALS"
	AssertionComparisonHasKey      AssertionComparison = "HAS_KEY"
	AssertionComparisonNotHasKey   AssertionComparison = "NOT_HAS_KEY"
	AssertionComparisonHasValue    AssertionComparison = "HAS_VALUE"
	AssertionComparisonNotHasValue AssertionComparison = "NOT_HAS_VALUE"
	AssertionComparisonIsEmpty     AssertionComparison = "IS_EMPTY"
	AssertionComparisonNotEmpty    AssertionComparison = "NOT_EMPTY"
	AssertionComparisonGreaterThan AssertionComparison = "GREATER_THAN"
	AssertionComparisonLessThan    AssertionComparison = "LESS_THAN"
	AssertionComparisonContains    AssertionComparison = "CONTAINS"
	AssertionComparisonNotContains AssertionComparison = "NOT_CONTAINS"
	AssertionComparisonIsNull      AssertionComparison = "IS_NULL"
	AssertionComparisonNotNull     AssertionComparison = "NOT_NULL"
)

// Assertion represents an assertion about an API response, which will be
// verified as part of the check.
type Assertion struct {
	Source     AssertionSource     `json:"source"`
	Property   string              `json:"property"`
	Comparison AssertionComparison `json:"comparison"`
	Target     string              `json:"target"`
	// Regex, when set on a HEADERS assertion, is matched against the header
	// value, and the assertion compares the first capture group instead of
	// the whole value; a value it does not match compares as empty. Empty
	// compares the whole value.
	Regex string `json:"regex"`
}
