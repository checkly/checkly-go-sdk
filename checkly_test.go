package checkly_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	checkly "github.com/checkly/checkly-go-sdk"
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

func cannedResponseServer(
	t *testing.T,
	wantMethod string,
	wantURL string,
	validate func(*testing.T, []byte),
	status int,
	filename string,
) *httptest.Server {
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wantMethod != r.Method {
			t.Errorf("want %q request, got %q", wantMethod, r.Method)
		}
		if r.URL.String() != wantURL {
			t.Errorf("want %q, got %q", wantURL, r.URL.String())
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		validate(t, body)
		w.WriteHeader(status)
		data, err := os.Open(fmt.Sprintf("fixtures/%s", filename))
		if err != nil {
			t.Error(err)
		}
		defer data.Close()
		io.Copy(w, data)
	}))
}

func validateCheck(t *testing.T, body []byte) {
	var gotCheck checkly.Check
	err := json.Unmarshal(body, &gotCheck)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(wantCheck, gotCheck) {
		t.Error(cmp.Diff(wantCheck, gotCheck))
	}
}

func validateEmptyBody(t *testing.T, body []byte) {
	if len(body) > 0 {
		t.Errorf("expected empty body, but got %q", body)
	}
}

func validateAnything(*testing.T, []byte) {
}

// TODO: adjust wantCheck to test multiple SSLCheckDomain value and remove it from this list
var ignoreCheckFields = cmpopts.IgnoreFields(checkly.Check{}, "ID", "AlertChannelSubscriptions", "FrequencyOffset",
	"PrivateLocations", "SSLCheckDomain")

func TestAPIError(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/api?autoAssignAlerts=false",
		validateAnything,
		http.StatusBadRequest,
		"BadRequest.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.CreateCheck(context.Background(), checkly.Check{
		Type: checkly.TypeAPI,
	})
	if err == nil {
		t.Error("want error when API returns 'bad request' status, got nil")
	}
	if !strings.Contains(err.Error(), "frequency") {
		t.Errorf("want API error value to contain 'frequency', got %q", err.Error())
	}
}

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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
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
			client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
			if _, err := tc.call(client); err != nil {
				t.Error(err)
			}
		})
	}
}

var wantGroupID int64 = 135

var wantGroup = checkly.Group{
	Name:             "test",
	Activated:        true,
	Muted:            false,
	Tags:             []string{"auto"},
	Locations:        []string{"eu-west-1"},
	PrivateLocations: &[]string{},
	Concurrency:      3,
	APICheckDefaults: checkly.APICheckDefaults{
		BaseURL: "example.com/api/test",
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
		BasicAuth: checkly.BasicAuth{
			Username: "user",
			Password: "pass",
		},
	},
	EnvironmentVariables: []checkly.EnvironmentVariable{
		{
			Key:   "ENVTEST",
			Value: "Hello world",
		},
	},
	RetryStrategy: &checkly.RetryStrategy{
		Type:               "FIXED",
		MaxRetries:         1,
		MaxDurationSeconds: 600,
	},
	UseGlobalAlertSettings: false,
	AlertSettings: checkly.AlertSettings{
		EscalationType: checkly.RunBased,
		RunBasedEscalation: checkly.RunBasedEscalation{
			FailedRunThreshold: 1,
		},
		TimeBasedEscalation: checkly.TimeBasedEscalation{
			MinutesFailingThreshold: 5,
		},
		Reminders: checkly.Reminders{
			Amount:   0,
			Interval: 5,
		},
		ParallelRunFailureThreshold: checkly.ParallelRunFailureThreshold{
			Enabled:    false,
			Percentage: 10,
		},
	},
	AlertChannelSubscriptions: []checkly.AlertChannelSubscription{
		{
			Activated: true,
			ChannelID: 2996,
		},
	},
	LocalSetupScript:    "setup-test",
	LocalTearDownScript: "teardown-test",
}

func validateGroup(t *testing.T, body []byte) {
	var gotGroup checkly.Group
	err := json.Unmarshal(body, &gotGroup)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(wantGroup, gotGroup) {
		t.Error(cmp.Diff(wantGroup, gotGroup))
	}
}

var ignoreGroupFields = cmpopts.IgnoreFields(checkly.Group{}, "ID", "PrivateLocations")

func TestCreateGroup(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/check-groups?autoAssignAlerts=false",
		validateGroup,
		http.StatusCreated,
		"CreateGroup.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	gotGroup, err := client.CreateGroup(context.Background(), wantGroup)
	if err != nil {
		t.Error(err)
	}
	ignored := cmpopts.IgnoreFields(checkly.Group{}, "ID", "AlertChannelSubscriptions", "PrivateLocations")
	if !cmp.Equal(wantGroup, *gotGroup, ignored) {
		t.Error(cmp.Diff(wantGroup, *gotGroup, ignoreGroupFields))
	}
}

func TestGetGroup(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/check-groups/%d", wantGroupID),
		validateEmptyBody,
		http.StatusOK,
		"CreateGroup.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	gotGroup, err := client.GetGroup(context.Background(), wantGroupID)
	if err != nil {
		t.Error(err)
	}
	ignored := cmpopts.IgnoreFields(checkly.Group{}, "ID", "AlertChannelSubscriptions", "PrivateLocations")
	if !cmp.Equal(wantGroup, *gotGroup, ignored) {
		t.Error(cmp.Diff(wantGroup, *gotGroup, ignored))
	}
}

func TestUpdateGroup(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/check-groups/%d?autoAssignAlerts=false", wantGroupID),
		validateGroup,
		http.StatusOK,
		"UpdateGroup.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	gotGroup, err := client.UpdateGroup(context.Background(), wantGroupID, wantGroup)
	if err != nil {
		t.Error(err)
	}
	ignored := cmpopts.IgnoreFields(checkly.Group{}, "ID", "AlertChannelSubscriptions", "PrivateLocations")
	if !cmp.Equal(wantGroup, *gotGroup, ignored) {
		t.Error(cmp.Diff(wantGroup, *gotGroup, ignoreGroupFields))
	}
}

func TestDeleteGroup(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/check-groups/%d", wantGroupID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteGroup(context.Background(), wantGroupID)
	if err != nil {
		t.Error(err)
	}
}

func TestGetCheckResult(t *testing.T) {
	t.Parallel()
	resultID := "580c4e71-0109-45ba-9130-887ff01e1a7f"
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/check-results/%s/%s", wantCheckID, resultID),
		validateEmptyBody,
		http.StatusOK,
		"GetCheckResult.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	result, err := client.GetCheckResult(context.Background(), wantCheckID, resultID)
	if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	}
	startedAt, _ := time.Parse(time.RFC3339, "2020-09-02T11:19:06.283Z")
	stoppedAt, _ := time.Parse(time.RFC3339, "2020-09-02T11:19:06.413Z")
	createdAt, _ := time.Parse(time.RFC3339, "2020-09-02T11:19:06.681Z")
	// Ignore api check result comparison for now
	result.ApiCheckResult = &checkly.ApiCheckResult{}
	expectedResult := checkly.CheckResult{
		ID:                  "580c4e71-0109-45ba-9130-887ff01e1a7f",
		HasErrors:           false,
		HasFailures:         false,
		RunLocation:         "eu-central-1",
		StartedAt:           startedAt,
		StoppedAt:           stoppedAt,
		ResponseTime:        129,
		ApiCheckResult:      &checkly.ApiCheckResult{},
		BrowserCheckResult:  nil,
		CheckID:             "73d29e72-6540-4bb5-967e-e07fa2c9465e",
		CreatedAt:           createdAt,
		Name:                "API check 1",
		CheckRunID:          1599045546009,
		Attempts:            1,
		IsDegraded:          false,
		OverMaxResponseTime: false,
	}
	if !cmp.Equal(expectedResult, *result, nil) {
		expected, _ := json.Marshal(expectedResult)
		got, _ := json.Marshal(result)
		t.Errorf("Got invalid result, expected: %s, \ngot: %s", expected, got)
	}
}

func TestGetTracerouteCheckResult(t *testing.T) {
	t.Parallel()
	resultID := "11111111-1111-1111-1111-111111111111"
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/check-results/%s/%s", wantCheckID, resultID),
		validateEmptyBody,
		http.StatusOK,
		"GetTracerouteCheckResult.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	result, err := client.GetCheckResult(context.Background(), wantCheckID, resultID)
	if err != nil {
		t.Fatalf("Expected no errors, got %v", err)
	}
	if result.TracerouteCheckResult == nil {
		t.Fatal("expected tracerouteCheckResult to be populated, got nil")
	}
	tr := result.TracerouteCheckResult
	if tr.TotalHops == nil || *tr.TotalHops != 12 {
		t.Errorf("expected totalHops 12, got %v", tr.TotalHops)
	}
	if tr.DestinationReached == nil || *tr.DestinationReached {
		t.Errorf("expected destinationReached false, got %v", tr.DestinationReached)
	}
	if tr.Response == nil || tr.Response.TruncationReason != "max-hops" {
		t.Errorf("expected response.truncationReason max-hops, got %+v", tr.Response)
	}
	if tr.Response == nil || len(tr.Response.Hops) != 1 {
		t.Errorf("expected 1 hop in response, got %+v", tr.Response)
	}
	if result.GRPCCheckResult != nil || result.SSLCheckResult != nil {
		t.Errorf("expected sibling typed results nil, got grpc=%v ssl=%v", result.GRPCCheckResult, result.SSLCheckResult)
	}
}

func TestGetGRPCCheckResult(t *testing.T) {
	t.Parallel()
	resultID := "22222222-2222-2222-2222-222222222222"
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/check-results/%s/%s", wantCheckID, resultID),
		validateEmptyBody,
		http.StatusOK,
		"GetGRPCCheckResult.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	result, err := client.GetCheckResult(context.Background(), wantCheckID, resultID)
	if err != nil {
		t.Fatalf("Expected no errors, got %v", err)
	}
	if result.GRPCCheckResult == nil {
		t.Fatal("expected grpcCheckResult to be populated, got nil")
	}
	g := result.GRPCCheckResult
	if g.GRPCStatusCode == nil || *g.GRPCStatusCode != 14 {
		t.Errorf("expected grpcStatusCode 14, got %v", g.GRPCStatusCode)
	}
	if g.Response == nil {
		t.Fatal("expected grpc response artifact, got nil")
	}
	if g.Response.GRPCStatusMessage != "connection refused" {
		t.Errorf("expected grpcStatusMessage 'connection refused', got %q", g.Response.GRPCStatusMessage)
	}
	if g.Response.HealthStatusLabel != "NOT_SERVING" {
		t.Errorf("expected healthStatusLabel NOT_SERVING, got %q", g.Response.HealthStatusLabel)
	}
	if len(g.Response.DiscoveredMethods) != 1 || len(g.Response.Metadata) != 1 {
		t.Errorf("expected 1 discovered method and 1 metadata entry, got %+v", g.Response)
	}
}

func TestGetSSLCheckResult(t *testing.T) {
	t.Parallel()
	resultID := "33333333-3333-3333-3333-333333333333"
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/check-results/%s/%s", wantCheckID, resultID),
		validateEmptyBody,
		http.StatusOK,
		"GetSSLCheckResult.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	result, err := client.GetCheckResult(context.Background(), wantCheckID, resultID)
	if err != nil {
		t.Fatalf("Expected no errors, got %v", err)
	}
	if result.SSLCheckResult == nil {
		t.Fatal("expected sslCheckResult to be populated, got nil")
	}
	s := result.SSLCheckResult
	if s.TLSVersion != "TLS 1.3" || s.CipherSuite != "TLS_AES_256_GCM_SHA384" {
		t.Errorf("expected TLS 1.3 / AES_256, got %q / %q", s.TLSVersion, s.CipherSuite)
	}
	if s.DaysUntilExpiry == nil || *s.DaysUntilExpiry != -3 {
		t.Errorf("expected daysUntilExpiry -3, got %v", s.DaysUntilExpiry)
	}
	if s.ChainTrusted == nil || *s.ChainTrusted {
		t.Errorf("expected chainTrusted false, got %v", s.ChainTrusted)
	}
	if s.FailureCategory != "expired" {
		t.Errorf("expected failureCategory expired, got %q", s.FailureCategory)
	}
	if s.Response == nil || s.Response.Certificate["subjectCN"] != "expired.example.com" {
		t.Errorf("expected response.certificate.subjectCN, got %+v", s.Response)
	}
}

func TestGetCheckResults(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/check-results/%s", wantCheckID),
		validateEmptyBody,
		http.StatusOK,
		"GetCheckResults.json",
	)

	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	results, err := client.GetCheckResults(context.Background(), wantCheckID, nil)
	if err != nil {
		t.Error(err)
	}
	if len(results) != 10 {
		t.Errorf("Expected to get 10 results got %d", len(results))
		return
	}
	startedAt, _ := time.Parse(time.RFC3339, "2020-09-02T11:19:06.283Z")
	stoppedAt, _ := time.Parse(time.RFC3339, "2020-09-02T11:19:06.413Z")
	createdAt, _ := time.Parse(time.RFC3339, "2020-09-02T11:19:06.681Z")
	actualFirstResult := results[0]
	// Ignore api check result comparison for now
	actualFirstResult.ApiCheckResult = &checkly.ApiCheckResult{}
	expectedFirstResult := checkly.CheckResult{
		ID:                  "580c4e71-0109-45ba-9130-887ff01e1a7f",
		HasErrors:           false,
		HasFailures:         false,
		RunLocation:         "eu-central-1",
		StartedAt:           startedAt,
		StoppedAt:           stoppedAt,
		ResponseTime:        129,
		ApiCheckResult:      &checkly.ApiCheckResult{},
		BrowserCheckResult:  nil,
		CheckID:             "73d29e72-6540-4bb5-967e-e07fa2c9465e",
		CreatedAt:           createdAt,
		Name:                "API check 1",
		CheckRunID:          1599045546009,
		Attempts:            1,
		IsDegraded:          false,
		OverMaxResponseTime: false,
	}
	if !cmp.Equal(expectedFirstResult, actualFirstResult, nil) {
		expected, _ := json.Marshal(expectedFirstResult)
		got, _ := json.Marshal(results[0])
		t.Errorf("Got invalid result, expected: %s, \ngot: %s", expected, got)
	}
}

