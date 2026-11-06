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

func validateMultiStepCheck(t *testing.T, body []byte) {
	var payload checkly.MultiStepCheck
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	// The payload wrapper normalizes nil lists to empty lists, so treat them
	// as equal here; the exact serialization is asserted on the raw body
	// below.
	if !cmp.Equal(testMultiStepCheck, payload, cmpopts.EquateEmpty()) {
		t.Error(cmp.Diff(testMultiStepCheck, payload, cmpopts.EquateEmpty()))
	}
	assertRawBodyField(t, body, "checkType", `"MULTI_STEP"`)
	assertRawBodyField(t, body, "doubleCheck", `false`)
	assertRawBodyField(t, body, "groupId", `null`)
	assertRawBodyField(t, body, "tags", `["tag-1"]`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyField(t, body, "scriptPath", `"__checks__/multistep.spec.js"`)
	assertRawBodyField(t, body, "dependencies", `[{"path":"helpers.js","content":"module.exports = {}"}]`)
}

var testMultiStepCheckScriptPath = "__checks__/multistep.spec.js"

var testMultiStepCheck = checkly.MultiStepCheck{
	ID:               "3013e5f5-b950-4364-b1a4-6b3ecb1e1c2e",
	Name:             "Multi-step Check #1",
	Frequency:        10,
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
	Script:                 "console.log('test')",
	ScriptPath:             &testMultiStepCheckScriptPath,
	Dependencies: []checkly.ScriptDependency{
		{
			Path:    "helpers.js",
			Content: "module.exports = {}",
		},
	},
	EnvironmentVariables: []checkly.EnvironmentVariable{
		{
			Key:   "FOO",
			Value: "bar",
		},
	},
	RetryStrategy: &checkly.RetryStrategy{
		Type:               "LINEAR",
		MaxRetries:         2,
		BaseBackoffSeconds: 60,
		MaxDurationSeconds: 600,
		SameRegion:         true,
	},
}

var ignoreMultiStepCheckFields = cmpopts.IgnoreFields(
	checkly.MultiStepCheck{},
	"ID",
	"AlertChannelSubscriptions",
	"CreatedAt",
	"UpdatedAt",
)

// The create and update endpoints don't include scriptPath or dependencies in
// their responses; only GET /v1/checks/{id} returns them.
var ignoreMultiStepCheckWriteResponseFields = cmpopts.IgnoreFields(
	checkly.MultiStepCheck{},
	"ScriptPath",
	"Dependencies",
)

func TestCreateMultiStepCheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/multistep?autoAssignAlerts=false",
		validateMultiStepCheck,
		http.StatusCreated,
		"CreateMultiStepCheck.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.CreateMultiStepCheck(context.Background(), testMultiStepCheck)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testMultiStepCheck, *response, ignoreMultiStepCheckFields, ignoreMultiStepCheckWriteResponseFields) {
		t.Error(cmp.Diff(testMultiStepCheck, *response, ignoreMultiStepCheckFields, ignoreMultiStepCheckWriteResponseFields))
	}
}

func TestDeleteMultiStepCheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/checks/%s", testMultiStepCheck.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	err := client.DeleteMultiStepCheck(context.Background(), testMultiStepCheck.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateMultiStepCheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/multistep/%s?autoAssignAlerts=false", testMultiStepCheck.ID),
		validateMultiStepCheck,
		http.StatusOK,
		"UpdateMultiStepCheck.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	_, err := client.UpdateMultiStepCheck(context.Background(), testMultiStepCheck.ID, testMultiStepCheck)
	if err != nil {
		t.Error(err)
	}
}

func TestGetMultiStepCheck(t *testing.T) {
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/checks/%s?includeDependencies=true", testMultiStepCheck.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetMultiStepCheck.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.GetMultiStepCheck(context.Background(), testMultiStepCheck.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testMultiStepCheck, *response, ignoreMultiStepCheckFields) {
		t.Error(cmp.Diff(testMultiStepCheck, *response, ignoreMultiStepCheckFields))
	}
}

func TestCreateMultiStepCheckSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/multistep?autoAssignAlerts=false",
		validateBrowserCheckEmptyPayload,
		http.StatusCreated,
		"CreateMultiStepCheck.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	_, err := client.CreateMultiStepCheck(context.Background(), checkly.MultiStepCheck{
		Name:   "Multi-step Check #1",
		Script: "console.log('test')",
	})
	if err != nil {
		t.Error(err)
	}
}
