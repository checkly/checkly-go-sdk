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

// --- traceroute monitor -----------------------------------------------------

func validateTracerouteMonitor(t *testing.T, body []byte) {
	var payload checkly.TracerouteMonitor
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testTracerouteMonitor, payload, ignoreTracerouteMonitorFields) {
		t.Error(cmp.Diff(testTracerouteMonitor, payload, ignoreTracerouteMonitorFields))
	}
}

var testTracerouteMonitor = checkly.TracerouteMonitor{
	ID:               "b2c3d4e5-1111-2222-3333-444455556666",
	Name:             "Traceroute Monitor #1",
	Frequency:        10,
	FrequencyOffset:  80,
	Activated:        false,
	Muted:            false,
	Locations:        []string{"us-east-1"},
	PrivateLocations: &[]string{},
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
	Request: checkly.TracerouteRequest{
		URL:      "traceroute.example.com",
		Protocol: "ICMP",
		// No Port: the ICMP protocol strips it from the wire (see MarshalJSON).
		IPFamily:       "IPv4",
		MaxHops:        30,
		MaxUnknownHops: 5,
		Timeout:        30,
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

var ignoreTracerouteMonitorFields = cmpopts.IgnoreFields(
	checkly.TracerouteMonitor{},
	"ID",
	"AlertChannelSubscriptions",
	"CreatedAt",
	"UpdatedAt",
)

func TestCreateTracerouteMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks/traceroute?autoAssignAlerts=false",
		validateTracerouteMonitor,
		http.StatusCreated,
		"CreateTracerouteMonitor.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.CreateTracerouteMonitor(context.Background(), testTracerouteMonitor)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testTracerouteMonitor, *response, ignoreTracerouteMonitorFields) {
		t.Error(cmp.Diff(testTracerouteMonitor, *response, ignoreTracerouteMonitorFields))
	}
}

func TestDeleteTracerouteMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/checks/%s", testTracerouteMonitor.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	err := client.DeleteTracerouteMonitor(context.Background(), testTracerouteMonitor.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateTracerouteMonitor(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/traceroute/%s?autoAssignAlerts=false", testTracerouteMonitor.ID),
		validateTracerouteMonitor,
		http.StatusOK,
		"UpdateTracerouteMonitor.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	_, err := client.UpdateTracerouteMonitor(context.Background(), testTracerouteMonitor.ID, testTracerouteMonitor)
	if err != nil {
		t.Error(err)
	}
}

func TestGetTracerouteMonitor(t *testing.T) {
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/checks/%s", testTracerouteMonitor.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetTracerouteMonitor.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.GetTracerouteMonitor(context.Background(), testTracerouteMonitor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testTracerouteMonitor, *response, ignoreTracerouteMonitorFields) {
		t.Error(cmp.Diff(testTracerouteMonitor, *response, ignoreTracerouteMonitorFields))
	}
}

// --- forbidden-field (anti-pattern) tests -----------------------------------

// TestTracerouteRequestStripsPortForICMP asserts the wire body omits `port`
// for an ICMP probe (mirroring the public API's Joi `.strip()`), while keeping
// it for non-ICMP protocols.
func TestTracerouteRequestStripsPortForICMP(t *testing.T) {
	t.Parallel()
	icmp, err := json.Marshal(checkly.TracerouteRequest{
		URL:      "traceroute.example.com",
		Protocol: "ICMP",
		Port:     33434,
	})
	if err != nil {
		t.Fatal(err)
	}
	var icmpBody map[string]interface{}
	if err := json.Unmarshal(icmp, &icmpBody); err != nil {
		t.Fatal(err)
	}
	if _, ok := icmpBody["port"]; ok {
		t.Errorf("ICMP traceroute body must omit `port`, got: %s", icmp)
	}

	tcp, err := json.Marshal(checkly.TracerouteRequest{
		URL:      "traceroute.example.com",
		Protocol: "TCP",
		Port:     33434,
	})
	if err != nil {
		t.Fatal(err)
	}
	var tcpBody map[string]interface{}
	if err := json.Unmarshal(tcp, &tcpBody); err != nil {
		t.Fatal(err)
	}
	if _, ok := tcpBody["port"]; !ok {
		t.Errorf("non-ICMP traceroute body must keep `port`, got: %s", tcp)
	}
}

