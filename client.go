package checkly

import (
	"context"
	"io"
	"net/http"
	"time"
)

// DefaultBaseURL is the root of the Checkly public API, which a client uses
// unless WithBaseURL sets another.
const DefaultBaseURL = "https://api.checklyhq.com"

// defaultSource is the x-checkly-source header a client sends unless
// WithSource sets another.
const defaultSource = "go-sdk"

// Client is an interface that implements Checkly's API
type Client interface {
	GetHeartbeatMonitor(
		ctx context.Context,
		ID string,
	) (*HeartbeatMonitor, error)

	// CreateCheck creates a new check with the specified details.
	// It returns the newly-created check, or an error.
	//
	// Deprecated: use CreateAPICheck, CreateBrowserCheck,
	// CreateMultiStepCheck or the create method of the monitor type, which
	// take the fields of their check type.
	CreateCheck(
		ctx context.Context,
		check Check,
	) (*Check, error)

	// CreateHeartbeatMonitor creates a new heartbeat monitor with the
	// specified details. It returns the newly-created monitor, or an error.
	CreateHeartbeatMonitor(
		ctx context.Context,
		monitor HeartbeatMonitor,
	) (*HeartbeatMonitor, error)

	// CreateTCPMonitor creates a new TCP monitor with the specified details.
	// It returns the newly-created monitor, or an error.
	CreateTCPMonitor(
		ctx context.Context,
		monitor TCPMonitor,
	) (*TCPMonitor, error)

	// CreateGRPCMonitor creates a new gRPC monitor with the specified details.
	// It returns the newly-created monitor, or an error.
	CreateGRPCMonitor(
		ctx context.Context,
		monitor GRPCMonitor,
	) (*GRPCMonitor, error)

	// CreateTracerouteMonitor creates a new traceroute monitor with the
	// specified details. It returns the newly-created monitor, or an error.
	CreateTracerouteMonitor(
		ctx context.Context,
		monitor TracerouteMonitor,
	) (*TracerouteMonitor, error)

	// CreateSSLMonitor creates a new SSL monitor with the specified details.
	// It returns the newly-created monitor, or an error.
	CreateSSLMonitor(
		ctx context.Context,
		monitor SSLMonitor,
	) (*SSLMonitor, error)

	// CreateURLMonitor creates a new URL monitor with the specified details.
	// It returns the newly-created monitor, or an error.
	CreateURLMonitor(
		ctx context.Context,
		monitor URLMonitor,
	) (*URLMonitor, error)

	// CreateDNSMonitor creates a new DNS monitor with the specified details.
	// It returns the newly-created monitor, or an error.
	CreateDNSMonitor(
		ctx context.Context,
		monitor DNSMonitor,
	) (*DNSMonitor, error)

	// CreateICMPMonitor creates a new ICMP monitor with the specified details.
	// It returns the newly-created monitor, or an error.
	CreateICMPMonitor(
		ctx context.Context,
		monitor ICMPMonitor,
	) (*ICMPMonitor, error)

	// CreatePlaywrightCheck creates a new Playwright check.
	CreatePlaywrightCheck(
		ctx context.Context,
		check PlaywrightCheck,
	) (*PlaywrightCheck, error)

	// CreateAPICheck creates a new API check with the specified details.
	// It returns the newly-created check, or an error.
	CreateAPICheck(
		ctx context.Context,
		check APICheck,
	) (*APICheck, error)

	// CreateBrowserCheck creates a new browser check with the specified
	// details. It returns the newly-created check, or an error.
	CreateBrowserCheck(
		ctx context.Context,
		check BrowserCheck,
	) (*BrowserCheck, error)

	// CreateMultiStepCheck creates a new multi-step check with the specified
	// details. It returns the newly-created check, or an error.
	CreateMultiStepCheck(
		ctx context.Context,
		check MultiStepCheck,
	) (*MultiStepCheck, error)

	// UpdateCheck updates an existing check with the specified details.
	// It returns the updated check, or an error.
	//
	// Deprecated: use UpdateAPICheck, UpdateBrowserCheck,
	// UpdateMultiStepCheck or the update method of the monitor type.
	// UpdateCheck uses the generic check update
	// endpoint, which the API has deprecated.
	UpdateCheck(
		ctx context.Context,
		ID string,
		check Check,
	) (*Check, error)

	// UpdateHeartbeatMonitor updates an existing heartbeat monitor with the
	// specified details. It returns the updated monitor, or an error.
	UpdateHeartbeatMonitor(
		ctx context.Context,
		ID string,
		monitor HeartbeatMonitor,
	) (*HeartbeatMonitor, error)

	// UpdateTCPMonitor updates an existing TCP monitor with the specified
	// details. It returns the updated monitor, or an error.
	UpdateTCPMonitor(
		ctx context.Context,
		ID string,
		monitor TCPMonitor,
	) (*TCPMonitor, error)

	// UpdateGRPCMonitor updates an existing gRPC monitor with the specified
	// details. It returns the updated monitor, or an error.
	UpdateGRPCMonitor(
		ctx context.Context,
		ID string,
		monitor GRPCMonitor,
	) (*GRPCMonitor, error)

	// UpdateTracerouteMonitor updates an existing traceroute monitor with the
	// specified details. It returns the updated monitor, or an error.
	UpdateTracerouteMonitor(
		ctx context.Context,
		ID string,
		monitor TracerouteMonitor,
	) (*TracerouteMonitor, error)

	// UpdateSSLMonitor updates an existing SSL monitor with the specified
	// details. It returns the updated monitor, or an error.
	UpdateSSLMonitor(
		ctx context.Context,
		ID string,
		monitor SSLMonitor,
	) (*SSLMonitor, error)

	// UpdateURLMonitor updates an existing URL monitor with the specified details.
	// It returns the updated monitor, or an error.
	UpdateURLMonitor(
		ctx context.Context,
		ID string,
		monitor URLMonitor,
	) (*URLMonitor, error)

	// UpdateDNSMonitor updates an existing DNS monitor with the specified details.
	// It returns the updated monitor, or an error.
	UpdateDNSMonitor(
		ctx context.Context,
		ID string,
		monitor DNSMonitor,
	) (*DNSMonitor, error)

	// UpdateICMPMonitor updates an existing ICMP monitor with the specified details.
	// It returns the updated monitor, or an error.
	UpdateICMPMonitor(
		ctx context.Context,
		ID string,
		monitor ICMPMonitor,
	) (*ICMPMonitor, error)

	// UpdatePlaywrightCheck updates an existing Playwright check.
	UpdatePlaywrightCheck(
		ctx context.Context,
		ID string,
		check PlaywrightCheck,
	) (*PlaywrightCheck, error)

	// UpdateAPICheck updates an existing API check with the specified details.
	// It returns the updated check, or an error.
	UpdateAPICheck(
		ctx context.Context,
		ID string,
		check APICheck,
	) (*APICheck, error)

	// UpdateBrowserCheck updates an existing browser check with the specified
	// details. It returns the updated check, or an error.
	UpdateBrowserCheck(
		ctx context.Context,
		ID string,
		check BrowserCheck,
	) (*BrowserCheck, error)

	// UpdateMultiStepCheck updates an existing multi-step check with the
	// specified details. It returns the updated check, or an error.
	UpdateMultiStepCheck(
		ctx context.Context,
		ID string,
		check MultiStepCheck,
	) (*MultiStepCheck, error)

	// DeleteCheck deletes the check with the specified ID.
	//
	// Deprecated: use DeleteAPICheck, DeleteBrowserCheck, DeleteMultiStepCheck
	// or the delete method of the monitor type.
	DeleteCheck(
		ctx context.Context,
		ID string,
	) error

	// DeleteHeartbeatMonitor deletes the monitor with the specified ID.
	DeleteHeartbeatMonitor(
		ctx context.Context,
		ID string,
	) error

	// DeleteTCPMonitor deletes the monitor with the specified ID.
	DeleteTCPMonitor(
		ctx context.Context,
		ID string,
	) error

	// DeleteGRPCMonitor deletes the monitor with the specified ID.
	DeleteGRPCMonitor(
		ctx context.Context,
		ID string,
	) error

	// DeleteTracerouteMonitor deletes the monitor with the specified ID.
	DeleteTracerouteMonitor(
		ctx context.Context,
		ID string,
	) error

	// DeleteSSLMonitor deletes the monitor with the specified ID.
	DeleteSSLMonitor(
		ctx context.Context,
		ID string,
	) error

	// DeleteURLMonitor deletes the monitor with the specified ID.
	DeleteURLMonitor(
		ctx context.Context,
		ID string,
	) error

	// DeleteDNSMonitor deletes the monitor with the specified ID.
	DeleteDNSMonitor(
		ctx context.Context,
		ID string,
	) error

	// DeleteICMPMonitor deletes the monitor with the specified ID.
	DeleteICMPMonitor(
		ctx context.Context,
		ID string,
	) error

	// DeletePlaywrightCheck deletes an existing Playwright check.
	DeletePlaywrightCheck(
		ctx context.Context,
		ID string,
	) error

	// DeleteAPICheck deletes the API check with the specified ID. It uses
	// the generic check endpoint and does not verify the check type.
	DeleteAPICheck(
		ctx context.Context,
		ID string,
	) error

	// DeleteBrowserCheck deletes the browser check with the specified ID. It uses
	// the generic check endpoint and does not verify the check type.
	DeleteBrowserCheck(
		ctx context.Context,
		ID string,
	) error

	// DeleteMultiStepCheck deletes the multi-step check with the specified ID. It uses
	// the generic check endpoint and does not verify the check type.
	DeleteMultiStepCheck(
		ctx context.Context,
		ID string,
	) error

	// GetCheck takes the ID of an existing check, and returns the check
	// parameters, or an error.
	//
	// Deprecated: use GetAPICheck, GetBrowserCheck, GetMultiStepCheck or the
	// get method of the monitor type.
	GetCheck(
		ctx context.Context,
		ID string,
	) (*Check, error)

	// GetTCPMonitor takes the ID of an existing TCP monitor, and returns the
	// monitor parameters, or an error.
	GetTCPMonitor(
		ctx context.Context,
		ID string,
	) (*TCPMonitor, error)

	// GetGRPCMonitor takes the ID of an existing gRPC monitor, and returns the
	// monitor parameters, or an error.
	GetGRPCMonitor(
		ctx context.Context,
		ID string,
	) (*GRPCMonitor, error)

	// GetTracerouteMonitor takes the ID of an existing traceroute monitor,
	// and returns the monitor parameters, or an error.
	GetTracerouteMonitor(
		ctx context.Context,
		ID string,
	) (*TracerouteMonitor, error)

	// GetSSLMonitor takes the ID of an existing SSL monitor, and returns the
	// monitor parameters, or an error.
	GetSSLMonitor(
		ctx context.Context,
		ID string,
	) (*SSLMonitor, error)

	// GetURLMonitor takes the ID of an existing URL monitor, and returns the
	// monitor parameters, or an error.
	GetURLMonitor(
		ctx context.Context,
		ID string,
	) (*URLMonitor, error)

	// GetDNSMonitor takes the ID of an existing DNS monitor, and returns the
	// monitor parameters, or an error.
	GetDNSMonitor(
		ctx context.Context,
		ID string,
	) (*DNSMonitor, error)

	// GetICMPMonitor takes the ID of an existing ICMP monitor, and returns the
	// monitor parameters, or an error.
	GetICMPMonitor(
		ctx context.Context,
		ID string,
	) (*ICMPMonitor, error)

	// GetPlaywrightCheck retrieves an existing Playwright check.
	GetPlaywrightCheck(
		ctx context.Context,
		ID string,
	) (*PlaywrightCheck, error)

	// GetAPICheck takes the ID of an existing API check, and returns the
	// check parameters, or an error.
	GetAPICheck(
		ctx context.Context,
		ID string,
	) (*APICheck, error)

	// GetBrowserCheck takes the ID of an existing browser check, and returns
	// the check parameters, or an error.
	GetBrowserCheck(
		ctx context.Context,
		ID string,
	) (*BrowserCheck, error)

	// GetMultiStepCheck takes the ID of an existing multi-step check, and
	// returns the check parameters, or an error.
	GetMultiStepCheck(
		ctx context.Context,
		ID string,
	) (*MultiStepCheck, error)

	// CreateGroup creates a new check group with the specified details.
	// It returns the newly-created group, or an error.
	CreateGroup(
		ctx context.Context,
		group Group,
	) (*Group, error)

	// CreateGroupV2 creates a new check group with the specified details.
	// It returns the newly-created group, or an error.
	CreateGroupV2(
		ctx context.Context,
		group GroupV2,
	) (*GroupV2, error)

	// GetGroup takes the ID of an existing check group, and returns the
	// corresponding group, or an error.
	GetGroup(
		ctx context.Context,
		ID int64,
	) (*Group, error)

	// GetGroupV2 takes the ID of an existing check group, and returns the
	// corresponding group, or an error.
	GetGroupV2(
		ctx context.Context,
		ID int64,
	) (*GroupV2, error)

	// UpdateGroup takes the ID of an existing check group, and updates the
	// corresponding check group to match the supplied group. It returns the updated
	// group, or an error.
	UpdateGroup(
		ctx context.Context,
		ID int64,
		group Group,
	) (*Group, error)

	// UpdateGroupV2 takes the ID of an existing check group, and updates the
	// corresponding check group to match the supplied group. It returns the updated
	// group, or an error.
	UpdateGroupV2(
		ctx context.Context,
		ID int64,
		group GroupV2,
	) (*GroupV2, error)

	// DeleteGroup deletes the check group with the specified ID. It returns a
	DeleteGroup(
		ctx context.Context,
		ID int64,
	) error

	// DeleteGroupV2 deletes the check group with the specified ID. It returns a
	DeleteGroupV2(
		ctx context.Context,
		ID int64,
	) error

	// GetCheckResult gets a specific Check result, or it returns an error.
	GetCheckResult(
		ctx context.Context,
		checkID,
		checkResultID string,
	) (*CheckResult, error)

	// GetCheckResults gets the results of the given Check
	GetCheckResults(
		ctx context.Context,
		checkID string,
		filters *CheckResultsFilter,
	) ([]CheckResult, error)

	// CreateSnippet creates a new snippet with the specified details. It returns
	// the newly-created snippet, or an error.
	CreateSnippet(
		ctx context.Context,
		snippet Snippet,
	) (*Snippet, error)

	// GetSnippet takes the ID of an existing snippet, and returns the
	// corresponding snippet, or an error.
	GetSnippet(
		ctx context.Context,
		ID int64,
	) (*Snippet, error)

	// UpdateSnippet takes the ID of an existing snippet, and updates the
	// corresponding snippet to match the supplied snippet. It returns the updated
	// snippet, or an error.
	UpdateSnippet(
		ctx context.Context,
		ID int64,
		snippet Snippet,
	) (*Snippet, error)

	// DeleteSnippet deletes the snippet with the specified ID. It returns a
	DeleteSnippet(
		ctx context.Context,
		ID int64,
	) error

	// CreateEnvironmentVariable creates a new environment variable with the
	// specified details.  It returns the newly-created environment variable,
	// or an error.
	CreateEnvironmentVariable(
		ctx context.Context,
		envVar EnvironmentVariable,
	) (*EnvironmentVariable, error)

	// GetEnvironmentVariable takes the ID of an existing environment variable, and returns the
	// corresponding environment variable, or an error.
	GetEnvironmentVariable(
		ctx context.Context,
		key string,
	) (*EnvironmentVariable, error)

	// UpdateEnvironmentVariable takes the ID of an existing environment variable, and updates the
	// corresponding environment variable to match the supplied environment variable. It returns the updated
	// environment variable, or an error.
	UpdateEnvironmentVariable(
		ctx context.Context,
		key string,
		envVar EnvironmentVariable,
	) (*EnvironmentVariable, error)

	// DeleteEnvironmentVariable deletes the environment variable with the specified ID. It returns a
	DeleteEnvironmentVariable(
		ctx context.Context,
		key string,
	) error

	// CreateAlertChannel creates a new alert channel with the specified details. It returns
	// the newly-created alert channel, or an error.
	CreateAlertChannel(
		ctx context.Context,
		ac AlertChannel,
	) (*AlertChannel, error)

	// GetAlertChannel takes the ID of an existing alert channel, and returns the
	// corresponding alert channel, or an error.
	GetAlertChannel(
		ctx context.Context,
		ID int64,
	) (*AlertChannel, error)

	// UpdateAlertChannel takes the ID of an existing alert channel, and updates the
	// corresponding alert channel to match the supplied alert channel. It returns the updated
	// alert channel, or an error.
	UpdateAlertChannel(
		ctx context.Context,
		ID int64,
		ac AlertChannel,
	) (*AlertChannel, error)

	// DeleteAlertChannel deletes the alert channel with the specified ID.
	DeleteAlertChannel(
		ctx context.Context,
		ID int64,
	) error

	// CreateDashboard creates a new dashboard with the specified details.
	CreateDashboard(
		ctx context.Context,
		dashboard Dashboard,
	) (*Dashboard, error)

	// GetDashboard takes the ID of an existing dashboard and returns it
	GetDashboard(
		ctx context.Context,
		ID string,
	) (*Dashboard, error)

	// UpdateDashboard takes the ID of an existing dashboard, and updates the
	// corresponding dashboard to match the supplied dashboard.
	UpdateDashboard(
		ctx context.Context,
		ID string,
		dashboard Dashboard,
	) (*Dashboard, error)

	// DeleteDashboard deletes the dashboard with the specified ID.
	DeleteDashboard(
		ctx context.Context,
		ID string,
	) error

	// CreateMaintenanceWindow creates a new maintenance window with the specified details.
	CreateMaintenanceWindow(
		ctx context.Context,
		mw MaintenanceWindow,
	) (*MaintenanceWindow, error)

	// GetMaintenanceWindow takes the ID of an existing maintenance window and returns it
	GetMaintenanceWindow(
		ctx context.Context,
		ID int64,
	) (*MaintenanceWindow, error)

	// UpdateMaintenanceWindow takes the ID of an existing maintenance window, and updates the
	// corresponding maintenance window to match the supplied maintenance window.
	UpdateMaintenanceWindow(
		ctx context.Context,
		ID int64,
		mw MaintenanceWindow,
	) (*MaintenanceWindow, error)

	// DeleteMaintenanceWindow deletes the maintenance window with the specified ID.
	DeleteMaintenanceWindow(
		ctx context.Context,
		ID int64,
	) error

	// CreatePrivateLocation creates a new private location with the specified details.
	CreatePrivateLocation(
		ctx context.Context,
		pl PrivateLocation,
	) (*PrivateLocation, error)

	// GetPrivateLocation takes the ID of an existing private location and returns it
	GetPrivateLocation(
		ctx context.Context,
		ID string,
	) (*PrivateLocation, error)

	// UpdatePrivateLocation takes the ID of an existing private location and updates it
	// to match the new one.
	UpdatePrivateLocation(
		ctx context.Context,
		ID string,
		pl PrivateLocation,
	) (*PrivateLocation, error)

	// DeletePrivateLocation deletes the private location with the specified ID.
	DeletePrivateLocation(
		ctx context.Context,
		ID string,
	) error

	// CreateClientCertificate creates a new client certificate and returns
	// the created resource.
	CreateClientCertificate(
		ctx context.Context,
		cs ClientCertificate,
	) (*ClientCertificate, error)

	// GetClientCertificate retrieves a client certificate.
	GetClientCertificate(
		ctx context.Context,
		ID string,
	) (*ClientCertificate, error)

	// DeleteClientCertificate deletes a client certificate.
	DeleteClientCertificate(
		ctx context.Context,
		ID string,
	) error

	// CreateStatusPage creates a new status page and returns the created
	// resource.
	CreateStatusPage(
		ctx context.Context,
		page StatusPage,
	) (*StatusPage, error)

	// GetStatusPage retrieves a status page.
	GetStatusPage(
		ctx context.Context,
		ID string,
	) (*StatusPage, error)

	// UpdateStatusPage updates a status page.
	UpdateStatusPage(
		ctx context.Context,
		ID string,
		page StatusPage,
	) (*StatusPage, error)

	// DeleteStatusPage deletes a status page.
	DeleteStatusPage(
		ctx context.Context,
		ID string,
	) error

	// CreateStatusPageService creates a new status page service and returns
	// the created resource.
	CreateStatusPageService(
		ctx context.Context,
		service StatusPageService,
	) (*StatusPageService, error)

	// GetStatusPageService retrieves a status page service.
	GetStatusPageService(
		ctx context.Context,
		ID string,
	) (*StatusPageService, error)

	// UpdateStatusPageService updates a status page service.
	UpdateStatusPageService(
		ctx context.Context,
		ID string,
		service StatusPageService,
	) (*StatusPageService, error)

	// DeleteStatusPageService deletes a status page service.
	DeleteStatusPageService(
		ctx context.Context,
		ID string,
	) error

	// CreateStatusPageV3 creates a new v3 status page and returns the
	// created resource.
	CreateStatusPageV3(
		ctx context.Context,
		page StatusPageV3,
	) (*StatusPageV3, error)

	// GetStatusPageV3 retrieves a v3 status page.
	GetStatusPageV3(
		ctx context.Context,
		ID string,
	) (*StatusPageV3, error)

	// UpdateStatusPageV3 updates a v3 status page.
	UpdateStatusPageV3(
		ctx context.Context,
		ID string,
		page StatusPageV3,
	) (*StatusPageV3, error)

	// DeleteStatusPageV3 deletes a v3 status page.
	DeleteStatusPageV3(
		ctx context.Context,
		ID string,
	) error

	// CreateStatusPageComponentV3 creates a new component on a v3 status
	// page and returns the created resource.
	CreateStatusPageComponentV3(
		ctx context.Context,
		statusPageID string,
		component StatusPageComponentV3,
	) (*StatusPageComponentV3, error)

	// GetStatusPageComponentV3 retrieves a component of a v3 status page.
	GetStatusPageComponentV3(
		ctx context.Context,
		statusPageID string,
		ID string,
	) (*StatusPageComponentV3, error)

	// UpdateStatusPageComponentV3 updates a component of a v3 status page.
	UpdateStatusPageComponentV3(
		ctx context.Context,
		statusPageID string,
		ID string,
		component StatusPageComponentV3,
	) (*StatusPageComponentV3, error)

	// DeleteStatusPageComponentV3 deletes a component of a v3 status page.
	// Child components are detached from a deleted GROUP, not deleted.
	DeleteStatusPageComponentV3(
		ctx context.Context,
		statusPageID string,
		ID string,
	) error

	// CreateStatusPageAutomationRuleV3 creates a new automation rule on a
	// v3 status page and returns the created resource.
	CreateStatusPageAutomationRuleV3(
		ctx context.Context,
		statusPageID string,
		rule StatusPageAutomationRuleV3,
	) (*StatusPageAutomationRuleV3, error)

	// GetStatusPageAutomationRuleV3 retrieves an automation rule of a v3
	// status page.
	GetStatusPageAutomationRuleV3(
		ctx context.Context,
		statusPageID string,
		ID string,
	) (*StatusPageAutomationRuleV3, error)

	// UpdateStatusPageAutomationRuleV3 updates an automation rule of a v3
	// status page.
	UpdateStatusPageAutomationRuleV3(
		ctx context.Context,
		statusPageID string,
		ID string,
		rule StatusPageAutomationRuleV3,
	) (*StatusPageAutomationRuleV3, error)

	// DeleteStatusPageAutomationRuleV3 deletes an automation rule of a v3
	// status page.
	DeleteStatusPageAutomationRuleV3(
		ctx context.Context,
		statusPageID string,
		ID string,
	) error

	// GetRuntime returns the runtime with the specified ID, or an error.
	GetRuntime(
		ctx context.Context,
		ID string,
	) (*Runtime, error)

	GetStaticIPs(ctx context.Context) ([]StaticIP, error)

	// UploadCodeBundle uploads a code bundle which can be used with
	// Playwright checks.
	UploadCodeBundle(
		ctx context.Context,
		data io.Reader,
		size int64,
		options UploadCodeBundleOptions,
	) (*CodeBundle, error)

	// PeekCodeBundle peeks the associated metadata of a code bundle without
	// downloading the code bundle.
	PeekCodeBundle(
		ctx context.Context,
		key string,
	) (*CodeBundle, error)
}

