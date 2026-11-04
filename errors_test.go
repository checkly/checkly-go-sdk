package checkly_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	checkly "github.com/checkly/checkly-go-sdk"
)

// errorReply is a canned response for the error tests.
type errorReply struct {
	status int
	header http.Header
	body   string
}

// newErrorServer answers every request with reply and returns a client for
// it. The client is set up to skip the waits between rate-limit retries.
func newErrorServer(t *testing.T, reply errorReply) checkly.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		for k, v := range reply.header {
			w.Header()[k] = v
		}
		w.WriteHeader(reply.status)
		io.WriteString(w, reply.body)
	}))
	t.Cleanup(srv.Close)
	c := checkly.NewClient(srv.URL, "test-key", srv.Client(), nil)
	checkly.SkipRetryWaits(c)
	return c
}

func readFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("fixtures/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestErrorMessages(t *testing.T) {
	t.Parallel()
	html := "<html>\n<head><title>502 Bad Gateway</title></head>\n<body>\n" +
		strings.Repeat("<center>nginx</center>\n", 20) + "</body>\n</html>\n"
	tests := []struct {
		name      string
		accountID string
		reply     errorReply
		want      string
		is        []error
		isNot     []error
	}{
		{
			name:  "boom 404",
			reply: errorReply{status: 404, body: `{"statusCode":404,"error":"Not Found","message":"Not Found"}`},
			want:  "GET /v1/checks/abc: 404 Not Found",
			is:    []error{checkly.ErrNotFound},
			isNot: []error{checkly.ErrForbidden, checkly.ErrConflict},
		},
		{
			name:  "validation error with keys",
			reply: errorReply{status: 400, body: readFixture(t, "BadRequest.json")},
			want:  `GET /v1/checks/abc: 400 Bad Request: child "frequency" fails because ["frequency" must be one of [5, 10, 15, 30, 60, 720, 1440]] (keys: frequency)`,
			isNot: []error{checkly.ErrNotFound},
		},
		{
			name:  "bad token",
			reply: errorReply{status: 401, body: `{"statusCode":401,"error":"Unauthorized","message":"Bad token"}`},
			want:  "GET /v1/checks/abc: 401 Unauthorized: Bad token. Check the API key.",
			is:    []error{checkly.ErrUnauthorized},
		},
		{
			name:  "plan limit",
			reply: errorReply{status: 402, body: `{"statusCode":402,"error":"Payment Required","message":"You have reached the maximum number of API checks on your plan."}`},
			want:  "GET /v1/checks/abc: 402 Payment Required: You have reached the maximum number of API checks on your plan.",
			is:    []error{checkly.ErrPlanLimit},
		},
		{
			// Every refusal for an inaccessible account also carries the
			// refused-by header; the error code decides the hint.
			name:      "account not accessible",
			accountID: "acc-123",
			reply: errorReply{
				status: 403,
				header: http.Header{"X-Checkly-Refused-By": {"permission"}},
				body:   `{"statusCode":403,"error":"Forbidden","message":"The selected account is not accessible with the provided credentials.","errorCode":"ERR_ACCOUNT_NOT_ACCESSIBLE"}`,
			},
			want:  `GET /v1/checks/abc: 403 Forbidden [ERR_ACCOUNT_NOT_ACCESSIBLE]: The selected account is not accessible with the provided credentials. Check the account ID "acc-123".`,
			is:    []error{checkly.ErrAccountNotAccessible, checkly.ErrForbidden},
			isNot: []error{checkly.ErrUnauthorized},
		},
		{
			name: "account not accessible without an account ID",
			reply: errorReply{
				status: 403,
				header: http.Header{"X-Checkly-Refused-By": {"permission"}},
				body:   `{"statusCode":403,"error":"Forbidden","message":"The selected account is not accessible with the provided credentials.","errorCode":"ERR_ACCOUNT_NOT_ACCESSIBLE"}`,
			},
			want: "GET /v1/checks/abc: 403 Forbidden [ERR_ACCOUNT_NOT_ACCESSIBLE]: The selected account is not accessible with the provided credentials.",
			is:   []error{checkly.ErrAccountNotAccessible, checkly.ErrForbidden},
		},
		{
			name: "refused by role",
			reply: errorReply{
				status: 403,
				header: http.Header{"X-Checkly-Refused-By": {"permission"}},
				body:   `{"statusCode":403,"error":"Forbidden","message":"Forbidden"}`,
			},
			want:  "GET /v1/checks/abc: 403 Forbidden. The API key's role does not permit this action.",
			is:    []error{checkly.ErrForbidden},
			isNot: []error{checkly.ErrAccountNotAccessible},
		},
		{
			// Only ERR_ACCOUNT_NOT_ACCESSIBLE gets a hint; another error
			// code still suppresses the role hint of the refused-by header.
			name: "refused with another error code",
			reply: errorReply{
				status: 403,
				header: http.Header{"X-Checkly-Refused-By": {"permission"}},
				body:   `{"statusCode":403,"error":"Forbidden","message":"Not allowed.","errorCode":"ERR_SOMETHING_ELSE"}`,
			},
			want:  "GET /v1/checks/abc: 403 Forbidden [ERR_SOMETHING_ELSE]: Not allowed.",
			is:    []error{checkly.ErrForbidden},
			isNot: []error{checkly.ErrAccountNotAccessible},
		},
		{
			name:  "forbidden without a hint",
			reply: errorReply{status: 403, body: `{"statusCode":403,"error":"Forbidden","message":"Your account is blocked."}`},
			want:  "GET /v1/checks/abc: 403 Forbidden: Your account is blocked.",
			is:    []error{checkly.ErrForbidden},
		},
		{
			name:  "conflict",
			reply: errorReply{status: 409, body: `{"statusCode":409,"error":"Conflict","message":"The environment variable already exists"}`},
			want:  "GET /v1/checks/abc: 409 Conflict: The environment variable already exists",
			is:    []error{checkly.ErrConflict},
		},
		{
			name: "rate limited after the retries",
			reply: errorReply{
				status: 429,
				header: http.Header{"Retry-After": {"12"}},
				body:   `{"statusCode":429,"error":"Too Many Requests","message":"Rate limit exceeded"}`,
			},
			want: "GET /v1/checks/abc: 429 Too Many Requests: Rate limit exceeded (retried 4 times; retry after 12s)",
			is:   []error{checkly.ErrRateLimited},
		},
		{
			// The client caps its own waits, but reports the server's hint
			// as it is.
			name: "rate limited with a long hint",
			reply: errorReply{
				status: 429,
				header: http.Header{"Retry-After": {"120"}},
				body:   `{"statusCode":429,"error":"Too Many Requests","message":"Rate limit exceeded"}`,
			},
			want: "GET /v1/checks/abc: 429 Too Many Requests: Rate limit exceeded (retried 4 times; retry after 2m0s)",
			is:   []error{checkly.ErrRateLimited},
		},
		{
			name:  "boom 500",
			reply: errorReply{status: 500, body: `{"statusCode":500,"error":"Internal Server Error","message":"An internal server error occurred"}`},
			want:  "GET /v1/checks/abc: 500 Internal Server Error: An internal server error occurred",
			isNot: []error{checkly.ErrNotFound, checkly.ErrForbidden, checkly.ErrRateLimited},
		},
		{
			name:  "html from a proxy is collapsed and truncated",
			reply: errorReply{status: 502, header: http.Header{"Content-Type": {"text/html"}}, body: html},
			want: "GET /v1/checks/abc: 502 Bad Gateway: " +
				(`<html> <head><title>502 Bad Gateway</title></head> <body> ` + strings.Repeat("<center>nginx</center> ", 20))[:500] +
				"… (truncated)",
		},
		{
			// A JSON message is collapsed and truncated like a body that is
			// not JSON, without splitting the multi-byte "é" at the cut.
			name:  "long json message",
			reply: errorReply{status: 400, body: `{"statusCode":400,"error":"Bad Request","message":"` + strings.Repeat("x", 499) + `é tail\n\n  more"}`},
			want:  "GET /v1/checks/abc: 400 Bad Request: " + strings.Repeat("x", 499) + "… (truncated)",
		},
		{
			name:  "json message whitespace",
			reply: errorReply{status: 402, body: `{"statusCode":402,"error":"Payment Required","message":"[a] over the limit\n[b]  over the limit"}`},
			want:  "GET /v1/checks/abc: 402 Payment Required: [a] over the limit [b] over the limit",
			is:    []error{checkly.ErrPlanLimit},
		},
		{
			name:  "plain text",
			reply: errorReply{status: 413, header: http.Header{"Content-Type": {"text/plain"}}, body: "Payload content length\n  greater than maximum allowed\n"},
			want:  "GET /v1/checks/abc: 413 Request Entity Too Large: Payload content length greater than maximum allowed",
		},
		{
			name:  "json with a reason but no message",
			reply: errorReply{status: 418, body: `{"statusCode":418,"error":"Short And Stout"}`},
			want:  "GET /v1/checks/abc: 418 Short And Stout",
		},
		{
			name:  "empty body",
			reply: errorReply{status: 404},
			want:  "GET /v1/checks/abc: 404 Not Found",
			is:    []error{checkly.ErrNotFound},
		},
		{
			name:  "unexpected success status",
			reply: errorReply{status: 201, body: `{"name":"SECRET-RESPONSE-MARKER"}`},
			want:  "GET /v1/checks/abc: unexpected 201 Created response, want 200 OK",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := newErrorServer(t, tt.reply)
			if tt.accountID != "" {
				c.SetAccountId(tt.accountID)
			}
			_, err := c.Get(context.Background(), "abc")
			if err == nil {
				t.Fatal("want error, got nil")
			}
			if err.Error() != tt.want {
				t.Errorf("want error\n%q\ngot\n%q", tt.want, err.Error())
			}
			var apiErr *checkly.APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != tt.reply.status {
				t.Fatalf("want *APIError with status %d, got %#v", tt.reply.status, err)
			}
			// A success body is a resource, which may hold secrets, not an
			// error message.
			if tt.reply.status < 300 && apiErr.Message != "" {
				t.Errorf("want no Message for a success status, got %q", apiErr.Message)
			}
			// The sentinels also match through wrapping.
			wrapped := fmt.Errorf("reading check: %w", err)
			for _, target := range tt.is {
				if !errors.Is(wrapped, target) {
					t.Errorf("want errors.Is(err, %v)", target)
				}
			}
			for _, target := range tt.isNot {
				if errors.Is(wrapped, target) {
					t.Errorf("want !errors.Is(err, %v)", target)
				}
			}
			if got, want := checkly.IsNotFound(wrapped), tt.reply.status == 404; got != want {
				t.Errorf("want IsNotFound %v, got %v", want, got)
			}
		})
	}
}