// tracerouteRequestDefaultKeys are the request keys that create leaves out
// when unset, so the API stores its defaults, and that update sends as null.
var tracerouteRequestDefaultKeys = []string{"protocol", "port", "maxHops", "maxUnknownHops", "timeout"}

// validateTracerouteMonitorEmptyPayload asserts that the keys the update
// endpoint would otherwise keep at their stored value are sent for an empty
// monitor, on create and update alike.
func validateTracerouteMonitorEmptyPayload(t *testing.T, body []byte) {
	assertRawBodyField(t, body, "tags", `[]`)
	assertRawBodyField(t, body, "groupOrder", `null`)
	assertRawBodyField(t, body, "alertChannelSubscriptions", `[]`)
	assertRawBodyField(t, body, "locations", `[]`)
	assertRawBodyField(t, body, "privateLocations", `[]`)
}

func TestTracerouteMonitorSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	monitor := checkly.TracerouteMonitor{
		Name:    "Traceroute Monitor #1",
		Request: checkly.TracerouteRequest{URL: "traceroute.example.com"},
	}
	runPayloadCases(t, []payloadCase{
		{
			name:    "CreateTracerouteMonitor",
			method:  http.MethodPost,
			url:     "/v1/checks/traceroute?autoAssignAlerts=false",
			status:  http.StatusCreated,
			fixture: "CreateTracerouteMonitor.json",
			validate: func(t *testing.T, body []byte) {
				validateTracerouteMonitorEmptyPayload(t, body)
				for _, key := range tracerouteRequestDefaultKeys {
					assertRawBodyPathAbsent(t, body, "request", key)
				}
			},
			call: func(ctx context.Context, c checkly.Client) error {
				_, err := c.CreateTracerouteMonitor(ctx, monitor)
				return err
			},
		},
		{
			name:    "UpdateTracerouteMonitor",
			method:  http.MethodPut,
			url:     fmt.Sprintf("/v1/checks/traceroute/%s?autoAssignAlerts=false", testTracerouteMonitor.ID),
			status:  http.StatusOK,
			fixture: "UpdateTracerouteMonitor.json",
			validate: func(t *testing.T, body []byte) {
				validateTracerouteMonitorEmptyPayload(t, body)
				for _, key := range tracerouteRequestDefaultKeys {
					assertRawBodyPath(t, body, `null`, "request", key)
				}
				assertRawBodyPath(t, body, `[]`, "request", "assertions")
				// A null IP family or PTR lookup would not mean the default.
				assertRawBodyPathAbsent(t, body, "request", "ipFamily")
				assertRawBodyPathAbsent(t, body, "request", "ptrLookup")
			},
			call: func(ctx context.Context, c checkly.Client) error {
				_, err := c.UpdateTracerouteMonitor(ctx, testTracerouteMonitor.ID, monitor)
				return err
			},
		},
	})
}

func TestUpdateTracerouteMonitorSerializesSetValues(t *testing.T) {
	t.Parallel()
	ptrLookup := false
	request := checkly.TracerouteRequest{
		URL:            "traceroute.example.com",
		Protocol:       "UDP",
		Port:           33434,
		IPFamily:       "IPv6",
		MaxHops:        20,
		MaxUnknownHops: 10,
		PtrLookup:      &ptrLookup,
		Timeout:        5,
	}
	icmp := request
	icmp.Protocol = "ICMP"
	update := func(request checkly.TracerouteRequest) func(context.Context, checkly.Client) error {
		return func(ctx context.Context, c checkly.Client) error {
			_, err := c.UpdateTracerouteMonitor(ctx, testTracerouteMonitor.ID, checkly.TracerouteMonitor{
				Name:    "Traceroute Monitor #1",
				Request: request,
			})
			return err
		}
	}
	url := fmt.Sprintf("/v1/checks/traceroute/%s?autoAssignAlerts=false", testTracerouteMonitor.ID)
	runPayloadCases(t, []payloadCase{
		{
			name:    "UDP",
			method:  http.MethodPut,
			url:     url,
			status:  http.StatusOK,
			fixture: "UpdateTracerouteMonitor.json",
			validate: func(t *testing.T, body []byte) {
				assertRawBodyPath(t, body, `"traceroute.example.com"`, "request", "url")
				assertRawBodyPath(t, body, `"UDP"`, "request", "protocol")
				assertRawBodyPath(t, body, `33434`, "request", "port")
				assertRawBodyPath(t, body, `"IPv6"`, "request", "ipFamily")
				assertRawBodyPath(t, body, `20`, "request", "maxHops")
				assertRawBodyPath(t, body, `10`, "request", "maxUnknownHops")
				assertRawBodyPath(t, body, `false`, "request", "ptrLookup")
				assertRawBodyPath(t, body, `5`, "request", "timeout")
			},
			call: update(request),
		},
		{
			// The API discards the port of an ICMP probe.
			name:    "ICMP",
			method:  http.MethodPut,
			url:     url,
			status:  http.StatusOK,
			fixture: "UpdateTracerouteMonitor.json",
			validate: func(t *testing.T, body []byte) {
				assertRawBodyPath(t, body, `"ICMP"`, "request", "protocol")
				assertRawBodyPath(t, body, `null`, "request", "port")
			},
			call: update(icmp),
		},
	})
}

