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

var wantGroupID int64 = 135

var wantGroup = checkly.GroupV2{
	Name:             "test",
	Activated:        true,
	Muted:            false,
	Tags:             []string{"auto"},
	Locations:        []string{"eu-west-1"},
	PrivateLocations: &[]string{},
	Concurrency:      3,
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
				Source:     checkly.AssertionSourceStatusCode,
				Comparison: checkly.AssertionComparisonEquals,
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
	RetryStrategy: &checkly.RetryStrategy{
		Type:               "FIXED",
		MaxRetries:         1,
		MaxDurationSeconds: 600,
	},
	UseGlobalAlertSettings: new(false),
	AlertSettings: &checkly.AlertSettings{
		EscalationType: checkly.EscalationTypeRunBased,
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
		ParallelRunFailureThreshold: checkly.ParallelRunFailureThreshold{
			Enabled:    false,
			Percentage: 10,
		},
	},
	AlertChannelSubscriptions: []checkly.AlertChannelSubscription{
		{
			Activated: true,
			ChannelID: 2996,
		},
	},
	LocalSetupScript:    new("setup-test"),
	LocalTearDownScript: new("teardown-test"),
}

func validateGroup(t *testing.T, body []byte) {
	var gotGroup checkly.GroupV2
	err := json.Unmarshal(body, &gotGroup)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(wantGroup, gotGroup) {
		t.Error(cmp.Diff(wantGroup, gotGroup))
	}
}

var ignoreGroupFields = cmpopts.IgnoreFields(checkly.GroupV2{}, "ID", "PrivateLocations", "CreatedAt", "UpdatedAt")

func TestCreateGroupV2(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v2/check-groups?autoAssignAlerts=false",
		validateGroup,
		http.StatusCreated,
		"CreateGroup.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	gotGroup, err := client.CreateGroupV2(context.Background(), wantGroup)
	if err != nil {
		t.Error(err)
	}
	ignored := cmpopts.IgnoreFields(checkly.GroupV2{}, "ID", "AlertChannelSubscriptions", "PrivateLocations", "CreatedAt", "UpdatedAt")
	if !cmp.Equal(wantGroup, *gotGroup, ignored) {
		t.Error(cmp.Diff(wantGroup, *gotGroup, ignoreGroupFields))
	}
}

func TestGetGroupV2(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/check-groups/%d", wantGroupID),
		validateEmptyBody,
		http.StatusOK,
		"CreateGroup.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	gotGroup, err := client.GetGroupV2(context.Background(), wantGroupID)
	if err != nil {
		t.Error(err)
	}
	ignored := cmpopts.IgnoreFields(checkly.GroupV2{}, "ID", "AlertChannelSubscriptions", "PrivateLocations", "CreatedAt", "UpdatedAt")
	if !cmp.Equal(wantGroup, *gotGroup, ignored) {
		t.Error(cmp.Diff(wantGroup, *gotGroup, ignored))
	}
}

func TestUpdateGroupV2(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v2/check-groups/%d?autoAssignAlerts=false", wantGroupID),
		validateGroup,
		http.StatusOK,
		"UpdateGroup.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	gotGroup, err := client.UpdateGroupV2(context.Background(), wantGroupID, wantGroup)
	if err != nil {
		t.Error(err)
	}
	ignored := cmpopts.IgnoreFields(checkly.GroupV2{}, "ID", "AlertChannelSubscriptions", "PrivateLocations", "CreatedAt", "UpdatedAt")
	if !cmp.Equal(wantGroup, *gotGroup, ignored) {
		t.Error(cmp.Diff(wantGroup, *gotGroup, ignoreGroupFields))
	}
}

func validateGroupEmptyPayload(t *testing.T, body []byte) {
	assertRawBodyField(t, body, "tags", `[]`)
	assertRawBodyField(t, body, "locations", `[]`)
	assertRawBodyField(t, body, "privateLocations", `[]`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyField(t, body, "setupSnippetId", `null`)
	assertRawBodyField(t, body, "tearDownSnippetId", `null`)
	assertRawBodyField(t, body, "localSetupScript", `null`)
	assertRawBodyField(t, body, "localTearDownScript", `null`)
	assertRawBodyField(t, body, "runParallel", `null`)
	assertRawBodyField(t, body, "useGlobalAlertSettings", `null`)
	assertRawBodyField(t, body, "alertSettings", `null`)
	assertRawBodyField(t, body, "retryStrategy", `null`)
}

func TestUpdateGroupV2SerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		"/v2/check-groups/17?autoAssignAlerts=false",
		validateGroupEmptyPayload,
		http.StatusOK,
		"UpdateGroup.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	_, err := client.UpdateGroupV2(context.Background(), 17, checkly.GroupV2{
		Name: "Group #1",
	})
	if err != nil {
		t.Error(err)
	}
}

