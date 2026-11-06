package checkly_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/checkly/checkly-go-sdk/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

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
	ID:               "ab1c2d3e-e4f5-6789-0abc-def012345678",
	Name:             "ICMP Monitor #1",
	Frequency:        10,
	FrequencyOffset:  20,
	Activated:        true,
	Muted:            false,
	Locations:        []string{"us-east-1"},
	PrivateLocations: &[]string{},
	Tags: []string{
		"tag-1",
	},
	DegradedPacketLossThreshold: 5,
	MaxPacketLossThreshold:      50,
	AlertSettings: &checkly.AlertSettings{
		EscalationType: checkly.EscalationTypeRunBased,
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.GetICMPMonitor(context.Background(), testICMPMonitor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testICMPMonitor, *response, ignoreICMPMonitorFields) {
		t.Error(cmp.Diff(testICMPMonitor, *response, ignoreICMPMonitorFields))
	}
}

func validateICMPMonitorEmptyPayload(t *testing.T, body []byte) {
	assertRawBodyField(t, body, "tags", `[]`)
	// The API rejects null assertions.
	assertRawNestedBodyField(t, body, "request", "assertions", `[]`)
	assertRawBodyField(t, body, "groupOrder", `null`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyField(t, body, "locations", `[]`)
	assertRawBodyField(t, body, "privateLocations", `[]`)
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
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

func TestICMPMonitorSendsPrivateLocations(t *testing.T) {
	t.Parallel()
	monitor := testICMPMonitor
	monitor.PrivateLocations = &[]string{"my-private-location"}
	validate := func(t *testing.T, body []byte) {
		assertRawBodyField(t, body, "privateLocations", `["my-private-location"]`)
	}
	runPayloadCases(t, []payloadCase{
		{
			name:     "CreateICMPMonitor",
			method:   http.MethodPost,
			url:      "/v1/checks/icmp?autoAssignAlerts=false",
			status:   http.StatusCreated,
			fixture:  "CreateICMPMonitor.json",
			validate: validate,
			call: func(ctx context.Context, c checkly.Client) error {
				_, err := c.CreateICMPMonitor(ctx, monitor)
				return err
			},
		},
		{
			name:     "UpdateICMPMonitor",
			method:   http.MethodPut,
			url:      fmt.Sprintf("/v1/checks/icmp/%s?autoAssignAlerts=false", monitor.ID),
			status:   http.StatusOK,
			fixture:  "UpdateICMPMonitor.json",
			validate: validate,
			call: func(ctx context.Context, c checkly.Client) error {
				_, err := c.UpdateICMPMonitor(ctx, monitor.ID, monitor)
				return err
			},
		},
	})
}