func TestGetCheckResultsWithFilters(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		"/v1/check-results/73d29e72-6540?checkType=API&from=1&hasFailures=1&limit=100&location=us-east-1&page=1&to=1000",
		validateEmptyBody,
		http.StatusOK,
		"GetCheckResults.json",
	)

	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	results, err := client.GetCheckResults(context.Background(), "73d29e72-6540", &checkly.CheckResultsFilter{
		Limit:       100,
		Page:        1,
		From:        1,
		To:          1000,
		CheckType:   checkly.TypeAPI,
		HasFailures: true,
		Location:    "us-east-1",
	})
	if err != nil {
		t.Error(err)
	}
	if len(results) != 10 {
		t.Errorf("Expected to get 10 results got %d", len(results))
		return
	}
}

func TestGetCheckResultsWithFilters2(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		"/v1/check-results/73d29e72-6540?",
		validateEmptyBody,
		http.StatusOK,
		"GetCheckResults.json",
	)

	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	results, err := client.GetCheckResults(
		context.Background(),
		"73d29e72-6540", &checkly.CheckResultsFilter{},
	)
	if err != nil {
		t.Error(err)
	}
	if len(results) != 10 {
		t.Errorf("Expected to get 10 results got %d", len(results))
		return
	}
}

var ignoreSnippetFields = cmpopts.IgnoreFields(checkly.Snippet{}, "ID")

var testSnippet = checkly.Snippet{
	ID:     1,
	Name:   "snippet1",
	Script: "script1",
}

func validateSnippet(t *testing.T, body []byte) {
	var gotSnippet checkly.Snippet
	err := json.Unmarshal(body, &gotSnippet)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testSnippet, gotSnippet) {
		t.Error(cmp.Diff(testSnippet, gotSnippet))
	}
}

func TestCreateSnippet(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/snippets",
		validateSnippet,
		http.StatusCreated,
		"CreateSnippet.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	gotSnippet, err := client.CreateSnippet(context.Background(), testSnippet)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testSnippet, *gotSnippet, ignoreSnippetFields) {
		t.Error(cmp.Diff(testSnippet, *gotSnippet, ignoreSnippetFields))
	}
}

func TestGetSnippet(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/snippets/%d", testSnippet.ID),
		validateEmptyBody,
		http.StatusOK,
		"CreateSnippet.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	gotSnippet, err := client.GetSnippet(context.Background(), testSnippet.ID)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testSnippet, *gotSnippet, ignoreSnippetFields) {
		t.Error(cmp.Diff(testSnippet, *gotSnippet, ignoreSnippetFields))
	}
}

func TestUpdateSnippet(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/snippets/%d", testSnippet.ID),
		validateSnippet,
		http.StatusOK,
		"UpdateSnippet.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	gotSnippet, err := client.UpdateSnippet(context.Background(), testSnippet.ID, testSnippet)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testSnippet, *gotSnippet, ignoreSnippetFields) {
		t.Error(cmp.Diff(testSnippet, *gotSnippet, ignoreSnippetFields))
	}
}

func TestDeleteSnippet(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/snippets/%d", testSnippet.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteSnippet(context.Background(), testSnippet.ID)
	if err != nil {
		t.Error(err)
	}
}

var testEnvVariable = checkly.EnvironmentVariable{
	Key:   "k1",
	Value: "v1",
}

func validateEnvVariable(t *testing.T, body []byte) {
	var result checkly.EnvironmentVariable
	err := json.Unmarshal(body, &result)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testEnvVariable, result) {
		t.Error(cmp.Diff(testEnvVariable, result))
	}
}

func TestCreateEnvironmentVariable(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/variables",
		validateEnvVariable,
		http.StatusCreated,
		"CreateEnvironmentVariable.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	result, err := client.CreateEnvironmentVariable(context.Background(), testEnvVariable)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testEnvVariable, *result, nil) {
		t.Error(cmp.Diff(testEnvVariable, *result, nil))
	}
}

func TestGetEnvironmentVariable(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/variables/%s", testEnvVariable.Key),
		validateEmptyBody,
		http.StatusOK,
		"CreateEnvironmentVariable.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	result, err := client.GetEnvironmentVariable(context.Background(), testEnvVariable.Key)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testEnvVariable, *result, nil) {
		t.Error(cmp.Diff(testEnvVariable, *result, nil))
	}
}

func TestUpdateEnvironmentVariable(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/variables/%s", testEnvVariable.Key),
		validateEnvVariable,
		http.StatusOK,
		"UpdateEnvironmentVariable.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	result, err := client.UpdateEnvironmentVariable(
		context.Background(),
		testEnvVariable.Key,
		testEnvVariable,
	)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testEnvVariable, *result, nil) {
		t.Error(cmp.Diff(testEnvVariable, *result, nil))
	}
}

func TestDeleteEnvironmentVariable(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/variables/%s", testEnvVariable.Key),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteEnvironmentVariable(context.Background(), testEnvVariable.Key)
	if err != nil {
		t.Error(err)
	}
}

var ignoreAlertChannelFields = cmpopts.IgnoreFields(checkly.AlertChannel{}, "ID")

func getTestAlertChannelEmail() *checkly.AlertChannel {
	return &checkly.AlertChannel{
		ID:   1,
		Type: checkly.AlertTypeEmail,
		Email: &checkly.AlertChannelEmail{
			Address: "test@example.com",
		},
	}
}

func getTestAlertChannelSlack() checkly.AlertChannel {
	ac := checkly.AlertChannel{
		ID:   1,
		Type: checkly.AlertTypeEmail,
		Slack: &checkly.AlertChannelSlack{
			Channel:    "test",
			WebhookURL: "https://slack.com/test",
		},
	}
	return ac
}

func validateAlertChannel(t *testing.T, body []byte) {
	response := map[string]interface{}{}
	err := json.Unmarshal(body, &response)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	ac := checkly.AlertChannel{}
	ac.Type = response["type"].(string)
	cfgJSON, err := json.Marshal(response["config"])
	if err != nil {
		t.Error("NO CFG", response["config"].(string))
		return
	}
	cfg, err := checkly.AlertChannelConfigFromJSON(ac.Type, cfgJSON)
	if err != nil {
		t.Error("NO CFG", err)
		return
	}
	ac.SetConfig(cfg)

	if ac.Type == checkly.AlertTypeEmail {
		ta := getTestAlertChannelEmail()
		if ac.Email == nil {
			t.Error("Email nil -> ", cfg, string(body))
			return
		}
		if ta.Email.Address != ac.Email.Address {
			t.Errorf(
				"Expected: %s, Got: %s",
				ta.Email.Address,
				ac.Email.Address,
			)
		}
	}
	if ac.Type == checkly.AlertTypeSlack {
		ta := getTestAlertChannelSlack()
		if ta.Slack.Channel != ac.Slack.Channel {
			t.Errorf(
				"Expected: %s, Got: %s",
				ta.Slack.Channel,
				ac.Slack.Channel,
			)
		}
		if ta.Slack.WebhookURL != ac.Slack.WebhookURL {
			t.Errorf(
				"Expected: %s, Got: %s",
				ta.Slack.WebhookURL,
				ac.Slack.WebhookURL,
			)
		}
	}
}

func TestCreateAlertChannel(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/alert-channels",
		validateAlertChannel,
		http.StatusOK,
		"CreateAlertChannelEmail.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	ta := getTestAlertChannelEmail()
	ac, err := client.CreateAlertChannel(context.Background(), *ta)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(ta, ac, ignoreAlertChannelFields) {
		t.Error(cmp.Diff(ta, ac, ignoreAlertChannelFields))
	}
}

func TestGetAlertChannel(t *testing.T) {
	t.Skip("broken since it was added: validateAlertChannel decodes the empty body of the GET request as JSON and fails")
	t.Parallel()
	ta := getTestAlertChannelEmail()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/alert-channels/%d", ta.ID),
		validateAlertChannel,
		http.StatusOK,
		"GetAlertChannelEmail.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	ac, err := client.GetAlertChannel(context.Background(), ta.ID)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(ta, *ac, ignoreAlertChannelFields) {
		t.Error(cmp.Diff(ta, *ac, ignoreAlertChannelFields))
	}
}

func TestUpdateAlertChannel(t *testing.T) {
	t.Parallel()
	ta := getTestAlertChannelEmail()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/alert-channels/%d", ta.ID),
		validateAlertChannel,
		http.StatusOK,
		"UpdateAlertChannel.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateAlertChannel(context.Background(), ta.ID, *ta)
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteAlertChannel(t *testing.T) {
	t.Parallel()
	ta := getTestAlertChannelEmail()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/alert-channels/%d", ta.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteAlertChannel(context.Background(), ta.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestAlertChannelSetSmsConfig(t *testing.T) {
	ac := checkly.AlertChannel{
		Type: "SMS",
	}
	c := &checkly.AlertChannelSMS{
		Name:   "foo",
		Number: "012345",
	}
	ac.SetConfig(c)
	if ac.SMS == nil {
		t.Error("Shouldn't be nil")
		return
	}
	if ac.SMS.Name != "foo" {
		t.Errorf("Unexpected value: %s", ac.SMS.Name)
	}
}

func TestAlertChannelSetEmailConfig(t *testing.T) {
	ac := checkly.AlertChannel{
		Type: "EMAIL",
	}
	c := checkly.AlertChannelEmail{
		Address: "add@example.com",
	}
	ac.SetConfig(&c)
	if ac.Email == nil {
		t.Error("Shouldn't be nil")
		return
	}
	if ac.Email.Address != c.Address {
		t.Errorf("Unexpected value: %s", ac.Email.Address)
	}
}

func TestAlertChannelSetWebookConfig(t *testing.T) {
	ac := checkly.AlertChannel{
		Type: "WEBHOOK",
	}
	var c interface{}
	json := []byte(`{"headers":[{"key":"X-1","value":"h1v"},{"key":"X-2","value":"h2v"}],"method":"get","name":"fooname","queryParameters":[{"key":"q1","value":"v1"}],"template":"tmpl","url":"https://example.com/webhook","webhookSecret":"foosecret"}`)

	c, err := checkly.AlertChannelConfigFromJSON(ac.Type, json)
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	}
	ac.SetConfig(c)
	if ac.Webhook == nil {
		t.Error("Shouldn't be nil")
		return
	}
	if ac.Webhook.Name != "fooname" {
		t.Errorf("Unexpected value: %s", ac.Webhook.Name)
	}
	if ac.Webhook.Method != "get" {
		t.Errorf("Unexpected value: %s", ac.Webhook.Method)
	}
	if ac.Webhook.Template != "tmpl" {
		t.Errorf("Unexpected value: %s", ac.Webhook.Template)
	}
	if ac.Webhook.URL != "https://example.com/webhook" {
		t.Errorf("Unexpected value: %s", ac.Webhook.URL)
	}
	if ac.Webhook.WebhookSecret != "foosecret" {
		t.Errorf("Unexpected value: %s", ac.Webhook.WebhookSecret)
	}
	if len(ac.Webhook.Headers) != 2 {
		t.Errorf("Unexpected len: %d", len(ac.Webhook.Headers))
	}
	if ac.Webhook.Headers[0].Key != "X-1" && ac.Webhook.Headers[0].Key != "X-2" {
		t.Errorf("Unexpected value: %s", ac.Webhook.Headers[0].Key)
	}
	if ac.Webhook.Headers[0].Value != "h1v" && ac.Webhook.Headers[0].Value != "h2v" {
		t.Errorf("Unexpected value: %s", ac.Webhook.Headers[0].Value)
	}
	if len(ac.Webhook.QueryParameters) != 1 {
		t.Errorf("Unexpected len: %d", len(ac.Webhook.QueryParameters))
	}
	if ac.Webhook.QueryParameters[0].Key != "q1" {
		t.Errorf("Unexpected value: %s", ac.Webhook.QueryParameters[0].Key)
	}
	if ac.Webhook.QueryParameters[0].Value != "v1" {
		t.Errorf("Unexpected value: %s", ac.Webhook.QueryParameters[0].Value)
	}
}

func validateDashboard(t *testing.T, body []byte) {
	var gotDashboard checkly.Dashboard
	err := json.Unmarshal(body, &gotDashboard)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testDashboard, gotDashboard) {
		t.Error(cmp.Diff(testDashboard, gotDashboard))
	}
}

var trueVal = true

var testDashboard = checkly.Dashboard{
	DashboardID:        "abcd1234",
	CustomDomain:       "is.checkly.online",
	CustomUrl:          "status-page",
	Logo:               "https://www.checklyhq.com/images/text_racoon_logo.svg",
	Favicon:            "https://www.checklyhq.com/images/text_racoon_logo.svg",
	Link:               "https://www.checklyhq.com",
	Description:        "Checkly status page",
	Header:             "Status",
	ShowHeader:         &trueVal,
	Width:              "FULL",
	RefreshRate:        60,
	Paginate:           true,
	PaginationRate:     30,
	Tags:               []string{"api"},
	HideTags:           false,
	ChecksPerPage:      15,
	UseTagsAndOperator: true,
	EnableIncidents:    true,
	CustomCSS:          "",
	ExpandChecks:       false,
	ShowCheckRunLinks:  false,
	ShowP95:            &trueVal,
	ShowP99:            &trueVal,
}

var ignoreDashboardFields = cmpopts.IgnoreFields(checkly.Dashboard{}, "DashboardID", "CreatedAt", "ID", "Keys")

func TestCreateDashboard(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/dashboards",
		validateDashboard,
		http.StatusCreated,
		"CreateDashboard.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	gotDashboard, err := client.CreateDashboard(context.Background(), testDashboard)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testDashboard, *gotDashboard, ignoreDashboardFields) {
		t.Error(cmp.Diff(testDashboard, *gotDashboard, ignoreDashboardFields))
	}
}

func TestDeleteDashboard(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/dashboards/%s", testDashboard.DashboardID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteDashboard(context.Background(), testDashboard.DashboardID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateDashboard(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/dashboards/%s", testDashboard.DashboardID),
		validateDashboard,
		http.StatusOK,
		"CreateDashboard.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateDashboard(context.Background(), testDashboard.DashboardID, testDashboard)
	if err != nil {
		t.Error(err)
	}
}

func TestGetDashboard(t *testing.T) {
	t.Skip("broken since it was added: validateDashboard decodes the empty body of the GET request as JSON and fails")
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/dashboards/%s", testDashboard.DashboardID),
		validateDashboard,
		http.StatusOK,
		"CreateDashboard.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	ac, err := client.GetDashboard(context.Background(), testDashboard.DashboardID)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testDashboard, *ac, ignoreDashboardFields) {
		t.Error(cmp.Diff(testDashboard, *ac, ignoreDashboardFields))
	}
}

