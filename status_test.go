package checkly_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	checkly "github.com/checkly/checkly-go-sdk"
)

// statusCase describes one call of a Client method for TestSuccessStatuses:
// the requests it sends, the response statuses it treats as success, and the
// canned body that lets a successful response decode.
type statusCase struct {
	// requests lists the "METHOD /path?query" lines the method sends, in
	// order. Only GetStaticIPs sends more than one.
	requests []string
	// leadStatuses are the statuses answered to the first requests, before
	// the status under test applies to the remaining ones. They let a case
	// test how a later request of a multi-request method handles a status.
	leadStatuses []int
	// notRetried marks methods that bypass the rate-limit retries, so that a
	// 429 response fails them instead of being re-sent.
	notRetried bool
	// apiError marks methods that report a rejected status as an *APIError
	// with that status. The others only have to return some error.
	apiError bool
	// accept lists the response statuses the method treats as success.
	accept []int
	// body is the canned response body for every status but 204, which
	// carries no body. It defaults to "{}".
	body string
	call func(ctx context.Context, c checkly.Client) error
}

var (
	statusOK          = []int{http.StatusOK}
	statusCreated     = []int{http.StatusCreated}
	statusOKOrCreated = []int{http.StatusOK, http.StatusCreated}
	statusNoContent   = []int{http.StatusNoContent}
)

// candidateStatuses are the statuses every method is tried with. A method must
// succeed for each status in its accept list and fail for the others. 429 is
// left out because the client retries it; TestRateLimitRetryPerMethod covers it.
var candidateStatuses = []int{
	http.StatusOK,
	http.StatusCreated,
	http.StatusAccepted,
	http.StatusNoContent,
	http.StatusBadRequest,
	http.StatusUnauthorized,
	http.StatusForbidden,
	http.StatusNotFound,
	http.StatusConflict,
	http.StatusInternalServerError,
}

const (
	testID      = "id1"
	testIntID   = int64(7)
	testStatusP = "sp1"
)

func ignoreResult[T any](_ T, err error) error { return err }