func TestAPIErrorFields(t *testing.T) {
	t.Parallel()
	c := newErrorServer(t, errorReply{
		status: 429,
		header: http.Header{"Retry-After": {"12"}},
		body:   `{"statusCode":429,"error":"Too Many Requests","message":"Rate limit exceeded"}`,
	})
	_, err := c.Get(context.Background(), "abc?x=1")
	var apiErr *checkly.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("want *APIError, got %v", err)
	}
	if apiErr.Method != "GET" || apiErr.Path != "/v1/checks/abc" || apiErr.Message != "Rate limit exceeded" || apiErr.RetryAfter != 12*time.Second {
		t.Errorf("unexpected fields %#v", apiErr)
	}
}

// TestAPIErrorLiteral checks that an APIError built by a caller, as in tests
// that fake the client, renders and classifies sensibly.
func TestAPIErrorLiteral(t *testing.T) {
	t.Parallel()
	err := &checkly.APIError{StatusCode: 404}
	if got, want := err.Error(), "404 Not Found"; got != want {
		t.Errorf("want %q, got %q", want, got)
	}
	if !checkly.IsNotFound(fmt.Errorf("x: %w", err)) {
		t.Error("want IsNotFound")
	}
	if checkly.IsNotFound(&checkly.APIError{StatusCode: 400, Message: "check 404 not found"}) {
		t.Error("want a 400 not to match IsNotFound, whatever its message")
	}
	if checkly.IsNotFound(errors.New("GET /v1/checks/404: 500 Internal Server Error")) {
		t.Error("want an untyped error not to match IsNotFound")
	}
	if checkly.IsNotFound(nil) {
		t.Error("want nil not to match IsNotFound")
	}
}

