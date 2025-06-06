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
	"path"
	"path/filepath"
	"reflect"
	"strconv"
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
	Script: "foo",
	EnvironmentVariables: []checkly.EnvironmentVariable{
		{
			Key:   "ENVTEST",
			Value: "Hello world",
		},
	},
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
	"AlertSettings.SSLCertificates", "PrivateLocations", "SSLCheckDomain")

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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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

func TestCreate(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks?autoAssignAlerts=false",
		validateCheck,
		http.StatusCreated,
		"CreateCheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
	gotCheck, err := client.Create(context.Background(), wantCheck)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(wantCheck, *gotCheck, ignoreCheckFields) {
		t.Error(cmp.Diff(wantCheck, *gotCheck, ignoreCheckFields))
	}
}

func TestGet(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/checks/%s", wantCheckID),
		validateEmptyBody,
		http.StatusOK,
		"GetCheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
	gotCheck, err := client.Get(context.Background(), wantCheckID)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(wantCheck, *gotCheck, ignoreCheckFields) {
		t.Error(cmp.Diff(wantCheck, *gotCheck, ignoreCheckFields))
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/%s?autoAssignAlerts=false", wantCheckID),
		validateCheck,
		http.StatusOK,
		"UpdateCheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
	gotCheck, err := client.Update(context.Background(), wantCheckID, wantCheck)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(wantCheck, *gotCheck, ignoreCheckFields) {
		t.Error(cmp.Diff(wantCheck, *gotCheck, ignoreCheckFields))
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/checks/%s", wantCheckID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
	err := client.Delete(context.Background(), wantCheckID)
	if err != nil {
		t.Error(err)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
	err := client.DeleteCheck(context.Background(), wantCheckID)
	if err != nil {
		t.Error(err)
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

var ignoreGroupFields = cmpopts.IgnoreFields(checkly.Group{}, "ID", "AlertSettings.SSLCertificates", "PrivateLocations")

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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	return
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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

var testDashboard = checkly.Dashboard{
	DashboardID:        "abcd1234",
	CustomDomain:       "is.checkly.online",
	CustomUrl:          "status-page",
	Logo:               "https://www.checklyhq.com/images/text_racoon_logo.svg",
	Favicon:            "https://www.checklyhq.com/images/text_racoon_logo.svg",
	Link:               "https://www.checklyhq.com",
	Description:        "Checkly status page",
	Header:             "Status",
	Width:              "FULL",
	RefreshRate:        60,
	Paginate:           true,
	PaginationRate:     30,
	Tags:               []string{"api"},
	HideTags:           false,
	ChecksPerPage:      15,
	UseTagsAndOperator: true,
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
	_, err := client.UpdateDashboard(context.Background(), testDashboard.DashboardID, testDashboard)
	if err != nil {
		t.Error(err)
	}
}

func TestGetDashboard(t *testing.T) {
	return
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/dashboards/%s", testDashboard.DashboardID),
		validateDashboard,
		http.StatusOK,
		"CreateDashboard.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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

var testMaintenanceWindow = checkly.MaintenanceWindow{
	ID:             1,
	Name:           "TEST",
	StartsAt:       "2014-08-24T00:00:00.000Z",
	EndsAt:         "2014-08-24T00:00:00.000Z",
	RepeatUnit:     "MONTH",
	RepeatEndsAt:   "2014-08-24T00:00:00.000Z",
	RepeatInterval: 10,
	CreatedAt:      "2013-08-24",
	UpdatedAt:      "2014-08-24",
	Tags:           []string{"string"},
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
	_, err := client.UpdateMaintenanceWindow(context.Background(), testMaintenanceWindow.ID, testMaintenanceWindow)
	if err != nil {
		t.Error(err)
	}
}

func TestGetMaintenanceWindow(t *testing.T) {
	return
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/maintenance-windows/%d", testMaintenanceWindow.ID),
		validateMaintenanceWindow,
		http.StatusOK,
		"CreateMaintenanceWindow.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
	mw, err := client.GetMaintenanceWindow(context.Background(), testMaintenanceWindow.ID)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testMaintenanceWindow, *mw, ignoreMaintenanceWindowFields) {
		t.Error(cmp.Diff(testMaintenanceWindow, *mw, ignoreMaintenanceWindowFields))
	}
}

var testTriggerCheck = checkly.TriggerCheck{
	ID:        1,
	CheckId:   "721d28d6-149f-4f32-95e1-e497b23156f4",
	Token:     "MDMnt4oPjBBZ",
	CreatedAt: "2013-08-24",
	UpdatedAt: "2013-08-24",
	CalledAt:  "2013-08-24",
	URL:       "https://127.0.0.1:35647/checks/721d28d6-149f-4f32-95e1-e497b23156f4/trigger/MDMnt4oPjBBZ",
}

var ignoreTriggerCheck = cmpopts.IgnoreFields(checkly.TriggerCheck{}, "ID", "CreatedAt", "UpdatedAt", "CalledAt", "URL")

func TestCreateTriggerCheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		fmt.Sprintf("/v1/triggers/checks/%s", testTriggerCheck.CheckId),
		validateEmptyBody,
		http.StatusCreated,
		"CreateTriggerCheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
	gotTriggerCheck, err := client.CreateTriggerCheck(context.Background(), testTriggerCheck.CheckId)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testTriggerCheck, *gotTriggerCheck, ignoreTriggerCheck) {
		t.Error(cmp.Diff(testTriggerCheck, *gotTriggerCheck, ignoreTriggerCheck))
	}
}
func TestDeleteTriggerCheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		path.Join("/v1/triggers/checks/", testTriggerCheck.CheckId),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
	err := client.DeleteTriggerCheck(context.Background(), testTriggerCheck.CheckId)
	if err != nil {
		t.Error(err)
	}
}
func TestGetTriggerCheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/triggers/checks/%s", testTriggerCheck.CheckId),
		validateEmptyBody,
		http.StatusOK,
		"CreateTriggerCheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
	gotTriggerCheck, err := client.GetTriggerCheck(context.Background(), testTriggerCheck.CheckId)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testTriggerCheck, *gotTriggerCheck, ignoreTriggerCheck) {
		t.Error(cmp.Diff(testTriggerCheck, *gotTriggerCheck, ignoreTriggerCheck))
	}
}

