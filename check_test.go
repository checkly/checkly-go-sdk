package checkly_test

import (
	"context"
	"encoding/json"
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
	Type:      checkly.TypeAPI,
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
				Source:     checkly.StatusCode,
				Comparison: checkly.Equals,
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
		EscalationType: checkly.RunBased,
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

func validateCheck(t *testing.T, body []byte) {
	var gotCheck checkly.Check
	err := json.Unmarshal(body, &gotCheck)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	// The payload sends unset lists as [], which decodes to an empty slice.
	if !cmp.Equal(wantCheck, gotCheck, cmpopts.EquateEmpty()) {
		t.Error(cmp.Diff(wantCheck, gotCheck, cmpopts.EquateEmpty()))
	}
}

// TODO: adjust wantCheck to test multiple SSLCheckDomain value and remove it from this list
var ignoreCheckFields = cmpopts.IgnoreFields(checkly.Check{}, "ID", "AlertChannelSubscriptions", "FrequencyOffset",
	"PrivateLocations", "SSLCheckDomain", "CreatedAt", "UpdatedAt")

func TestCreateCheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/api?autoAssignAlerts=false",
		validateCheck,
		http.StatusCreated,
		"CreateCheck.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	gotCheck, err := client.CreateCheck(context.Background(), wantCheck)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(wantCheck, *gotCheck, ignoreCheckFields) {
		t.Error(cmp.Diff(wantCheck, *gotCheck, ignoreCheckFields))
	}
}

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

func TestUpdateCheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/%s?autoAssignAlerts=false", wantCheckID),
		validateCheck,
		http.StatusOK,
		"UpdateCheck.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	gotCheck, err := client.UpdateCheck(context.Background(), wantCheckID, wantCheck)
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

// TestCheckScriptAndEnvironmentVariablesPayload verifies that CreateCheck and
// UpdateCheck send a Check's Script and EnvironmentVariables. wantCheck is an
// API check, whose endpoint does not store these fields, so it leaves them
// unset and a browser check covers them here instead.
func TestCheckScriptAndEnvironmentVariablesPayload(t *testing.T) {
	t.Parallel()
	check := checkly.Check{
		Name:      "browser",
		Type:      checkly.TypeBrowser,
		Frequency: 10,
		Locations: []string{"eu-west-1"},
		Script:    "await page.goto('https://example.com')",
		EnvironmentVariables: []checkly.EnvironmentVariable{
			{Key: "ENVTEST", Value: "Hello world"},
		},
	}
	validate := func(t *testing.T, body []byte) {
		var got struct {
			Script               *string                       `json:"script"`
			EnvironmentVariables []checkly.EnvironmentVariable `json:"environmentVariables"`
		}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("decoding error for data %q: %v", body, err)
		}
		if got.Script == nil || *got.Script != check.Script {
			t.Errorf("want script %q in the payload, got %s", check.Script, body)
		}
		if !cmp.Equal(check.EnvironmentVariables, got.EnvironmentVariables) {
			t.Error(cmp.Diff(check.EnvironmentVariables, got.EnvironmentVariables))
		}
	}
	cases := []struct {
		name    string
		method  string
		url     string
		status  int
		fixture string
		call    func(checkly.Client) (*checkly.Check, error)
	}{
		{
			name:    "CreateCheck",
			method:  http.MethodPost,
			url:     "/v1/checks/browser?autoAssignAlerts=false",
			status:  http.StatusCreated,
			fixture: "CreateBrowserCheck.json",
			call: func(c checkly.Client) (*checkly.Check, error) {
				return c.CreateCheck(context.Background(), check)
			},
		},
		{
			name:    "UpdateCheck",
			method:  http.MethodPut,
			url:     fmt.Sprintf("/v1/checks/%s?autoAssignAlerts=false", wantCheckID),
			status:  http.StatusOK,
			fixture: "UpdateBrowserCheck.json",
			call: func(c checkly.Client) (*checkly.Check, error) {
				return c.UpdateCheck(context.Background(), wantCheckID, check)
			},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ts := cannedResponseServer(t, tc.method, tc.url, validate, tc.status, tc.fixture)
			defer ts.Close()
			client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
			if _, err := tc.call(client); err != nil {
				t.Error(err)
			}
		})
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
		fmt.Fprintf(w, `{"id":%q,"checkType":%q}`, id, checkType)
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

// validateCheckEmptyPayload asserts that the keys the update endpoint would
// otherwise keep at their stored value are sent for an empty check: lists as
// [], the snippet IDs, local scripts and group order as null, and the alert
// settings' zero values.
func validateCheckEmptyPayload(t *testing.T, body []byte) {
	assertRawBodyField(t, body, "tags", `[]`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyField(t, body, "setupSnippetId", `null`)
	assertRawBodyField(t, body, "tearDownSnippetId", `null`)
	assertRawBodyField(t, body, "localSetupScript", `null`)
	assertRawBodyField(t, body, "localTearDownScript", `null`)
	assertRawBodyField(t, body, "groupOrder", `null`)
	assertRawBodyPath(t, body, `0`, "alertSettings", "reminders", "amount")
	assertRawBodyPath(t, body, `false`, "alertSettings", "parallelRunFailureThreshold", "enabled")
}

func TestCheckSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	check := checkly.Check{Name: "test", Type: checkly.TypeAPI}
	runPayloadCases(t, []payloadCase{
		{
			name:     "CreateCheck",
			method:   http.MethodPost,
			url:      "/v1/checks/api?autoAssignAlerts=false",
			status:   http.StatusCreated,
			fixture:  "CreateCheck.json",
			validate: validateCheckEmptyPayload,
			call: func(ctx context.Context, c checkly.Client) error {
				_, err := c.CreateCheck(ctx, check)
				return err
			},
		},
		{
			name:     "UpdateCheck",
			method:   http.MethodPut,
			url:      fmt.Sprintf("/v1/checks/%s?autoAssignAlerts=false", wantCheckID),
			status:   http.StatusOK,
			fixture:  "UpdateCheck.json",
			validate: validateCheckEmptyPayload,
			call: func(ctx context.Context, c checkly.Client) error {
				_, err := c.UpdateCheck(ctx, wantCheckID, check)
				return err
			},
		},
	})
}