func TestTracerouteMonitorSendsPrivateLocations(t *testing.T) {
	t.Parallel()
	monitor := testTracerouteMonitor
	monitor.PrivateLocations = &[]string{"my-private-location"}
	validate := func(t *testing.T, body []byte) {
		assertRawBodyField(t, body, "privateLocations", `["my-private-location"]`)
	}
	runPayloadCases(t, []payloadCase{
		{
			name:     "CreateTracerouteMonitor",
			method:   http.MethodPost,
			url:      "/v1/checks/traceroute?autoAssignAlerts=false",
			status:   http.StatusCreated,
			fixture:  "CreateTracerouteMonitor.json",
			validate: validate,
			call: func(ctx context.Context, c checkly.Client) error {
				_, err := c.CreateTracerouteMonitor(ctx, monitor)
				return err
			},
		},
		{
			name:     "UpdateTracerouteMonitor",
			method:   http.MethodPut,
			url:      fmt.Sprintf("/v1/checks/traceroute/%s?autoAssignAlerts=false", monitor.ID),
			status:   http.StatusOK,
			fixture:  "UpdateTracerouteMonitor.json",
			validate: validate,
			call: func(ctx context.Context, c checkly.Client) error {
				_, err := c.UpdateTracerouteMonitor(ctx, monitor.ID, monitor)
				return err
			},
		},
	})
}

// --- no maximum response time ------------------------------------------------

// TestTracerouteMonitorSendsMaxResponseTime asserts that NoMaxResponseTime
// sends a maxResponseTime of 0 on create and update, while a zero
// MaxResponseTime without it stays left out of the request.
func TestTracerouteMonitorSendsMaxResponseTime(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		monitor checkly.TracerouteMonitor
		want    string // "" means left out
	}{
		{"no limit", checkly.TracerouteMonitor{NoMaxResponseTime: true}, `0`},
		{"no limit with degraded time", checkly.TracerouteMonitor{NoMaxResponseTime: true, DegradedResponseTime: 25000}, `0`},
		{"limit", checkly.TracerouteMonitor{MaxResponseTime: 5000}, `5000`},
		{"unset", checkly.TracerouteMonitor{}, ``},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			monitor := tc.monitor
			monitor.Name = "Traceroute Monitor #1"
			monitor.Request = checkly.TracerouteRequest{URL: "traceroute.example.com"}
			for _, call := range []struct {
				name   string
				status int
				call   func(checkly.Client) error
			}{
				{"create", http.StatusCreated, func(c checkly.Client) error {
					_, err := c.CreateTracerouteMonitor(context.Background(), monitor)
					return err
				}},
				{"update", http.StatusOK, func(c checkly.Client) error {
					_, err := c.UpdateTracerouteMonitor(context.Background(), testTracerouteMonitor.ID, monitor)
					return err
				}},
			} {
				c, lastBody := newRecordingServer(t, call.status, `{"checkType":"TRACEROUTE"}`)
				if err := call.call(c); err != nil {
					t.Fatalf("%s: %v", call.name, err)
				}
				body := lastBody()
				if tc.want == "" {
					assertRawBodyKeyAbsent(t, body, "maxResponseTime")
				} else {
					assertRawBodyField(t, body, "maxResponseTime", tc.want)
				}
				if tc.monitor.DegradedResponseTime != 0 {
					assertRawBodyField(t, body, "degradedResponseTime", fmt.Sprint(tc.monitor.DegradedResponseTime))
				}
				if t.Failed() {
					t.Logf("%s body: %s", call.name, body)
				}
			}
		})
	}
}

