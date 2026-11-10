package checkly

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Sentinel errors for the API failures callers commonly branch on. An
// *APIError matches them with errors.Is according to its status code (and,
// for ErrAccountNotAccessible and ErrValidation, its body), also when wrapped:
//
//	if errors.Is(err, checkly.ErrNotFound) { ... }
var (
	// ErrBadRequest matches any 400 response.
	ErrBadRequest = errors.New("bad request")
	// ErrValidation matches a 400 response whose body has a validation block,
	// which the API sends when a request fails schema validation, such as a
	// missing field or a value of the wrong type. [APIError].Details lists
	// the fields it names, if any. Such an error also matches ErrBadRequest.
	// Most other rule failures, such as a date range that ends before it
	// starts, have no block, so they match only ErrBadRequest and only their
	// Message says what is wrong.
	ErrValidation = errors.New("validation failed")
	// ErrNotFound matches a 404 response.
	ErrNotFound = errors.New("not found")
	// ErrUnauthorized matches a 401 response: the server rejected the API
	// key. A client without an API key fails with ErrMissingAPIKey instead,
	// before sending the request.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden matches any 403 response.
	ErrForbidden = errors.New("forbidden")
	// ErrAccountNotAccessible matches a 403 response with the error code
	// ERR_ACCOUNT_NOT_ACCESSIBLE: the API key has no access to the account
	// passed to WithAuth. Such an error also matches ErrForbidden.
	ErrAccountNotAccessible = errors.New("account not accessible")
	// ErrPlanLimit matches a 402 response: the action exceeds the account's
	// plan.
	ErrPlanLimit = errors.New("plan limit reached")
	// ErrConflict matches a 409 response, such as creating a resource that
	// already exists.
	ErrConflict = errors.New("conflict")
	// ErrRateLimited matches a 429 response that was still rate limited after
	// the client's retries.
	ErrRateLimited = errors.New("rate limited")
)

// Sentinel errors for auth that cannot authenticate a request. They are not
// *APIError values and do not match ErrUnauthorized, which means the server
// rejected the key.
var (
	// ErrMissingAPIKey is matched by the error an authenticated method of a
	// client without auth returns, without sending the request, prefixed
	// with the request as in "GET /v1/checks/abc: no API key configured; set
	// one with WithAuth". It also matches the error NewClient returns for
	// WithAuth with an empty API key and for WithAuthFromEnv when
	// CHECKLY_API_KEY is not set.
	ErrMissingAPIKey = errors.New("no API key configured; set one with WithAuth")
	// ErrMissingAccountID is matched by the error NewClient returns for
	// WithAuth with an empty account ID and for WithAuthFromEnv when
	// CHECKLY_ACCOUNT_ID is not set.
	ErrMissingAccountID = errors.New("an API key needs an account ID")
	// ErrInvalidAccountID is matched by the error NewClient returns when
	// WithAuth or WithAuthFromEnv gets an account ID that is not a UUID, such
	// as an API key passed in its place.
	ErrInvalidAccountID = errors.New("the account ID is not a UUID")
)

// ErrCheckTypeMismatch is matched by the error a check or monitor getter, such
// as GetAPICheck or GetURLMonitor, returns when the ID belongs to a check of
// another type, prefixed with the request and naming both types as in
// "GET /v1/checks/abc: check is of type \"BROWSER\", not \"API\"". It is not
// an *APIError: the API answered successfully.
var ErrCheckTypeMismatch = errors.New("check type mismatch")

// ErrUnsupportedCheckType is matched by the [*UnsupportedCheckTypeError] that
// GetCheck returns for a check of a type the SDK has no resource type for.
var ErrUnsupportedCheckType = errors.New("unsupported check type")

// UnsupportedCheckTypeError is returned by GetCheck for a check of a type the
// SDK has no resource type for, such as [CheckTypeAgentic]. It matches
// [ErrUnsupportedCheckType]. It is not an *APIError: the API answered
// successfully.
type UnsupportedCheckTypeError struct {
	// ID is the check's UUID.
	ID string
	// Type is the check type the API returned.
	Type CheckType
}

func (e *UnsupportedCheckTypeError) Error() string {
	return fmt.Sprintf("GET %s: check is of unsupported type %q", errorPath("/v1/checks/"+e.ID), e.Type)
}

// Unwrap returns [ErrUnsupportedCheckType].
func (e *UnsupportedCheckTypeError) Unwrap() error { return ErrUnsupportedCheckType }

// IsNotFound reports whether err is, or wraps, an API error for a 404
// response. It is equivalent to errors.Is(err, ErrNotFound).
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// errorCodeAccountNotAccessible is the backend's error code for a request
// whose API key has no access to the selected account.
const errorCodeAccountNotAccessible = "ERR_ACCOUNT_NOT_ACCESSIBLE"

