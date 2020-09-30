package checkly_test

import (
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

	checkly "github.com/checkly/checkly-go-sdk"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
	UseGlobalAlertSettings: false,
	DegradedResponseTime:   15000,
	MaxResponseTime:        30000,
	GroupID:                0,
	GroupOrder:             0,
}

func cannedResponseServer(t *testing.T, wantMethod string, wantURL string, validate func(*testing.T, []byte), status int, filename string) *httptest.Server {
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wantMethod != r.Method {
			t.Errorf("want %q request, got %q", wantMethod, r.Method)
		}
		if r.URL.String() != wantURL {
			t.Errorf("want %q, got %q", wantURL, r.URL.String())
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		validate(t, body)
		w.WriteHeader(status)
		data, err := os.Open(fmt.Sprintf("testdata/%s", filename))
		if err != nil {
			t.Fatal(err)
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

var ignoreCheckFields = cmpopts.IgnoreFields(checkly.Check{}, "ID")

func TestAPIError(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks",
		validateAnything,
		http.StatusBadRequest,
		"BadRequest.json",
	)
	defer ts.Close()
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	_, err := client.Create(checkly.Check{})
	if err == nil {
		t.Fatal("want error when API returns 'bad request' status, got nil")
	}
	if !strings.Contains(err.Error(), "frequency") {
		t.Errorf("want API error value to contain 'frequency', got %q", err.Error())
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks",
		validateCheck,
		http.StatusCreated,
		"CreateCheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	gotCheck, err := client.Create(wantCheck)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(wantCheck, gotCheck, ignoreCheckFields) {
		t.Error(cmp.Diff(wantCheck, gotCheck, ignoreCheckFields))
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
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	gotCheck, err := client.Get(wantCheckID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(wantCheck, gotCheck, ignoreCheckFields) {
		t.Error(cmp.Diff(wantCheck, gotCheck, ignoreCheckFields))
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/%s", wantCheckID),
		validateCheck,
		http.StatusOK,
		"UpdateCheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	gotCheck, err := client.Update(wantCheckID, wantCheck)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(wantCheck, gotCheck, ignoreCheckFields) {
		t.Error(cmp.Diff(wantCheck, gotCheck, ignoreCheckFields))
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
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	err := client.Delete(wantCheckID)
	if err != nil {
		t.Fatal(err)
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
	AlertChannelSubscriptions: []checkly.Subscription{
		{
			Activated:      true,
			AlertChannelID: 2996,
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
		"/v1/check-groups",
		validateGroup,
		http.StatusCreated,
		"CreateGroup.json",
	)
	defer ts.Close()
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	gotGroup, err := client.CreateGroup(wantGroup)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(wantGroup, gotGroup, ignoreGroupFields) {
		t.Error(cmp.Diff(wantGroup, gotGroup, ignoreGroupFields))
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
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	gotGroup, err := client.GetGroup(wantGroupID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(wantGroup, gotGroup, ignoreGroupFields) {
		t.Error(cmp.Diff(wantGroup, gotGroup, ignoreGroupFields))
	}
}

func TestUpdateGroup(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/check-groups/%d", wantGroupID),
		validateGroup,
		http.StatusOK,
		"UpdateGroup.json",
	)
	defer ts.Close()
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	gotGroup, err := client.UpdateGroup(wantGroupID, wantGroup)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(wantGroup, gotGroup, ignoreGroupFields) {
		t.Error(cmp.Diff(wantGroup, gotGroup, ignoreGroupFields))
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
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	err := client.DeleteGroup(wantGroupID)
	if err != nil {
		t.Fatal(err)
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
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	result, err := client.GetCheckResult(wantCheckID, resultID)
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
	if !cmp.Equal(expectedResult, result, nil) {
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
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	results, err := client.GetCheckResults(wantCheckID, nil)
	if err != nil {
		t.Fatal(err)
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
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	results, err := client.GetCheckResults("73d29e72-6540", &checkly.CheckResultsFilter{
		Limit:       100,
		Page:        1,
		From:        1,
		To:          1000,
		CheckType:   checkly.TypeAPI,
		HasFailures: true,
		Location:    "us-east-1",
	})
	if err != nil {
		t.Fatal(err)
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
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	results, err := client.GetCheckResults("73d29e72-6540", &checkly.CheckResultsFilter{})
	if err != nil {
		t.Fatal(err)
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
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	gotSnippet, err := client.CreateSnippet(testSnippet)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testSnippet, gotSnippet, ignoreSnippetFields) {
		t.Error(cmp.Diff(testSnippet, gotSnippet, ignoreSnippetFields))
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
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	gotSnippet, err := client.GetSnippet(testSnippet.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testSnippet, gotSnippet, ignoreSnippetFields) {
		t.Error(cmp.Diff(testSnippet, gotSnippet, ignoreSnippetFields))
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
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	gotSnippet, err := client.UpdateSnippet(testSnippet.ID, testSnippet)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testSnippet, gotSnippet, ignoreSnippetFields) {
		t.Error(cmp.Diff(testSnippet, gotSnippet, ignoreSnippetFields))
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
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	err := client.DeleteSnippet(testSnippet.ID)
	if err != nil {
		t.Fatal(err)
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
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	result, err := client.CreateEnvironmentVariable(testEnvVariable)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testEnvVariable, result, nil) {
		t.Error(cmp.Diff(testEnvVariable, result, nil))
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
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	result, err := client.GetEnvironmentVariable(testEnvVariable.Key)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testEnvVariable, result, nil) {
		t.Error(cmp.Diff(testEnvVariable, result, nil))
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
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	result, err := client.UpdateEnvironmentVariable(testEnvVariable.Key, testEnvVariable)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testEnvVariable, result, nil) {
		t.Error(cmp.Diff(testEnvVariable, result, nil))
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
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	err := client.DeleteEnvironmentVariable(testEnvVariable.Key)
	if err != nil {
		t.Fatal(err)
	}
}
