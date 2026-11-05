package checkly_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	if !cmp.Equal(wantCheck, gotCheck) {
		t.Error(cmp.Diff(wantCheck, gotCheck))
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
			client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
			if err := tt.get(client); err == nil || !strings.Contains(err.Error(), "check is of type") {
				t.Errorf("expected a check type error, got %v", err)
			}
		})
	}
}
