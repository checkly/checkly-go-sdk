package checkly_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	checkly "github.com/checkly/checkly-go-sdk"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var wantCheckID = "73d29e72-6540-4bb5-967e-e07fa2c9465e"

var wantCheck = checkly.Check{
	Name:      "test",
	Type:      checkly.CheckTypeAPI,
	Frequency: 10,
	Activated: true,
	Muted:     false,
	RetryStrategy: &checkly.RetryStrategy{
		Type:               "FIXED",
		MaxRetries:         1,
		MaxDurationSeconds: 600,
	},
	ShouldFail:       false,
	Locations:        []string{"eu-west-1"},
	PrivateLocations: &[]string{},
	Request: checkly.Request{
		Method: http.MethodGet,
		URL:    "https://example.com",
		Headers: []checkly.KeyValue{
			{
				Key:   "X-Test",
				Value: "foo",
			},
		},
		QueryParameters: []checkly.KeyValue{
			{
				Key:   "query",
				Value: "foo",
			},
		},
		Assertions: []checkly.Assertion{
			{
				Source:     checkly.AssertionSourceStatusCode,
				Comparison: checkly.AssertionComparisonEquals,
				Target:     "200",
			},
		},
		Body:     "",
		BodyType: "NONE",
		BasicAuth: &checkly.BasicAuth{
			Username: "",
			Password: "",
		},
		IPFamily: "IPv4",
	},
	// Script and EnvironmentVariables are left unset: the API check endpoint
	// (/v1/checks/api) does not accept them and drops them from the payload,
	// so an API check never stores them.
	Tags: []string{
		"foo",
		"bar",
	},
	SSLCheckDomain:      "example.com",
	LocalSetupScript:    "setitup",
	LocalTearDownScript: "tearitdown",
	AlertSettings: checkly.AlertSettings{
		EscalationType: checkly.EscalationTypeRunBased,
		RunBasedEscalation: checkly.RunBasedEscalation{
			FailedRunThreshold: 1,
		},
		TimeBasedEscalation: checkly.TimeBasedEscalation{
			MinutesFailingThreshold: 5,
		},
		Reminders: checkly.Reminders{
			Interval: 5,
		},
		ParallelRunFailureThreshold: checkly.ParallelRunFailureThreshold{
			Enabled:    false,
			Percentage: 10,
		},
	},
	UseGlobalAlertSettings:    false,
	DegradedResponseTime:      15000,
	MaxResponseTime:           30000,
	GroupID:                   0,
	GroupOrder:                0,
	AlertChannelSubscriptions: nil,
}

// TODO: adjust wantCheck to test multiple SSLCheckDomain value and remove it from this list
var ignoreCheckFields = cmpopts.IgnoreFields(checkly.Check{}, "ID", "AlertChannelSubscriptions", "FrequencyOffset",
	"PrivateLocations", "SSLCheckDomain", "CreatedAt", "UpdatedAt")

func TestGetCheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/checks/%s", wantCheckID),
		validateEmptyBody,
		http.StatusOK,
		"GetCheck.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	gotCheck, err := client.GetCheck(context.Background(), wantCheckID)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(wantCheck, *gotCheck, ignoreCheckFields) {
		t.Error(cmp.Diff(wantCheck, *gotCheck, ignoreCheckFields))
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

// checkBodyServer answers a GET of /v1/checks/{id}{query} with a check of
// the given checkType.
func checkBodyServer(t *testing.T, id, query, checkType string) *httptest.Server {
	t.Helper()
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if want := "/v1/checks/" + id + query; r.Method != http.MethodGet || r.URL.String() != want {
			t.Errorf("want GET %q, got %s %q", want, r.Method, r.URL.String())
		}
		_, _ = fmt.Fprintf(w, `{"id":%q,"checkType":%q}`, id, checkType)
	}))
	t.Cleanup(ts.Close)
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
