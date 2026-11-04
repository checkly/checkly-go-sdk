package checkly

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Check type constants
type CheckType string

// TypeBrowser is used to identify a browser check.
const TypeBrowser = "BROWSER"

// TypeAPI is used to identify an API check.
const TypeAPI = "API"

// TypeHeartbeat is used to identify a browser check.
const TypeHeartbeat = "HEARTBEAT"

// Check represents the parameters for an existing check.
//
// Deprecated: use APICheck, BrowserCheck, MultiStepCheck or a monitor type
// such as HeartbeatMonitor, with the client methods of their type.
type Check struct {
	ID                        string                     `json:"id"`
	Name                      string                     `json:"name"`
	Description               *string                    `json:"description"`
	Type                      string                     `json:"checkType"`
	Frequency                 int                        `json:"frequency"`
	FrequencyOffset           int                        `json:"frequencyOffset,omitempty"`
	Activated                 bool                       `json:"activated"`
	Muted                     bool                       `json:"muted"`
	ShouldFail                bool                       `json:"shouldFail"`
	RunParallel               bool                       `json:"runParallel"`
	Locations                 []string                   `json:"locations"`
	DegradedResponseTime      int                        `json:"degradedResponseTime"`
	MaxResponseTime           int                        `json:"maxResponseTime"`
	Script                    string                     `json:"script,omitempty"`
	EnvironmentVariables      []EnvironmentVariable      `json:"environmentVariables"`
	Tags                      []string                   `json:"tags,omitempty"`
	SSLCheckDomain            string                     `json:"sslCheckDomain"`
	SetupSnippetID            int64                      `json:"setupSnippetId,omitempty"`
	TearDownSnippetID         int64                      `json:"tearDownSnippetId,omitempty"`
	LocalSetupScript          string                     `json:"localSetupScript,omitempty"`
	LocalTearDownScript       string                     `json:"localTearDownScript,omitempty"`
	AlertSettings             AlertSettings              `json:"alertSettings,omitempty"`
	UseGlobalAlertSettings    bool                       `json:"useGlobalAlertSettings"`
	Request                   Request                    `json:"request"`
	Heartbeat                 Heartbeat                  `json:"heartbeat"`
	GroupID                   int64                      `json:"groupId,omitempty"`
	GroupOrder                int                        `json:"groupOrder,omitempty"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	CreatedAt                 time.Time                  `json:"createdAt"`
	UpdatedAt                 time.Time                  `json:"updatedAt"`

	// Pointers
	PrivateLocations *[]string        `json:"privateLocations"`
	RuntimeID        *string          `json:"runtimeId"`
	RetryStrategy    *RetryStrategy   `json:"retryStrategy"`
	TriggerIncident  *IncidentTrigger `json:"triggerIncident"`

	// Deprecated: this property will be removed in future versions. Please use RetryStrategy instead.
	DoubleCheck bool `json:"doubleCheck"`
}

// CreateCheck creates a new check with the specified details. It returns the
// newly-created check, or an error.
func (c *client) CreateCheck(
	ctx context.Context,
	check Check,
) (*Check, error) {
	var endpoint string
	switch check.Type {
	case "BROWSER":
		endpoint = "/v1/checks/browser"
	case "API":
		endpoint = "/v1/checks/api"
	case "HEARTBEAT":
		endpoint = "/v1/checks/heartbeat"
	case "MULTI_STEP":
		endpoint = "/v1/checks/multistep"
	case "TCP":
		return nil, fmt.Errorf("user error: use CreateTCPMonitor to create TCP monitors")
	case "URL":
		return nil, fmt.Errorf("user error: use CreateURLMonitor to create URL monitors")
	case "DNS":
		return nil, fmt.Errorf("user error: use CreateDNSMonitor to create DNS monitors")
	case "ICMP":
		return nil, fmt.Errorf("user error: use CreateICMPMonitor to create ICMP monitors")
	case "GRPC":
		return nil, fmt.Errorf("user error: use CreateGRPCMonitor to create GRPC monitors")
	case "TRACEROUTE":
		return nil, fmt.Errorf("user error: use CreateTracerouteMonitor to create TRACEROUTE monitors")
	case "SSL":
		return nil, fmt.Errorf("user error: use CreateSSLMonitor to create SSL monitors")
	default:
		return nil, fmt.Errorf("unknown check type: %s", check.Type)
	}
	var result Check
	err := c.callJSON(ctx, http.MethodPost, withAutoAssignAlertsFlag(endpoint), createCheckPayload(check), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type checkPayload struct {
	Check
	GroupID *int64 `json:"groupId"`
}

func createCheckPayload(check Check) checkPayload {
	payload := checkPayload{
		Check: check,
	}

	// GroupID must be null if empty or the group will not get unset on update.
	if check.GroupID != 0 {
		payload.GroupID = &check.GroupID
	}

	// A nil value for a list will cause the backend to not update the value.
	// We must send empty lists instead.
	if check.Locations == nil {
		payload.Locations = []string{}
	}

	if check.PrivateLocations == nil {
		payload.PrivateLocations = &[]string{}
	}

	return payload
}

// GetCheck takes the ID of an existing check, and returns the check
// parameters, or an error.
func (c *client) GetCheck(
	ctx context.Context,
	ID string,
) (*Check, error) {
	result := Check{}
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/checks/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateCheck updates an existing check with the specified details. It
// returns the updated check, or an error.
func (c *client) UpdateCheck(
	ctx context.Context,
	ID string,
	check Check,
) (*Check, error) {
	var result Check
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/%s", ID)), createCheckPayload(check), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteCheck deletes the check with the specified ID.
func (c *client) DeleteCheck(
	ctx context.Context,
	ID string,
) error {
	_, err := c.call(ctx, http.MethodDelete, fmt.Sprintf("/v1/checks/%s", ID), nil, http.StatusNoContent)
	return err
}

// getCheckOfType fetches the check at path and decodes it into result. The
// generic check endpoint returns checks of every type, so the response's
// checkType is compared with the expected one: decoding a check of another
// type would silently drop its type-specific fields.
func (c *client) getCheckOfType(
	ctx context.Context,
	path string,
	checkType string,
	result any,
) error {
	res, err := c.call(ctx, http.MethodGet, path, nil, http.StatusOK)
	if err != nil {
		return err
	}
	var header struct {
		CheckType string `json:"checkType"`
	}
	if err := decode(res, &header); err != nil {
		return err
	}
	if header.CheckType != checkType {
		return fmt.Errorf("%s %s: check is of type %q, not %q", res.method, res.path, header.CheckType, checkType)
	}
	return decode(res, result)
}

func withAutoAssignAlertsFlag(url string) string {
	flag := "autoAssignAlerts=false"
	if strings.Contains(url, "?") {
		return url + "&" + flag
	}
	return url + "?" + flag
}
