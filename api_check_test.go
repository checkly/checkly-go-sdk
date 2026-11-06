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

func validateAPICheck(t *testing.T, body []byte) {
	var payload checkly.APICheck
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	// The payload wrapper normalizes nil lists to empty lists, so treat them
	// as equal here; the exact serialization is asserted on the raw body
	// below.
	if !cmp.Equal(testAPICheck, payload, cmpopts.EquateEmpty()) {
		t.Error(cmp.Diff(testAPICheck, payload, cmpopts.EquateEmpty()))
	}
	assertRawBodyField(t, body, "checkType", `"API"`)
	assertRawBodyField(t, body, "doubleCheck", `false`)
	assertRawBodyField(t, body, "groupId", `null`)
	assertRawBodyField(t, body, "tags", `["tag-1"]`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyField(t, body, "setupSnippetId", `123`)
	assertRawBodyField(t, body, "tearDownSnippetId", `456`)
	assertRawBodyField(t, body, "degradedResponseTime", `5000`)
	assertRawBodyField(t, body, "maxResponseTime", `20000`)
}

var (
	testAPICheckDescription         = "API check description"
	testAPICheckLocalSetupScript    = "console.log('setup')"
	testAPICheckLocalTearDownScript = "console.log('teardown')"
)

var testAPICheck = checkly.APICheck{
	ID:               "80f57ec9-2ee7-4152-b6ae-e1f5c47eae86",
	Name:             "API Check #1",
	Description:      &testAPICheckDescription,
	Frequency:        10,
	FrequencyOffset:  17,
	Activated:        true,
	Muted:            false,
	ShouldFail:       false,
	Locations:        []string{"us-east-1", "us-west-1"},
	PrivateLocations: &[]string{},
	Tags:             []string{"tag-1"},
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
	SetupSnippetID:         123,
	TearDownSnippetID:      456,
	LocalSetupScript:       &testAPICheckLocalSetupScript,
	LocalTearDownScript:    &testAPICheckLocalTearDownScript,
	RetryStrategy: &checkly.RetryStrategy{
		Type:               "LINEAR",
		MaxRetries:         2,
		BaseBackoffSeconds: 60,
		MaxDurationSeconds: 600,
		SameRegion:         true,
	},
	Request: checkly.Request{
		Method:          "GET",
		URL:             "https://api.checklyhq.com",
		FollowRedirects: true,
		SkipSSL:         false,
		Headers: []checkly.KeyValue{
			{
				Key:   "X-Test",
				Value: "foo",
			},
		},
		QueryParameters: []checkly.KeyValue{
			{
				Key:   "query",
				Value: "param",
			},
		},
		Assertions: []checkly.Assertion{
			{
				Source:     "STATUS_CODE",
				Comparison: "EQUALS",
				Target:     "200",
			},
		},
	},
}

var ignoreAPICheckFields = cmpopts.IgnoreFields(
	checkly.APICheck{},
	"ID",
	"AlertChannelSubscriptions",
	"CreatedAt",
	"UpdatedAt",
)

func TestCreateAPICheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/api?autoAssignAlerts=false",
		validateAPICheck,
		http.StatusCreated,
		"CreateAPICheck.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.CreateAPICheck(context.Background(), testAPICheck)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testAPICheck, *response, ignoreAPICheckFields) {
		t.Error(cmp.Diff(testAPICheck, *response, ignoreAPICheckFields))
	}
}

func TestDeleteAPICheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/checks/%s", testAPICheck.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	err := client.DeleteAPICheck(context.Background(), testAPICheck.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateAPICheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/api/%s?autoAssignAlerts=false", testAPICheck.ID),
		validateAPICheck,
		http.StatusOK,
		"UpdateAPICheck.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	_, err := client.UpdateAPICheck(context.Background(), testAPICheck.ID, testAPICheck)
	if err != nil {
		t.Error(err)
	}
}

func TestGetAPICheck(t *testing.T) {
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/checks/%s", testAPICheck.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetAPICheck.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.GetAPICheck(context.Background(), testAPICheck.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testAPICheck, *response, ignoreAPICheckFields) {
		t.Error(cmp.Diff(testAPICheck, *response, ignoreAPICheckFields))
	}
}

// The update endpoints are PATCH-style: keys omitted from the payload keep
// their server-side value. The following tests verify that clearable fields
// are always serialized, using null or an empty list, when unset on the
// struct, so that they can actually be reset via an update.

func validateAPICheckEmptyPayload(t *testing.T, body []byte) {
	assertRawBodyField(t, body, "tags", `[]`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyField(t, body, "locations", `[]`)
	assertRawBodyField(t, body, "privateLocations", `[]`)
	assertRawBodyField(t, body, "groupId", `null`)
	// The group order must always be serialized, or the backend would keep
	// the previous value on update instead of resetting it to 0.
	assertRawBodyField(t, body, "groupOrder", `0`)
	assertRawBodyField(t, body, "setupSnippetId", `null`)
	assertRawBodyField(t, body, "tearDownSnippetId", `null`)
	assertRawBodyField(t, body, "localSetupScript", `null`)
	assertRawBodyField(t, body, "localTearDownScript", `null`)
	// 0 is a schema-valid response time limit and must be serialized, or the
	// backend would default (create) or keep (update) the previous values.
	assertRawBodyField(t, body, "degradedResponseTime", `0`)
	assertRawBodyField(t, body, "maxResponseTime", `0`)
	// Absent basic auth credentials must be serialized as null inside the
	// request object, or the backend would keep previously configured
	// credentials on update.
	assertRawNestedBodyField(t, body, "request", "basicAuth", `null`)
	// The API rejects null for the request lists.
	assertRawNestedBodyField(t, body, "request", "headers", `[]`)
	assertRawNestedBodyField(t, body, "request", "queryParameters", `[]`)
	assertRawNestedBodyField(t, body, "request", "assertions", `[]`)
}

func TestAPICheckSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	check := checkly.APICheck{Name: "API Check #1"}
	runPayloadCases(t, []payloadCase{
		{
			name:     "CreateAPICheck",
			method:   http.MethodPost,
			url:      "/v1/checks/api?autoAssignAlerts=false",
			status:   http.StatusCreated,
			fixture:  "CreateAPICheck.json",
			validate: validateAPICheckEmptyPayload,
			call: func(ctx context.Context, c checkly.Client) error {
				_, err := c.CreateAPICheck(ctx, check)
				return err
			},
		},
		{
			name:     "UpdateAPICheck",
			method:   http.MethodPut,
			url:      fmt.Sprintf("/v1/checks/api/%s?autoAssignAlerts=false", testAPICheck.ID),
			status:   http.StatusOK,
			fixture:  "UpdateAPICheck.json",
			validate: validateAPICheckEmptyPayload,
			call: func(ctx context.Context, c checkly.Client) error {
				_, err := c.UpdateAPICheck(ctx, testAPICheck.ID, check)
				return err
			},
		},
	})
}