// maxMessageExcerpt is the number of bytes of an error message kept in
// APIError.Message: the message field of a JSON error body, or a body that is
// not JSON, typically an HTML page from a proxy. Real messages are rarely
// longer, and the cap limits how much of an echoed request the error shows.
const maxMessageExcerpt = 500

// APIError is returned when the API answers with a status the method does not
// accept. Use errors.As to inspect it, or errors.Is with the sentinel errors
// above to classify it.
//
// The error text has the form
//
//	METHOD /v1/path: 404 Not Found[ [ERROR_CODE]][: message][ (keys: ...)]
//
// and may end with a sentence suggesting what to check. It never contains the
// API key, request headers, the request body or, for an unexpected success
// status, the response body. Message is the backend's own error text, which
// for a validation error can quote a submitted value.
type APIError struct {
	// Method is the HTTP method of the request, such as "GET".
	Method string
	// Path is the request path without the query string, such as
	// "/v1/checks/abc".
	Path string
	// StatusCode is the HTTP status of the response.
	StatusCode int
	// Code is the machine-readable error code from the response body, such
	// as "ERR_ACCOUNT_NOT_ACCESSIBLE", if any.
	Code string
	// Message is the error message from the response body, or the body
	// itself if it is not JSON, with whitespace collapsed and truncated to
	// 500 bytes.
	Message string
	// Details lists the request fields a validation error refers to, from
	// the validation.keys field of the response body. It is non-nil, though
	// possibly empty, when the body has a validation block, and that is
	// what [ErrValidation] matches.
	Details []string
	// RetryAfter is how long the server asked the client to wait before
	// retrying a rate-limited (429) request, if it said so in a Retry-After
	// or X-RateLimit-Reset header. It is the server's value, not capped like
	// the client's own waits between retries.
	RetryAfter time.Duration

	// reason is the status reason phrase from the response body, used in
	// place of http.StatusText when set.
	reason string
	// want lists the statuses the method accepts. It is only used to
	// describe an unexpected success status.
	want []int
	// retries is how many times the request was re-sent, after a 429 or a
	// broken connection.
	retries int
	// hint is a sentence suggesting what to check.
	hint string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	var b strings.Builder
	if prefix := strings.TrimSpace(e.Method + " " + e.Path); prefix != "" {
		b.WriteString(prefix)
		b.WriteString(": ")
	}
	if e.StatusCode >= 200 && e.StatusCode < 300 {
		fmt.Fprintf(&b, "unexpected %s response", statusLine(e.StatusCode, ""))
		for i, status := range e.want {
			if i == 0 {
				b.WriteString(", want ")
			} else {
				b.WriteString(" or ")
			}
			b.WriteString(statusLine(status, ""))
		}
		return b.String()
	}
	b.WriteString(statusLine(e.StatusCode, e.reason))
	if e.Code != "" {
		fmt.Fprintf(&b, " [%s]", e.Code)
	}
	if e.Message != "" && !strings.EqualFold(e.Message, e.reasonPhrase()) {
		b.WriteString(": ")
		b.WriteString(e.Message)
	}
	if len(e.Details) > 0 {
		fmt.Fprintf(&b, " (keys: %s)", strings.Join(e.Details, ", "))
	}
	// The retry count explains, for example, a 404 from a re-sent DELETE
	// whose first attempt succeeded but lost its response.
	var parts []string
	if e.retries > 0 {
		parts = append(parts, fmt.Sprintf("retried %d %s", e.retries, plural(e.retries, "time", "times")))
	}
	if e.StatusCode == http.StatusTooManyRequests && e.RetryAfter > 0 {
		parts = append(parts, fmt.Sprintf("retry after %v", ceilSeconds(e.RetryAfter)))
	}
	if len(parts) > 0 {
		fmt.Fprintf(&b, " (%s)", strings.Join(parts, "; "))
	}
	if e.hint != "" {
		// The hint is a sentence of its own. The message usually is one
		// too, so only add a full stop when it lacks one.
		if s := b.String(); !strings.HasSuffix(s, ".") && !strings.HasSuffix(s, "!") && !strings.HasSuffix(s, "?") {
			b.WriteString(".")
		}
		b.WriteString(" ")
		b.WriteString(e.hint)
	}
	return b.String()
}