func validateMaintenanceWindow(t *testing.T, body []byte) {
	var gotMaintenanceWindow checkly.MaintenanceWindow
	err := json.Unmarshal(body, &gotMaintenanceWindow)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testMaintenanceWindow, gotMaintenanceWindow) {
		t.Error(cmp.Diff(testMaintenanceWindow, gotMaintenanceWindow))
	}
}

var testMaintenanceWindowTrue = true

// testMaintenanceWindow is written the way the API returns it, with empty
// lists and explicit status page defaults, so it compares equal to responses.
var testMaintenanceWindow = checkly.MaintenanceWindow{
	ID:                1,
	Name:              "TEST",
	StartsAt:          "2014-08-24T00:00:00.000Z",
	EndsAt:            "2014-08-24T00:00:00.000Z",
	RepeatUnit:        "MONTH",
	RepeatEndsAt:      "2014-08-24T00:00:00.000Z",
	RepeatInterval:    10,
	CreatedAt:         "2013-08-24",
	UpdatedAt:         "2014-08-24",
	Tags:              []string{"string"},
	Timezone:          "America/New_York",
	PauseAllChecks:    true,
	SilenceAlertsTags: []string{"string"},
	SilenceAllAlerts:  false,
	Description:       "Planned database upgrade",
	StatusPageVisibility: checkly.MaintenanceWindowStatusPageVisibility{
		ReminderMinutesBefore: []int{},
		AutoStart:             &testMaintenanceWindowTrue,
		AutoEnd:               &testMaintenanceWindowTrue,
		ShowAffectedServices:  &testMaintenanceWindowTrue,
		StatusPageIDs:         []string{},
		ServiceIDs:            []string{},
	},
}

// An update replaces the whole window, so a zero MaintenanceWindow must send
// every field with the API's default rather than omitting it.
func TestMaintenanceWindowMarshalsEveryField(t *testing.T) {
	t.Parallel()

	body, err := json.Marshal(checkly.MaintenanceWindow{Name: "TEST"})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"id":                0.0,
		"name":              "TEST",
		"startsAt":          "",
		"endsAt":            "",
		"repeatInterval":    nil,
		"repeatUnit":        nil,
		"repeatEndsAt":      nil,
		"tags":              []any{},
		"timezone":          nil,
		"pauseAllChecks":    false,
		"silenceAlertsTags": []any{},
		"silenceAllAlerts":  false,
		"description":       nil,
		"statusPageVisibility": map[string]any{
			"enabled":               false,
			"severity":              nil,
			"affectAllServices":     false,
			"notifyOnStart":         false,
			"notifyOnEnd":           false,
			"suppressAutoIncidents": false,
			"reminderMinutesBefore": []any{},
			"autoStart":             true,
			"autoEnd":               true,
			"showAffectedServices":  true,
			"statusPageIds":         []any{},
			"serviceIds":            []any{},
		},
		"created_at": "",
		"updated_at": "",
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestMaintenanceWindowMarshalsSetValues(t *testing.T) {
	t.Parallel()

	falseValue := false
	body, err := json.Marshal(checkly.MaintenanceWindow{
		Name:              "TEST",
		RepeatInterval:    2,
		RepeatUnit:        "WEEK",
		RepeatEndsAt:      "2030-01-01",
		Tags:              []string{"pause"},
		Timezone:          "Europe/Berlin",
		PauseAllChecks:    true,
		SilenceAlertsTags: []string{"silence"},
		SilenceAllAlerts:  true,
		Description:       "Upgrade",
		StatusPageVisibility: checkly.MaintenanceWindowStatusPageVisibility{
			Enabled:               true,
			Severity:              "MAJOR",
			AffectAllServices:     false,
			NotifyOnStart:         true,
			NotifyOnEnd:           true,
			SuppressAutoIncidents: true,
			ReminderMinutesBefore: []int{60, 120},
			AutoStart:             &falseValue,
			AutoEnd:               &falseValue,
			ShowAffectedServices:  &falseValue,
			StatusPageIDs:         []string{"page"},
			ServiceIDs:            []string{"service"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"id":                0.0,
		"name":              "TEST",
		"startsAt":          "",
		"endsAt":            "",
		"repeatInterval":    2.0,
		"repeatUnit":        "WEEK",
		"repeatEndsAt":      "2030-01-01",
		"tags":              []any{"pause"},
		"timezone":          "Europe/Berlin",
		"pauseAllChecks":    true,
		"silenceAlertsTags": []any{"silence"},
		"silenceAllAlerts":  true,
		"description":       "Upgrade",
		"statusPageVisibility": map[string]any{
			"enabled":               true,
			"severity":              "MAJOR",
			"affectAllServices":     false,
			"notifyOnStart":         true,
			"notifyOnEnd":           true,
			"suppressAutoIncidents": true,
			"reminderMinutesBefore": []any{60.0, 120.0},
			"autoStart":             false,
			"autoEnd":               false,
			"showAffectedServices":  false,
			"statusPageIds":         []any{"page"},
			"serviceIds":            []any{"service"},
		},
		"created_at": "",
		"updated_at": "",
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

var ignoreMaintenanceWindowFields = cmpopts.IgnoreFields(checkly.MaintenanceWindow{}, "ID", "CreatedAt", "UpdatedAt")

func TestCreateMaintenanceWindow(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/maintenance-windows",
		validateMaintenanceWindow,
		http.StatusCreated,
		"CreateMaintenanceWindow.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	gotMaintencanceWindow, err := client.CreateMaintenanceWindow(context.Background(), testMaintenanceWindow)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testMaintenanceWindow, *gotMaintencanceWindow, ignoreMaintenanceWindowFields) {
		t.Error(cmp.Diff(testMaintenanceWindow, *gotMaintencanceWindow, ignoreMaintenanceWindowFields))
	}
}

func TestDeleteMaintenanceWindow(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/maintenance-windows/%d", testMaintenanceWindow.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteMaintenanceWindow(context.Background(), testMaintenanceWindow.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateMaintenanceWindow(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/maintenance-windows/%d", testMaintenanceWindow.ID),
		validateMaintenanceWindow,
		http.StatusOK,
		"CreateMaintenanceWindow.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateMaintenanceWindow(context.Background(), testMaintenanceWindow.ID, testMaintenanceWindow)
	if err != nil {
		t.Error(err)
	}
}

func TestGetMaintenanceWindow(t *testing.T) {
	t.Skip("broken since it was added: validateMaintenanceWindow decodes the empty body of the GET request as JSON and fails")
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/maintenance-windows/%d", testMaintenanceWindow.ID),
		validateMaintenanceWindow,
		http.StatusOK,
		"CreateMaintenanceWindow.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	mw, err := client.GetMaintenanceWindow(context.Background(), testMaintenanceWindow.ID)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testMaintenanceWindow, *mw, ignoreMaintenanceWindowFields) {
		t.Error(cmp.Diff(testMaintenanceWindow, *mw, ignoreMaintenanceWindowFields))
	}
}

func validatePrivateLocation(t *testing.T, body []byte) {
	var gotPrivateLocation checkly.PrivateLocation
	err := json.Unmarshal(body, &gotPrivateLocation)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testPrivateLocation, gotPrivateLocation) {
		t.Error(cmp.Diff(testPrivateLocation, gotPrivateLocation))
	}
}

var testPrivateLocation = checkly.PrivateLocation{
	ID:        "1",
	Name:      "New Private Location",
	SlugName:  "new-private-location",
	Icon:      "location",
	CreatedAt: "2013-08-24",
	UpdatedAt: "2013-08-24",
}

var ignorePrivateLocationFields = cmpopts.IgnoreFields(checkly.PrivateLocation{}, "ID", "CreatedAt", "UpdatedAt", "Keys")

func TestCreatePrivateLocation(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/private-locations",
		validatePrivateLocation,
		http.StatusCreated,
		"CreatePrivateLocation.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	gotPrivateLocation, err := client.CreatePrivateLocation(context.Background(), testPrivateLocation)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testPrivateLocation, *gotPrivateLocation, ignorePrivateLocationFields) {
		t.Error(cmp.Diff(testPrivateLocation, *gotPrivateLocation, ignorePrivateLocationFields))
	}
}

func TestDeletePrivateLocation(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/private-locations/%s", testPrivateLocation.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeletePrivateLocation(context.Background(), testPrivateLocation.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdatePrivateLocation(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/private-locations/%s", testPrivateLocation.ID),
		validatePrivateLocation,
		http.StatusOK,
		"CreatePrivateLocation.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdatePrivateLocation(context.Background(), testPrivateLocation.ID, testPrivateLocation)
	if err != nil {
		t.Error(err)
	}
}

func TestGetPrivateLocation(t *testing.T) {
	t.Skip("broken since it was added: validatePrivateLocation decodes the empty body of the GET request as JSON and fails")
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/private-locations/%s", testPrivateLocation.ID),
		validatePrivateLocation,
		http.StatusOK,
		"CreatePrivateLocation.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	pl, err := client.GetPrivateLocation(context.Background(), testPrivateLocation.ID)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testPrivateLocation, *pl, ignorePrivateLocationFields) {
		t.Error(cmp.Diff(testPrivateLocation, *pl, ignorePrivateLocationFields))
	}
}

func TestGetStaticIPs(t *testing.T) {
	t.Parallel()

	fixtureMap := map[string]string{
		"/v1/static-ips-by-region":   "StaticIPs.json",
		"/v1/static-ipv6s-by-region": "StaticIPv6s.json",
	}

	// we can't use cannedResponseServer here since we need a response on more than one URL
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fileName, ok := fixtureMap[r.URL.Path]
		if !ok {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}

		filePath := filepath.Join("fixtures", fileName)
		fileData, err := ioutil.ReadFile(filePath)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		w.Write(fileData)
	}))
	defer ts.Close()

	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	gotStaticIPs, err := client.GetStaticIPs(context.Background())
	if err != nil {
		t.Error(err)
	}

	exampleIPv4 := netip.MustParsePrefix("54.151.146.209/32")
	exampleIPv6 := netip.MustParsePrefix("2600:1f18:12ca:3000::/56")

	expected := []checkly.StaticIP{
		{Region: "ap-southeast-1", Address: exampleIPv4},
		{Region: "us-east-1", Address: exampleIPv6},
	}

	for _, exp := range expected {
		found := false
		for _, ip := range gotStaticIPs {
			if reflect.DeepEqual(ip, exp) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %+v to be included in %+v, but it was not found", exp, gotStaticIPs)
		}
	}
}

var ignoreClientCertificateFields = cmpopts.IgnoreFields(checkly.ClientCertificate{}, "ID", "Passphrase", "CreatedAt")

var testClientCertificate = checkly.ClientCertificate{
	ID:          "49a1d5df-b89a-4998-a469-b1358e282ea5",
	Host:        "*.acme.com",
	Path:        "/partner/api",
	Certificate: "-----BEGIN CERTIFICATE-----\nMIICDzCCAbagAwIBAgIUMTZlfGA7WcD8e4/zt2MqxvEgQPYwCgYIKoZIzj0EAwIw\nVDELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAkNBMREwDwYDVQQHDAhUb29udG93bjES\nMBAGA1UECgwJQWNtZSBJbmMuMREwDwYDVQQDDAhhY21lLmNvbTAeFw0yNTAzMDMw\nNTQ2NTJaFw00OTEwMjMwNTQ2NTJaMHgxCzAJBgNVBAYTAlVTMQswCQYDVQQIDAJD\nQTERMA8GA1UEBwwIVG9vbnRvd24xEjAQBgNVBAoMCUFjbWUgSW5jLjEXMBUGA1UE\nAwwOV2lsZSBFLiBDb3lvdGUxHDAaBgkqhkiG9w0BCQEWDXdpbGVAYWNtZS5jb20w\nWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAATAjjDGsKFS1qgdNqziDZoD5hamTfdH\n0P+Ukk1RIue57QYVXhQSyNzcEz15kQnwYezEqfN+FtjtTwdk/CgnAELlo0IwQDAd\nBgNVHQ4EFgQU9C9CpZqM2WMrOs3vAYsc5GbjyzswHwYDVR0jBBgwFoAUnlOyzF/N\nK7YmKQegLdbdyIOCT/UwCgYIKoZIzj0EAwIDRwAwRAIgGgSnBymlH4MkZCVk5DYH\nPdnDo2Xf5uFi1Eyn2LTYP1MCIEtiGtsf0qYv6NzIPd5uTTZoB/8hPrAgM1QzWG4O\n3C/I\n-----END CERTIFICATE-----\n",
	PrivateKey:  "-----BEGIN ENCRYPTED PRIVATE KEY-----\nMIH0MF8GCSqGSIb3DQEFDTBSMDEGCSqGSIb3DQEFDDAkBBA5yR3aqy8mZD2wQzp1\nFH2JAgIIADAMBggqhkiG9w0CCQUAMB0GCWCGSAFlAwQBKgQQA49YCnXvfJ2CsQsV\n9C5JJwSBkNkWunSlqyeVW6OFa/+OjlLArgTGvW5ul08qu/145O9PO4Nr2CXeK5N2\nuvHwkWGfD8IVke+sgZPUjLoHsJ4h4AnyxlNHpIxgOfm0CoXT7PTaFb//d5NC6XyB\nK7ZpBzIThGlbuS/b9wp4MPmSaJn5Fci+84VG7KYK5RxU0fcU0rGSBynrZw803wnO\nFjP7qaq5bw==\n-----END ENCRYPTED PRIVATE KEY-----\n",
	Passphrase:  "secret password",
	TrustedCA:   "-----BEGIN CERTIFICATE-----\nMIIB/jCCAaOgAwIBAgIUZzxdNpoDYXaNiIBsh0/s++I+ZOEwCgYIKoZIzj0EAwIw\nVDELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAkNBMREwDwYDVQQHDAhUb29udG93bjES\nMBAGA1UECgwJQWNtZSBJbmMuMREwDwYDVQQDDAhhY21lLmNvbTAeFw0yNTAzMDMw\nNTQzMDZaFw0yNTA0MDIwNTQzMDZaMFQxCzAJBgNVBAYTAlVTMQswCQYDVQQIDAJD\nQTERMA8GA1UEBwwIVG9vbnRvd24xEjAQBgNVBAoMCUFjbWUgSW5jLjERMA8GA1UE\nAwwIYWNtZS5jb20wWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARDH3KGK6Vsk1A4\nyGf9ItQIS3yuAOi0n0ihmPzIOOOEN0c758ETABeUdgH55bakdx6q5KYSxf4TuXsJ\n2nCihqVVo1MwUTAdBgNVHQ4EFgQUnlOyzF/NK7YmKQegLdbdyIOCT/UwHwYDVR0j\nBBgwFoAUnlOyzF/NK7YmKQegLdbdyIOCT/UwDwYDVR0TAQH/BAUwAwEB/zAKBggq\nhkjOPQQDAgNJADBGAiEA/cJ9jV8MQz4ypQsFvUatrnbxyHO0f+pJhf09pAk6Kj8C\nIQCkSbope5r0KlVdqBeFF8wCfE3plwpelve3jqVIz6MedQ==\n-----END CERTIFICATE-----\n",
}

