package checkly

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// CheckType is the type of a check or monitor, in
// [CheckResultsFilter].CheckType and [UnsupportedCheckTypeError].Type.
type CheckType string

// Check types, the values of [CheckResultsFilter].CheckType.
const (
	// CheckTypeAgentic is the type of agentic checks, which the SDK has no
	// resource type for. It can be used as CheckResultsFilter.CheckType;
	// GetCheck returns an error matching [ErrUnsupportedCheckType] for such a
	// check.
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

// Check is a check or monitor of any type, as returned by [Client.GetCheck]: a
// pointer to [APICheck], [BrowserCheck], [MultiStepCheck], [PlaywrightCheck],
// [HeartbeatMonitor], [URLMonitor], [TCPMonitor], [DNSMonitor], [ICMPMonitor],
// [GRPCMonitor], [SSLMonitor] or [TracerouteMonitor]. Use a type switch to read
// it; future SDK versions may add types, so give the switch a default case.
//
// Only the pointer types implement Check, so a type switch case on a
// non-pointer type fails to compile instead of never matching.
type Check interface {
	isCheck()
}

func (c *client) GetCheck(
	ctx context.Context,
	ID string,
) (Check, error) {
	// The type is not known before the response, and browser and multi-step
	// checks omit their script dependencies without includeDependencies.
	res, err := c.call(ctx, http.MethodGet, fmt.Sprintf("/v1/checks/%s?includeDependencies=true", ID), nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	checkType, err := decodeCheckType(res)
	if err != nil {
		return nil, err
	}
	var check Check
	switch checkType {
	case CheckTypeAPI:
		check = &APICheck{}
	case CheckTypeBrowser:
		check = &BrowserCheck{}
	case CheckTypeMultiStep:
		check = &MultiStepCheck{}
	case CheckTypePlaywright:
		check = &PlaywrightCheck{}
	case CheckTypeHeartbeat:
		check = &HeartbeatMonitor{}
	case CheckTypeURL:
		check = &URLMonitor{}
	case CheckTypeTCP:
		check = &TCPMonitor{}
	case CheckTypeDNS:
		check = &DNSMonitor{}
	case CheckTypeICMP:
		check = &ICMPMonitor{}
	case CheckTypeGRPC:
		check = &GRPCMonitor{}
	case CheckTypeSSL:
		check = &SSLMonitor{}
	case CheckTypeTraceroute:
		check = &TracerouteMonitor{}
	default:
		return nil, &UnsupportedCheckTypeError{ID: ID, Type: checkType}
	}
	if err := decode(res, check); err != nil {
		return nil, err
	}
	return check, nil
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
	got, err := decodeCheckType(res)
	if err != nil {
		return err
	}
	if got != checkType {
		return &messageError{
			msg:     fmt.Sprintf("%s %s: check is of type %q, not %q", res.method, res.path, got, checkType),
			wrapped: ErrCheckTypeMismatch,
		}
	}
	return decode(res, result)
}

// decodeCheckType reads the checkType of a check response.
func decodeCheckType(res rawResponse) (CheckType, error) {
	var header struct {
		CheckType CheckType `json:"checkType"`
	}
	if err := decode(res, &header); err != nil {
		return "", err
	}
	return header.CheckType, nil
}

func withAutoAssignAlertsFlag(url string) string {
	flag := "autoAssignAlerts=false"
	if strings.Contains(url, "?") {
		return url + "&" + flag
	}
	return url + "?" + flag
}
