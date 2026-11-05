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

// Client is a client for the Checkly public API, with methods for the
// resources of a Checkly account. [NewClient] returns one; it is not
// modified afterwards, so it is safe for concurrent use. The package
// documentation describes the errors, cancellation, retries and update
// behaviour that all methods share.
type Client interface {
	// CreateCheck creates a check from the generic [Check] fields and returns
	// it with its ID. Type must be [CheckTypeAPI], [CheckTypeBrowser],
	// [CheckTypeHeartbeat] or [CheckTypeMultiStep]; for any other type it
	// returns an error without sending a request.
	//
	// Deprecated: use CreateAPICheck, CreateBrowserCheck,
	// CreateMultiStepCheck or the create method of the monitor type, which
	// take the fields of their check type.
	CreateCheck(
		ctx context.Context,
		check Check,
	) (*Check, error)

	// GetCheck returns the check with the given ID, whatever its type, as a
	// [Check].
	//
	// Deprecated: use GetAPICheck, GetBrowserCheck, GetMultiStepCheck or the
	// get method of the monitor type.
	GetCheck(
		ctx context.Context,
		ID string,
	) (*Check, error)

	// UpdateCheck updates the check with the given ID and returns it.
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

	// DeleteCheck deletes the check with the given ID, whatever its type.
	//
	// Deprecated: use DeleteAPICheck, DeleteBrowserCheck, DeleteMultiStepCheck
	// or the delete method of the monitor type.
	DeleteCheck(
		ctx context.Context,
		ID string,
	) error

	// CreateAPICheck creates an API check and returns it with its ID.
	CreateAPICheck(
		ctx context.Context,
		check APICheck,
	) (*APICheck, error)

	// GetAPICheck returns the API check with the given ID. It returns an error,
	// not an [*APIError], when the ID belongs to a check of another type.
	GetAPICheck(
		ctx context.Context,
		ID string,
	) (*APICheck, error)

	// UpdateAPICheck updates the API check with the given ID and returns it.
	UpdateAPICheck(
		ctx context.Context,
		ID string,
		check APICheck,
	) (*APICheck, error)

	// DeleteAPICheck deletes the API check with the given ID. It uses the
	// generic check endpoint and does not verify the check type.
	DeleteAPICheck(
		ctx context.Context,
		ID string,
	) error

	// CreateBrowserCheck creates a browser check and returns it with its ID.
	CreateBrowserCheck(
		ctx context.Context,
		check BrowserCheck,
	) (*BrowserCheck, error)

	// GetBrowserCheck returns the browser check with the given ID, including its
	// script dependencies. It returns an error, not an [*APIError], when the ID
	// belongs to a check of another type.
	GetBrowserCheck(
		ctx context.Context,
		ID string,
	) (*BrowserCheck, error)

	// UpdateBrowserCheck updates the browser check with the given ID and returns
	// it.
	UpdateBrowserCheck(
		ctx context.Context,
		ID string,
		check BrowserCheck,
	) (*BrowserCheck, error)

	// DeleteBrowserCheck deletes the browser check with the given ID. It uses
	// the generic check endpoint and does not verify the check type.
	DeleteBrowserCheck(
		ctx context.Context,
		ID string,
	) error

	// CreateMultiStepCheck creates a multi-step check and returns it with its ID.
	CreateMultiStepCheck(
		ctx context.Context,
		check MultiStepCheck,
	) (*MultiStepCheck, error)

	// GetMultiStepCheck returns the multi-step check with the given ID,
	// including its script dependencies. It returns an error, not an
	// [*APIError], when the ID belongs to a check of another type.
	GetMultiStepCheck(
		ctx context.Context,
		ID string,
	) (*MultiStepCheck, error)

	// UpdateMultiStepCheck updates the multi-step check with the given ID and
	// returns it.
	UpdateMultiStepCheck(
		ctx context.Context,
		ID string,
		check MultiStepCheck,
	) (*MultiStepCheck, error)

	// DeleteMultiStepCheck deletes the multi-step check with the given ID. It
	// uses the generic check endpoint and does not verify the check type.
	DeleteMultiStepCheck(
		ctx context.Context,
		ID string,
	) error

	// CreatePlaywrightCheck creates a Playwright check and returns it with its
	// ID.
	CreatePlaywrightCheck(
		ctx context.Context,
		check PlaywrightCheck,
	) (*PlaywrightCheck, error)

	// GetPlaywrightCheck returns the Playwright check with the given ID. It does
	// not verify the check type.
	GetPlaywrightCheck(
		ctx context.Context,
		ID string,
	) (*PlaywrightCheck, error)

	// UpdatePlaywrightCheck updates the Playwright check with the given ID and
	// returns it. An empty Tags, a nil AlertSettings, or a zero GroupID or
	// GroupOrder, which are left out of the request, resets that field to the
	// API default.
	UpdatePlaywrightCheck(
		ctx context.Context,
		ID string,
		check PlaywrightCheck,
	) (*PlaywrightCheck, error)

	// DeletePlaywrightCheck deletes the Playwright check with the given ID. It
	// uses the generic check endpoint and does not verify the check type.
	DeletePlaywrightCheck(
		ctx context.Context,
		ID string,
	) error

	// UploadCodeBundle uploads size bytes read from data as a code bundle for
	// Playwright checks and returns the stored bundle's metadata.
	UploadCodeBundle(
		ctx context.Context,
		data io.Reader,
		size int64,
		options UploadCodeBundleOptions,
	) (*CodeBundle, error)

	// PeekCodeBundle returns the metadata of the code bundle stored under key,
	// such as its size and checksum, without downloading it. When there is no
	// bundle for the key, including when the key belongs to another account,
	// the error matches [ErrCodeBundleNotFound].
	PeekCodeBundle(
		ctx context.Context,
		key string,
	) (*CodeBundle, error)

	// CreateHeartbeatMonitor creates a heartbeat monitor and returns it with its
	// ID.
	CreateHeartbeatMonitor(
		ctx context.Context,
		monitor HeartbeatMonitor,
	) (*HeartbeatMonitor, error)

	// GetHeartbeatMonitor returns the heartbeat monitor with the given ID. It
	// does not verify the check type.
	GetHeartbeatMonitor(
		ctx context.Context,
		ID string,
	) (*HeartbeatMonitor, error)

	// UpdateHeartbeatMonitor updates the heartbeat monitor with the given ID and
	// returns it.
	UpdateHeartbeatMonitor(
		ctx context.Context,
		ID string,
		monitor HeartbeatMonitor,
	) (*HeartbeatMonitor, error)

	// DeleteHeartbeatMonitor deletes the heartbeat monitor with the given ID. It
	// uses the generic check endpoint and does not verify the check type.
	DeleteHeartbeatMonitor(
		ctx context.Context,
		ID string,
	) error

	// CreateURLMonitor creates a URL monitor and returns it with its ID.
	CreateURLMonitor(
		ctx context.Context,
		monitor URLMonitor,
	) (*URLMonitor, error)

	// GetURLMonitor returns the URL monitor with the given ID. It does not
	// verify the check type.
	GetURLMonitor(
		ctx context.Context,
		ID string,
	) (*URLMonitor, error)

	// UpdateURLMonitor updates the URL monitor with the given ID and returns it.
	UpdateURLMonitor(
		ctx context.Context,
		ID string,
		monitor URLMonitor,
	) (*URLMonitor, error)

	// DeleteURLMonitor deletes the URL monitor with the given ID. It uses the
	// generic check endpoint and does not verify the check type.
	DeleteURLMonitor(
		ctx context.Context,
		ID string,
	) error

	// CreateTCPMonitor creates a TCP monitor and returns it with its ID.
	CreateTCPMonitor(
		ctx context.Context,
		monitor TCPMonitor,
	) (*TCPMonitor, error)

	// GetTCPMonitor returns the TCP monitor with the given ID. It does not
	// verify the check type.
	GetTCPMonitor(
		ctx context.Context,
		ID string,
	) (*TCPMonitor, error)

	// UpdateTCPMonitor updates the TCP monitor with the given ID and returns it.
	UpdateTCPMonitor(
		ctx context.Context,
		ID string,
		monitor TCPMonitor,
	) (*TCPMonitor, error)

	// DeleteTCPMonitor deletes the TCP monitor with the given ID. It uses the
	// generic check endpoint and does not verify the check type.
	DeleteTCPMonitor(
		ctx context.Context,
		ID string,
	) error

	// CreateDNSMonitor creates a DNS monitor and returns it with its ID.
	CreateDNSMonitor(
		ctx context.Context,
		monitor DNSMonitor,
	) (*DNSMonitor, error)

	// GetDNSMonitor returns the DNS monitor with the given ID. It does not
	// verify the check type.
	GetDNSMonitor(
		ctx context.Context,
		ID string,
	) (*DNSMonitor, error)

	// UpdateDNSMonitor updates the DNS monitor with the given ID and returns it.
	// A nil AlertSettings, which is left out of the request, resets the alert
	// settings to the API default.
	UpdateDNSMonitor(
		ctx context.Context,
		ID string,
		monitor DNSMonitor,
	) (*DNSMonitor, error)

	// DeleteDNSMonitor deletes the DNS monitor with the given ID. It uses the
	// generic check endpoint and does not verify the check type.
	DeleteDNSMonitor(
		ctx context.Context,
		ID string,
	) error

	// CreateICMPMonitor creates an ICMP monitor and returns it with its ID.
	CreateICMPMonitor(
		ctx context.Context,
		monitor ICMPMonitor,
	) (*ICMPMonitor, error)

	// GetICMPMonitor returns the ICMP monitor with the given ID. It does not
	// verify the check type.
	GetICMPMonitor(
		ctx context.Context,
		ID string,
	) (*ICMPMonitor, error)

	// UpdateICMPMonitor updates the ICMP monitor with the given ID and returns
	// it.
	UpdateICMPMonitor(
		ctx context.Context,
		ID string,
		monitor ICMPMonitor,
	) (*ICMPMonitor, error)

	// DeleteICMPMonitor deletes the ICMP monitor with the given ID. It uses the
	// generic check endpoint and does not verify the check type.
	DeleteICMPMonitor(
		ctx context.Context,
		ID string,
	) error

	// CreateGRPCMonitor creates a gRPC monitor and returns it with its ID.
	CreateGRPCMonitor(
		ctx context.Context,
		monitor GRPCMonitor,
	) (*GRPCMonitor, error)

	// GetGRPCMonitor returns the gRPC monitor with the given ID. It does not
	// verify the check type.
	GetGRPCMonitor(
		ctx context.Context,
		ID string,
	) (*GRPCMonitor, error)

	// UpdateGRPCMonitor updates the gRPC monitor with the given ID and returns
	// it.
	UpdateGRPCMonitor(
		ctx context.Context,
		ID string,
		monitor GRPCMonitor,
	) (*GRPCMonitor, error)

	// DeleteGRPCMonitor deletes the gRPC monitor with the given ID. It uses the
	// generic check endpoint and does not verify the check type.
	DeleteGRPCMonitor(
		ctx context.Context,
		ID string,
	) error

	// CreateTracerouteMonitor creates a traceroute monitor and returns it with
	// its ID.
	CreateTracerouteMonitor(
		ctx context.Context,
		monitor TracerouteMonitor,
	) (*TracerouteMonitor, error)

	// GetTracerouteMonitor returns the traceroute monitor with the given ID. It
	// does not verify the check type.
	GetTracerouteMonitor(
		ctx context.Context,
		ID string,
	) (*TracerouteMonitor, error)

	// UpdateTracerouteMonitor updates the traceroute monitor with the given ID
	// and returns it.
	UpdateTracerouteMonitor(
		ctx context.Context,
		ID string,
		monitor TracerouteMonitor,
	) (*TracerouteMonitor, error)

	// DeleteTracerouteMonitor deletes the traceroute monitor with the given ID.
	// It uses the generic check endpoint and does not verify the check type.
	DeleteTracerouteMonitor(
		ctx context.Context,
		ID string,
	) error

	// CreateSSLMonitor creates an SSL monitor and returns it with its ID.
	CreateSSLMonitor(
		ctx context.Context,
		monitor SSLMonitor,
	) (*SSLMonitor, error)

	// GetSSLMonitor returns the SSL monitor with the given ID. It does not
	// verify the check type.
	GetSSLMonitor(
		ctx context.Context,
		ID string,
	) (*SSLMonitor, error)

	// UpdateSSLMonitor updates the SSL monitor with the given ID and returns it.
	UpdateSSLMonitor(
		ctx context.Context,
		ID string,
		monitor SSLMonitor,
	) (*SSLMonitor, error)

	// DeleteSSLMonitor deletes the SSL monitor with the given ID. It uses the
	// generic check endpoint and does not verify the check type.
	DeleteSSLMonitor(
		ctx context.Context,
		ID string,
	) error

	// CreateGroup creates a check group and returns it with its ID.
	CreateGroup(
		ctx context.Context,
		group Group,
	) (*Group, error)

	// GetGroup returns the check group with the given ID.
	GetGroup(
		ctx context.Context,
		ID int64,
	) (*Group, error)

	// UpdateGroup updates the check group with the given ID and returns it.
	UpdateGroup(
		ctx context.Context,
		ID int64,
		group Group,
	) (*Group, error)

	// DeleteGroup deletes the check group with the given ID.
	DeleteGroup(
		ctx context.Context,
		ID int64,
	) error

	// CreateGroupV2 creates a check group through the v2 check group endpoint
	// and returns it with its ID.
	CreateGroupV2(
		ctx context.Context,
		group GroupV2,
	) (*GroupV2, error)

	// GetGroupV2 returns the check group with the given ID as a [GroupV2].
	GetGroupV2(
		ctx context.Context,
		ID int64,
	) (*GroupV2, error)

	// UpdateGroupV2 updates the check group with the given ID through the v2
	// check group endpoint and returns it.
	UpdateGroupV2(
		ctx context.Context,
		ID int64,
		group GroupV2,
	) (*GroupV2, error)

	// DeleteGroupV2 deletes the check group with the given ID, like
	// DeleteGroup.
	DeleteGroupV2(
		ctx context.Context,
		ID int64,
	) error

	// GetCheckResult returns the result with ID checkResultID of the check with
	// ID checkID.
	GetCheckResult(
		ctx context.Context,
		checkID,
		checkResultID string,
	) (*CheckResult, error)

	// GetCheckResults returns one page of the final results of the check with ID
	// checkID, narrowed by filters, which may be nil; [CheckResultsFilter] gives
	// the defaults and limits. It uses the v1 check results endpoint, which the
	// API has deprecated.
	GetCheckResults(
		ctx context.Context,
		checkID string,
		filters *CheckResultsFilter,
	) ([]CheckResult, error)

	// CreateSnippet creates a snippet and returns it with its ID.
	CreateSnippet(
		ctx context.Context,
		snippet Snippet,
	) (*Snippet, error)

	// GetSnippet returns the snippet with the given ID.
	GetSnippet(
		ctx context.Context,
		ID int64,
	) (*Snippet, error)

	// UpdateSnippet replaces the name and script of the snippet with the given ID
	// and returns it.
	UpdateSnippet(
		ctx context.Context,
		ID int64,
		snippet Snippet,
	) (*Snippet, error)

	// DeleteSnippet deletes the snippet with the given ID.
	DeleteSnippet(
		ctx context.Context,
		ID int64,
	) error

	// CreateEnvironmentVariable creates an environment variable and returns it.
	CreateEnvironmentVariable(
		ctx context.Context,
		envVar EnvironmentVariable,
	) (*EnvironmentVariable, error)

	// GetEnvironmentVariable returns the environment variable with the given
	// key.
	GetEnvironmentVariable(
		ctx context.Context,
		key string,
	) (*EnvironmentVariable, error)

	// UpdateEnvironmentVariable updates the environment variable with the given
	// key and returns it. Value must be set, because the API rejects an empty
	// value and a Get of a secret variable returns none. The API rejects a
	// change of the key and turning a secret variable into a non-secret one.
	UpdateEnvironmentVariable(
		ctx context.Context,
		key string,
		envVar EnvironmentVariable,
	) (*EnvironmentVariable, error)

	// DeleteEnvironmentVariable deletes the environment variable with the given
	// key.
	DeleteEnvironmentVariable(
		ctx context.Context,
		key string,
	) error

	// CreateAlertChannel creates an alert channel and returns it with its ID.
	CreateAlertChannel(
		ctx context.Context,
		ac AlertChannel,
	) (*AlertChannel, error)

	// GetAlertChannel returns the alert channel with the given ID.
	GetAlertChannel(
		ctx context.Context,
		ID int64,
	) (*AlertChannel, error)

	// UpdateAlertChannel updates the alert channel with the given ID and returns
	// it. A nil SSLExpiry or SSLExpiryThreshold resets it to the API default,
	// while a nil SendRecovery, SendFailure or SendDegraded keeps its current
	// value.
	UpdateAlertChannel(
		ctx context.Context,
		ID int64,
		ac AlertChannel,
	) (*AlertChannel, error)

	// DeleteAlertChannel deletes the alert channel with the given ID.
	DeleteAlertChannel(
		ctx context.Context,
		ID int64,
	) error

	// CreateDashboard creates a dashboard and returns it with its IDs.
	CreateDashboard(
		ctx context.Context,
		dashboard Dashboard,
	) (*Dashboard, error)

	// GetDashboard returns the dashboard whose [Dashboard.DashboardID] is ID.
	GetDashboard(
		ctx context.Context,
		ID string,
	) (*Dashboard, error)

	// UpdateDashboard updates the dashboard whose [Dashboard.DashboardID] is ID
	// and returns it. A nil ShowHeader, ShowP95 or ShowP99, a false IsPrivate,
	// HideTags or UseTagsAndOperator, an empty Width or Tags, or a zero
	// ChecksPerPage resets that field to the API default, while an empty
	// Header, Description, Logo, Link, Favicon or CustomDomain keeps its
	// current value.
	UpdateDashboard(
		ctx context.Context,
		ID string,
		dashboard Dashboard,
	) (*Dashboard, error)

	// DeleteDashboard deletes the dashboard whose [Dashboard.DashboardID] is ID.
	DeleteDashboard(
		ctx context.Context,
		ID string,
	) error

	// CreateMaintenanceWindow creates a maintenance window and returns it with
	// its ID.
	CreateMaintenanceWindow(
		ctx context.Context,
		mw MaintenanceWindow,
	) (*MaintenanceWindow, error)

	// GetMaintenanceWindow returns the maintenance window with the given ID.
	GetMaintenanceWindow(
		ctx context.Context,
		ID int64,
	) (*MaintenanceWindow, error)

	// UpdateMaintenanceWindow updates the maintenance window with the given ID
	// and returns it.
	UpdateMaintenanceWindow(
		ctx context.Context,
		ID int64,
		mw MaintenanceWindow,
	) (*MaintenanceWindow, error)

	// DeleteMaintenanceWindow deletes the maintenance window with the given ID.
	DeleteMaintenanceWindow(
		ctx context.Context,
		ID int64,
	) error

	// CreatePrivateLocation creates a private location and returns it with its
	// ID.
	CreatePrivateLocation(
		ctx context.Context,
		pl PrivateLocation,
	) (*PrivateLocation, error)

	// GetPrivateLocation returns the private location with the given ID.
	GetPrivateLocation(
		ctx context.Context,
		ID string,
	) (*PrivateLocation, error)

	// UpdatePrivateLocation updates the private location with the given ID and
	// returns it. The slug name cannot be changed.
	UpdatePrivateLocation(
		ctx context.Context,
		ID string,
		pl PrivateLocation,
	) (*PrivateLocation, error)

	// DeletePrivateLocation deletes the private location with the given ID.
	DeletePrivateLocation(
		ctx context.Context,
		ID string,
	) error

	// CreateClientCertificate creates a client certificate and returns it with
	// its ID.
	CreateClientCertificate(
		ctx context.Context,
		cs ClientCertificate,
	) (*ClientCertificate, error)

	// GetClientCertificate returns the client certificate with the given ID.
	GetClientCertificate(
		ctx context.Context,
		ID string,
	) (*ClientCertificate, error)

	// DeleteClientCertificate deletes the client certificate with the given ID.
	DeleteClientCertificate(
		ctx context.Context,
		ID string,
	) error

	// CreateStatusPage creates a status page and returns it with its ID.
	CreateStatusPage(
		ctx context.Context,
		page StatusPage,
	) (*StatusPage, error)

	// GetStatusPage returns the status page with the given ID.
	GetStatusPage(
		ctx context.Context,
		ID string,
	) (*StatusPage, error)

	// UpdateStatusPage updates the status page with the given ID and returns
	// it. An empty DefaultTheme, which is left out of the request, resets the
	// theme to [StatusPageThemeAuto].
	UpdateStatusPage(
		ctx context.Context,
		ID string,
		page StatusPage,
	) (*StatusPage, error)

	// DeleteStatusPage deletes the status page with the given ID.
	DeleteStatusPage(
		ctx context.Context,
		ID string,
	) error

	// CreateStatusPageService creates a status page service and returns it with
	// its ID.
	CreateStatusPageService(
		ctx context.Context,
		service StatusPageService,
	) (*StatusPageService, error)

	// GetStatusPageService returns the status page service with the given ID.
	GetStatusPageService(
		ctx context.Context,
		ID string,
	) (*StatusPageService, error)

	// UpdateStatusPageService renames the status page service with the given ID
	// and returns it.
	UpdateStatusPageService(
		ctx context.Context,
		ID string,
		service StatusPageService,
	) (*StatusPageService, error)

	// DeleteStatusPageService deletes the status page service with the given
	// ID.
	DeleteStatusPageService(
		ctx context.Context,
		ID string,
	) error

	// CreateStatusPageV3 creates a v3 status page and returns it with its ID.
	CreateStatusPageV3(
		ctx context.Context,
		page StatusPageV3,
	) (*StatusPageV3, error)

	// GetStatusPageV3 returns the v3 status page with the given ID.
	GetStatusPageV3(
		ctx context.Context,
		ID string,
	) (*StatusPageV3, error)

	// UpdateStatusPageV3 updates the v3 status page with the given ID and
	// returns it. An empty DefaultTheme, which is left out of the request,
	// resets the theme to [StatusPageThemeAuto].
	UpdateStatusPageV3(
		ctx context.Context,
		ID string,
		page StatusPageV3,
	) (*StatusPageV3, error)

	// DeleteStatusPageV3 deletes the v3 status page with the given ID.
	DeleteStatusPageV3(
		ctx context.Context,
		ID string,
	) error

	// CreateStatusPageComponentV3 adds a component to the v3 status page with ID
	// statusPageID and returns it with its ID.
	CreateStatusPageComponentV3(
		ctx context.Context,
		statusPageID string,
		component StatusPageComponentV3,
	) (*StatusPageComponentV3, error)

	// GetStatusPageComponentV3 returns the component with the given ID of the v3
	// status page with ID statusPageID.
	GetStatusPageComponentV3(
		ctx context.Context,
		statusPageID string,
		ID string,
	) (*StatusPageComponentV3, error)

	// UpdateStatusPageComponentV3 updates the component with the given ID of
	// the v3 status page with ID statusPageID and returns it. An empty Type
	// resets the type to [StatusPageComponentV3TypeService], a nil
	// Configuration resets the configuration to the type's default, and an
	// empty Description or ParentID removes it.
	UpdateStatusPageComponentV3(
		ctx context.Context,
		statusPageID string,
		ID string,
		component StatusPageComponentV3,
	) (*StatusPageComponentV3, error)

	// DeleteStatusPageComponentV3 deletes the component with the given ID of
	// the v3 status page with ID statusPageID. Child components are detached
	// from a deleted [StatusPageComponentV3TypeGroup], not deleted.
	DeleteStatusPageComponentV3(
		ctx context.Context,
		statusPageID string,
		ID string,
	) error

	// CreateStatusPageAutomationRuleV3 adds an automation rule to the v3 status
	// page with ID statusPageID and returns it with its ID.
	CreateStatusPageAutomationRuleV3(
		ctx context.Context,
		statusPageID string,
		rule StatusPageAutomationRuleV3,
	) (*StatusPageAutomationRuleV3, error)

	// GetStatusPageAutomationRuleV3 returns the automation rule with the given
	// ID of the v3 status page with ID statusPageID.
	GetStatusPageAutomationRuleV3(
		ctx context.Context,
		statusPageID string,
		ID string,
	) (*StatusPageAutomationRuleV3, error)

	// UpdateStatusPageAutomationRuleV3 updates the automation rule with the
	// given ID of the v3 status page with ID statusPageID and returns it. An
	// empty Components removes the rule's components.
	UpdateStatusPageAutomationRuleV3(
		ctx context.Context,
		statusPageID string,
		ID string,
		rule StatusPageAutomationRuleV3,
	) (*StatusPageAutomationRuleV3, error)

	// DeleteStatusPageAutomationRuleV3 deletes the automation rule with the
	// given ID of the v3 status page with ID statusPageID.
	DeleteStatusPageAutomationRuleV3(
		ctx context.Context,
		statusPageID string,
		ID string,
	) error

	// GetRuntime returns the runtime with the given ID, such as "2025.04". The
	// endpoint is public, so the request is sent without credentials.
	GetRuntime(
		ctx context.Context,
		ID string,
	) (*Runtime, error)

	// GetStaticIPs returns the IP ranges that checks run from, per region: IPv6
	// prefixes, and IPv4 addresses as /32 prefixes. The endpoints are public, so
	// the requests are sent without credentials.
	GetStaticIPs(ctx context.Context) ([]StaticIP, error)
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
// (ErrInvalidAccountID).
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