func TestDecodeErrorOmitsBody(t *testing.T) {
	t.Parallel()
	c := newErrorServer(t, errorReply{status: 200, body: `{"frequency":"SECRET-RESPONSE-MARKER"}`})
	_, err := c.Get(context.Background(), "abc")
	if err == nil {
		t.Fatal("want decode error, got nil")
	}
	msg := err.Error()
	if strings.Contains(msg, "SECRET-RESPONSE-MARKER") {
		t.Errorf("want the response body left out of the error, got %q", msg)
	}
	if want := "GET /v1/checks/abc: decoding 200 response (38 bytes): "; !strings.HasPrefix(msg, want) {
		t.Errorf("want error starting with %q, got %q", want, msg)
	}
	var apiErr *checkly.APIError
	if errors.As(err, &apiErr) {
		t.Errorf("want a decode error not to be an *APIError, got %v", apiErr)
	}
}

// TestRateLimitRetryAfter checks that RetryAfter is the server's uncapped
// hint in each header form, and that the error text rounds it up to whole
// seconds.
func TestRateLimitRetryAfter(t *testing.T) {
	t.Parallel()
	inTwoMinutes := time.Now().Add(2 * time.Minute)
	tests := []struct {
		name     string
		header   http.Header
		min, max time.Duration
		text     string
	}{
		{"retry-after seconds", http.Header{"Retry-After": {"120"}}, 120 * time.Second, 120 * time.Second, "retry after 2m0s"},
		{"retry-after date", http.Header{"Retry-After": {inTwoMinutes.UTC().Format(http.TimeFormat)}}, 110 * time.Second, 120 * time.Second, "retry after "},
		{"reset", http.Header{"X-Ratelimit-Reset": {strconv.FormatInt(inTwoMinutes.Unix(), 10)}}, 110 * time.Second, 120 * time.Second, "retry after "},
		// A wait under a second is rounded up, not shown as "0s".
		{"short reset", http.Header{"X-Ratelimit-Reset": {strconv.FormatInt(time.Now().Unix()+1, 10)}}, time.Nanosecond, time.Second, "retry after 1s)"},
		// A hint centuries ahead saturates at the maximum Duration, which
		// cannot be rounded up to whole seconds without overflowing.
		{"far-future reset", http.Header{"X-Ratelimit-Reset": {"99999999999999"}}, math.MaxInt64, math.MaxInt64, "retry after 2562047h47m16.854775807s)"},
		{"far-future retry-after date", http.Header{"Retry-After": {"Fri, 31 Dec 9999 23:59:59 GMT"}}, math.MaxInt64, math.MaxInt64, "retry after 2562047h47m16.854775807s)"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := newErrorServer(t, errorReply{status: 429, header: tt.header})
			_, err := c.Get(context.Background(), "abc")
			var apiErr *checkly.APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("want *APIError, got %v", err)
			}
			if apiErr.RetryAfter < tt.min || apiErr.RetryAfter > tt.max {
				t.Errorf("want RetryAfter within [%v, %v], got %v", tt.min, tt.max, apiErr.RetryAfter)
			}
			if !strings.Contains(err.Error(), tt.text) || strings.Contains(err.Error(), "retry after 0s") {
				t.Errorf("want error containing %q, got %q", tt.text, err)
			}
		})
	}
}

