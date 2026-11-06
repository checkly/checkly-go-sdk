package checkly

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// CheckType is the type of a check or monitor, in [Check].Type and
// [CheckResultsFilter].CheckType.
type CheckType string

// Check types, the values of [Check].Type and [CheckResultsFilter].CheckType.
const (
	// CheckTypeAgentic is the type of agentic checks, which the SDK has no
	// resource type for. It can be used as CheckResultsFilter.CheckType and
	// appears in Check.Type of checks read with GetCheck.
	CheckTypeAgentic    CheckType = "AGENTIC"
	CheckTypeAPI        CheckType = "API"
	CheckTypeBrowser    CheckType = "BROWSER"
	CheckTypeHeartbeat  CheckType = "HEARTBEAT"
	CheckTypeICMP       CheckType = "ICMP"
	CheckTypeMultiStep  CheckType = "MULTI_STEP"
	CheckTypeTCP        CheckType = "TCP"
	CheckTypePlaywright CheckType = "PLAYWRIGHT"
	CheckTypeTraceroute CheckType = "TRACEROUTE"
	CheckTypeURL        CheckType = "URL"
	CheckTypeDNS        CheckType = "DNS"
	CheckTypeSSL        CheckType = "SSL"
	CheckTypeGRPC       CheckType = "GRPC"
)

// Check is a check of any type, as returned by [Client.GetCheck].
//
// Deprecated: use APICheck, BrowserCheck, MultiStepCheck or a monitor type
// such as HeartbeatMonitor, with the client methods of their type.
type Check struct {
	// ID is the check's UUID, assigned by the API.
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	// Type is the check type, such as [CheckTypeAPI].
	Type CheckType `json:"checkType"`
	// Frequency is the number of minutes between runs.
	Frequency int `json:"frequency"`
	// FrequencyOffset is 10, 20 or 30 when Frequency is 0, for a run every
	// that many seconds. For a higher Frequency the API assigns the offset,
	// but rejects a value sent above Frequency * 10 for a Frequency up to 60,
	// or above Frequency / 5 rounded up for a higher one. FrequencyOffset is
	// left out of the request when zero.
	FrequencyOffset int  `json:"frequencyOffset,omitempty"`
	Activated       bool `json:"activated"`
	// Muted stops notifications when the check fails or recovers.
	Muted bool `json:"muted"`
	// ShouldFail inverts the result: a run that would fail passes, and the
	// other way around.
	ShouldFail bool `json:"shouldFail"`
	// RunParallel runs the check in all its locations in parallel.
	RunParallel bool `json:"runParallel"`
	// Locations are the public locations to run from, as region codes such
	// as "us-east-1". Nil is sent as an empty list.
	Locations []string `json:"locations"`
	// DegradedResponseTime and MaxResponseTime are the degraded and failing
	// response time limits in milliseconds.
	DegradedResponseTime int `json:"degradedResponseTime"`
	MaxResponseTime      int `json:"maxResponseTime"`
	// Script is left out of the request when empty.
	Script               string                `json:"script,omitempty"`
	EnvironmentVariables []EnvironmentVariable `json:"environmentVariables"`
	Tags                 []string              `json:"tags,omitempty"`
	SSLCheckDomain       string                `json:"sslCheckDomain"`
	// SetupSnippetID and TearDownSnippetID are the IDs of the [Snippet] run
	// in the setup and teardown phase of an API check; 0 means none.
	SetupSnippetID    int64 `json:"setupSnippetId,omitempty"`
	TearDownSnippetID int64 `json:"tearDownSnippetId,omitempty"`
	// LocalSetupScript and LocalTearDownScript are Node.js code run in the
	// setup and teardown phase of an API check; empty means none.
	LocalSetupScript    string        `json:"localSetupScript,omitempty"`
	LocalTearDownScript string        `json:"localTearDownScript,omitempty"`
	AlertSettings       AlertSettings `json:"alertSettings,omitempty"`
	// UseGlobalAlertSettings applies the account's alert settings instead
	// of AlertSettings.
	UseGlobalAlertSettings bool `json:"useGlobalAlertSettings"`
	// Request is the request of an API check.
	Request Request `json:"request"`
	// Heartbeat is the configuration of a heartbeat monitor.
	Heartbeat Heartbeat `json:"heartbeat"`
	// GroupID is the ID of the check group the check belongs to; 0 means
	// none.
	GroupID int64 `json:"groupId,omitempty"`
	// GroupOrder is the check's position in its group, which sets the run
	// order when the group is triggered from the API or CI/CD.
	GroupOrder int `json:"groupOrder,omitempty"`
	// AlertChannelSubscriptions are the alert channels that receive the
	// check's alerts. The SDK disables automatic assignment, so an empty
	// list leaves the check without any.
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	CreatedAt                 time.Time                  `json:"created_at"`
	UpdatedAt                 time.Time                  `json:"updated_at"`

	// PrivateLocations are the slug names of the account's private
	// locations to run from. Nil is sent as an empty list.
	PrivateLocations *[]string `json:"privateLocations"`
	// RuntimeID is the runtime version the check runs with, such as
	// "2025.04".
	RuntimeID *string `json:"runtimeId"`
	// RetryStrategy sets how failed runs are retried. Nil disables retries.
	// In a group whose RetryStrategy is not [RetryStrategyTypeFallback], the
	// group's strategy applies instead.
	RetryStrategy *RetryStrategy `json:"retryStrategy"`
	// TriggerIncident makes the check open and resolve an incident on a
	// status page service, based on its alert configuration. Nil turns it
	// off.
	TriggerIncident *IncidentTrigger `json:"triggerIncident"`
}

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
// type would silently drop its type-specific fields, so it fails with an error
// matching ErrCheckTypeMismatch instead.
func (c *client) getCheckOfType(
	ctx context.Context,
	path string,
	checkType CheckType,
	result any,
) error {
	res, err := c.call(ctx, http.MethodGet, path, nil, http.StatusOK)
	if err != nil {
		return err
	}
	var header struct {
		CheckType CheckType `json:"checkType"`
	}
	if err := decode(res, &header); err != nil {
		return err
	}
	if header.CheckType != checkType {
		return &messageError{
			msg:     fmt.Sprintf("%s %s: check is of type %q, not %q", res.method, res.path, header.CheckType, checkType),
			wrapped: ErrCheckTypeMismatch,
		}
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