// Is reports whether e matches one of the sentinel errors of this package.
func (e *APIError) Is(target error) bool {
	switch target {
	case ErrBadRequest:
		return e.StatusCode == http.StatusBadRequest
	case ErrValidation:
		return e.StatusCode == http.StatusBadRequest && e.Details != nil
	case ErrNotFound:
		return e.StatusCode == http.StatusNotFound
	case ErrUnauthorized:
		return e.StatusCode == http.StatusUnauthorized
	case ErrForbidden:
		return e.StatusCode == http.StatusForbidden
	case ErrAccountNotAccessible:
		return e.StatusCode == http.StatusForbidden && e.Code == errorCodeAccountNotAccessible
	case ErrPlanLimit:
		return e.StatusCode == http.StatusPaymentRequired
	case ErrConflict:
		return e.StatusCode == http.StatusConflict
	case ErrRateLimited:
		return e.StatusCode == http.StatusTooManyRequests
	}
	return false
}

// reasonPhrase returns the reason phrase of the response status.
func (e *APIError) reasonPhrase() string {
	if e.reason != "" {
		return e.reason
	}
	return http.StatusText(e.StatusCode)
}

// statusLine renders a status as "404 Not Found", using reason if set.
func statusLine(status int, reason string) string {
	if reason == "" {
		reason = http.StatusText(status)
	}
	return strings.TrimSpace(strconv.Itoa(status) + " " + reason)
}

// ceilSeconds rounds a positive d up to whole seconds, so that a short wait is
// never shown as "0s". A d too close to the maximum Duration to round up, as
// for a hint dated centuries ahead, is returned unchanged.
func ceilSeconds(d time.Duration) time.Duration {
	secs := d / time.Second
	if d%time.Second != 0 {
		secs++
	}
	if secs > math.MaxInt64/time.Second {
		return d
	}
	return secs * time.Second
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// errorBody is the error response body of the API: a Boom error payload,
// optionally with an error code and, for validation errors, the offending
// keys.
type errorBody struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	ErrorCode  string `json:"errorCode"`
	Validation *struct {
		Keys []string `json:"keys"`
	} `json:"validation"`
}

// newAPIError describes res, a response whose status the method does not
// accept. want lists the accepted statuses and accountID is the account the
// client selected, if any; both only feed the error text.
func newAPIError(res rawResponse, want []int, accountID string) *APIError {
	status, header, body := res.status, res.header, res.body
	e := &APIError{
		Method:     res.method,
		Path:       res.path,
		StatusCode: status,
		want:       want,
		retries:    res.retries,
	}
	if status >= 200 && status < 300 {
		// A success body is a resource, not an error message, and may
		// hold secrets.
		return e
	}

	var parsed errorBody
	if trimmed := strings.TrimSpace(string(body)); trimmed != "" {
		if err := json.Unmarshal(body, &parsed); err == nil {
			e.reason = parsed.Error
			e.Code = parsed.ErrorCode
			e.Message = excerpt(parsed.Message)
			if parsed.Validation != nil {
				e.Details = emptyIfNil(parsed.Validation.Keys)
			}
		} else {
			e.Message = excerpt(trimmed)
		}
	}

	switch status {
	case http.StatusUnauthorized:
		e.hint = "Check the API key."
	case http.StatusForbidden:
		// Every refusal for an inaccessible account also carries the
		// refused-by header, so the error code takes precedence.
		if e.Code == errorCodeAccountNotAccessible {
			if accountID != "" {
				e.hint = fmt.Sprintf("Check the account ID %q.", accountID)
			}
		} else if e.Code == "" && header.Get("X-Checkly-Refused-By") == "permission" {
			e.hint = "The API key's role does not permit this action."
		}
	case http.StatusTooManyRequests:
		e.RetryAfter, _ = rateLimitHint(header, time.Now())
	}
	return e
}

// excerpt collapses the whitespace in s and truncates it to
// maxMessageExcerpt bytes, without splitting a UTF-8 sequence.
func excerpt(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= maxMessageExcerpt {
		return s
	}
	cut := maxMessageExcerpt
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + "… (truncated)"
}

// transportError is returned when a request gets no response: the connection
// failed, the context ended, or the like. It names the API host, which helps
// to spot a misconfigured base URL, and unwraps to the *url.Error, so that
// errors.Is(err, context.DeadlineExceeded) and errors.As with a *url.Error
// both work.
type transportError struct {
	method string
	path   string
	host   string
	err    *url.Error
}

func (e *transportError) Error() string {
	return fmt.Sprintf("%s %s: request to %s failed: %v", e.method, e.path, e.host, e.err.Err)
}

func (e *transportError) Unwrap() error {
	return e.err
}

// messageError is an error with its own text that wraps another error, so
// that errors.Is and errors.As find the wrapped one while Error() shows only
// msg. It lets an error match a sentinel with a more specific text, and keeps
// a wrapped error's text, which may quote secrets, out of the message.
type messageError struct {
	msg     string
	wrapped error
}

func (e *messageError) Error() string { return e.msg }

func (e *messageError) Unwrap() error { return e.wrapped }