// TestErrorOmitsSecrets checks that the SDK puts neither the API key nor the
// request payload into an error. When the server itself echoes them in its
// error body, they reach the error only as part of the backend's message.
func TestErrorOmitsSecrets(t *testing.T) {
	t.Parallel()
	const apiKey = "SECRET-API-KEY-MARKER"
	const payload = "SECRET-PAYLOAD-MARKER"
	secrets := []string{apiKey, payload}
	tests := []struct {
		name string
		// reply builds the error body from the request's Authorization
		// header and body.
		reply func(auth, body string) string
	}{
		{"no echo", func(string, string) string {
			return `{"statusCode":500,"error":"Internal Server Error","message":"An internal server error occurred"}`
		}},
		{"echo as text", func(auth, body string) string {
			return "auth " + auth + " body " + body
		}},
		{"echo in json message", func(auth, body string) string {
			msg, _ := json.Marshal(map[string]any{
				"statusCode": 500,
				"error":      "Internal Server Error",
				"message":    "auth " + auth + " body " + body,
			})
			return string(msg)
		}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				w.WriteHeader(500)
				io.WriteString(w, tt.reply(r.Header.Get("Authorization"), string(body)))
			}))
			t.Cleanup(srv.Close)
			c := checkly.NewClient(srv.URL, apiKey, srv.Client(), nil)
			c.SetAccountId("acc-123")
			_, err := c.CreateCheck(context.Background(), checkly.Check{
				Type:                 checkly.TypeAPI,
				Name:                 payload,
				EnvironmentVariables: []checkly.EnvironmentVariable{{Key: "TOKEN", Value: payload}},
			})
			var apiErr *checkly.APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("want *APIError, got %v", err)
			}
			// The message is capped, which limits how much of an echoed
			// request it can show.
			if max := 500 + len("… (truncated)"); len(apiErr.Message) > max {
				t.Errorf("want Message of at most %d bytes, got %d", max, len(apiErr.Message))
			}
			// Whatever the server echoed is only in the backend's message;
			// the rest of the error text holds no secret.
			rest := strings.Replace(err.Error(), apiErr.Message, "", 1)
			for _, secret := range secrets {
				if strings.Contains(rest, secret) {
					t.Errorf("want %s only inside the backend message, got %q", secret, err)
				}
			}
			if tt.name == "no echo" {
				want := "POST /v1/checks/api: 500 Internal Server Error: An internal server error occurred"
				if err.Error() != want {
					t.Errorf("want %q, got %q", want, err)
				}
			}
		})
	}
}