// statusCases pins the success statuses, HTTP methods and paths of every
// Client method that makes a request. A key is the method name, optionally
// followed by "/" and a variant name when one method needs several cases.
var statusCases = map[string]statusCase{
	// Checks
	"Create": {
		requests: []string{"POST /v1/checks?autoAssignAlerts=false"},
		accept:   statusCreated,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.Create(ctx, checkly.Check{Type: checkly.TypeAPI}))
		},
	},
	"Update": {
		requests: []string{"PUT /v1/checks/id1?autoAssignAlerts=false"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.Update(ctx, testID, checkly.Check{Type: checkly.TypeAPI}))
		},
	},
	"Delete":              deleteCheckCase(checkly.Client.Delete),
	"Get":                 getCheckCase(checkly.Client.Get),
	"GetHeartbeatCheck":   getCheckCase(checkly.Client.GetHeartbeatCheck),
	"GetHeartbeatMonitor": getCheckCase(checkly.Client.GetHeartbeatMonitor),
	"CreateCheck": {
		requests: []string{"POST /v1/checks/api?autoAssignAlerts=false"},
		accept:   statusCreated,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateCheck(ctx, checkly.Check{Type: checkly.TypeAPI}))
		},
	},
	"CreateHeartbeat": {
		requests: []string{"POST /v1/checks/heartbeat?autoAssignAlerts=false"},
		accept:   statusCreated,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateHeartbeat(ctx, checkly.HeartbeatCheck{}))
		},
	},
	"CreateHeartbeatMonitor": {
		requests: []string{"POST /v1/checks/heartbeat?autoAssignAlerts=false"},
		accept:   statusCreated,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateHeartbeatMonitor(ctx, checkly.HeartbeatMonitor{}))
		},
	},
	"CreateTCPCheck": {
		requests: []string{"POST /v1/checks/tcp?autoAssignAlerts=false"},
		accept:   statusCreated,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateTCPCheck(ctx, checkly.TCPCheck{}))
		},
	},
	"CreateTCPMonitor": {
		requests: []string{"POST /v1/checks/tcp?autoAssignAlerts=false"},
		accept:   statusCreated,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateTCPMonitor(ctx, checkly.TCPMonitor{}))
		},
	},
	"CreateGRPCMonitor": {
		requests: []string{"POST /v1/checks/grpc?autoAssignAlerts=false"},
		accept:   statusCreated,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateGRPCMonitor(ctx, checkly.GRPCMonitor{}))
		},
	},
	"CreateTracerouteMonitor": {
		requests: []string{"POST /v1/checks/traceroute?autoAssignAlerts=false"},
		accept:   statusCreated,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateTracerouteMonitor(ctx, checkly.TracerouteMonitor{}))
		},
	},
	"CreateSSLMonitor": {
		requests: []string{"POST /v1/checks/ssl?autoAssignAlerts=false"},
		accept:   statusCreated,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateSSLMonitor(ctx, checkly.SSLMonitor{}))
		},
	},
	"CreateURLMonitor": {
		requests: []string{"POST /v1/checks/url?autoAssignAlerts=false"},
		accept:   statusCreated,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateURLMonitor(ctx, checkly.URLMonitor{}))
		},
	},
	"CreateDNSMonitor": {
		requests: []string{"POST /v1/checks/dns?autoAssignAlerts=false"},
		accept:   statusCreated,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateDNSMonitor(ctx, checkly.DNSMonitor{}))
		},
	},
	"CreateICMPMonitor": {
		requests: []string{"POST /v1/checks/icmp?autoAssignAlerts=false"},
		accept:   statusCreated,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateICMPMonitor(ctx, checkly.ICMPMonitor{}))
		},
	},
	"CreatePlaywrightCheck": {
		requests: []string{"POST /v1/checks/playwright?autoAssignAlerts=false"},
		accept:   statusCreated,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreatePlaywrightCheck(ctx, checkly.PlaywrightCheck{}))
		},
	},
	"CreateAPICheck": {
		requests: []string{"POST /v1/checks/api?autoAssignAlerts=false"},
		accept:   statusCreated,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateAPICheck(ctx, checkly.APICheck{}))
		},
	},
	"CreateBrowserCheck": {
		requests: []string{"POST /v1/checks/browser?autoAssignAlerts=false"},
		accept:   statusCreated,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateBrowserCheck(ctx, checkly.BrowserCheck{}))
		},
	},
	"CreateMultiStepCheck": {
		requests: []string{"POST /v1/checks/multistep?autoAssignAlerts=false"},
		accept:   statusCreated,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateMultiStepCheck(ctx, checkly.MultiStepCheck{}))
		},
	},
	"UpdateCheck": {
		requests: []string{"PUT /v1/checks/id1?autoAssignAlerts=false"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateCheck(ctx, testID, checkly.Check{Type: checkly.TypeAPI}))
		},
	},
	"UpdateHeartbeat": {
		requests: []string{"PUT /v1/checks/heartbeat/id1?autoAssignAlerts=false"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateHeartbeat(ctx, testID, checkly.HeartbeatCheck{}))
		},
	},
	"UpdateHeartbeatMonitor": {
		requests: []string{"PUT /v1/checks/heartbeat/id1?autoAssignAlerts=false"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateHeartbeatMonitor(ctx, testID, checkly.HeartbeatMonitor{}))
		},
	},
	"UpdateTCPCheck": {
		requests: []string{"PUT /v1/checks/tcp/id1?autoAssignAlerts=false"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateTCPCheck(ctx, testID, checkly.TCPCheck{}))
		},
	},
	"UpdateTCPMonitor": {
		requests: []string{"PUT /v1/checks/tcp/id1?autoAssignAlerts=false"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateTCPMonitor(ctx, testID, checkly.TCPMonitor{}))
		},
	},
	"UpdateGRPCMonitor": {
		requests: []string{"PUT /v1/checks/grpc/id1?autoAssignAlerts=false"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateGRPCMonitor(ctx, testID, checkly.GRPCMonitor{}))
		},
	},
	"UpdateTracerouteMonitor": {
		requests: []string{"PUT /v1/checks/traceroute/id1?autoAssignAlerts=false"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateTracerouteMonitor(ctx, testID, checkly.TracerouteMonitor{}))
		},
	},
	"UpdateSSLMonitor": {
		requests: []string{"PUT /v1/checks/ssl/id1?autoAssignAlerts=false"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateSSLMonitor(ctx, testID, checkly.SSLMonitor{}))
		},
	},
	"UpdateURLMonitor": {
		requests: []string{"PUT /v1/checks/url/id1?autoAssignAlerts=false"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateURLMonitor(ctx, testID, checkly.URLMonitor{}))
		},
	},
	"UpdateDNSMonitor": {
		requests: []string{"PUT /v1/checks/dns/id1?autoAssignAlerts=false"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateDNSMonitor(ctx, testID, checkly.DNSMonitor{}))
		},
	},
	"UpdateICMPMonitor": {
		requests: []string{"PUT /v1/checks/icmp/id1?autoAssignAlerts=false"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateICMPMonitor(ctx, testID, checkly.ICMPMonitor{}))
		},
	},
	"UpdatePlaywrightCheck": {
		requests: []string{"PUT /v1/checks/playwright/id1?autoAssignAlerts=false"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdatePlaywrightCheck(ctx, testID, checkly.PlaywrightCheck{}))
		},
	},
	"UpdateAPICheck": {
		requests: []string{"PUT /v1/checks/api/id1?autoAssignAlerts=false"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateAPICheck(ctx, testID, checkly.APICheck{}))
		},
	},
	"UpdateBrowserCheck": {
		requests: []string{"PUT /v1/checks/browser/id1?autoAssignAlerts=false"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateBrowserCheck(ctx, testID, checkly.BrowserCheck{}))
		},
	},
	"UpdateMultiStepCheck": {
		requests: []string{"PUT /v1/checks/multistep/id1?autoAssignAlerts=false"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateMultiStepCheck(ctx, testID, checkly.MultiStepCheck{}))
		},
	},
	"DeleteCheck":             deleteCheckCase(checkly.Client.DeleteCheck),
	"DeleteHeartbeatMonitor":  deleteCheckCase(checkly.Client.DeleteHeartbeatMonitor),
	"DeleteTCPMonitor":        deleteCheckCase(checkly.Client.DeleteTCPMonitor),
	"DeleteGRPCMonitor":       deleteCheckCase(checkly.Client.DeleteGRPCMonitor),
	"DeleteTracerouteMonitor": deleteCheckCase(checkly.Client.DeleteTracerouteMonitor),
	"DeleteSSLMonitor":        deleteCheckCase(checkly.Client.DeleteSSLMonitor),
	"DeleteURLMonitor":        deleteCheckCase(checkly.Client.DeleteURLMonitor),
	"DeleteDNSMonitor":        deleteCheckCase(checkly.Client.DeleteDNSMonitor),
	"DeleteICMPMonitor":       deleteCheckCase(checkly.Client.DeleteICMPMonitor),
	"DeletePlaywrightCheck":   deleteCheckCase(checkly.Client.DeletePlaywrightCheck),
	"DeleteAPICheck":          deleteCheckCase(checkly.Client.DeleteAPICheck),
	"DeleteBrowserCheck":      deleteCheckCase(checkly.Client.DeleteBrowserCheck),
	"DeleteMultiStepCheck":    deleteCheckCase(checkly.Client.DeleteMultiStepCheck),
	"GetCheck":                getCheckCase(checkly.Client.GetCheck),
	"GetTCPCheck":             getCheckCase(checkly.Client.GetTCPCheck),
	"GetTCPMonitor":           getCheckCase(checkly.Client.GetTCPMonitor),
	"GetGRPCMonitor":          getCheckCase(checkly.Client.GetGRPCMonitor),
	"GetTracerouteMonitor":    getCheckCase(checkly.Client.GetTracerouteMonitor),
	"GetSSLMonitor":           getCheckCase(checkly.Client.GetSSLMonitor),
	"GetURLMonitor":           getCheckCase(checkly.Client.GetURLMonitor),
	"GetDNSMonitor":           getCheckCase(checkly.Client.GetDNSMonitor),
	"GetICMPMonitor":          getCheckCase(checkly.Client.GetICMPMonitor),
	"GetPlaywrightCheck":      getCheckCase(checkly.Client.GetPlaywrightCheck),
	// The per-type getters reject a response whose checkType differs from
	// theirs, so the canned body has to name the right type.
	"GetAPICheck": {
		requests: []string{"GET /v1/checks/id1"},
		accept:   statusOK,
		apiError: true,
		body:     `{"checkType":"API"}`,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetAPICheck(ctx, testID))
		},
	},
	"GetBrowserCheck": {
		requests: []string{"GET /v1/checks/id1?includeDependencies=true"},
		accept:   statusOK,
		apiError: true,
		body:     `{"checkType":"BROWSER"}`,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetBrowserCheck(ctx, testID))
		},
	},
	"GetMultiStepCheck": {
		requests: []string{"GET /v1/checks/id1?includeDependencies=true"},
		accept:   statusOK,
		apiError: true,
		body:     `{"checkType":"MULTI_STEP"}`,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetMultiStepCheck(ctx, testID))
		},
	},

	// Groups
	"CreateGroup": {
		requests: []string{"POST /v1/check-groups?autoAssignAlerts=false"},
		accept:   statusCreated,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateGroup(ctx, checkly.Group{}))
		},
	},
	"CreateGroupV2": {
		requests: []string{"POST /v2/check-groups?autoAssignAlerts=false"},
		accept:   statusCreated,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateGroupV2(ctx, checkly.GroupV2{}))
		},
	},
	"GetGroup": {
		requests: []string{"GET /v1/check-groups/7"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetGroup(ctx, testIntID))
		},
	},
	"GetGroupV2": {
		requests: []string{"GET /v1/check-groups/7"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetGroupV2(ctx, testIntID))
		},
	},
	"UpdateGroup": {
		requests: []string{"PUT /v1/check-groups/7?autoAssignAlerts=false"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateGroup(ctx, testIntID, checkly.Group{}))
		},
	},
	"UpdateGroupV2": {
		requests: []string{"PUT /v2/check-groups/7?autoAssignAlerts=false"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateGroupV2(ctx, testIntID, checkly.GroupV2{}))
		},
	},
	"DeleteGroup": {
		requests: []string{"DELETE /v1/check-groups/7"},
		accept:   statusNoContent,
		apiError: true,
		call:     func(ctx context.Context, c checkly.Client) error { return c.DeleteGroup(ctx, testIntID) },
	},
	"DeleteGroupV2": {
		requests: []string{"DELETE /v1/check-groups/7"},
		accept:   statusNoContent,
		apiError: true,
		call:     func(ctx context.Context, c checkly.Client) error { return c.DeleteGroupV2(ctx, testIntID) },
	},

	// Check results
	"GetCheckResult": {
		requests: []string{"GET /v1/check-results/id1/res1"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetCheckResult(ctx, testID, "res1"))
		},
	},
	"GetCheckResults/no filter": {
		requests: []string{"GET /v1/check-results/id1"},
		accept:   statusOK,
		apiError: true,
		body:     `[]`,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetCheckResults(ctx, testID, nil))
		},
	},
	"GetCheckResults/all filters": {
		requests: []string{"GET /v1/check-results/id1?checkType=API&from=1700000000&hasFailures=1&limit=5&location=eu-west-1&page=2&to=1700003600"},
		accept:   statusOK,
		apiError: true,
		body:     `[]`,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetCheckResults(ctx, testID, &checkly.CheckResultsFilter{
				Limit:       5,
				Page:        2,
				Location:    "eu-west-1",
				From:        1700000000,
				To:          1700003600,
				CheckType:   checkly.TypeAPI,
				HasFailures: true,
			}))
		},
	},

	// Snippets
	"CreateSnippet": {
		requests: []string{"POST /v1/snippets"},
		accept:   statusCreated,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateSnippet(ctx, checkly.Snippet{}))
		},
	},
	"GetSnippet": {
		requests: []string{"GET /v1/snippets/7"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetSnippet(ctx, testIntID))
		},
	},
	"UpdateSnippet": {
		requests: []string{"PUT /v1/snippets/7"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateSnippet(ctx, testIntID, checkly.Snippet{}))
		},
	},
	"DeleteSnippet": {
		requests: []string{"DELETE /v1/snippets/7"},
		accept:   statusNoContent,
		call:     func(ctx context.Context, c checkly.Client) error { return c.DeleteSnippet(ctx, testIntID) },
	},

	// Environment variables
	"CreateEnvironmentVariable": {
		requests: []string{"POST /v1/variables"},
		accept:   statusCreated,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateEnvironmentVariable(ctx, checkly.EnvironmentVariable{}))
		},
	},
	"GetEnvironmentVariable": {
		requests: []string{"GET /v1/variables/KEY"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetEnvironmentVariable(ctx, "KEY"))
		},
	},
	"UpdateEnvironmentVariable": {
		requests: []string{"PUT /v1/variables/KEY"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateEnvironmentVariable(ctx, "KEY", checkly.EnvironmentVariable{}))
		},
	},
	"DeleteEnvironmentVariable": {
		requests: []string{"DELETE /v1/variables/KEY"},
		accept:   statusNoContent,
		call:     func(ctx context.Context, c checkly.Client) error { return c.DeleteEnvironmentVariable(ctx, "KEY") },
	},

	// Alert channels
	"CreateAlertChannel": {
		requests: []string{"POST /v1/alert-channels"},
		accept:   statusOKOrCreated,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateAlertChannel(ctx, checkly.AlertChannel{}))
		},
	},
	"GetAlertChannel": {
		requests: []string{"GET /v1/alert-channels/7"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetAlertChannel(ctx, testIntID))
		},
	},
	"UpdateAlertChannel": {
		requests: []string{"PUT /v1/alert-channels/7"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateAlertChannel(ctx, testIntID, checkly.AlertChannel{}))
		},
	},
	"DeleteAlertChannel": {
		requests: []string{"DELETE /v1/alert-channels/7"},
		accept:   statusNoContent,
		call:     func(ctx context.Context, c checkly.Client) error { return c.DeleteAlertChannel(ctx, testIntID) },
	},

	// Dashboards
	"CreateDashboard": {
		requests: []string{"POST /v1/dashboards"},
		accept:   statusOKOrCreated,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateDashboard(ctx, checkly.Dashboard{}))
		},
	},
	"GetDashboard": {
		requests: []string{"GET /v1/dashboards/id1"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetDashboard(ctx, testID))
		},
	},
	"UpdateDashboard": {
		requests: []string{"PUT /v1/dashboards/id1"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateDashboard(ctx, testID, checkly.Dashboard{}))
		},
	},
	"DeleteDashboard": {
		requests: []string{"DELETE /v1/dashboards/id1"},
		accept:   statusNoContent,
		call:     func(ctx context.Context, c checkly.Client) error { return c.DeleteDashboard(ctx, testID) },
	},

	// Maintenance windows
	"CreateMaintenanceWindow": {
		requests: []string{"POST /v1/maintenance-windows"},
		accept:   statusOKOrCreated,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateMaintenanceWindow(ctx, checkly.MaintenanceWindow{}))
		},
	},
	"GetMaintenanceWindow": {
		requests: []string{"GET /v1/maintenance-windows/7"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetMaintenanceWindow(ctx, testIntID))
		},
	},
	"UpdateMaintenanceWindow": {
		requests: []string{"PUT /v1/maintenance-windows/7"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateMaintenanceWindow(ctx, testIntID, checkly.MaintenanceWindow{}))
		},
	},
	"DeleteMaintenanceWindow": {
		requests: []string{"DELETE /v1/maintenance-windows/7"},
		accept:   statusNoContent,
		call:     func(ctx context.Context, c checkly.Client) error { return c.DeleteMaintenanceWindow(ctx, testIntID) },
	},

	// Private locations
	"CreatePrivateLocation": {
		requests: []string{"POST /v1/private-locations"},
		accept:   statusOKOrCreated,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreatePrivateLocation(ctx, checkly.PrivateLocation{}))
		},
	},
	"GetPrivateLocation": {
		requests: []string{"GET /v1/private-locations/id1"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetPrivateLocation(ctx, testID))
		},
	},
	"UpdatePrivateLocation": {
		requests: []string{"PUT /v1/private-locations/id1"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdatePrivateLocation(ctx, testID, checkly.PrivateLocation{}))
		},
	},
	"DeletePrivateLocation": {
		requests: []string{"DELETE /v1/private-locations/id1"},
		accept:   statusNoContent,
		call:     func(ctx context.Context, c checkly.Client) error { return c.DeletePrivateLocation(ctx, testID) },
	},

	// Triggers
	"CreateTriggerCheck": {
		requests: []string{"POST /v1/triggers/checks/id1"},
		accept:   statusOKOrCreated,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateTriggerCheck(ctx, testID))
		},
	},
	"GetTriggerCheck": {
		requests: []string{"GET /v1/triggers/checks/id1"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetTriggerCheck(ctx, testID))
		},
	},
	"DeleteTriggerCheck": {
		requests: []string{"DELETE /v1/triggers/checks/id1"},
		accept:   statusNoContent,
		call:     func(ctx context.Context, c checkly.Client) error { return c.DeleteTriggerCheck(ctx, testID) },
	},
	"CreateTriggerGroup": {
		requests: []string{"POST /v1/triggers/check-groups/7"},
		accept:   statusOKOrCreated,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateTriggerGroup(ctx, testIntID))
		},
	},
	"GetTriggerGroup": {
		requests: []string{"GET /v1/triggers/check-groups/7"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetTriggerGroup(ctx, testIntID))
		},
	},
	"DeleteTriggerGroup": {
		requests: []string{"DELETE /v1/triggers/check-groups/7"},
		accept:   statusNoContent,
		call:     func(ctx context.Context, c checkly.Client) error { return c.DeleteTriggerGroup(ctx, testIntID) },
	},

	// Client certificates
	"CreateClientCertificate": {
		requests: []string{"POST /v1/client-certificates"},
		accept:   statusCreated,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateClientCertificate(ctx, checkly.ClientCertificate{}))
		},
	},
	"GetClientCertificate": {
		requests: []string{"GET /v1/client-certificates/id1"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetClientCertificate(ctx, testID))
		},
	},
	"DeleteClientCertificate": {
		requests: []string{"DELETE /v1/client-certificates/id1"},
		accept:   statusNoContent,
		call:     func(ctx context.Context, c checkly.Client) error { return c.DeleteClientCertificate(ctx, testID) },
	},

	// Status pages (v1)
	"CreateStatusPage": {
		requests: []string{"POST /v1/status-pages"},
		accept:   statusCreated,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateStatusPage(ctx, checkly.StatusPage{}))
		},
	},
	"GetStatusPage": {
		requests: []string{"GET /v1/status-pages/id1"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetStatusPage(ctx, testID))
		},
	},
	"UpdateStatusPage": {
		requests: []string{"PUT /v1/status-pages/id1"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateStatusPage(ctx, testID, checkly.StatusPage{}))
		},
	},
	"DeleteStatusPage": {
		requests: []string{"DELETE /v1/status-pages/id1"},
		accept:   statusNoContent,
		call:     func(ctx context.Context, c checkly.Client) error { return c.DeleteStatusPage(ctx, testID) },
	},
	"CreateStatusPageService": {
		requests: []string{"POST /v1/status-pages/services"},
		accept:   statusCreated,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateStatusPageService(ctx, checkly.StatusPageService{}))
		},
	},
	"GetStatusPageService": {
		requests: []string{"GET /v1/status-pages/services/id1"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetStatusPageService(ctx, testID))
		},
	},
	"UpdateStatusPageService": {
		requests: []string{"PUT /v1/status-pages/services/id1"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateStatusPageService(ctx, testID, checkly.StatusPageService{}))
		},
	},
	"DeleteStatusPageService": {
		requests: []string{"DELETE /v1/status-pages/services/id1"},
		accept:   statusNoContent,
		call:     func(ctx context.Context, c checkly.Client) error { return c.DeleteStatusPageService(ctx, testID) },
	},

	// Status pages (v3)
	"CreateStatusPageV3": {
		requests: []string{"POST /v3/status-pages"},
		accept:   statusCreated,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateStatusPageV3(ctx, checkly.StatusPageV3{}))
		},
	},
	"GetStatusPageV3": {
		requests: []string{"GET /v3/status-pages/sp1"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetStatusPageV3(ctx, testStatusP))
		},
	},
	"UpdateStatusPageV3": {
		requests: []string{"PUT /v3/status-pages/sp1"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateStatusPageV3(ctx, testStatusP, checkly.StatusPageV3{}))
		},
	},
	"DeleteStatusPageV3": {
		requests: []string{"DELETE /v3/status-pages/sp1"},
		accept:   statusNoContent,
		call:     func(ctx context.Context, c checkly.Client) error { return c.DeleteStatusPageV3(ctx, testStatusP) },
	},
	"CreateStatusPageComponentV3": {
		requests: []string{"POST /v3/status-pages/sp1/components"},
		accept:   statusCreated,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateStatusPageComponentV3(ctx, testStatusP, checkly.StatusPageComponentV3{}))
		},
	},
	"GetStatusPageComponentV3": {
		requests: []string{"GET /v3/status-pages/sp1/components/id1"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetStatusPageComponentV3(ctx, testStatusP, testID))
		},
	},
	"UpdateStatusPageComponentV3": {
		requests: []string{"PUT /v3/status-pages/sp1/components/id1"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateStatusPageComponentV3(ctx, testStatusP, testID, checkly.StatusPageComponentV3{}))
		},
	},
	"DeleteStatusPageComponentV3": {
		requests: []string{"DELETE /v3/status-pages/sp1/components/id1"},
		accept:   statusNoContent,
		call: func(ctx context.Context, c checkly.Client) error {
			return c.DeleteStatusPageComponentV3(ctx, testStatusP, testID)
		},
	},
	"CreateStatusPageAutomationRuleV3": {
		requests: []string{"POST /v3/status-pages/sp1/automation-rules"},
		accept:   statusCreated,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.CreateStatusPageAutomationRuleV3(ctx, testStatusP, checkly.StatusPageAutomationRuleV3{}))
		},
	},
	"GetStatusPageAutomationRuleV3": {
		requests: []string{"GET /v3/status-pages/sp1/automation-rules/id1"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetStatusPageAutomationRuleV3(ctx, testStatusP, testID))
		},
	},
	"UpdateStatusPageAutomationRuleV3": {
		requests: []string{"PUT /v3/status-pages/sp1/automation-rules/id1"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.UpdateStatusPageAutomationRuleV3(ctx, testStatusP, testID, checkly.StatusPageAutomationRuleV3{}))
		},
	},
	"DeleteStatusPageAutomationRuleV3": {
		requests: []string{"DELETE /v3/status-pages/sp1/automation-rules/id1"},
		accept:   statusNoContent,
		call: func(ctx context.Context, c checkly.Client) error {
			return c.DeleteStatusPageAutomationRuleV3(ctx, testStatusP, testID)
		},
	},

	// Runtimes and static IPs
	"GetRuntime": {
		requests: []string{"GET /v1/runtimes/2024.02"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetRuntime(ctx, "2024.02"))
		},
	},
	// Both static IP requests decode a JSON object, so "{}" serves both. The
	// first case varies the status of the IPv6 request, the second that of
	// the IPv4 request that follows a successful IPv6 one.
	"GetStaticIPs/first request": {
		requests: []string{"GET /v1/static-ipv6s-by-region", "GET /v1/static-ips-by-region"},
		accept:   statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetStaticIPs(ctx))
		},
	},
	"GetStaticIPs/second request": {
		requests:     []string{"GET /v1/static-ipv6s-by-region", "GET /v1/static-ips-by-region"},
		leadStatuses: []int{http.StatusOK},
		accept:       statusOK,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.GetStaticIPs(ctx))
		},
	},

	// Code bundles. These requests bypass the rate-limit retries and go to
	// the /next API.
	"UploadCodeBundle": {
		requests:   []string{"POST /next/checkly-storage/upload-code-bundle"},
		accept:     statusOK,
		notRetried: true,
		call: func(ctx context.Context, c checkly.Client) error {
			data := "bundle"
			return ignoreResult(c.UploadCodeBundle(ctx, strings.NewReader(data), int64(len(data)), checkly.UploadCodeBundleOptions{}))
		},
	},
	"PeekCodeBundle": {
		requests:   []string{"POST /next/checkly-storage/peek-code-bundle"},
		accept:     statusOK,
		notRetried: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(c.PeekCodeBundle(ctx, "bundle-key"))
		},
	},
}

