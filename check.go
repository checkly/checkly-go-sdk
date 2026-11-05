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

// Check represents the parameters for an existing check.
//
// Deprecated: use APICheck, BrowserCheck, MultiStepCheck or a monitor type
// such as HeartbeatMonitor, with the client methods of their type.
type Check struct {
	ID                        string                     `json:"id"`
	Name                      string                     `json:"name"`
	Description               *string                    `json:"description"`
	Type                      CheckType                  `json:"checkType"`
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
	CreatedAt                 time.Time                  `json:"created_at"`
	UpdatedAt                 time.Time                  `json:"updated_at"`

	PrivateLocations *[]string        `json:"privateLocations"`
	RuntimeID        *string          `json:"runtimeId"`
	RetryStrategy    *RetryStrategy   `json:"retryStrategy"`
	TriggerIncident  *IncidentTrigger `json:"triggerIncident"`

	// Deprecated: this property will be removed in future versions. Please use RetryStrategy instead.
	DoubleCheck bool `json:"doubleCheck"`
}

func (c *client) CreateCheck(
	ctx context.Context,
	check Check,
) (*Check, error) {
	var endpoint string
	switch check.Type {
	case CheckTypeBrowser:
		endpoint = "/v1/checks/browser"
	case CheckTypeAPI:
		endpoint = "/v1/checks/api"
	case CheckTypeHeartbeat:
		endpoint = "/v1/checks/heartbeat"
	case CheckTypeMultiStep:
		endpoint = "/v1/checks/multistep"
	case CheckTypeTCP:
		return nil, fmt.Errorf("user error: use CreateTCPMonitor to create TCP monitors")
	case CheckTypeURL:
		return nil, fmt.Errorf("user error: use CreateURLMonitor to create URL monitors")
	case CheckTypeDNS:
		return nil, fmt.Errorf("user error: use CreateDNSMonitor to create DNS monitors")
	case CheckTypeICMP:
		return nil, fmt.Errorf("user error: use CreateICMPMonitor to create ICMP monitors")
	case CheckTypeGRPC:
		return nil, fmt.Errorf("user error: use CreateGRPCMonitor to create GRPC monitors")
	case CheckTypeTraceroute:
		return nil, fmt.Errorf("user error: use CreateTracerouteMonitor to create TRACEROUTE monitors")
	case CheckTypeSSL:
		return nil, fmt.Errorf("user error: use CreateSSLMonitor to create SSL monitors")
	case CheckTypePlaywright:
		return nil, fmt.Errorf("user error: use CreatePlaywrightCheck to create PLAYWRIGHT checks")
	case CheckTypeAgentic:
		return nil, fmt.Errorf("user error: the SDK cannot create AGENTIC checks")
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
	// The update endpoint keeps the stored value of an omitted key. The
	// fields below shadow embedded fields whose zero value would be omitted,
	// so that an update resets them: lists are sent empty, and the rest as
	// null, which is also their default on create.
	Tags                      []string                   `json:"tags"`
	SetupSnippetID            *int64                     `json:"setupSnippetId"`
	TearDownSnippetID         *int64                     `json:"tearDownSnippetId"`
	LocalSetupScript          *string                    `json:"localSetupScript"`
	LocalTearDownScript       *string                    `json:"localTearDownScript"`
	GroupOrder                *int                       `json:"groupOrder"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions"`
}

func createCheckPayload(check Check) checkPayload {
	payload := checkPayload{
		Check:                     check,
		Tags:                      emptyIfNil(check.Tags),
		SetupSnippetID:            nullIfZero(check.SetupSnippetID),
		TearDownSnippetID:         nullIfZero(check.TearDownSnippetID),
		LocalSetupScript:          nullIfZero(check.LocalSetupScript),
		LocalTearDownScript:       nullIfZero(check.LocalTearDownScript),
		GroupOrder:                nullIfZero(check.GroupOrder),
		AlertChannelSubscriptions: emptyIfNil(check.AlertChannelSubscriptions),
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