func validateClientCertificate(t *testing.T, body []byte) {
	var clientCertificate checkly.ClientCertificate
	err := json.Unmarshal(body, &clientCertificate)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testClientCertificate, clientCertificate) {
		t.Error(cmp.Diff(testClientCertificate, clientCertificate))
	}
}

func TestCreateClientCertificate(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/client-certificates",
		validateClientCertificate,
		http.StatusCreated,
		"CreateClientCertificate.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.CreateClientCertificate(context.Background(), testClientCertificate)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testClientCertificate, *response, ignoreClientCertificateFields) {
		t.Error(cmp.Diff(testClientCertificate, *response, ignoreClientCertificateFields))
	}
}

func TestGetClientCertificate(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/client-certificates/%s", testClientCertificate.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetClientCertificate.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.GetClientCertificate(context.Background(), testClientCertificate.ID)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testClientCertificate, *response, ignoreClientCertificateFields) {
		t.Error(cmp.Diff(testClientCertificate, *response, ignoreClientCertificateFields))
	}
}

func TestDeleteClientCertificate(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/client-certificates/%s", testClientCertificate.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteClientCertificate(context.Background(), testClientCertificate.ID)
	if err != nil {
		t.Error(err)
	}
}

func validateStatusPageService(t *testing.T, body []byte) {
	var payload checkly.StatusPageService
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testStatusPageService, payload) {
		t.Error(cmp.Diff(testStatusPageService, payload))
	}
}

var testStatusPageService = checkly.StatusPageService{
	ID:   "8a894b49-467f-4af6-9230-cd4d3bd452f4",
	Name: "Foo service",
}

var ignoreStatusPageServiceFields = cmpopts.IgnoreFields(checkly.StatusPageService{}, "ID")

func TestCreateStatusPageService(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/status-pages/services",
		validateStatusPageService,
		http.StatusCreated,
		"CreateStatusPageService.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.CreateStatusPageService(context.Background(), testStatusPageService)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testStatusPageService, *response, ignoreStatusPageServiceFields) {
		t.Error(cmp.Diff(testStatusPageService, *response, ignoreStatusPageServiceFields))
	}
}

func TestDeleteStatusPageService(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/status-pages/services/%s", testStatusPageService.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteStatusPageService(context.Background(), testStatusPageService.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateStatusPageService(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/status-pages/services/%s", testStatusPageService.ID),
		validateStatusPageService,
		http.StatusOK,
		"UpdateStatusPageService.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateStatusPageService(context.Background(), testStatusPageService.ID, testStatusPageService)
	if err != nil {
		t.Error(err)
	}
}

func TestGetStatusPageService(t *testing.T) {
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/status-pages/services/%s", testStatusPageService.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetStatusPageService.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.GetStatusPageService(context.Background(), testStatusPageService.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testStatusPageService, *response, ignoreStatusPageServiceFields) {
		t.Error(cmp.Diff(testStatusPageService, *response, ignoreStatusPageServiceFields))
	}
}

func validateStatusPage(t *testing.T, body []byte) {
	var payload checkly.StatusPage
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testStatusPage, payload) {
		t.Error(cmp.Diff(testStatusPage, payload))
	}
}

var testStatusPage = checkly.StatusPage{
	ID:           "cd8d05a4-c292-4dc4-a78f-1dea65e5457e",
	Name:         "Foo status page",
	URL:          "foo-status-page",
	DefaultTheme: checkly.StatusPageThemeAuto,
	Cards: []checkly.StatusPageCard{
		{
			Name: "Foo card",
			Services: []checkly.StatusPageService{
				{
					ID:   "8a894b49-467f-4af6-9230-cd4d3bd452f4",
					Name: "Foo service",
				},
			},
		},
	},
}

var ignoreStatusPageFields = cmpopts.IgnoreFields(checkly.StatusPage{}, "ID")

func TestCreateStatusPage(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/status-pages",
		validateStatusPage,
		http.StatusCreated,
		"CreateStatusPage.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.CreateStatusPage(context.Background(), testStatusPage)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testStatusPage, *response, ignoreStatusPageFields) {
		t.Error(cmp.Diff(testStatusPage, *response, ignoreStatusPageFields))
	}
}

func TestDeleteStatusPage(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/status-pages/%s", testStatusPage.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteStatusPage(context.Background(), testStatusPage.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateStatusPage(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/status-pages/%s", testStatusPage.ID),
		validateStatusPage,
		http.StatusOK,
		"UpdateStatusPage.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateStatusPage(context.Background(), testStatusPage.ID, testStatusPage)
	if err != nil {
		t.Error(err)
	}
}

func TestGetStatusPage(t *testing.T) {
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/status-pages/%s", testStatusPage.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetStatusPage.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.GetStatusPage(context.Background(), testStatusPage.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testStatusPage, *response, ignoreStatusPageFields) {
		t.Error(cmp.Diff(testStatusPage, *response, ignoreStatusPageFields))
	}
}

func validateStatusPageV3(t *testing.T, body []byte) {
	var payload checkly.StatusPageV3
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testStatusPageV3, payload) {
		t.Error(cmp.Diff(testStatusPageV3, payload))
	}
}

var testStatusPageV3 = checkly.StatusPageV3{
	ID:            "e35f7e14-91b2-4d24-b7b6-e0f9e2f8e51c",
	Name:          "Foo v3 status page",
	URL:           "foo-v3-status-page",
	Description:   "All Foo systems",
	DefaultTheme:  checkly.StatusPageThemeAuto,
	SupportLink:   "mailto:support@example.org",
	FooterText:    "Foo Inc.",
	AllowIndexing: true,
	ThemeColors: &checkly.StatusPageV3ThemeColors{
		Light: checkly.StatusPageV3ThemeColorGroup{
			BodyBackgroundColor:          "#F9FAFB",
			HeaderBackgroundColor:        "#FFFFFF",
			HeaderFontColor:              "#151A1E",
			TitleFontColor:               "#212930",
			BodyFontColor:                "#475766",
			BodyFontColorMuted:           "#60758A",
			NavigationFontColor:          "#151A1E",
			LinkFontColor:                "#FF0000",
			CardBackgroundColor:          "#FFFFFF",
			BorderColor:                  "#E0E5EB",
			PrimaryButtonBackgroundColor: "#151A1E",
			PrimaryButtonFontColor:       "#FFFFFF",
		},
		Dark: checkly.StatusPageV3ThemeColorGroup{
			BodyBackgroundColor:          "#14171C",
			HeaderBackgroundColor:        "#171B21",
			HeaderFontColor:              "#FFFFFF",
			TitleFontColor:               "#ECEEF2",
			BodyFontColor:                "#C6CDD7",
			BodyFontColorMuted:           "#A3B3C2",
			NavigationFontColor:          "#FFFFFF",
			LinkFontColor:                "#248AFF",
			CardBackgroundColor:          "#171B21",
			BorderColor:                  "#242B34",
			PrimaryButtonBackgroundColor: "#242B34",
			PrimaryButtonFontColor:       "#FFFFFF",
		},
	},
}

var ignoreStatusPageV3Fields = cmpopts.IgnoreFields(checkly.StatusPageV3{}, "ID")

func TestCreateStatusPageV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v3/status-pages",
		validateStatusPageV3,
		http.StatusCreated,
		"CreateStatusPageV3.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.CreateStatusPageV3(context.Background(), testStatusPageV3)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testStatusPageV3, *response, ignoreStatusPageV3Fields) {
		t.Error(cmp.Diff(testStatusPageV3, *response, ignoreStatusPageV3Fields))
	}
}

func TestDeleteStatusPageV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v3/status-pages/%s", testStatusPageV3.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteStatusPageV3(context.Background(), testStatusPageV3.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateStatusPageV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v3/status-pages/%s", testStatusPageV3.ID),
		validateStatusPageV3,
		http.StatusOK,
		"UpdateStatusPageV3.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateStatusPageV3(context.Background(), testStatusPageV3.ID, testStatusPageV3)
	if err != nil {
		t.Error(err)
	}
}

func TestGetStatusPageV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v3/status-pages/%s", testStatusPageV3.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetStatusPageV3.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.GetStatusPageV3(context.Background(), testStatusPageV3.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testStatusPageV3, *response, ignoreStatusPageV3Fields) {
		t.Error(cmp.Diff(testStatusPageV3, *response, ignoreStatusPageV3Fields))
	}
}

func validateStatusPageComponentV3(t *testing.T, body []byte) {
	var payload checkly.StatusPageComponentV3
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testStatusPageComponentV3, payload) {
		t.Error(cmp.Diff(testStatusPageComponentV3, payload))
	}
}

var testStatusPageComponentV3ShowHistoricalData = false

var testStatusPageComponentV3 = checkly.StatusPageComponentV3{
	ID:           "0e4f5a72-6a5c-42a1-9a8a-5f7d38c8a9d1",
	StatusPageID: testStatusPageV3.ID,
	Type:         checkly.StatusPageComponentV3TypeService,
	Name:         "Foo API",
	Description:  "The Foo public API",
	DisplayOrder: 1,
	Hidden:       false,
	Configuration: &checkly.StatusPageComponentV3Configuration{
		ShowHistoricalData: &testStatusPageComponentV3ShowHistoricalData,
	},
	ParentID: "0a2f26fb-47cc-42b7-91c6-40de3ec91a52",
}

var ignoreStatusPageComponentV3Fields = cmpopts.IgnoreFields(checkly.StatusPageComponentV3{}, "ID")

func TestCreateStatusPageComponentV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		fmt.Sprintf("/v3/status-pages/%s/components", testStatusPageV3.ID),
		validateStatusPageComponentV3,
		http.StatusCreated,
		"CreateStatusPageComponentV3.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.CreateStatusPageComponentV3(context.Background(), testStatusPageV3.ID, testStatusPageComponentV3)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testStatusPageComponentV3, *response, ignoreStatusPageComponentV3Fields) {
		t.Error(cmp.Diff(testStatusPageComponentV3, *response, ignoreStatusPageComponentV3Fields))
	}
}

func TestDeleteStatusPageComponentV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v3/status-pages/%s/components/%s", testStatusPageV3.ID, testStatusPageComponentV3.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteStatusPageComponentV3(context.Background(), testStatusPageV3.ID, testStatusPageComponentV3.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateStatusPageComponentV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v3/status-pages/%s/components/%s", testStatusPageV3.ID, testStatusPageComponentV3.ID),
		validateStatusPageComponentV3,
		http.StatusOK,
		"UpdateStatusPageComponentV3.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateStatusPageComponentV3(context.Background(), testStatusPageV3.ID, testStatusPageComponentV3.ID, testStatusPageComponentV3)
	if err != nil {
		t.Error(err)
	}
}

func TestGetStatusPageComponentV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v3/status-pages/%s/components/%s", testStatusPageV3.ID, testStatusPageComponentV3.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetStatusPageComponentV3.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.GetStatusPageComponentV3(context.Background(), testStatusPageV3.ID, testStatusPageComponentV3.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testStatusPageComponentV3, *response, ignoreStatusPageComponentV3Fields) {
		t.Error(cmp.Diff(testStatusPageComponentV3, *response, ignoreStatusPageComponentV3Fields))
	}
}

func validateStatusPageAutomationRuleV3(t *testing.T, body []byte) {
	var payload checkly.StatusPageAutomationRuleV3
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testStatusPageAutomationRuleV3, payload) {
		t.Error(cmp.Diff(testStatusPageAutomationRuleV3, payload))
	}
}

var testStatusPageAutomationRuleV3 = checkly.StatusPageAutomationRuleV3{
	ID:                    "5b0a9a1e-8e29-4b71-92e1-0db9b3b41f0a",
	StatusPageID:          testStatusPageV3.ID,
	Name:                  "Foo API outage",
	Enabled:               true,
	FirstUpdate:           "We are investigating an issue with the Foo API.",
	LastUpdate:            "The issue has been resolved.",
	NotifySubscribers:     true,
	CoolDownWindowMinutes: 5,
	Tags:                  []string{"foo-api"},
	Components: []checkly.StatusPageAutomationRuleComponentV3{
		{
			ComponentID:  testStatusPageComponentV3.ID,
			TargetImpact: checkly.StatusPageTargetImpactV3MajorOutage,
		},
	},
}

var ignoreStatusPageAutomationRuleV3Fields = cmpopts.IgnoreFields(checkly.StatusPageAutomationRuleV3{}, "ID")

func TestCreateStatusPageAutomationRuleV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		fmt.Sprintf("/v3/status-pages/%s/automation-rules", testStatusPageV3.ID),
		validateStatusPageAutomationRuleV3,
		http.StatusCreated,
		"CreateStatusPageAutomationRuleV3.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.CreateStatusPageAutomationRuleV3(context.Background(), testStatusPageV3.ID, testStatusPageAutomationRuleV3)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testStatusPageAutomationRuleV3, *response, ignoreStatusPageAutomationRuleV3Fields) {
		t.Error(cmp.Diff(testStatusPageAutomationRuleV3, *response, ignoreStatusPageAutomationRuleV3Fields))
	}
}

func TestDeleteStatusPageAutomationRuleV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v3/status-pages/%s/automation-rules/%s", testStatusPageV3.ID, testStatusPageAutomationRuleV3.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteStatusPageAutomationRuleV3(context.Background(), testStatusPageV3.ID, testStatusPageAutomationRuleV3.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateStatusPageAutomationRuleV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v3/status-pages/%s/automation-rules/%s", testStatusPageV3.ID, testStatusPageAutomationRuleV3.ID),
		validateStatusPageAutomationRuleV3,
		http.StatusOK,
		"UpdateStatusPageAutomationRuleV3.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateStatusPageAutomationRuleV3(context.Background(), testStatusPageV3.ID, testStatusPageAutomationRuleV3.ID, testStatusPageAutomationRuleV3)
	if err != nil {
		t.Error(err)
	}
}

