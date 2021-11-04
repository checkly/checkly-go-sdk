package checkly_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	checkly "github.com/checkly/checkly-go-sdk"
)

var wantCheckID = "73d29e72-6540-4bb5-967e-e07fa2c9465e"

var wantCheck = checkly.Check{
	Name:        "test",
	Type:        checkly.TypeAPI,
	Frequency:   10,
	Activated:   true,
	Muted:       false,
	DoubleCheck: true,
	SSLCheck:    true,
	ShouldFail:  false,
	Locations:   []string{"eu-west-1"},
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
		SSLCertificates: checkly.SSLCertificates{
			Enabled:        false,
			AlertThreshold: 30,
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
		data, err := os.Open(fmt.Sprintf("testdata/%s", filename))
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

var ignoreCheckFields = cmpopts.IgnoreFields(checkly.Check{}, "ID", "AlertChannelSubscriptions", "FrequencyOffset")

func TestAPIError(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks?autoAssignAlerts=false",
		validateAnything,
		http.StatusBadRequest,
		"BadRequest.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
	_, err := client.Create(context.Background(), checkly.Check{})
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

var wantGroupID int64 = 135

var wantGroup = checkly.Group{
	Name:        "test",
	Activated:   true,
	Muted:       false,
	Tags:        []string{"auto"},
	Locations:   []string{"eu-west-1"},
	Concurrency: 3,
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
	DoubleCheck:            true,
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
		SSLCertificates: checkly.SSLCertificates{
			Enabled:        true,
			AlertThreshold: 30,
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

var ignoreGroupFields = cmpopts.IgnoreFields(checkly.Group{}, "ID")

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
	ignored := cmpopts.IgnoreFields(checkly.Group{}, "ID", "AlertChannelSubscriptions")
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
	ignored := cmpopts.IgnoreFields(checkly.Group{}, "ID", "AlertChannelSubscriptions")
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
	ignored := cmpopts.IgnoreFields(checkly.Group{}, "ID", "AlertChannelSubscriptions")
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
		t.Errorf("Expected no errors, got %w", err)
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
	ID:             10,
	CustomUrl:      "string",
	CustomDomain:   "string",
	Logo:           "string",
	Header:         "string",
	Width:          "FULL",
	RefreshRate:    60,
	Paginate:       true,
	PaginationRate: 30,
	Tags:           []string{"string"},
	HideTags:       false,
}

var ignoreDashboardFields = cmpopts.IgnoreFields(checkly.Dashboard{}, "ID")

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
		fmt.Sprintf("/v1/dashboards/%d", testDashboard.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
	err := client.DeleteDashboard(context.Background(), testDashboard.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateDashboard(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/dashboards/%d", testDashboard.ID),
		validateDashboard,
		http.StatusOK,
		"CreateDashboard.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
	_, err := client.UpdateDashboard(context.Background(), testDashboard.ID, testDashboard)
	if err != nil {
		t.Error(err)
	}
}

func TestGetDashboard(t *testing.T) {
	return
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/dashboards/%d", testDashboard.ID),
		validateDashboard,
		http.StatusOK,
		"CreateDashboard.json",
	)
	defer ts.Close()
	client := checkly.NewClient(ts.URL, "dummy-key", ts.Client(), nil)
	ac, err := client.GetDashboard(context.Background(), testDashboard.ID)
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
