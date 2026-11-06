package checkly_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	checkly "github.com/checkly/checkly-go-sdk/v2"
)

func validateTCPMonitorEmptyPayload(t *testing.T, body []byte) {
	assertRawBodyField(t, body, "tags", `[]`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyField(t, body, "locations", `[]`)
	assertRawBodyField(t, body, "privateLocations", `[]`)
	assertRawBodyField(t, body, "groupId", `null`)
	// 0 is a schema-valid response time limit and must be serialized, or the
	// backend would default (create) or keep (update) the previous values.
	assertRawBodyField(t, body, "degradedResponseTime", `0`)
	assertRawBodyField(t, body, "maxResponseTime", `0`)
	// 0 is the value the group order resets to and must be serialized, or
	// the backend would keep the previous value on update.
	assertRawBodyField(t, body, "groupOrder", `0`)

	// The endpoints are PATCH-style down to the nested request object, so
	// clearable request keys must always be serialized as well.
	assertRawNestedBodyField(t, body, "request", "assertions", `[]`)
	// data must NOT be present: the backend rejects an empty string and a
	// null value cannot clear it either, so an empty data value is omitted.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	assertRawBodyKeyAbsent(t, raw["request"], "data")
}

func TestCreateTCPMonitorSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/tcp?autoAssignAlerts=false",
		validateTCPMonitorEmptyPayload,
		http.StatusCreated,
		"CreateTCPMonitor.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	_, err := client.CreateTCPMonitor(context.Background(), checkly.TCPMonitor{
		Name: "TCP Monitor #1",
		Request: checkly.TCPRequest{
			Hostname: "example.com",
			Port:     443,
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateTCPMonitorSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		"/v1/checks/tcp/26900d4d-a852-4e34-9bd0-e40a7a5a2495?autoAssignAlerts=false",
		validateTCPMonitorEmptyPayload,
		http.StatusOK,
		"CreateTCPMonitor.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	_, err := client.UpdateTCPMonitor(
		context.Background(),
		"26900d4d-a852-4e34-9bd0-e40a7a5a2495",
		checkly.TCPMonitor{
			Name: "TCP Monitor #1",
			Request: checkly.TCPRequest{
				Hostname: "example.com",
				Port:     443,
			},
		},
	)
	if err != nil {
		t.Error(err)
	}
}