func TestTransportErrors(t *testing.T) {
	t.Parallel()
	t.Run("closed server", func(t *testing.T) {
		t.Parallel()
		srv := httptest.NewServer(http.NotFoundHandler())
		host := srv.Listener.Addr().String()
		srv.Close()
		c := checkly.NewClient(srv.URL, "test-key", nil, nil)
		err := c.DeleteCheck(context.Background(), "abc")
		want := "DELETE /v1/checks/abc: request to " + host + " failed: "
		if err == nil || !strings.HasPrefix(err.Error(), want) {
			t.Fatalf("want error starting with %q, got %v", want, err)
		}
		var urlErr *url.Error
		if !errors.As(err, &urlErr) {
			t.Errorf("want a *url.Error in the chain, got %v", err)
		}
		var apiErr *checkly.APIError
		if errors.As(err, &apiErr) {
			t.Errorf("want a transport error not to be an *APIError, got %v", apiErr)
		}
	})
	t.Run("cancelled with debug output", func(t *testing.T) {
		t.Parallel()
		srv := httptest.NewServer(http.NotFoundHandler())
		t.Cleanup(srv.Close)
		var debug strings.Builder
		c := checkly.NewClient(srv.URL, "test-key", srv.Client(), &debug)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := c.Get(ctx, "abc")
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("want context.Canceled, got %v", err)
		}
		var urlErr *url.Error
		if !errors.As(err, &urlErr) {
			t.Errorf("want a *url.Error in the chain, got %v", err)
		}
		want := "GET /v1/checks/abc: request to " + srv.Listener.Addr().String() + " failed: context canceled"
		if err.Error() != want {
			t.Errorf("want %q, got %q", want, err)
		}
		if !strings.Contains(debug.String(), "GET /v1/checks/abc") {
			t.Errorf("want the request in the debug output, got %q", debug.String())
		}
	})
	t.Run("deadline", func(t *testing.T) {
		t.Parallel()
		release := make(chan struct{})
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-release:
			case <-r.Context().Done():
			}
		}))
		t.Cleanup(srv.Close)
		t.Cleanup(func() { close(release) })
		c := checkly.NewClient(srv.URL, "test-key", srv.Client(), nil)
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		_, err := c.Get(ctx, "abc?x=1")
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("want context.DeadlineExceeded, got %v", err)
		}
		var urlErr *url.Error
		if !errors.As(err, &urlErr) {
			t.Errorf("want a *url.Error in the chain, got %v", err)
		}
		want := "GET /v1/checks/abc: request to " + srv.Listener.Addr().String() + " failed: context deadline exceeded"
		if err.Error() != want {
			t.Errorf("want %q, got %q", want, err)
		}
	})
}