// deleteCheckCase covers the check and monitor deleters, which all share the
// generic check endpoint and DeleteCheck's error handling.
func deleteCheckCase(del func(checkly.Client, context.Context, string) error) statusCase {
	return statusCase{
		requests: []string{"DELETE /v1/checks/id1"},
		accept:   statusNoContent,
		apiError: true,
		call:     func(ctx context.Context, c checkly.Client) error { return del(c, ctx, testID) },
	}
}

// getCheckCase covers the check and monitor getters that read the generic
// check endpoint without checking the type of the returned check.
func getCheckCase[T any](get func(checkly.Client, context.Context, string) (T, error)) statusCase {
	return statusCase{
		requests: []string{"GET /v1/checks/id1"},
		accept:   statusOK,
		apiError: true,
		call: func(ctx context.Context, c checkly.Client) error {
			return ignoreResult(get(c, ctx, testID))
		},
	}
}

// methodsWithoutStatusCase are the Client methods TestSuccessStatuses does
// not cover, because they make no request.
var methodsWithoutStatusCase = []string{
	"SetAccountId",
	"SetChecklySource",
}

// statusCaseMethod returns the Client method name of a statusCases key.
func statusCaseMethod(key string) string {
	method, _, _ := strings.Cut(key, "/")
	return method
}