func TestCreateGroupV2SerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v2/check-groups?autoAssignAlerts=false",
		validateGroupEmptyPayload,
		http.StatusCreated,
		"CreateGroup.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	_, err := client.CreateGroupV2(context.Background(), checkly.GroupV2{
		Name: "Group #1",
	})
	if err != nil {
		t.Error(err)
	}
}

// Zero values inside a set AlertSettings are sent, not left out.
func TestCreateGroupV2SerializesZeroAlertSettingsFields(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v2/check-groups?autoAssignAlerts=false",
		func(t *testing.T, body []byte) {
			assertRawBodyPath(t, body, `0`, "alertSettings", "reminders", "amount")
			assertRawBodyPath(t, body, `false`, "alertSettings", "parallelRunFailureThreshold", "enabled")
		},
		http.StatusCreated,
		"CreateGroup.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	_, err := client.CreateGroupV2(context.Background(), checkly.GroupV2{
		Name: "Group #1",
		AlertSettings: &checkly.AlertSettings{
			EscalationType:     checkly.EscalationTypeRunBased,
			RunBasedEscalation: checkly.RunBasedEscalation{FailedRunThreshold: 1},
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateGroupV2SerializesSetValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		"/v2/check-groups/17?autoAssignAlerts=false",
		func(t *testing.T, body []byte) {
			assertRawBodyField(t, body, "setupSnippetId", `5`)
			assertRawBodyField(t, body, "tearDownSnippetId", `6`)
			assertRawBodyField(t, body, "localSetupScript", `"setup"`)
			assertRawBodyField(t, body, "localTearDownScript", `"teardown"`)
			assertRawBodyField(t, body, "alertChannelSubscriptions", `[{"alertChannelId":42,"activated":true}]`)
			assertRawBodyField(t, body, "tags", `["prod"]`)
		},
		http.StatusOK,
		"UpdateGroup.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	_, err := client.UpdateGroupV2(context.Background(), 17, checkly.GroupV2{
		Name:                "Group #1",
		SetupSnippetID:      new(int64(5)),
		TearDownSnippetID:   new(int64(6)),
		LocalSetupScript:    new("setup"),
		LocalTearDownScript: new("teardown"),
		Tags:                []string{"prod"},
		AlertChannelSubscriptions: []checkly.AlertChannelSubscription{
			{
				ChannelID: 42,
				Activated: true,
			},
		},
	})
	if err != nil {
		t.Error(err)
	}
}

// GetGroupV2 reads through the v1 route, whose response carries the v2 group
// shape plus projectBindings. The fixture is such a response, with the FALLBACK
// retry strategy and nulls for the nullable settings.
func TestGetGroupV2NullableSettings(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/check-groups/%d", wantGroupID),
		validateEmptyBody,
		http.StatusOK,
		"GetGroupV2.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	gotGroup, err := client.GetGroupV2(context.Background(), wantGroupID)
	if err != nil {
		t.Fatal(err)
	}
	runParallel := false
	want := checkly.GroupV2{
		ID:          156,
		Name:        "test",
		Activated:   true,
		RunParallel: &runParallel,
		Tags:        []string{"auto"},
		Locations:   []string{"eu-west-1"},
		Concurrency: 3,
		APICheckDefaults: checkly.APICheckDefaults{
			BaseURL:         "example.com/api/test",
			Headers:         []checkly.KeyValue{},
			QueryParameters: []checkly.KeyValue{},
			Assertions:      []checkly.Assertion{},
		},
		EnvironmentVariables: []checkly.EnvironmentVariable{},
		AlertChannelSubscriptions: []checkly.AlertChannelSubscription{
			{Activated: true, ChannelID: 2996},
		},
		PrivateLocations: &[]string{},
		RetryStrategy:    &checkly.RetryStrategy{Type: "FALLBACK"},
	}
	// This test is about the v1 read shape; timestamp decoding is tested
	// separately.
	ignored := cmpopts.IgnoreFields(checkly.GroupV2{}, "CreatedAt", "UpdatedAt")
	if !cmp.Equal(want, *gotGroup, ignored) {
		t.Error(cmp.Diff(want, *gotGroup, ignored))
	}
}

// DeleteGroupV2 deletes through the v1 route, as the API has no v2 one.
func TestDeleteGroupV2(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/check-groups/%d", wantGroupID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	if err := client.DeleteGroupV2(context.Background(), wantGroupID); err != nil {
		t.Error(err)
	}
}
