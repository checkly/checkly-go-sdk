package checkly

// Assertion source constants

// StatusCode identifies the HTTP status code as an assertion source.
const StatusCode = "STATUS_CODE"

// JSONBody identifies the JSON body data as an assertion source.
const JSONBody = "JSON_BODY"

// TextBody identifies the response body text as an assertion source.
const TextBody = "TEXT_BODY"

// Headers identifies the HTTP headers as an assertion source.
const Headers = "HEADERS"

// ResponseTime identifies the response time as an assertion source.
const ResponseTime = "RESPONSE_TIME"

// ResponseData identifies the response data of a TCP check as an assertion source.
const ResponseData = "RESPONSE_DATA"

// Certificate identifies a certificate field selector (via the assertion
// property) as an assertion source, for use with an SSL monitor.
const Certificate = "CERTIFICATE"

// Connection identifies a connection/handshake field selector (via the
// assertion property) as an assertion source, for use with an SSL monitor.
const Connection = "CONNECTION"

// JSONResponse identifies a JSONPath over the response (via the assertion
// property) as an assertion source, for use with SSL, ICMP and DNS monitors.
const JSONResponse = "JSON_RESPONSE"

// TextResponse identifies a regex applied to the serialized response (via the
// assertion property) as an assertion source, for use with an SSL monitor.
const TextResponse = "TEXT_RESPONSE"

// Assertion comparison constants

// Equals asserts that the source and target are equal.
const Equals = "EQUALS"

// NotEquals asserts that the source and target are not equal.
const NotEquals = "NOT_EQUALS"

// IsEmpty asserts that the source is empty.
const IsEmpty = "IS_EMPTY"

// NotEmpty asserts that the source is not empty.
const NotEmpty = "NOT_EMPTY"

// GreaterThan asserts that the source is greater than the target.
const GreaterThan = "GREATER_THAN"

// LessThan asserts that the source is less than the target.
const LessThan = "LESS_THAN"

// Contains asserts that the source contains a specified value.
const Contains = "CONTAINS"

// NotContains asserts that the source does not contain a specified value.
const NotContains = "NOT_CONTAINS"

// IsNull asserts that the source is null.
const IsNull = "IS_NULL"

// NotNull asserts that the source is not null.
const NotNull = "NOT_NULL"

// Assertion represents an assertion about an API response, which will be
// verified as part of the check.
type Assertion struct {
	Edit          bool   `json:"edit"`
	Order         int    `json:"order"`
	ArrayIndex    int    `json:"arrayIndex"`
	ArraySelector int    `json:"arraySelector"`
	Source        string `json:"source"`
	Property      string `json:"property"`
	Comparison    string `json:"comparison"`
	Target        string `json:"target"`
}
