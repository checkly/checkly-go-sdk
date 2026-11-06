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

// --- gRPC monitor -----------------------------------------------------------

func validateGRPCMonitor(t *testing.T, body []byte) {
	var payload checkly.GRPCMonitor
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	// The payload sends empty lists rather than omitting them, which decode
	// to empty rather than nil slices.
	if !cmp.Equal(testGRPCMonitor, payload, ignoreGRPCMonitorFields, cmpopts.EquateEmpty()) {
		t.Error(cmp.Diff(testGRPCMonitor, payload, ignoreGRPCMonitorFields, cmpopts.EquateEmpty()))
	}
}

var testGRPCMonitor = checkly.GRPCMonitor{
	ID:              "a1b2c3d4-1111-2222-3333-444455556666",
	Name:            "gRPC Monitor #1",
	Frequency:       10,
	FrequencyOffset: 80,
	Activated:       false,
	Muted:           false,
	Locations:       []string{"us-east-1"},
	Tags: []string{
		"tag-1",
	},
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
	DegradedResponseTime:   3000,
	MaxResponseTime:        5000,
	RetryStrategy:          nil,
	PrivateLocations:       &[]string{},
	Request: checkly.GRPCRequest{
		URL: "grpc.example.com",
		// interface{} port decodes from JSON as a float64, so the expected
		// value must use the same concrete type for cmp to match.
		Port:     float64(443),
		IPFamily: "IPv4",
		SkipSSL:  false,
		Timeout:  30,
		GRPCConfig: checkly.GRPCConfig{
			Mode:   "BEHAVIOR",
			TLS:    true,
			Method: "/grpc.health.v1.Health/Check",
		},
		Assertions: []checkly.Assertion{
			{
				Source:     "GRPC_STATUS",
				Target:     "0",
				Property:   "",
				Comparison: "EQUALS",
			},
		},
	},
}

var ignoreGRPCMonitorFields = cmpopts.IgnoreFields(
	checkly.GRPCMonitor{},
	"ID",
	"AlertChannelSubscriptions",
	"CreatedAt",
	"UpdatedAt",
)

func TestCreateGRPCMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/grpc?autoAssignAlerts=false",
		validateGRPCMonitor,
		http.StatusCreated,
		"CreateGRPCMonitor.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.CreateGRPCMonitor(context.Background(), testGRPCMonitor)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testGRPCMonitor, *response, ignoreGRPCMonitorFields) {
		t.Error(cmp.Diff(testGRPCMonitor, *response, ignoreGRPCMonitorFields))
	}
}

func TestDeleteGRPCMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/checks/%s", testGRPCMonitor.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	err := client.DeleteGRPCMonitor(context.Background(), testGRPCMonitor.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateGRPCMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/grpc/%s?autoAssignAlerts=false", testGRPCMonitor.ID),
		validateGRPCMonitor,
		http.StatusOK,
		"UpdateGRPCMonitor.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	_, err := client.UpdateGRPCMonitor(context.Background(), testGRPCMonitor.ID, testGRPCMonitor)
	if err != nil {
		t.Error(err)
	}
}

func TestGetGRPCMonitor(t *testing.T) {
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/checks/%s", testGRPCMonitor.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetGRPCMonitor.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.GetGRPCMonitor(context.Background(), testGRPCMonitor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testGRPCMonitor, *response, ignoreGRPCMonitorFields) {
		t.Error(cmp.Diff(testGRPCMonitor, *response, ignoreGRPCMonitorFields))
	}
}

// TestGRPCConfigModeForbiddenFields asserts the BEHAVIOR-only keys are dropped
// from a HEALTH-mode config and the HEALTH-only `service` key is dropped from a
// BEHAVIOR-mode config — the mode-exclusive fields must never reach the wire.
func TestGRPCConfigModeForbiddenFields(t *testing.T) {
	t.Parallel()
	health, err := json.Marshal(checkly.GRPCConfig{
		Mode:    "HEALTH",
		TLS:     true,
		Service: "grpc.health.v1.Health",
	})
	if err != nil {
		t.Fatal(err)
	}
	var healthBody map[string]interface{}
	if err := json.Unmarshal(health, &healthBody); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"encoding", "method", "serviceDefinition", "protoContent", "bfbsContent", "message"} {
		if _, ok := healthBody[k]; ok {
			t.Errorf("HEALTH-mode gRPC config must omit %q, got: %s", k, health)
		}
	}
	if _, ok := healthBody["service"]; !ok {
		t.Errorf("HEALTH-mode gRPC config must keep `service`, got: %s", health)
	}

	behavior, err := json.Marshal(checkly.GRPCConfig{
		Mode:              "BEHAVIOR",
		TLS:               true,
		Method:            "/grpc.health.v1.Health/Check",
		ServiceDefinition: "syntax = \"proto3\";",
	})
	if err != nil {
		t.Fatal(err)
	}
	var behaviorBody map[string]interface{}
	if err := json.Unmarshal(behavior, &behaviorBody); err != nil {
		t.Fatal(err)
	}
	if _, ok := behaviorBody["service"]; ok {
		t.Errorf("BEHAVIOR-mode gRPC config must omit `service`, got: %s", behavior)
	}
	for _, k := range []string{"method", "serviceDefinition"} {
		if _, ok := behaviorBody[k]; !ok {
			t.Errorf("BEHAVIOR-mode gRPC config must keep %q, got: %s", k, behavior)
		}
	}
}