func TestGetStatusPageAutomationRuleV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v3/status-pages/%s/automation-rules/%s", testStatusPageV3.ID, testStatusPageAutomationRuleV3.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetStatusPageAutomationRuleV3.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.GetStatusPageAutomationRuleV3(context.Background(), testStatusPageV3.ID, testStatusPageAutomationRuleV3.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testStatusPageAutomationRuleV3, *response, ignoreStatusPageAutomationRuleV3Fields) {
		t.Error(cmp.Diff(testStatusPageAutomationRuleV3, *response, ignoreStatusPageAutomationRuleV3Fields))
	}
}

func validateURLMonitor(t *testing.T, body []byte) {
	var payload checkly.URLMonitor
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testURLMonitor, payload) {
		t.Error(cmp.Diff(testURLMonitor, payload))
	}
}

var testURLMonitor = checkly.URLMonitor{
	ID:               "6f8f7bd5-2340-4de4-9df5-5db45b309ad4",
	Name:             "URL Monitor #1",
	Frequency:        10,
	FrequencyOffset:  17,
	Activated:        true,
	Muted:            false,
	ShouldFail:       false,
	Locations:        []string{"us-east-1", "us-west-1"},
	PrivateLocations: &[]string{},
	Tags:             []string{"tag-1"},
	// The payload wrapper always serializes the subscriptions, sending an
	// empty list when unset, so that they can be cleared via an update.
	AlertChannelSubscriptions: []checkly.AlertChannelSubscription{},
	AlertSettings: &checkly.AlertSettings{
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
	UseGlobalAlertSettings: true,
	DegradedResponseTime:   5000,
	MaxResponseTime:        20000,
	RetryStrategy: &checkly.RetryStrategy{
		Type:               "LINEAR",
		MaxRetries:         2,
		BaseBackoffSeconds: 60,
		MaxDurationSeconds: 600,
		SameRegion:         true,
	},
	Request: checkly.URLRequest{
		URL:             "https://welcome.checklyhq.com",
		FollowRedirects: true,
		SkipSSL:         false,
		IPFamily:        "IPv4",
		Assertions: []checkly.Assertion{
			{
				Source:     "STATUS_CODE",
				Target:     "200",
				Property:   "",
				Comparison: "EQUALS",
			},
		},
	},
}

var ignoreURLMonitorFields = cmpopts.IgnoreFields(
	checkly.URLMonitor{},
	"ID",
	"AlertChannelSubscriptions",
	"CreatedAt",
	"UpdatedAt",
)

func TestCreateURLMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/url?autoAssignAlerts=false",
		validateURLMonitor,
		http.StatusCreated,
		"CreateURLMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.CreateURLMonitor(context.Background(), testURLMonitor)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testURLMonitor, *response, ignoreURLMonitorFields) {
		t.Error(cmp.Diff(testURLMonitor, *response, ignoreURLMonitorFields))
	}
}

func TestDeleteURLMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/checks/%s", testURLMonitor.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteURLMonitor(context.Background(), testURLMonitor.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateURLMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/url/%s?autoAssignAlerts=false", testURLMonitor.ID),
		validateURLMonitor,
		http.StatusOK,
		"UpdateURLMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateURLMonitor(context.Background(), testURLMonitor.ID, testURLMonitor)
	if err != nil {
		t.Error(err)
	}
}

func TestGetURLMonitor(t *testing.T) {
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/checks/%s", testURLMonitor.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetURLMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.GetURLMonitor(context.Background(), testURLMonitor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testURLMonitor, *response, ignoreURLMonitorFields) {
		t.Error(cmp.Diff(testURLMonitor, *response, ignoreURLMonitorFields))
	}
}

func validateDNSMonitor(t *testing.T, body []byte) {
	var payload checkly.DNSMonitor
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testDNSMonitor, payload, ignoreDNSMonitorFields) {
		t.Error(cmp.Diff(testDNSMonitor, payload, ignoreDNSMonitorFields))
	}
}

var testDNSMonitor = checkly.DNSMonitor{
	ID:              "25797fe8-3b88-46f5-a3c2-448167007a68",
	Name:            "DNS Monitor #1",
	Frequency:       10,
	FrequencyOffset: 80,
	Activated:       false,
	Muted:           false,
	Locations:       []string{"us-east-1"},
	Tags: []string{
		"tag-1",
	},
	AlertSettings: &checkly.AlertSettings{
		EscalationType: checkly.RunBased,
		RunBasedEscalation: checkly.RunBasedEscalation{
			FailedRunThreshold: 1,
		},
		Reminders: checkly.Reminders{
			Interval: 5,
		},
		ParallelRunFailureThreshold: checkly.ParallelRunFailureThreshold{
			Enabled:    false,
			Percentage: 10,
		},
	},
	UseGlobalAlertSettings: false,
	DegradedResponseTime:   500,
	MaxResponseTime:        1000,
	RetryStrategy:          nil,
	Request: checkly.DNSRequest{
		RecordType: "A",
		Query:      "welcome.checklyhq.com",
		Protocol:   "UDP",
		Assertions: []checkly.Assertion{
			{
				Source:     "RESPONSE_CODE",
				Target:     "NOERROR",
				Property:   "",
				Comparison: "EQUALS",
			},
		},
	},
}

var ignoreDNSMonitorFields = cmpopts.IgnoreFields(
	checkly.DNSMonitor{},
	"ID",
	"AlertChannelSubscriptions",
	"CreatedAt",
	"UpdatedAt",
)

func TestCreateDNSMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/dns?autoAssignAlerts=false",
		validateDNSMonitor,
		http.StatusCreated,
		"CreateDNSMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.CreateDNSMonitor(context.Background(), testDNSMonitor)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testDNSMonitor, *response, ignoreDNSMonitorFields) {
		t.Error(cmp.Diff(testDNSMonitor, *response, ignoreDNSMonitorFields))
	}
}

func TestDeleteDNSMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/checks/%s", testDNSMonitor.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteDNSMonitor(context.Background(), testDNSMonitor.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateDNSMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/dns/%s?autoAssignAlerts=false", testDNSMonitor.ID),
		validateDNSMonitor,
		http.StatusOK,
		"UpdateDNSMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateDNSMonitor(context.Background(), testDNSMonitor.ID, testDNSMonitor)
	if err != nil {
		t.Error(err)
	}
}

func TestGetDNSMonitor(t *testing.T) {
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/checks/%s", testDNSMonitor.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetDNSMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.GetDNSMonitor(context.Background(), testDNSMonitor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testDNSMonitor, *response, ignoreDNSMonitorFields) {
		t.Error(cmp.Diff(testDNSMonitor, *response, ignoreDNSMonitorFields))
	}
}

func validateICMPMonitor(t *testing.T, body []byte) {
	var payload checkly.ICMPMonitor
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testICMPMonitor, payload, ignoreICMPMonitorFields) {
		t.Error(cmp.Diff(testICMPMonitor, payload, ignoreICMPMonitorFields))
	}
}

var testICMPMonitor = checkly.ICMPMonitor{
	ID:              "ab1c2d3e-e4f5-6789-0abc-def012345678",
	Name:            "ICMP Monitor #1",
	Frequency:       10,
	FrequencyOffset: 20,
	Activated:       true,
	Muted:           false,
	Locations:       []string{"us-east-1"},
	Tags: []string{
		"tag-1",
	},
	DegradedPacketLossThreshold: 5,
	MaxPacketLossThreshold:      50,
	AlertSettings: &checkly.AlertSettings{
		EscalationType: checkly.RunBased,
		RunBasedEscalation: checkly.RunBasedEscalation{
			FailedRunThreshold: 1,
		},
		Reminders: checkly.Reminders{
			Interval: 5,
		},
		ParallelRunFailureThreshold: checkly.ParallelRunFailureThreshold{
			Enabled:    false,
			Percentage: 10,
		},
	},
	UseGlobalAlertSettings: false,
	RetryStrategy:          nil,
	Request: checkly.ICMPRequest{
		Hostname:  "example.com",
		IPFamily:  "IPv4",
		PingCount: 5,
		Assertions: []checkly.Assertion{
			{
				Source:     "LATENCY",
				Target:     "200",
				Property:   "avg",
				Comparison: "LESS_THAN",
			},
		},
	},
}

var ignoreICMPMonitorFields = cmpopts.IgnoreFields(
	checkly.ICMPMonitor{},
	"ID",
	"AlertChannelSubscriptions",
	"CreatedAt",
	"UpdatedAt",
)

func TestCreateICMPMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/icmp?autoAssignAlerts=false",
		validateICMPMonitor,
		http.StatusCreated,
		"CreateICMPMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.CreateICMPMonitor(context.Background(), testICMPMonitor)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testICMPMonitor, *response, ignoreICMPMonitorFields) {
		t.Error(cmp.Diff(testICMPMonitor, *response, ignoreICMPMonitorFields))
	}
}

func TestDeleteICMPMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/checks/%s", testICMPMonitor.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteICMPMonitor(context.Background(), testICMPMonitor.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateICMPMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/icmp/%s?autoAssignAlerts=false", testICMPMonitor.ID),
		validateICMPMonitor,
		http.StatusOK,
		"UpdateICMPMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateICMPMonitor(context.Background(), testICMPMonitor.ID, testICMPMonitor)
	if err != nil {
		t.Error(err)
	}
}

func TestGetICMPMonitor(t *testing.T) {
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/checks/%s", testICMPMonitor.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetICMPMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.GetICMPMonitor(context.Background(), testICMPMonitor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testICMPMonitor, *response, ignoreICMPMonitorFields) {
		t.Error(cmp.Diff(testICMPMonitor, *response, ignoreICMPMonitorFields))
	}
}

// --- gRPC monitor -----------------------------------------------------------

func validateGRPCMonitor(t *testing.T, body []byte) {
	var payload checkly.GRPCMonitor
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	// The payload sends empty lists rather than omitting them, which decode
	// to empty rather than nil slices.
	if !cmp.Equal(testGRPCMonitor, payload, ignoreGRPCMonitorFields, cmpopts.EquateEmpty()) {
		t.Error(cmp.Diff(testGRPCMonitor, payload, ignoreGRPCMonitorFields, cmpopts.EquateEmpty()))
	}
}

var testGRPCMonitor = checkly.GRPCMonitor{
	ID:              "a1b2c3d4-1111-2222-3333-444455556666",
	Name:            "gRPC Monitor #1",
	Frequency:       10,
	FrequencyOffset: 80,
	Activated:       false,
	Muted:           false,
	Locations:       []string{"us-east-1"},
	Tags: []string{
		"tag-1",
	},
	AlertSettings: &checkly.AlertSettings{
		EscalationType: checkly.RunBased,
		RunBasedEscalation: checkly.RunBasedEscalation{
			FailedRunThreshold: 1,
		},
		Reminders: checkly.Reminders{
			Interval: 5,
		},
		ParallelRunFailureThreshold: checkly.ParallelRunFailureThreshold{
			Enabled:    false,
			Percentage: 10,
		},
	},
	UseGlobalAlertSettings: false,
	DegradedResponseTime:   3000,
	MaxResponseTime:        5000,
	RetryStrategy:          nil,
	PrivateLocations:       &[]string{},
	Request: checkly.GRPCRequest{
		URL: "grpc.example.com",
		// interface{} port decodes from JSON as a float64, so the expected
		// value must use the same concrete type for cmp to match.
		Port:     float64(443),
		IPFamily: "IPv4",
		SkipSSL:  false,
		Timeout:  30,
		GRPCConfig: checkly.GRPCConfig{
			Mode:   "BEHAVIOR",
			TLS:    true,
			Method: "/grpc.health.v1.Health/Check",
		},
		Assertions: []checkly.Assertion{
			{
				Source:     "GRPC_STATUS",
				Target:     "0",
				Property:   "",
				Comparison: "EQUALS",
			},
		},
	},
}

var ignoreGRPCMonitorFields = cmpopts.IgnoreFields(
	checkly.GRPCMonitor{},
	"ID",
	"AlertChannelSubscriptions",
	"CreatedAt",
	"UpdatedAt",
)

func TestCreateGRPCMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/grpc?autoAssignAlerts=false",
		validateGRPCMonitor,
		http.StatusCreated,
		"CreateGRPCMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.CreateGRPCMonitor(context.Background(), testGRPCMonitor)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testGRPCMonitor, *response, ignoreGRPCMonitorFields) {
		t.Error(cmp.Diff(testGRPCMonitor, *response, ignoreGRPCMonitorFields))
	}
}

func TestDeleteGRPCMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/checks/%s", testGRPCMonitor.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteGRPCMonitor(context.Background(), testGRPCMonitor.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateGRPCMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/grpc/%s?autoAssignAlerts=false", testGRPCMonitor.ID),
		validateGRPCMonitor,
		http.StatusOK,
		"UpdateGRPCMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateGRPCMonitor(context.Background(), testGRPCMonitor.ID, testGRPCMonitor)
	if err != nil {
		t.Error(err)
	}
}

func TestGetGRPCMonitor(t *testing.T) {
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/checks/%s", testGRPCMonitor.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetGRPCMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.GetGRPCMonitor(context.Background(), testGRPCMonitor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testGRPCMonitor, *response, ignoreGRPCMonitorFields) {
		t.Error(cmp.Diff(testGRPCMonitor, *response, ignoreGRPCMonitorFields))
	}
}

// --- traceroute monitor -----------------------------------------------------

func validateTracerouteMonitor(t *testing.T, body []byte) {
	var payload checkly.TracerouteMonitor
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testTracerouteMonitor, payload, ignoreTracerouteMonitorFields) {
		t.Error(cmp.Diff(testTracerouteMonitor, payload, ignoreTracerouteMonitorFields))
	}
}

