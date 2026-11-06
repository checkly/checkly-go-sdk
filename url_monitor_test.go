package checkly_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	checkly "github.com/checkly/checkly-go-sdk/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

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
		EscalationType: checkly.EscalationTypeRunBased,
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.GetURLMonitor(context.Background(), testURLMonitor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testURLMonitor, *response, ignoreURLMonitorFields) {
		t.Error(cmp.Diff(testURLMonitor, *response, ignoreURLMonitorFields))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
