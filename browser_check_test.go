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

func validateBrowserCheck(t *testing.T, body []byte) {
	var payload checkly.BrowserCheck
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	// The payload wrapper normalizes nil lists to empty lists, so treat them
	// as equal here; the exact serialization is asserted on the raw body
	// below.
	if !cmp.Equal(testBrowserCheck, payload, cmpopts.EquateEmpty()) {
		t.Error(cmp.Diff(testBrowserCheck, payload, cmpopts.EquateEmpty()))
	}
	assertRawBodyField(t, body, "checkType", `"BROWSER"`)
	assertRawBodyField(t, body, "doubleCheck", `false`)
	assertRawBodyField(t, body, "groupId", `null`)
	assertRawBodyField(t, body, "tags", `["tag-1"]`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyField(t, body, "scriptPath", `"__checks__/browser.spec.js"`)
	assertRawBodyField(t, body, "dependencies", `[{"path":"helpers.js","content":"module.exports = {}"}]`)
}

var testBrowserCheckScriptPath = "__checks__/browser.spec.js"

var testBrowserCheck = checkly.BrowserCheck{
	ID:               "0895dfd1-6b18-4cc9-a2c9-67ec9b45f7bd",
	Name:             "Browser Check #1",
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
	ScriptPath:             &testBrowserCheckScriptPath,
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
	SSLCheckDomain: "www.example.com",
	RetryStrategy: &checkly.RetryStrategy{
		Type:               "LINEAR",
		MaxRetries:         2,
		BaseBackoffSeconds: 60,
		MaxDurationSeconds: 600,
		SameRegion:         true,
	},
}

var ignoreBrowserCheckFields = cmpopts.IgnoreFields(
	checkly.BrowserCheck{},
	"ID",
	"AlertChannelSubscriptions",
	"CreatedAt",
	"UpdatedAt",
)

// The create and update endpoints don't include scriptPath or dependencies in
// their responses; only GET /v1/checks/{id} returns them.
var ignoreBrowserCheckWriteResponseFields = cmpopts.IgnoreFields(
	checkly.BrowserCheck{},
	"ScriptPath",
	"Dependencies",
)

func TestCreateBrowserCheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/browser?autoAssignAlerts=false",
		validateBrowserCheck,
		http.StatusCreated,
		"CreateBrowserCheck.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.CreateBrowserCheck(context.Background(), testBrowserCheck)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testBrowserCheck, *response, ignoreBrowserCheckFields, ignoreBrowserCheckWriteResponseFields) {
		t.Error(cmp.Diff(testBrowserCheck, *response, ignoreBrowserCheckFields, ignoreBrowserCheckWriteResponseFields))
	}
}

func TestDeleteBrowserCheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/checks/%s", testBrowserCheck.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	err := client.DeleteBrowserCheck(context.Background(), testBrowserCheck.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateBrowserCheck(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/browser/%s?autoAssignAlerts=false", testBrowserCheck.ID),
		validateBrowserCheck,
		http.StatusOK,
		"UpdateBrowserCheck.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	_, err := client.UpdateBrowserCheck(context.Background(), testBrowserCheck.ID, testBrowserCheck)
	if err != nil {
		t.Error(err)
	}
}

func TestGetBrowserCheck(t *testing.T) {
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/checks/%s?includeDependencies=true", testBrowserCheck.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetBrowserCheck.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.GetBrowserCheck(context.Background(), testBrowserCheck.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testBrowserCheck, *response, ignoreBrowserCheckFields) {
		t.Error(cmp.Diff(testBrowserCheck, *response, ignoreBrowserCheckFields))
	}
}

func validateBrowserCheckEmptyPayload(t *testing.T, body []byte) {
	assertRawBodyField(t, body, "tags", `[]`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyField(t, body, "locations", `[]`)
	assertRawBodyField(t, body, "privateLocations", `[]`)
	assertRawBodyField(t, body, "groupId", `null`)
	// The group order must always be serialized, or the backend would keep
	// the previous value on update instead of resetting it to 0.
	assertRawBodyField(t, body, "groupOrder", `0`)
	assertRawBodyField(t, body, "dependencies", `[]`)
	assertRawBodyField(t, body, "scriptPath", `null`)
}

func TestCreateBrowserCheckSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/browser?autoAssignAlerts=false",
		validateBrowserCheckEmptyPayload,
		http.StatusCreated,
		"CreateBrowserCheck.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	_, err := client.CreateBrowserCheck(context.Background(), checkly.BrowserCheck{
		Name:   "Browser Check #1",
		Script: "console.log('test')",
	})
	if err != nil {
		t.Error(err)
	}
}