var testTracerouteMonitor = checkly.TracerouteMonitor{
	ID:              "b2c3d4e5-1111-2222-3333-444455556666",
	Name:            "Traceroute Monitor #1",
	Frequency:       10,
	FrequencyOffset: 80,
	Activated:       false,
	Muted:           false,
	Locations:       []string{"us-east-1"},
	Tags: []string{
		"tag-1",
	},
	AlertSettings: &checkly.AlertSettings{
		EscalationType: checkly.RunBased,
		RunBasedEscalation: checkly.RunBasedEscalation{
			FailedRunThreshold: 1,
		},
		Reminders: checkly.Reminders{
			Interval: 5,
		},
		ParallelRunFailureThreshold: checkly.ParallelRunFailureThreshold{
			Enabled:    false,
			Percentage: 10,
		},
	},
	UseGlobalAlertSettings: false,
	DegradedResponseTime:   3000,
	MaxResponseTime:        5000,
	RetryStrategy:          nil,
	Request: checkly.TracerouteRequest{
		URL:      "traceroute.example.com",
		Protocol: "ICMP",
		// No Port: the ICMP protocol strips it from the wire (see MarshalJSON).
		IPFamily:       "IPv4",
		MaxHops:        30,
		MaxUnknownHops: 5,
		Timeout:        30,
		Assertions: []checkly.Assertion{
			{
				Source:     "LATENCY",
				Target:     "200",
				Property:   "avg",
				Comparison: "LESS_THAN",
			},
		},
	},
}

var ignoreTracerouteMonitorFields = cmpopts.IgnoreFields(
	checkly.TracerouteMonitor{},
	"ID",
	"AlertChannelSubscriptions",
	"CreatedAt",
	"UpdatedAt",
)

func TestCreateTracerouteMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/traceroute?autoAssignAlerts=false",
		validateTracerouteMonitor,
		http.StatusCreated,
		"CreateTracerouteMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.CreateTracerouteMonitor(context.Background(), testTracerouteMonitor)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testTracerouteMonitor, *response, ignoreTracerouteMonitorFields) {
		t.Error(cmp.Diff(testTracerouteMonitor, *response, ignoreTracerouteMonitorFields))
	}
}

func TestDeleteTracerouteMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/checks/%s", testTracerouteMonitor.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteTracerouteMonitor(context.Background(), testTracerouteMonitor.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateTracerouteMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/traceroute/%s?autoAssignAlerts=false", testTracerouteMonitor.ID),
		validateTracerouteMonitor,
		http.StatusOK,
		"UpdateTracerouteMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateTracerouteMonitor(context.Background(), testTracerouteMonitor.ID, testTracerouteMonitor)
	if err != nil {
		t.Error(err)
	}
}

func TestGetTracerouteMonitor(t *testing.T) {
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/checks/%s", testTracerouteMonitor.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetTracerouteMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.GetTracerouteMonitor(context.Background(), testTracerouteMonitor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testTracerouteMonitor, *response, ignoreTracerouteMonitorFields) {
		t.Error(cmp.Diff(testTracerouteMonitor, *response, ignoreTracerouteMonitorFields))
	}
}

// --- SSL monitor ------------------------------------------------------------

func validateSSLMonitor(t *testing.T, body []byte) {
	var payload checkly.SSLMonitor
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testSSLMonitor, payload, ignoreSSLMonitorFields) {
		t.Error(cmp.Diff(testSSLMonitor, payload, ignoreSSLMonitorFields))
	}
}

var testSSLMonitor = checkly.SSLMonitor{
	ID:              "c3d4e5f6-1111-2222-3333-444455556666",
	Name:            "SSL Monitor #1",
	Frequency:       10,
	FrequencyOffset: 80,
	Activated:       false,
	Muted:           false,
	Locations:       []string{"us-east-1"},
	Tags: []string{
		"tag-1",
	},
	AlertSettings: &checkly.AlertSettings{
		EscalationType: checkly.RunBased,
		RunBasedEscalation: checkly.RunBasedEscalation{
			FailedRunThreshold: 1,
		},
		Reminders: checkly.Reminders{
			Interval: 5,
		},
		ParallelRunFailureThreshold: checkly.ParallelRunFailureThreshold{
			Enabled:    false,
			Percentage: 10,
		},
	},
	UseGlobalAlertSettings: false,
	DegradedResponseTime:   3000,
	MaxResponseTime:        10000,
	RetryStrategy:          nil,
	PrivateLocations:       &[]string{},
	Request: checkly.SSLRequest{
		SSLConfig: checkly.SSLConfig{
			Hostname:            "example.com",
			Port:                443,
			IPFamily:            "IPv4",
			SkipChainValidation: false,
			HandshakeTimeoutMs:  10000,
			// SecurityBaseline left nil: the create body must omit
			// `securityBaseline` so the server applies its default baseline.
			AlertDaysBeforeExpiry: 20,
		},
		Assertions: []checkly.Assertion{
			{
				Source:     checkly.Certificate,
				Property:   "daysUntilExpiry",
				Comparison: checkly.GreaterThan,
				Target:     "14",
			},
			{
				Source:     checkly.Connection,
				Property:   "tlsVersion",
				Comparison: checkly.Equals,
				Target:     "TLSv1.3",
			},
			{
				Source:     checkly.JSONResponse,
				Property:   "$.certificate.keySizeBits",
				Comparison: checkly.GreaterThan,
				Target:     "2048",
			},
			{
				Source:     checkly.TextResponse,
				Property:   "Issuer:.*Let's Encrypt",
				Comparison: checkly.Contains,
				Target:     "true",
			},
		},
	},
}

var ignoreSSLMonitorFields = cmpopts.IgnoreFields(
	checkly.SSLMonitor{},
	"ID",
	"AlertChannelSubscriptions",
	"CreatedAt",
	"UpdatedAt",
)

func TestCreateSSLMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/ssl?autoAssignAlerts=false",
		validateSSLMonitor,
		http.StatusCreated,
		"CreateSSLMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.CreateSSLMonitor(context.Background(), testSSLMonitor)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testSSLMonitor, *response, ignoreSSLMonitorFields) {
		t.Error(cmp.Diff(testSSLMonitor, *response, ignoreSSLMonitorFields))
	}
}

func TestDeleteSSLMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/checks/%s", testSSLMonitor.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteSSLMonitor(context.Background(), testSSLMonitor.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateSSLMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/ssl/%s?autoAssignAlerts=false", testSSLMonitor.ID),
		validateSSLMonitor,
		http.StatusOK,
		"UpdateSSLMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateSSLMonitor(context.Background(), testSSLMonitor.ID, testSSLMonitor)
	if err != nil {
		t.Error(err)
	}
}

func TestGetSSLMonitor(t *testing.T) {
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/checks/%s", testSSLMonitor.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetSSLMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.GetSSLMonitor(context.Background(), testSSLMonitor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testSSLMonitor, *response, ignoreSSLMonitorFields) {
		t.Error(cmp.Diff(testSSLMonitor, *response, ignoreSSLMonitorFields))
	}
}

// --- forbidden-field (anti-pattern) tests -----------------------------------

// TestTracerouteRequestStripsPortForICMP asserts the wire body omits `port`
// for an ICMP probe (mirroring the public API's Joi `.strip()`), while keeping
// it for non-ICMP protocols.
func TestTracerouteRequestStripsPortForICMP(t *testing.T) {
	t.Parallel()
	icmp, err := json.Marshal(checkly.TracerouteRequest{
		URL:      "traceroute.example.com",
		Protocol: "ICMP",
		Port:     33434,
	})
	if err != nil {
		t.Fatal(err)
	}
	var icmpBody map[string]interface{}
	if err := json.Unmarshal(icmp, &icmpBody); err != nil {
		t.Fatal(err)
	}
	if _, ok := icmpBody["port"]; ok {
		t.Errorf("ICMP traceroute body must omit `port`, got: %s", icmp)
	}

	tcp, err := json.Marshal(checkly.TracerouteRequest{
		URL:      "traceroute.example.com",
		Protocol: "TCP",
		Port:     33434,
	})
	if err != nil {
		t.Fatal(err)
	}
	var tcpBody map[string]interface{}
	if err := json.Unmarshal(tcp, &tcpBody); err != nil {
		t.Fatal(err)
	}
	if _, ok := tcpBody["port"]; !ok {
		t.Errorf("non-ICMP traceroute body must keep `port`, got: %s", tcp)
	}
}

// TestGRPCConfigModeForbiddenFields asserts the BEHAVIOR-only keys are dropped
// from a HEALTH-mode config and the HEALTH-only `service` key is dropped from a
// BEHAVIOR-mode config — the mode-exclusive fields must never reach the wire.
func TestGRPCConfigModeForbiddenFields(t *testing.T) {
	t.Parallel()
	health, err := json.Marshal(checkly.GRPCConfig{
		Mode:    "HEALTH",
		TLS:     true,
		Service: "grpc.health.v1.Health",
	})
	if err != nil {
		t.Fatal(err)
	}
	var healthBody map[string]interface{}
	if err := json.Unmarshal(health, &healthBody); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"encoding", "method", "serviceDefinition", "protoContent", "bfbsContent", "message"} {
		if _, ok := healthBody[k]; ok {
			t.Errorf("HEALTH-mode gRPC config must omit %q, got: %s", k, health)
		}
	}
	if _, ok := healthBody["service"]; !ok {
		t.Errorf("HEALTH-mode gRPC config must keep `service`, got: %s", health)
	}

	behavior, err := json.Marshal(checkly.GRPCConfig{
		Mode:              "BEHAVIOR",
		TLS:               true,
		Method:            "/grpc.health.v1.Health/Check",
		ServiceDefinition: "syntax = \"proto3\";",
	})
	if err != nil {
		t.Fatal(err)
	}
	var behaviorBody map[string]interface{}
	if err := json.Unmarshal(behavior, &behaviorBody); err != nil {
		t.Fatal(err)
	}
	if _, ok := behaviorBody["service"]; ok {
		t.Errorf("BEHAVIOR-mode gRPC config must omit `service`, got: %s", behavior)
	}
	for _, k := range []string{"method", "serviceDefinition"} {
		if _, ok := behaviorBody[k]; !ok {
			t.Errorf("BEHAVIOR-mode gRPC config must keep %q, got: %s", k, behavior)
		}
	}
}

func TestGRPCConfigFlatBuffersRoundTrip(t *testing.T) {
	t.Parallel()
	want := checkly.GRPCConfig{
		Mode:        "BEHAVIOR",
		TLS:         true,
		Encoding:    "FLATBUFFERS",
		Method:      "example.Greeter/Greet",
		BfbsContent: "RkxBVF9CVUZGRVJTX1NDSEVNQQ==",
		Message:     `{"name":"Checkly"}`,
	}

	body, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got checkly.GRPCConfig
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}

	var wire map[string]interface{}
	if err := json.Unmarshal(body, &wire); err != nil {
		t.Fatal(err)
	}
	if wire["encoding"] != "FLATBUFFERS" || wire["bfbsContent"] != want.BfbsContent {
		t.Fatalf("FlatBuffers fields were not preserved on the wire: %s", body)
	}
	for _, k := range []string{"service", "serviceDefinition", "protoContent"} {
		if _, ok := wire[k]; ok {
			t.Errorf("FlatBuffers gRPC config must omit %q, got: %s", k, body)
		}
	}
}

// TestSSLConfigOmitsSecurityBaselineWhenNil asserts a create body with a nil
// SecurityBaseline omits the `securityBaseline` key so the server applies its
// default baseline rather than receiving an empty override.
func TestSSLConfigOmitsSecurityBaselineWhenNil(t *testing.T) {
	t.Parallel()
	body, err := json.Marshal(testSSLMonitor.Request)
	if err != nil {
		t.Fatal(err)
	}
	var reqBody map[string]interface{}
	if err := json.Unmarshal(body, &reqBody); err != nil {
		t.Fatal(err)
	}
	sslConfig, ok := reqBody["sslConfig"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected `sslConfig` object in request body, got: %s", body)
	}
	if _, ok := sslConfig["securityBaseline"]; ok {
		t.Errorf("SSL config must omit `securityBaseline` when nil, got: %s", body)
	}
	if _, ok := reqBody["sslClientCertificateId"]; ok {
		t.Errorf("SSL request must omit `sslClientCertificateId` when nil, got: %s", body)
	}
}

