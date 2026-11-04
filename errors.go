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
// for ErrAccountNotAccessible, its error code), also when wrapped:
//
//	if errors.Is(err, checkly.ErrNotFound) { ... }
var (
	// ErrNotFound matches a 404 response.
	ErrNotFound = errors.New("not found")
	// ErrUnauthorized matches a 401 response: the API key is missing or
	// invalid.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden matches any 403 response.
	ErrForbidden = errors.New("forbidden")
	// ErrAccountNotAccessible matches a 403 response with the error code
	// ERR_ACCOUNT_NOT_ACCESSIBLE: the API key has no access to the account
	// selected with SetAccountId. Such an error also matches ErrForbidden.
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
	// the validation.keys field of the response body.
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
	// retries is how many times a rate-limited request was re-sent.
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
	if e.StatusCode == http.StatusTooManyRequests {
		var parts []string
		if e.retries > 0 {
			parts = append(parts, fmt.Sprintf("retried %d %s", e.retries, plural(e.retries, "time", "times")))
		}
		if e.RetryAfter > 0 {
			parts = append(parts, fmt.Sprintf("retry after %v", ceilSeconds(e.RetryAfter)))
		}
		if len(parts) > 0 {
			fmt.Fprintf(&b, " (%s)", strings.Join(parts, "; "))
		}
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
	Validation struct {
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
			e.Details = parsed.Validation.Keys
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
