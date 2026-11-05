package checkly_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	checkly "github.com/checkly/checkly-go-sdk"
	"github.com/google/go-cmp/cmp"
)

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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	result, err := client.GetCheckResult(context.Background(), wantCheckID, resultID)
	if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	}
	startedAt, _ := time.Parse(time.RFC3339, "2020-09-02T11:19:06.283Z")
	stoppedAt, _ := time.Parse(time.RFC3339, "2020-09-02T11:19:06.413Z")
	createdAt, _ := time.Parse(time.RFC3339, "2020-09-02T11:19:06.681Z")
	// Ignore api check result comparison for now
	result.APICheckResult = &checkly.APICheckResult{}
	expectedResult := checkly.CheckResult{
		ID:                  "580c4e71-0109-45ba-9130-887ff01e1a7f",
		HasErrors:           false,
		HasFailures:         false,
		RunLocation:         "eu-central-1",
		StartedAt:           startedAt,
		StoppedAt:           stoppedAt,
		ResponseTime:        129,
		APICheckResult:      &checkly.APICheckResult{},
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
	actualFirstResult.APICheckResult = &checkly.APICheckResult{}
	expectedFirstResult := checkly.CheckResult{
		ID:                  "580c4e71-0109-45ba-9130-887ff01e1a7f",
		HasErrors:           false,
		HasFailures:         false,
		RunLocation:         "eu-central-1",
		StartedAt:           startedAt,
		StoppedAt:           stoppedAt,
		ResponseTime:        129,
		APICheckResult:      &checkly.APICheckResult{},
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