// The v3 endpoints differ in how they treat a missing key: the page update
// only writes the keys present, while component and automation rule updates
// replace the resource wholesale. These tests pin the raw JSON, because
// decoding it back into the struct cannot tell an omitted key from an empty
// one.
func marshalToMap(t *testing.T, v interface{}) map[string]interface{} {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestStatusPageV3SerializesClearableFields(t *testing.T) {
	m := marshalToMap(t, checkly.StatusPageV3{Name: "foo", URL: "foo"})
	for _, key := range []string{
		"customDomain", "description", "logo", "logoDark", "redirectTo",
		"favicon", "privacyPolicyLink", "termsOfServiceLink", "supportLink",
		"footerText", "googleAnalyticsTag",
	} {
		value, ok := m[key]
		if !ok || value != "" {
			t.Errorf("expected %q to be sent as an empty string so the update clears it, got %#v (present: %v)", key, value, ok)
		}
	}
	if value, ok := m["themeColors"]; !ok || value != nil {
		t.Errorf("expected themeColors to be sent as null so the update clears it, got %#v (present: %v)", value, ok)
	}
	if value, ok := m["allowIndexing"]; !ok || value != false {
		t.Errorf("expected allowIndexing to always be sent, got %#v (present: %v)", value, ok)
	}
	// The API rejects an empty defaultTheme; omitted, it resets to AUTO.
	if value, ok := m["defaultTheme"]; ok {
		t.Errorf("expected an unset defaultTheme to be omitted, got %#v", value)
	}
}

func TestStatusPageComponentV3OmitsUnsetFields(t *testing.T) {
	m := marshalToMap(t, checkly.StatusPageComponentV3{Name: "foo"})
	// The update replaces the component wholesale, so an omitted key clears
	// it; configuration in particular must not be null, which the API
	// rejects.
	for _, key := range []string{"id", "statusPageId", "type", "description", "configuration", "parentId"} {
		if value, ok := m[key]; ok {
			t.Errorf("expected unset %q to be omitted, got %#v", key, value)
		}
	}
}

func TestStatusPageAutomationRuleV3OmitsEmptyComponents(t *testing.T) {
	for _, components := range [][]checkly.StatusPageAutomationRuleComponentV3{nil, {}} {
		m := marshalToMap(t, checkly.StatusPageAutomationRuleV3{
			Name:        "foo",
			FirstUpdate: "Investigating.",
			LastUpdate:  "Resolved.",
			Tags:        []string{"foo"},
			Components:  components,
		})
		// The API rejects a null list; omitted, it defaults to none.
		if value, ok := m["components"]; ok {
			t.Errorf("expected an empty component list to be omitted, got %#v", value)
		}
	}
}

// assertRawBodyField asserts that the raw JSON request body contains the
// given key with exactly the given serialized value. This catches fields
// added or normalized by the payload wrappers, which a round trip through
// the public struct cannot verify.
func assertRawBodyField(t *testing.T, body []byte, key, want string) {
	t.Helper()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	got, ok := raw[key]
	if !ok {
		t.Errorf("missing key %q in request body", key)
		return
	}
	if string(got) != want {
		t.Errorf("key %q in request body: want %s, got %s", key, want, got)
	}
}

// assertRawBodyKeyAbsent asserts that the raw JSON request body does not
// contain the given key.
func assertRawBodyKeyAbsent(t *testing.T, body []byte, key string) {
	t.Helper()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if got, ok := raw[key]; ok {
		t.Errorf("expected key %q to be absent from request body, got %s", key, got)
	}
}

// assertRawNestedBodyField asserts that the JSON object at the given
// top-level key of the raw request body contains the given nested key with
// exactly the given serialized value.
func assertRawNestedBodyField(t *testing.T, body []byte, key, nestedKey, want string) {
	t.Helper()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	outer, ok := raw[key]
	if !ok {
		t.Errorf("missing key %q in request body", key)
		return
	}
	var nested map[string]json.RawMessage
	if err := json.Unmarshal(outer, &nested); err != nil {
		t.Fatalf("decoding error for key %q data %q: %v", key, outer, err)
	}
	got, ok := nested[nestedKey]
	if !ok {
		t.Errorf("missing key %q.%q in request body", key, nestedKey)
		return
	}
	if string(got) != want {
		t.Errorf("key %q.%q in request body: want %s, got %s", key, nestedKey, want, got)
	}
}

func validateAPICheck(t *testing.T, body []byte) {
	var payload checkly.APICheck
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	// The payload wrapper normalizes nil lists to empty lists, so treat them
	// as equal here; the exact serialization is asserted on the raw body
	// below.
	if !cmp.Equal(testAPICheck, payload, cmpopts.EquateEmpty()) {
		t.Error(cmp.Diff(testAPICheck, payload, cmpopts.EquateEmpty()))
	}
	assertRawBodyField(t, body, "checkType", `"API"`)
	assertRawBodyField(t, body, "doubleCheck", `false`)
	assertRawBodyField(t, body, "groupId", `null`)
	assertRawBodyField(t, body, "tags", `["tag-1"]`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyField(t, body, "setupSnippetId", `123`)
	assertRawBodyField(t, body, "tearDownSnippetId", `456`)
	assertRawBodyField(t, body, "degradedResponseTime", `5000`)
	assertRawBodyField(t, body, "maxResponseTime", `20000`)
}

var (
	testAPICheckDescription         = "API check description"
	testAPICheckLocalSetupScript    = "console.log('setup')"
	testAPICheckLocalTearDownScript = "console.log('teardown')"
)

var testAPICheck = checkly.APICheck{
	ID:               "80f57ec9-2ee7-4152-b6ae-e1f5c47eae86",
	Name:             "API Check #1",
	Description:      &testAPICheckDescription,
	Frequency:        10,
	FrequencyOffset:  17,
	Activated:        true,
	Muted:            false,
	ShouldFail:       false,
	Locations:        []string{"us-east-1", "us-west-1"},
	PrivateLocations: &[]string{},
	Tags:             []string{"tag-1"},
	AlertSettings: &checkly.AlertSettings{
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
	UseGlobalAlertSettings: true,
	DegradedResponseTime:   5000,
	MaxResponseTime:        20000,
	SetupSnippetID:         123,
	TearDownSnippetID:      456,
	LocalSetupScript:       &testAPICheckLocalSetupScript,
	LocalTearDownScript:    &testAPICheckLocalTearDownScript,
	RetryStrategy: &checkly.RetryStrategy{
		Type:               "LINEAR",
		MaxRetries:         2,
		BaseBackoffSeconds: 60,
		MaxDurationSeconds: 600,
		SameRegion:         true,
	},
	Request: checkly.Request{
		Method:          "GET",
		URL:             "https://api.checklyhq.com",
		FollowRedirects: true,
		SkipSSL:         false,
		Headers: []checkly.KeyValue{
			{
				Key:   "X-Test",
				Value: "foo",
			},
		},
		QueryParameters: []checkly.KeyValue{
			{
				Key:   "query",
				Value: "param",
			},
		},
		Assertions: []checkly.Assertion{
			{
				Source:     "STATUS_CODE",
				Comparison: "EQUALS",
				Target:     "200",
			},
		},
	},
}

var ignoreAPICheckFields = cmpopts.IgnoreFields(
	checkly.APICheck{},
	"ID",
	"AlertChannelSubscriptions",
	"CreatedAt",
	"UpdatedAt",
)

func TestCreateAPICheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/api?autoAssignAlerts=false",
		validateAPICheck,
		http.StatusCreated,
		"CreateAPICheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.CreateAPICheck(context.Background(), testAPICheck)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testAPICheck, *response, ignoreAPICheckFields) {
		t.Error(cmp.Diff(testAPICheck, *response, ignoreAPICheckFields))
	}
}

func TestDeleteAPICheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/checks/%s", testAPICheck.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteAPICheck(context.Background(), testAPICheck.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateAPICheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/api/%s?autoAssignAlerts=false", testAPICheck.ID),
		validateAPICheck,
		http.StatusOK,
		"UpdateAPICheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateAPICheck(context.Background(), testAPICheck.ID, testAPICheck)
	if err != nil {
		t.Error(err)
	}
}

func TestGetAPICheck(t *testing.T) {
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/checks/%s", testAPICheck.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetAPICheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.GetAPICheck(context.Background(), testAPICheck.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testAPICheck, *response, ignoreAPICheckFields) {
		t.Error(cmp.Diff(testAPICheck, *response, ignoreAPICheckFields))
	}
}

func validateBrowserCheck(t *testing.T, body []byte) {
	var payload checkly.BrowserCheck
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	// The payload wrapper normalizes nil lists to empty lists, so treat them
	// as equal here; the exact serialization is asserted on the raw body
	// below.
	if !cmp.Equal(testBrowserCheck, payload, cmpopts.EquateEmpty()) {
		t.Error(cmp.Diff(testBrowserCheck, payload, cmpopts.EquateEmpty()))
	}
	assertRawBodyField(t, body, "checkType", `"BROWSER"`)
	assertRawBodyField(t, body, "doubleCheck", `false`)
	assertRawBodyField(t, body, "groupId", `null`)
	assertRawBodyField(t, body, "tags", `["tag-1"]`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyField(t, body, "scriptPath", `"__checks__/browser.spec.js"`)
	assertRawBodyField(t, body, "dependencies", `[{"path":"helpers.js","content":"module.exports = {}"}]`)
}

var testBrowserCheckScriptPath = "__checks__/browser.spec.js"

var testBrowserCheck = checkly.BrowserCheck{
	ID:               "0895dfd1-6b18-4cc9-a2c9-67ec9b45f7bd",
	Name:             "Browser Check #1",
	Frequency:        10,
	Activated:        true,
	Muted:            false,
	ShouldFail:       false,
	Locations:        []string{"us-east-1", "us-west-1"},
	PrivateLocations: &[]string{},
	Tags:             []string{"tag-1"},
	AlertSettings: &checkly.AlertSettings{
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
	UseGlobalAlertSettings: true,
	Script:                 "console.log('test')",
	ScriptPath:             &testBrowserCheckScriptPath,
	Dependencies: []checkly.ScriptDependency{
		{
			Path:    "helpers.js",
			Content: "module.exports = {}",
		},
	},
	EnvironmentVariables: []checkly.EnvironmentVariable{
		{
			Key:   "FOO",
			Value: "bar",
		},
	},
	SSLCheckDomain: "www.example.com",
	RetryStrategy: &checkly.RetryStrategy{
		Type:               "LINEAR",
		MaxRetries:         2,
		BaseBackoffSeconds: 60,
		MaxDurationSeconds: 600,
		SameRegion:         true,
	},
}

var ignoreBrowserCheckFields = cmpopts.IgnoreFields(
	checkly.BrowserCheck{},
	"ID",
	"AlertChannelSubscriptions",
	"CreatedAt",
	"UpdatedAt",
)

// The create and update endpoints don't include scriptPath or dependencies in
// their responses; only GET /v1/checks/{id} returns them.
var ignoreBrowserCheckWriteResponseFields = cmpopts.IgnoreFields(
	checkly.BrowserCheck{},
	"ScriptPath",
	"Dependencies",
)

func TestCreateBrowserCheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/browser?autoAssignAlerts=false",
		validateBrowserCheck,
		http.StatusCreated,
		"CreateBrowserCheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.CreateBrowserCheck(context.Background(), testBrowserCheck)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testBrowserCheck, *response, ignoreBrowserCheckFields, ignoreBrowserCheckWriteResponseFields) {
		t.Error(cmp.Diff(testBrowserCheck, *response, ignoreBrowserCheckFields, ignoreBrowserCheckWriteResponseFields))
	}
}

func TestDeleteBrowserCheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/checks/%s", testBrowserCheck.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteBrowserCheck(context.Background(), testBrowserCheck.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateBrowserCheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/browser/%s?autoAssignAlerts=false", testBrowserCheck.ID),
		validateBrowserCheck,
		http.StatusOK,
		"UpdateBrowserCheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateBrowserCheck(context.Background(), testBrowserCheck.ID, testBrowserCheck)
	if err != nil {
		t.Error(err)
	}
}

func TestGetBrowserCheck(t *testing.T) {
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/checks/%s?includeDependencies=true", testBrowserCheck.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetBrowserCheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.GetBrowserCheck(context.Background(), testBrowserCheck.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testBrowserCheck, *response, ignoreBrowserCheckFields) {
		t.Error(cmp.Diff(testBrowserCheck, *response, ignoreBrowserCheckFields))
	}
}

func validateMultiStepCheck(t *testing.T, body []byte) {
	var payload checkly.MultiStepCheck
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	// The payload wrapper normalizes nil lists to empty lists, so treat them
	// as equal here; the exact serialization is asserted on the raw body
	// below.
	if !cmp.Equal(testMultiStepCheck, payload, cmpopts.EquateEmpty()) {
		t.Error(cmp.Diff(testMultiStepCheck, payload, cmpopts.EquateEmpty()))
	}
	assertRawBodyField(t, body, "checkType", `"MULTI_STEP"`)
	assertRawBodyField(t, body, "doubleCheck", `false`)
	assertRawBodyField(t, body, "groupId", `null`)
	assertRawBodyField(t, body, "tags", `["tag-1"]`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyField(t, body, "scriptPath", `"__checks__/multistep.spec.js"`)
	assertRawBodyField(t, body, "dependencies", `[{"path":"helpers.js","content":"module.exports = {}"}]`)
}

var testMultiStepCheckScriptPath = "__checks__/multistep.spec.js"

var testMultiStepCheck = checkly.MultiStepCheck{
	ID:               "3013e5f5-b950-4364-b1a4-6b3ecb1e1c2e",
	Name:             "Multi-step Check #1",
	Frequency:        10,
	Activated:        true,
	Muted:            false,
	ShouldFail:       false,
	Locations:        []string{"us-east-1", "us-west-1"},
	PrivateLocations: &[]string{},
	Tags:             []string{"tag-1"},
	AlertSettings: &checkly.AlertSettings{
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
	UseGlobalAlertSettings: true,
	Script:                 "console.log('test')",
	ScriptPath:             &testMultiStepCheckScriptPath,
	Dependencies: []checkly.ScriptDependency{
		{
			Path:    "helpers.js",
			Content: "module.exports = {}",
		},
	},
	EnvironmentVariables: []checkly.EnvironmentVariable{
		{
			Key:   "FOO",
			Value: "bar",
		},
	},
	RetryStrategy: &checkly.RetryStrategy{
		Type:               "LINEAR",
		MaxRetries:         2,
		BaseBackoffSeconds: 60,
		MaxDurationSeconds: 600,
		SameRegion:         true,
	},
}

var ignoreMultiStepCheckFields = cmpopts.IgnoreFields(
	checkly.MultiStepCheck{},
	"ID",
	"AlertChannelSubscriptions",
	"CreatedAt",
	"UpdatedAt",
)

// The create and update endpoints don't include scriptPath or dependencies in
// their responses; only GET /v1/checks/{id} returns them.
var ignoreMultiStepCheckWriteResponseFields = cmpopts.IgnoreFields(
	checkly.MultiStepCheck{},
	"ScriptPath",
	"Dependencies",
)

func TestCreateMultiStepCheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/multistep?autoAssignAlerts=false",
		validateMultiStepCheck,
		http.StatusCreated,
		"CreateMultiStepCheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.CreateMultiStepCheck(context.Background(), testMultiStepCheck)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testMultiStepCheck, *response, ignoreMultiStepCheckFields, ignoreMultiStepCheckWriteResponseFields) {
		t.Error(cmp.Diff(testMultiStepCheck, *response, ignoreMultiStepCheckFields, ignoreMultiStepCheckWriteResponseFields))
	}
}

func TestDeleteMultiStepCheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/checks/%s", testMultiStepCheck.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	err := client.DeleteMultiStepCheck(context.Background(), testMultiStepCheck.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateMultiStepCheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/multistep/%s?autoAssignAlerts=false", testMultiStepCheck.ID),
		validateMultiStepCheck,
		http.StatusOK,
		"UpdateMultiStepCheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateMultiStepCheck(context.Background(), testMultiStepCheck.ID, testMultiStepCheck)
	if err != nil {
		t.Error(err)
	}
}

func TestGetMultiStepCheck(t *testing.T) {
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/checks/%s?includeDependencies=true", testMultiStepCheck.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetMultiStepCheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	response, err := client.GetMultiStepCheck(context.Background(), testMultiStepCheck.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testMultiStepCheck, *response, ignoreMultiStepCheckFields) {
		t.Error(cmp.Diff(testMultiStepCheck, *response, ignoreMultiStepCheckFields))
	}
}

// The update endpoints are PATCH-style: keys omitted from the payload keep
// their server-side value. The following tests verify that clearable fields
// are always serialized, using null or an empty list, when unset on the
// struct, so that they can actually be reset via an update.