// TestTracerouteMonitorRejectsNoMaxResponseTimeWithLimit asserts that
// NoMaxResponseTime together with a non-zero MaxResponseTime fails on create
// and update without sending a request.
func TestTracerouteMonitorRejectsNoMaxResponseTimeWithLimit(t *testing.T) {
	t.Parallel()
	c, lastBody := newRecordingServer(t, http.StatusOK, `{"checkType":"TRACEROUTE"}`)
	monitor := checkly.TracerouteMonitor{
		Name:              "Traceroute Monitor #1",
		NoMaxResponseTime: true,
		MaxResponseTime:   5000,
		Request:           checkly.TracerouteRequest{URL: "traceroute.example.com"},
	}
	_, err := c.CreateTracerouteMonitor(context.Background(), monitor)
	if want := "POST /v1/checks/traceroute: NoMaxResponseTime is set together with MaxResponseTime 5000; set only one"; err == nil || err.Error() != want {
		t.Errorf("create error = %v, want %q", err, want)
	}
	_, err = c.UpdateTracerouteMonitor(context.Background(), testTracerouteMonitor.ID, monitor)
	if want := "PUT /v1/checks/traceroute/" + testTracerouteMonitor.ID + ": NoMaxResponseTime is set together with MaxResponseTime 5000; set only one"; err == nil || err.Error() != want {
		t.Errorf("update error = %v, want %q", err, want)
	}
	if body := lastBody(); body != nil {
		t.Errorf("a request was sent: %s", body)
	}
}

// TestTracerouteMonitorDecodesNoMaxResponseTime asserts that a maxResponseTime
// of 0 read from the API sets NoMaxResponseTime, and that an update of the
// monitor read back sends 0 again.
func TestTracerouteMonitorDecodesNoMaxResponseTime(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		field  string
		want   checkly.TracerouteMonitor
		update string // the maxResponseTime an update of the read monitor sends; "" means left out
	}{
		{"zero", `,"maxResponseTime":0`, checkly.TracerouteMonitor{NoMaxResponseTime: true}, `0`},
		{"limit", `,"maxResponseTime":5000`, checkly.TracerouteMonitor{MaxResponseTime: 5000}, `5000`},
		{"null", `,"maxResponseTime":null`, checkly.TracerouteMonitor{}, ``},
		{"missing", ``, checkly.TracerouteMonitor{}, ``},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, lastBody := newRecordingServer(t, http.StatusOK, `{"checkType":"TRACEROUTE","request":{"url":"traceroute.example.com"}`+tc.field+`}`)
			got, err := c.GetTracerouteMonitor(context.Background(), testTracerouteMonitor.ID)
			if err != nil {
				t.Fatal(err)
			}
			want := tc.want
			want.Request = checkly.TracerouteRequest{URL: "traceroute.example.com"}
			if diff := cmp.Diff(want, *got); diff != "" {
				t.Errorf("decoded monitor (-want +got):\n%s", diff)
			}
			if _, err := c.UpdateTracerouteMonitor(context.Background(), testTracerouteMonitor.ID, *got); err != nil {
				t.Fatal(err)
			}
			if tc.update == "" {
				assertRawBodyKeyAbsent(t, lastBody(), "maxResponseTime")
			} else {
				assertRawBodyField(t, lastBody(), "maxResponseTime", tc.update)
			}
		})
	}
}

// TestTracerouteMonitorUnmarshalKeepsUnsetFields asserts that decoding into a
// monitor that already has values keeps those the JSON does not set, as the
// default decoder does.
func TestTracerouteMonitorUnmarshalKeepsUnsetFields(t *testing.T) {
	t.Parallel()
	monitor := checkly.TracerouteMonitor{Name: "kept", NoMaxResponseTime: true}
	if err := json.Unmarshal([]byte(`{"frequency":5}`), &monitor); err != nil {
		t.Fatal(err)
	}
	want := checkly.TracerouteMonitor{Name: "kept", Frequency: 5, NoMaxResponseTime: true}
	if diff := cmp.Diff(want, monitor); diff != "" {
		t.Errorf("decoded monitor (-want +got):\n%s", diff)
	}
}