// client implements Client. NewClient builds it from ClientOptions, and it is
// not modified afterwards, so it is safe for concurrent use.
type client struct {
	apiKey string
	// url is the API root, such as DefaultBaseURL, to which request paths
	// starting with "/" are appended.
	url       string
	accountID string
	// source is sent as the x-checkly-source header of every request.
	source     string
	httpClient *http.Client
	// sleep waits between rate-limit retries, returning early with ctx.Err()
	// when ctx is done. Tests replace it to avoid real delays.
	sleep func(context.Context, time.Duration) error
}

// NewClient constructs a Checkly API client. Without options it talks to
// DefaultBaseURL through http.DefaultClient and has no auth: the public
// methods GetStaticIPs and GetRuntime work, and every other method fails with
// ErrMissingAPIKey without sending a request. Nil options are skipped.
//
// NewClient returns a nil Client and the error of the first option with an
// invalid value, which starts with "checkly: " and names the option: a base
// URL WithBaseURL rejects, or auth WithAuth or WithAuthFromEnv rejects, an
// empty or unset API key (ErrMissingAPIKey) or account ID
// (ErrMissingAccountID), or an account ID that is not a UUID
// (ErrInvalidAccountID). The client is not modified after construction and
// is safe for concurrent use.
func NewClient(opts ...ClientOption) (Client, error) {
	c := &client{
		url:        DefaultBaseURL,
		source:     defaultSource,
		httpClient: http.DefaultClient,
		sleep:      sleepContext,
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}
