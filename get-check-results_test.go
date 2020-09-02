package checkly_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	checkly "github.com/checkly/checkly-go-sdk"
	"github.com/google/go-cmp/cmp"
)

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
	if !cmp.Equal(expectedFirstResult, results[0], nil) {
		expected, _ := json.Marshal(expectedFirstResult)
		got, _ := json.Marshal(results[0])
		t.Errorf("Got invalid result, expected: %s, \ngot: %s", expected, got)
	}
}

func TestGetCheckResultsWithFilters(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		"/v1/check-results/73d29e72-6540-4bb5-967e-e07fa2c9465e?checkType=API&from=1&hasFailures=1&limit=100&page=0&to=1000",
		validateEmptyBody,
		http.StatusOK,
		"GetCheckResults.json",
	)

	defer ts.Close()
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	results, err := client.GetCheckResults(wantCheckID, &checkly.CheckResultsFilter{
		Limit:       100,
		Page:        0,
		From:        1,
		To:          1000,
		CheckType:   checkly.TypeAPI,
		HasFailures: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 10 {
		t.Errorf("Expected to get 10 results got %d", len(results))
		return
	}
}
