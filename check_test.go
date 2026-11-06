package checkly_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/checkly/checkly-go-sdk/v2"
	"github.com/google/go-cmp/cmp"
)

var wantCheckID = "73d29e72-6540-4bb5-967e-e07fa2c9465e"

func TestGetCheck(t *testing.T) {
	t.Parallel()
	const id = "check-id"
	tests := []struct {
		checkType string
		// field is a JSON field specific to the type, which proves the
		// response was decoded into it.
		field string
		want  checkly.Check
	}{
		{"API", `"request":{"url":"https://example.com"}`, &checkly.APICheck{ID: id, Request: checkly.Request{URL: "https://example.com"}}},
		{"BROWSER", `"script":"browser"`, &checkly.BrowserCheck{ID: id, Script: "browser"}},
		{"MULTI_STEP", `"script":"multi-step"`, &checkly.MultiStepCheck{ID: id, Script: "multi-step"}},
		{"PLAYWRIGHT", `"codeBundlePath":"bundle.tar.gz"`, &checkly.PlaywrightCheck{ID: id, CodeBundlePath: "bundle.tar.gz"}},
		{"HEARTBEAT", `"heartbeat":{"period":5}`, &checkly.HeartbeatMonitor{ID: id, Heartbeat: checkly.Heartbeat{Period: 5}}},
		{"URL", `"request":{"url":"https://example.com"}`, &checkly.URLMonitor{ID: id, Request: checkly.URLRequest{URL: "https://example.com"}}},
		{"TCP", `"request":{"hostname":"example.com"}`, &checkly.TCPMonitor{ID: id, Request: checkly.TCPRequest{Hostname: "example.com"}}},
		{"DNS", `"request":{"query":"example.com"}`, &checkly.DNSMonitor{ID: id, Request: checkly.DNSRequest{Query: "example.com"}}},
		{"ICMP", `"request":{"hostname":"example.com"}`, &checkly.ICMPMonitor{ID: id, Request: checkly.ICMPRequest{Hostname: "example.com"}}},
		{"GRPC", `"request":{"url":"example.com"}`, &checkly.GRPCMonitor{ID: id, Request: checkly.GRPCRequest{URL: "example.com"}}},
		{"SSL", `"request":{"sslConfig":{"hostname":"example.com"}}`, &checkly.SSLMonitor{ID: id, Request: checkly.SSLRequest{SSLConfig: checkly.SSLConfig{Hostname: "example.com"}}}},
		{"TRACEROUTE", `"request":{"url":"example.com"}`, &checkly.TracerouteMonitor{ID: id, Request: checkly.TracerouteRequest{URL: "example.com"}}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.checkType, func(t *testing.T) {
			t.Parallel()
			body := fmt.Sprintf(`{"id":%q,"checkType":%q,%s}`, id, tt.checkType, tt.field)
			ts, requests := checkServer(t, id, "?includeDependencies=true", body)
			client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
			got, err := client.GetCheck(context.Background(), id)
			if err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(tt.want, got) {
				t.Error(cmp.Diff(tt.want, got))
			}
			if n := requests.Load(); n != 1 {
				t.Errorf("want 1 request, got %d", n)
			}
		})
	}
}

// TestUnsupportedCheckTypeErrorOmitsQuery checks that the error text shows the
// path without its query string, as other errors do, while ID stays as given.
func TestUnsupportedCheckTypeErrorOmitsQuery(t *testing.T) {
	t.Parallel()
	c := newErrorServer(t, errorReply{status: 200, body: `{"checkType":"AGENTIC"}`})
	_, err := c.GetCheck(context.Background(), "abc?x=1")
	var typeErr *checkly.UnsupportedCheckTypeError
	if !errors.As(err, &typeErr) {
		t.Fatalf("want *UnsupportedCheckTypeError, got %v", err)
	}
	if typeErr.ID != "abc?x=1" {
		t.Errorf("want ID %q, got %q", "abc?x=1", typeErr.ID)
	}
	want := `GET /v1/checks/abc: check is of unsupported type "AGENTIC"`
	if err.Error() != want {
		t.Errorf("want error %q, got %q", want, err.Error())
	}
}

// TestGetCheckHeaderDecodeError checks that a response whose checkType cannot
// be decoded is a decode error, not an unsupported check type, and that the
// error leaves the response body out.
func TestGetCheckHeaderDecodeError(t *testing.T) {
	t.Parallel()
	for name, body := range map[string]string{
		"wrongly typed checkType": `{"checkType":5,"name":"SECRET-RESPONSE-MARKER"}`,
		"not JSON":                `SECRET-RESPONSE-MARKER`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			c := newErrorServer(t, errorReply{status: 200, body: body})
			got, err := c.GetCheck(context.Background(), "abc")
			if err == nil {
				t.Fatal("want decode error, got nil")
			}
			if got != nil {
				t.Errorf("want no check, got %#v", got)
			}
			if errors.Is(err, checkly.ErrUnsupportedCheckType) {
				t.Errorf("want no match with ErrUnsupportedCheckType, got %v", err)
			}
			want := fmt.Sprintf("GET /v1/checks/abc: decoding 200 response (%d bytes): ", len(body))
			if !strings.HasPrefix(err.Error(), want) {
				t.Errorf("want error starting with %q, got %q", want, err)
			}
			if strings.Contains(err.Error(), "SECRET-RESPONSE-MARKER") {
				t.Errorf("want the response body left out of the error, got %q", err)
			}
		})
	}
}

