package checkly_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	checkly "github.com/checkly/checkly-go-sdk"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.GetSSLMonitor(context.Background(), testSSLMonitor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testSSLMonitor, *response, ignoreSSLMonitorFields) {
		t.Error(cmp.Diff(testSSLMonitor, *response, ignoreSSLMonitorFields))
	}
}

// TestSSLConfigOmitsSecurityBaselineWhenNil asserts a create body with a nil
// SecurityBaseline omits the `securityBaseline` key so the server applies its
// default baseline rather than receiving an empty override, while a nil
// SSLClientCertificateID is sent as null.
func TestSSLConfigOmitsSecurityBaselineWhenNil(t *testing.T) {
	t.Parallel()
	monitor := testSSLMonitor
	monitor.Request.SSLConfig.SecurityBaseline = nil
	monitor.Request.SSLClientCertificateID = nil
	runPayloadCases(t, []payloadCase{{
		name:    "CreateSSLMonitor",
		method:  http.MethodPost,
		url:     "/v1/checks/ssl?autoAssignAlerts=false",
		status:  http.StatusCreated,
		fixture: "CreateSSLMonitor.json",
		validate: func(t *testing.T, body []byte) {
			assertRawBodyPathAbsent(t, body, "request", "sslConfig", "securityBaseline")
			assertRawBodyPath(t, body, `null`, "request", "sslClientCertificateId")
		},
		call: func(ctx context.Context, c checkly.Client) error {
			_, err := c.CreateSSLMonitor(ctx, monitor)
			return err
		},
	}})
}

// validateSSLMonitorEmptyPayload asserts that the keys the update endpoint
// would otherwise keep at their stored value are sent for an empty monitor.
func validateSSLMonitorEmptyPayload(t *testing.T, body []byte) {
	assertRawBodyField(t, body, "tags", `[]`)
	assertRawBodyField(t, body, "groupOrder", `null`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyPath(t, body, `[]`, "request", "assertions")
	assertRawBodyPath(t, body, `null`, "request", "sslClientCertificateId")
	assertRawBodyPath(t, body, `null`, "request", "sslConfig", "serverName")
}

func TestSSLMonitorSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	monitor := checkly.SSLMonitor{
		Name:    "SSL Monitor #1",
		Request: checkly.SSLRequest{SSLConfig: checkly.SSLConfig{Hostname: "example.com"}},
	}
	runPayloadCases(t, []payloadCase{
		{
			name:     "CreateSSLMonitor",
			method:   http.MethodPost,
			url:      "/v1/checks/ssl?autoAssignAlerts=false",
			status:   http.StatusCreated,
			fixture:  "CreateSSLMonitor.json",
			validate: validateSSLMonitorEmptyPayload,
			call: func(ctx context.Context, c checkly.Client) error {
				_, err := c.CreateSSLMonitor(ctx, monitor)
				return err
			},
		},
		{
			name:     "UpdateSSLMonitor",
			method:   http.MethodPut,
			url:      fmt.Sprintf("/v1/checks/ssl/%s?autoAssignAlerts=false", testSSLMonitor.ID),
			status:   http.StatusOK,
			fixture:  "UpdateSSLMonitor.json",
			validate: validateSSLMonitorEmptyPayload,
			call: func(ctx context.Context, c checkly.Client) error {
				_, err := c.UpdateSSLMonitor(ctx, testSSLMonitor.ID, monitor)
				return err
			},
		},
	})
}