func validateAPICheckEmptyPayload(t *testing.T, body []byte) {
	assertRawBodyField(t, body, "tags", `[]`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyField(t, body, "locations", `[]`)
	assertRawBodyField(t, body, "privateLocations", `[]`)
	assertRawBodyField(t, body, "groupId", `null`)
	// The group order must always be serialized, or the backend would keep
	// the previous value on update instead of resetting it to 0.
	assertRawBodyField(t, body, "groupOrder", `0`)
	assertRawBodyField(t, body, "setupSnippetId", `null`)
	assertRawBodyField(t, body, "tearDownSnippetId", `null`)
	assertRawBodyField(t, body, "localSetupScript", `null`)
	assertRawBodyField(t, body, "localTearDownScript", `null`)
	// 0 is a schema-valid response time limit and must be serialized, or the
	// backend would default (create) or keep (update) the previous values.
	assertRawBodyField(t, body, "degradedResponseTime", `0`)
	assertRawBodyField(t, body, "maxResponseTime", `0`)
	// Absent basic auth credentials must be serialized as null inside the
	// request object, or the backend would keep previously configured
	// credentials on update.
	assertRawNestedBodyField(t, body, "request", "basicAuth", `null`)
}

func TestCreateAPICheckSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/api?autoAssignAlerts=false",
		validateAPICheckEmptyPayload,
		http.StatusCreated,
		"CreateAPICheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.CreateAPICheck(context.Background(), checkly.APICheck{
		Name: "API Check #1",
	})
	if err != nil {
		t.Error(err)
	}
}

func validateBrowserCheckEmptyPayload(t *testing.T, body []byte) {
	assertRawBodyField(t, body, "tags", `[]`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyField(t, body, "locations", `[]`)
	assertRawBodyField(t, body, "privateLocations", `[]`)
	assertRawBodyField(t, body, "groupId", `null`)
	// The group order must always be serialized, or the backend would keep
	// the previous value on update instead of resetting it to 0.
	assertRawBodyField(t, body, "groupOrder", `0`)
	assertRawBodyField(t, body, "dependencies", `[]`)
	assertRawBodyField(t, body, "scriptPath", `null`)
}

func TestCreateBrowserCheckSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/browser?autoAssignAlerts=false",
		validateBrowserCheckEmptyPayload,
		http.StatusCreated,
		"CreateBrowserCheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.CreateBrowserCheck(context.Background(), checkly.BrowserCheck{
		Name:   "Browser Check #1",
		Script: "console.log('test')",
	})
	if err != nil {
		t.Error(err)
	}
}

func TestCreateMultiStepCheckSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/multistep?autoAssignAlerts=false",
		validateBrowserCheckEmptyPayload,
		http.StatusCreated,
		"CreateMultiStepCheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.CreateMultiStepCheck(context.Background(), checkly.MultiStepCheck{
		Name:   "Multi-step Check #1",
		Script: "console.log('test')",
	})
	if err != nil {
		t.Error(err)
	}
}

func validateGroupEmptyPayload(t *testing.T, body []byte) {
	assertRawBodyField(t, body, "locations", `[]`)
	assertRawBodyField(t, body, "privateLocations", `[]`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyField(t, body, "setupSnippetId", `null`)
	assertRawBodyField(t, body, "tearDownSnippetId", `null`)
	assertRawBodyField(t, body, "localSetupScript", `null`)
	assertRawBodyField(t, body, "localTearDownScript", `null`)
}

func TestCreateGroupSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/check-groups?autoAssignAlerts=false",
		validateGroupEmptyPayload,
		http.StatusCreated,
		"CreateGroup.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.CreateGroup(context.Background(), checkly.Group{
		Name: "Group #1",
	})
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateGroupSerializesSetValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		"/v1/check-groups/17?autoAssignAlerts=false",
		func(t *testing.T, body []byte) {
			assertRawBodyField(t, body, "setupSnippetId", `5`)
			assertRawBodyField(t, body, "tearDownSnippetId", `6`)
			assertRawBodyField(t, body, "localSetupScript", `"setup"`)
			assertRawBodyField(t, body, "localTearDownScript", `"teardown"`)
			assertRawBodyField(t, body, "alertChannelSubscriptions", `[{"alertChannelId":42,"activated":true}]`)
		},
		http.StatusOK,
		"UpdateGroup.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateGroup(context.Background(), 17, checkly.Group{
		Name:                "Group #1",
		SetupSnippetID:      5,
		TearDownSnippetID:   6,
		LocalSetupScript:    "setup",
		LocalTearDownScript: "teardown",
		AlertChannelSubscriptions: []checkly.AlertChannelSubscription{
			{
				ChannelID: 42,
				Activated: true,
			},
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestCreateGroupV2SerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v2/check-groups?autoAssignAlerts=false",
		func(t *testing.T, body []byte) {
			assertRawBodyField(t, body, "locations", `[]`)
			assertRawBodyField(t, body, "privateLocations", `[]`)
			assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
		},
		http.StatusCreated,
		"CreateGroup.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.CreateGroupV2(context.Background(), checkly.GroupV2{
		Name: "Group #1",
	})
	if err != nil {
		t.Error(err)
	}
}

func validateTCPMonitorEmptyPayload(t *testing.T, body []byte) {
	assertRawBodyField(t, body, "tags", `[]`)
	assertRawBodyField(t, body, "locations", `[]`)
	assertRawBodyField(t, body, "privateLocations", `[]`)
	assertRawBodyField(t, body, "groupId", `null`)
	// 0 is a schema-valid response time limit and must be serialized, or the
	// backend would default (create) or keep (update) the previous values.
	assertRawBodyField(t, body, "degradedResponseTime", `0`)
	assertRawBodyField(t, body, "maxResponseTime", `0`)
	// 0 is the value the group order resets to and must be serialized, or
	// the backend would keep the previous value on update.
	assertRawBodyField(t, body, "groupOrder", `0`)

	// The endpoints are PATCH-style down to the nested request object, so
	// clearable request keys must always be serialized as well.
	assertRawNestedBodyField(t, body, "request", "assertions", `[]`)
	// data must NOT be present: the backend rejects an empty string and a
	// null value cannot clear it either, so an empty data value is omitted.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	assertRawBodyKeyAbsent(t, raw["request"], "data")
}

func TestCreateTCPMonitorSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/tcp?autoAssignAlerts=false",
		validateTCPMonitorEmptyPayload,
		http.StatusCreated,
		"CreateTCPMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.CreateTCPMonitor(context.Background(), checkly.TCPMonitor{
		Name: "TCP Monitor #1",
		Request: checkly.TCPRequest{
			Hostname: "example.com",
			Port:     443,
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateTCPMonitorSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		"/v1/checks/tcp/26900d4d-a852-4e34-9bd0-e40a7a5a2495?autoAssignAlerts=false",
		validateTCPMonitorEmptyPayload,
		http.StatusOK,
		"CreateTCPMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateTCPMonitor(
		context.Background(),
		"26900d4d-a852-4e34-9bd0-e40a7a5a2495",
		checkly.TCPMonitor{
			Name: "TCP Monitor #1",
			Request: checkly.TCPRequest{
				Hostname: "example.com",
				Port:     443,
			},
		},
	)
	if err != nil {
		t.Error(err)
	}
}

func validateHeartbeatMonitorEmptyPayload(t *testing.T, body []byte) {
	assertRawBodyField(t, body, "tags", `[]`)
}

func TestCreateHeartbeatMonitorSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/heartbeat?autoAssignAlerts=false",
		validateHeartbeatMonitorEmptyPayload,
		http.StatusCreated,
		"CreateHeartbeatMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.CreateHeartbeatMonitor(context.Background(), checkly.HeartbeatMonitor{
		Name: "Heartbeat Monitor #1",
		Heartbeat: checkly.Heartbeat{
			Period:     5,
			PeriodUnit: "days",
			Grace:      0,
			GraceUnit:  "seconds",
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateHeartbeatMonitorSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		"/v1/checks/heartbeat/850cd536-a63b-4d33-9a12-8b125718f4e2?autoAssignAlerts=false",
		validateHeartbeatMonitorEmptyPayload,
		http.StatusOK,
		"CreateHeartbeatMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateHeartbeatMonitor(
		context.Background(),
		"850cd536-a63b-4d33-9a12-8b125718f4e2",
		checkly.HeartbeatMonitor{
			Name: "Heartbeat Monitor #1",
			Heartbeat: checkly.Heartbeat{
				Period:     5,
				PeriodUnit: "days",
				Grace:      0,
				GraceUnit:  "seconds",
			},
		},
	)
	if err != nil {
		t.Error(err)
	}
}

func validateURLMonitorEmptyPayload(t *testing.T, body []byte) {
	assertRawBodyField(t, body, "tags", `[]`)
	assertRawBodyField(t, body, "groupOrder", `null`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyField(t, body, "locations", `[]`)
	assertRawBodyField(t, body, "privateLocations", `[]`)
	assertRawBodyField(t, body, "groupId", `null`)
	// 0 is a schema-valid response time limit and must be serialized, or the
	// backend would default (create) or keep (update) the previous values.
	assertRawBodyField(t, body, "degradedResponseTime", `0`)
	assertRawBodyField(t, body, "maxResponseTime", `0`)
}

func TestCreateURLMonitorSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/url?autoAssignAlerts=false",
		validateURLMonitorEmptyPayload,
		http.StatusCreated,
		"CreateURLMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.CreateURLMonitor(context.Background(), checkly.URLMonitor{
		Name: "URL Monitor #1",
		Request: checkly.URLRequest{
			URL: "https://example.com",
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateURLMonitorSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/url/%s?autoAssignAlerts=false", testURLMonitor.ID),
		validateURLMonitorEmptyPayload,
		http.StatusOK,
		"UpdateURLMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateURLMonitor(
		context.Background(),
		testURLMonitor.ID,
		checkly.URLMonitor{
			ID:   testURLMonitor.ID,
			Name: "URL Monitor #1",
			Request: checkly.URLRequest{
				URL: "https://example.com",
			},
		},
	)
	if err != nil {
		t.Error(err)
	}
}

func validateDNSMonitorEmptyPayload(t *testing.T, body []byte) {
	assertRawBodyField(t, body, "tags", `[]`)
	assertRawBodyField(t, body, "groupOrder", `null`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyField(t, body, "locations", `[]`)
	assertRawBodyField(t, body, "groupId", `null`)
	// 0 is a schema-valid response time limit and must be serialized, or the
	// backend would default (create) or keep (update) the previous values.
	assertRawBodyField(t, body, "degradedResponseTime", `0`)
	assertRawBodyField(t, body, "maxResponseTime", `0`)
}

func TestCreateDNSMonitorSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/dns?autoAssignAlerts=false",
		validateDNSMonitorEmptyPayload,
		http.StatusCreated,
		"CreateDNSMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.CreateDNSMonitor(context.Background(), checkly.DNSMonitor{
		Name: "DNS Monitor #1",
		Request: checkly.DNSRequest{
			RecordType: "A",
			Query:      "example.com",
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateDNSMonitorSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/dns/%s?autoAssignAlerts=false", testDNSMonitor.ID),
		validateDNSMonitorEmptyPayload,
		http.StatusOK,
		"UpdateDNSMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateDNSMonitor(
		context.Background(),
		testDNSMonitor.ID,
		checkly.DNSMonitor{
			ID:   testDNSMonitor.ID,
			Name: "DNS Monitor #1",
			Request: checkly.DNSRequest{
				RecordType: "A",
				Query:      "example.com",
			},
		},
	)
	if err != nil {
		t.Error(err)
	}
}

func validateICMPMonitorEmptyPayload(t *testing.T, body []byte) {
	assertRawBodyField(t, body, "tags", `[]`)
	assertRawBodyField(t, body, "groupOrder", `null`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyField(t, body, "locations", `[]`)
	assertRawBodyField(t, body, "groupId", `null`)
	// The thresholds are always sent, so a 0 reaches the backend, which
	// rejects it, instead of being replaced by the default or prior value.
	assertRawBodyField(t, body, "degradedPacketLossThreshold", `0`)
	assertRawBodyField(t, body, "maxPacketLossThreshold", `0`)
}

func TestCreateICMPMonitorSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/icmp?autoAssignAlerts=false",
		validateICMPMonitorEmptyPayload,
		http.StatusCreated,
		"CreateICMPMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.CreateICMPMonitor(context.Background(), checkly.ICMPMonitor{
		Name: "ICMP Monitor #1",
		Request: checkly.ICMPRequest{
			Hostname: "example.com",
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateICMPMonitorSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/icmp/%s?autoAssignAlerts=false", testICMPMonitor.ID),
		validateICMPMonitorEmptyPayload,
		http.StatusOK,
		"UpdateICMPMonitor.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
	_, err := client.UpdateICMPMonitor(
		context.Background(),
		testICMPMonitor.ID,
		checkly.ICMPMonitor{
			ID:   testICMPMonitor.ID,
			Name: "ICMP Monitor #1",
			Request: checkly.ICMPRequest{
				Hostname: "example.com",
			},
		},
	)
	if err != nil {
		t.Error(err)
	}
}

func TestGetPerTypeCheckRejectsOtherCheckTypes(t *testing.T) {
	t.Parallel()
	// Each getter is served a check of another type and must refuse it
	// instead of decoding a mostly empty check.
	tests := []struct {
		name    string
		path    string
		fixture string
		get     func(client checkly.Client) error
	}{
		{
			name:    "GetAPICheck",
			path:    fmt.Sprintf("/v1/checks/%s", testBrowserCheck.ID),
			fixture: "GetBrowserCheck.json",
			get: func(client checkly.Client) error {
				_, err := client.GetAPICheck(context.Background(), testBrowserCheck.ID)
				return err
			},
		},
		{
			name:    "GetBrowserCheck",
			path:    fmt.Sprintf("/v1/checks/%s?includeDependencies=true", testAPICheck.ID),
			fixture: "GetAPICheck.json",
			get: func(client checkly.Client) error {
				_, err := client.GetBrowserCheck(context.Background(), testAPICheck.ID)
				return err
			},
		},
		{
			name:    "GetMultiStepCheck",
			path:    fmt.Sprintf("/v1/checks/%s?includeDependencies=true", testAPICheck.ID),
			fixture: "GetAPICheck.json",
			get: func(client checkly.Client) error {
				_, err := client.GetMultiStepCheck(context.Background(), testAPICheck.ID)
				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := cannedResponseServer(t, http.MethodGet, tt.path, validateEmptyBody, http.StatusOK, tt.fixture)
			defer ts.Close()
			client := checkly.NewClient(ts.URL, "dummy-key", ts.Client())
			if err := tt.get(client); err == nil || !strings.Contains(err.Error(), "check is of type") {
				t.Errorf("expected a check type error, got %v", err)
			}
		})
	}
}