// TestStatusCasesCoverClient fails when a Client method has no entry in
// statusCases, so that a new method cannot skip the success-status check.
func TestStatusCasesCoverClient(t *testing.T) {
	t.Parallel()
	covered := map[string]bool{}
	for key := range statusCases {
		covered[statusCaseMethod(key)] = true
	}
	clientType := reflect.TypeOf((*checkly.Client)(nil)).Elem()
	methods := map[string]bool{}
	for i := 0; i < clientType.NumMethod(); i++ {
		name := clientType.Method(i).Name
		methods[name] = true
		excluded := slices.Contains(methodsWithoutStatusCase, name)
		if !covered[name] && !excluded {
			t.Errorf("Client method %s has no entry in statusCases", name)
		}
		if covered[name] && excluded {
			t.Errorf("Client method %s is both in statusCases and excluded from it", name)
		}
	}
	for name := range covered {
		if !methods[name] {
			t.Errorf("statusCases has an entry for %s, which is not a Client method", name)
		}
	}
}

// TestSuccessStatuses checks, for every Client method, that each status in
// its accept list succeeds, that every other candidate status is rejected, and
// that the method sends the expected requests.
func TestSuccessStatuses(t *testing.T) {
	t.Parallel()
	for name, tc := range statusCases {
		name, tc := name, tc
		for _, status := range candidateStatuses {
			status := status
			want := slices.Contains(tc.accept, status)
			t.Run(name+"/"+strconv.Itoa(status), func(t *testing.T) {
				t.Parallel()
				body := tc.body
				if body == "" {
					body = "{}"
				}
				statuses := append(slices.Clone(tc.leadStatuses), status)
				srv, got := newStatusServer(t, statuses, nil, body)
				c := checkly.NewClient(srv.URL, "test-key", srv.Client(), nil)

				err := tc.call(context.Background(), c)
				requests := got()

				if want {
					if err != nil {
						t.Fatalf("want success for status %d, got %v", status, err)
					}
					if !slices.Equal(requests, tc.requests) {
						t.Errorf("want requests %q, got %q", tc.requests, requests)
					}
					return
				}
				assertRejected(t, err, status, tc.apiError)
				// A rejected response ends the call, so no request after
				// the rejected one is sent.
				if wantSent := tc.requests[:len(statuses)]; !slices.Equal(requests, wantSent) {
					t.Errorf("want requests %q, got %q", wantSent, requests)
				}
			})
		}
	}
}