var testTriggerGroup = checkly.TriggerGroup{
	ID:        1,
	GroupId:   215,
	Token:     "MDMnt4oPjBBZ",
	CreatedAt: "2013-08-24",
	UpdatedAt: "2013-08-24",
	CalledAt:  "2013-08-24",
	URL:       "https://127.0.0.1:35647/check-groups/215/trigger/MDMnt4oPjBBZ",
}

var ignoreTriggerGroup = cmpopts.IgnoreFields(checkly.TriggerGroup{}, "ID", "CreatedAt", "UpdatedAt", "CalledAt", "URL")

func TestCreateTriggerGroup(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		fmt.Sprintf("/v1/triggers/check-groups/%d", testTriggerGroup.GroupId),
		validateEmptyBody,
		http.StatusCreated,
		"CreateTriggerGroup.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
	gotTriggerGroup, err := client.CreateTriggerGroup(context.Background(), testTriggerGroup.GroupId)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testTriggerGroup, *gotTriggerGroup, ignoreTriggerGroup) {
		t.Error(cmp.Diff(testTriggerGroup, *gotTriggerGroup, ignoreTriggerGroup))
	}
}
func TestDeleteTriggerGroup(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		path.Join("/v1/triggers/check-groups/", strconv.FormatInt(testTriggerGroup.GroupId, 10)),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
	err := client.DeleteTriggerGroup(context.Background(), testTriggerGroup.GroupId)
	if err != nil {
		t.Error(err)
	}
}
func TestGetTriggerGroup(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/triggers/check-groups/%d", testTriggerGroup.GroupId),
		validateEmptyBody,
		http.StatusOK,
		"CreateTriggerGroup.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
	gotTriggerGroup, err := client.GetTriggerGroup(context.Background(), testTriggerGroup.GroupId)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testTriggerGroup, *gotTriggerGroup, ignoreTriggerGroup) {
		t.Error(cmp.Diff(testTriggerGroup, *gotTriggerGroup, ignoreTriggerGroup))
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
	_, err := client.UpdatePrivateLocation(context.Background(), testPrivateLocation.ID, testPrivateLocation)
	if err != nil {
		t.Error(err)
	}
}

func TestGetPrivateLocation(t *testing.T) {
	return
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/private-locations/%s", testPrivateLocation.ID),
		validatePrivateLocation,
		http.StatusOK,
		"CreatePrivateLocation.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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

	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
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
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
	response, err := client.GetStatusPage(context.Background(), testStatusPage.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testStatusPage, *response, ignoreStatusPageFields) {
		t.Error(cmp.Diff(testStatusPage, *response, ignoreStatusPageFields))
	}
}
