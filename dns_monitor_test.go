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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.GetDNSMonitor(context.Background(), testDNSMonitor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testDNSMonitor, *response, ignoreDNSMonitorFields) {
		t.Error(cmp.Diff(testDNSMonitor, *response, ignoreDNSMonitorFields))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