// TestRateLimitRetryPerMethod checks, for every Client method, that a 429 response
// followed by an accepted status makes the method re-send the request and
// succeed, or, for the methods that bypass the retries, fail after one
// request.
func TestRateLimitRetryPerMethod(t *testing.T) {
	t.Parallel()
	header := http.Header{"Retry-After": {"0"}}
	for name, tc := range statusCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			body := tc.body
			if body == "" {
				body = "{}"
			}
			// The request under test gets the 429; the requests before it
			// get their lead statuses and every later one is accepted.
			retried := len(tc.leadStatuses)
			statuses := append(slices.Clone(tc.leadStatuses), http.StatusTooManyRequests, tc.accept[0])
			srv, got := newStatusServer(t, statuses, map[int]http.Header{http.StatusTooManyRequests: header}, body)
			c := checkly.NewClient(srv.URL, "test-key", srv.Client(), nil)
			checkly.SkipRetryWaits(c)

			err := tc.call(context.Background(), c)
			requests := got()

			if tc.notRetried {
				if err == nil {
					t.Error("want error for a 429 response, got nil")
				}
				if want := tc.requests[:retried+1]; !slices.Equal(requests, want) {
					t.Errorf("want requests %q, got %q", want, requests)
				}
				return
			}
			if err != nil {
				t.Fatalf("want success after a retried 429, got %v", err)
			}
			want := slices.Insert(slices.Clone(tc.requests), retried, tc.requests[retried])
			if !slices.Equal(requests, want) {
				t.Errorf("want requests %q, got %q", want, requests)
			}
		})
	}
}

