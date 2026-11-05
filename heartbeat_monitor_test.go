package checkly_test

import (
	"context"
	"net/http"
	"testing"

	checkly "github.com/checkly/checkly-go-sdk"
)

func validateHeartbeatMonitorEmptyPayload(t *testing.T, body []byte) {
	assertRawBodyField(t, body, "tags", `[]`)
}

func TestCreateHeartbeatMonitorSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/heartbeat?autoAssignAlerts=false",
		validateHeartbeatMonitorEmptyPayload,
		http.StatusCreated,
		"CreateHeartbeatMonitor.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	_, err := client.CreateHeartbeatMonitor(context.Background(), checkly.HeartbeatMonitor{
		Name: "Heartbeat Monitor #1",
		Heartbeat: checkly.Heartbeat{
			Period:     5,
			PeriodUnit: "days",
			Grace:      0,
			GraceUnit:  "seconds",
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateHeartbeatMonitorSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		"/v1/checks/heartbeat/850cd536-a63b-4d33-9a12-8b125718f4e2?autoAssignAlerts=false",
		validateHeartbeatMonitorEmptyPayload,
		http.StatusOK,
		"CreateHeartbeatMonitor.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	_, err := client.UpdateHeartbeatMonitor(
		context.Background(),
		"850cd536-a63b-4d33-9a12-8b125718f4e2",
		checkly.HeartbeatMonitor{
			Name: "Heartbeat Monitor #1",
			Heartbeat: checkly.Heartbeat{
				Period:     5,
				PeriodUnit: "days",
				Grace:      0,
				GraceUnit:  "seconds",
			},
		},
	)
	if err != nil {
		t.Error(err)
	}
}