func TestGRPCConfigFlatBuffersRoundTrip(t *testing.T) {
	t.Parallel()
	want := checkly.GRPCConfig{
		Mode:        "BEHAVIOR",
		TLS:         true,
		Encoding:    "FLATBUFFERS",
		Method:      "example.Greeter/Greet",
		BfbsContent: "RkxBVF9CVUZGRVJTX1NDSEVNQQ==",
		Message:     `{"name":"Checkly"}`,
	}

	body, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got checkly.GRPCConfig
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}

	var wire map[string]interface{}
	if err := json.Unmarshal(body, &wire); err != nil {
		t.Fatal(err)
	}
	if wire["encoding"] != "FLATBUFFERS" || wire["bfbsContent"] != want.BfbsContent {
		t.Fatalf("FlatBuffers fields were not preserved on the wire: %s", body)
	}
	for _, k := range []string{"service", "serviceDefinition", "protoContent"} {
		if _, ok := wire[k]; ok {
			t.Errorf("FlatBuffers gRPC config must omit %q, got: %s", k, body)
		}
	}
}

// validateGRPCMonitorEmptyPayload asserts that the keys the update endpoint
// would otherwise keep at their stored value are sent for an empty monitor,
// on create and update alike, and that an empty mode is sent as BEHAVIOR.
func validateGRPCMonitorEmptyPayload(t *testing.T, body []byte) {
	assertRawBodyField(t, body, "tags", `[]`)
	assertRawBodyField(t, body, "groupOrder", `null`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyPath(t, body, `"BEHAVIOR"`, "request", "grpcConfig", "mode")
}

func TestGRPCMonitorSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	monitor := checkly.GRPCMonitor{
		Name:    "gRPC Monitor #1",
		Request: checkly.GRPCRequest{URL: "grpc.example.com", Port: 443},
	}
	runPayloadCases(t, []payloadCase{
		{
			name:    "CreateGRPCMonitor",
			method:  http.MethodPost,
			url:     "/v1/checks/grpc?autoAssignAlerts=false",
			status:  http.StatusCreated,
			fixture: "CreateGRPCMonitor.json",
			validate: func(t *testing.T, body []byte) {
				validateGRPCMonitorEmptyPayload(t, body)
				// Create leaves an unset timeout out, so the API stores its
				// default rather than null.
				assertRawBodyPathAbsent(t, body, "request", "timeout")
			},
			call: func(ctx context.Context, c checkly.Client) error {
				_, err := c.CreateGRPCMonitor(ctx, monitor)
				return err
			},
		},
		{
			name:    "UpdateGRPCMonitor",
			method:  http.MethodPut,
			url:     fmt.Sprintf("/v1/checks/grpc/%s?autoAssignAlerts=false", testGRPCMonitor.ID),
			status:  http.StatusOK,
			fixture: "UpdateGRPCMonitor.json",
			validate: func(t *testing.T, body []byte) {
				validateGRPCMonitorEmptyPayload(t, body)
				assertRawBodyPath(t, body, `null`, "request", "timeout")
			},
			call: func(ctx context.Context, c checkly.Client) error {
				_, err := c.UpdateGRPCMonitor(ctx, testGRPCMonitor.ID, monitor)
				return err
			},
		},
	})
}

func TestUpdateGRPCMonitorSerializesSetTimeout(t *testing.T) {
	t.Parallel()
	monitor := checkly.GRPCMonitor{
		Name:    "gRPC Monitor #1",
		Request: checkly.GRPCRequest{URL: "grpc.example.com", Port: 443, Timeout: 30},
	}
	runPayloadCases(t, []payloadCase{{
		name:    "UpdateGRPCMonitor",
		method:  http.MethodPut,
		url:     fmt.Sprintf("/v1/checks/grpc/%s?autoAssignAlerts=false", testGRPCMonitor.ID),
		status:  http.StatusOK,
		fixture: "UpdateGRPCMonitor.json",
		validate: func(t *testing.T, body []byte) {
			assertRawBodyPath(t, body, `30`, "request", "timeout")
			assertRawBodyPath(t, body, `"grpc.example.com"`, "request", "url")
		},
		call: func(ctx context.Context, c checkly.Client) error {
			_, err := c.UpdateGRPCMonitor(ctx, testGRPCMonitor.ID, monitor)
			return err
		},
	}})
}