// assertRejected checks that a method reported an error for a response status
// it does not accept and, when apiError is set, that the error is an
// *APIError for that status.
func assertRejected(t *testing.T, err error, status int, apiError bool) {
	t.Helper()
	if err == nil {
		t.Fatalf("want error for status %d, got nil", status)
	}
	if !apiError {
		return
	}
	var apiErr *checkly.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("want *APIError for status %d, got %T: %v", status, err, err)
	}
	if apiErr.StatusCode != status {
		t.Errorf("want *APIError with status %d, got %d: %v", status, apiErr.StatusCode, err)
	}
}

// newStatusServer answers the nth request with the nth entry of statuses,
// repeating the last entry once they run out, with the headers given for that
// status, and with body unless the status is 204. The returned function lists
// the requests received so far as "METHOD /path?query" lines.
func newStatusServer(t *testing.T, statuses []int, headers map[int]http.Header, body string) (*httptest.Server, func() []string) {
	t.Helper()
	var (
		mu       sync.Mutex
		requests []string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		mu.Lock()
		requests = append(requests, r.Method+" "+r.URL.RequestURI())
		status := statuses[min(len(requests), len(statuses))-1]
		mu.Unlock()
		for k, v := range headers[status] {
			w.Header()[k] = v
		}
		w.WriteHeader(status)
		if status != http.StatusNoContent {
			io.WriteString(w, body)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, func() []string {
		mu.Lock()
		defer mu.Unlock()
		return slices.Clone(requests)
	}
}

// TestPeekCodeBundleNotFound checks that PeekCodeBundle reports both a 403 and
// a 404 as ErrCodeBundleNotFound: the backend answers 403 for a key that
// belongs to another account.
func TestPeekCodeBundleNotFound(t *testing.T) {
	t.Parallel()
	for _, status := range []int{http.StatusForbidden, http.StatusNotFound} {
		status := status
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			t.Parallel()
			srv, _ := newStatusServer(t, []int{status}, nil, `{}`)
			c := checkly.NewClient(srv.URL, "test-key", srv.Client(), nil)
			_, err := c.PeekCodeBundle(context.Background(), "bundle-key")
			if !errors.Is(err, checkly.ErrCodeBundleNotFound) {
				t.Errorf("want ErrCodeBundleNotFound for status %d, got %v", status, err)
			}
		})
	}
}