func TestGetCheckRejectsUnsupportedTypes(t *testing.T) {
	t.Parallel()
	const id = "check-id"
	for _, checkType := range []checkly.CheckType{checkly.CheckTypeAgentic, "NEW_TYPE", ""} {
		checkType := checkType
		t.Run(string(checkType), func(t *testing.T) {
			t.Parallel()
			ts := checkBodyServer(t, id, "?includeDependencies=true", string(checkType))
			client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
			got, err := client.GetCheck(context.Background(), id)
			if got != nil {
				t.Errorf("want no check, got %#v", got)
			}
			if !errors.Is(err, checkly.ErrUnsupportedCheckType) {
				t.Fatalf("want an error matching ErrUnsupportedCheckType, got %v", err)
			}
			var typeErr *checkly.UnsupportedCheckTypeError
			if !errors.As(err, &typeErr) {
				t.Fatalf("want *UnsupportedCheckTypeError, got %v", err)
			}
			if typeErr.ID != id || typeErr.Type != checkType {
				t.Errorf("want ID %q and type %q, got %#v", id, checkType, typeErr)
			}
			var apiErr *checkly.APIError
			if errors.As(err, &apiErr) {
				t.Errorf("want no *APIError, got %v", apiErr)
			}
			want := fmt.Sprintf("GET /v1/checks/%s: check is of unsupported type %q", id, checkType)
			if err.Error() != want {
				t.Errorf("want error %q, got %q", want, err.Error())
			}
		})
	}
}

func TestDeleteCheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/checks/%s", wantCheckID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	err := client.DeleteCheck(context.Background(), wantCheckID)
	if err != nil {
		t.Error(err)
	}
}

// typedCheckGetters lists every getter that only accepts a check of its own
// type, with the checkType it expects and the query it sends.
var typedCheckGetters = []struct {
	name      string
	checkType string
	query     string
	// get calls the getter and returns the decoded check's ID.
	get func(c checkly.Client, id string) (string, error)
}{
	{"GetAPICheck", "API", "", getID(checkly.Client.GetAPICheck)},
	{"GetBrowserCheck", "BROWSER", "?includeDependencies=true", getID(checkly.Client.GetBrowserCheck)},
	{"GetMultiStepCheck", "MULTI_STEP", "?includeDependencies=true", getID(checkly.Client.GetMultiStepCheck)},
	{"GetPlaywrightCheck", "PLAYWRIGHT", "", getID(checkly.Client.GetPlaywrightCheck)},
	{"GetHeartbeatMonitor", "HEARTBEAT", "", getID(checkly.Client.GetHeartbeatMonitor)},
	{"GetURLMonitor", "URL", "", getID(checkly.Client.GetURLMonitor)},
	{"GetTCPMonitor", "TCP", "", getID(checkly.Client.GetTCPMonitor)},
	{"GetDNSMonitor", "DNS", "", getID(checkly.Client.GetDNSMonitor)},
	{"GetICMPMonitor", "ICMP", "", getID(checkly.Client.GetICMPMonitor)},
	{"GetGRPCMonitor", "GRPC", "", getID(checkly.Client.GetGRPCMonitor)},
	{"GetTracerouteMonitor", "TRACEROUTE", "", getID(checkly.Client.GetTracerouteMonitor)},
	{"GetSSLMonitor", "SSL", "", getID(checkly.Client.GetSSLMonitor)},
}

// getID adapts a getter to return only the ID field of the check it decodes.
func getID[T any](get func(checkly.Client, context.Context, string) (*T, error)) func(checkly.Client, string) (string, error) {
	return func(c checkly.Client, id string) (string, error) {
		r, err := get(c, context.Background(), id)
		if err != nil {
			return "", err
		}
		return reflect.ValueOf(r).Elem().FieldByName("ID").String(), nil
	}
}

// checkServer answers a GET of /v1/checks/{id}{query} with body, and counts
// the requests it receives.
func checkServer(t *testing.T, id, query, body string) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	requests := &atomic.Int32{}
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if want := "/v1/checks/" + id + query; r.Method != http.MethodGet || r.URL.String() != want {
			t.Errorf("want GET %q, got %s %q", want, r.Method, r.URL.String())
		}
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(ts.Close)
	return ts, requests
}

// checkBodyServer answers a GET of /v1/checks/{id}{query} with a check of
// the given checkType.
func checkBodyServer(t *testing.T, id, query, checkType string) *httptest.Server {
	t.Helper()
	ts, _ := checkServer(t, id, query, fmt.Sprintf(`{"id":%q,"checkType":%q}`, id, checkType))
	return ts
}

func TestGetPerTypeCheckRejectsOtherCheckTypes(t *testing.T) {
	t.Parallel()
	// Each getter is served a check of another type and must refuse it
	// instead of decoding a mostly empty check.
	const id = "check-id"
	for _, tt := range typedCheckGetters {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			other := "API"
			if tt.checkType == other {
				other = "BROWSER"
			}
			ts := checkBodyServer(t, id, tt.query, other)
			client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
			_, err := tt.get(client, id)
			if !errors.Is(err, checkly.ErrCheckTypeMismatch) {
				t.Fatalf("want an error matching ErrCheckTypeMismatch, got %v", err)
			}
			var apiErr *checkly.APIError
			if errors.As(err, &apiErr) {
				t.Errorf("want no *APIError, got %v", apiErr)
			}
			want := fmt.Sprintf("GET /v1/checks/%s: check is of type %q, not %q", id, other, tt.checkType)
			if err.Error() != want {
				t.Errorf("want error %q, got %q", want, err.Error())
			}
		})
	}
}

func TestGetPerTypeCheckAcceptsOwnCheckType(t *testing.T) {
	t.Parallel()
	const id = "check-id"
	for _, tt := range typedCheckGetters {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ts := checkBodyServer(t, id, tt.query, tt.checkType)
			client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
			got, err := tt.get(client, id)
			if err != nil {
				t.Fatal(err)
			}
			if got != id {
				t.Errorf("want ID %q, got %q", id, got)
			}
		})
	}
}
