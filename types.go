package checkly

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/netip"
	"time"
)

// Client is an interface that implements Checkly's API
type Client interface {
	// Create creates a new check with the specified details.
	// It returns the newly-created check, or an error.
	//
	// Deprecated: method type would be removed in future versions,
	// use CreateCheck instead.
	Create(
		ctx context.Context,
		check Check,
	) (*Check, error)

	// Update updates an existing check with the specified details.
	// It returns the updated check, or an error.
	//
	// Deprecated: this method would be removed in future versions,
	// use UpdateCheck instead.
	Update(
		ctx context.Context,
		ID string,
		check Check,
	) (*Check, error)

	// Delete deletes the check with the specified ID.
	//
	// Deprecated: this method would be removed in future versions,
	// use DeleteCheck instead.
	Delete(
		ctx context.Context,
		ID string,
	) error

	// Get takes the ID of an existing check, and returns the check parameters,
	// or an error.
	//
	// Deprecated: this method would be removed in future versions,
	// use GetCheck instead.
	Get(
		ctx context.Context,
		ID string,
	) (*Check, error)

	// Deprecated: Use GetHeartbeatMonitor instead.
	GetHeartbeatCheck(
		ctx context.Context,
		ID string,
	) (*HeartbeatCheck, error)

	GetHeartbeatMonitor(
		ctx context.Context,
		ID string,
	) (*HeartbeatMonitor, error)

	// Create creates a new check with the specified details.
	// It returns the newly-created check, or an error.
	CreateCheck(
		ctx context.Context,
		check Check,
	) (*Check, error)

	// CreateHeartbeat creates a new heartbeat check with the specified details.
	// It returns the newly-created check, or an error.
	//
	// Deprecated: Use CreateHeartbeatMonitor instead.
	CreateHeartbeat(
		ctx context.Context,
		check HeartbeatCheck,
	) (*HeartbeatCheck, error)

	// CreateHeartbeatMonitor creates a new heartbeat monitor with the
	// specified details. It returns the newly-created monitor, or an error.
	CreateHeartbeatMonitor(
		ctx context.Context,
		monitor HeartbeatMonitor,
	) (*HeartbeatMonitor, error)

	// CreateTCPCheck creates a new TCP check with the specified details.
	// It returns the newly-created check, or an error.
	//
	// Deprecated: Use CreateTCPMonitor instead.
	CreateTCPCheck(
		ctx context.Context,
		check TCPCheck,
	) (*TCPCheck, error)

	// CreateTCPMonitor creates a new TCP monitor with the specified details.
	// It returns the newly-created monitor, or an error.
	CreateTCPMonitor(
		ctx context.Context,
		monitor TCPMonitor,
	) (*TCPMonitor, error)

	// CreateURLMonitor creates a new URL monitor with the specified details.
	// It returns the newly-created monitor, or an error.
	CreateURLMonitor(
		ctx context.Context,
		monitor URLMonitor,
	) (*URLMonitor, error)

	// Update updates an existing check with the specified details.
	// It returns the updated check, or an error.
	UpdateCheck(
		ctx context.Context,
		ID string,
		check Check,
	) (*Check, error)

	// UpdateHeartbeat updates an existing heartbeat check with the specified details.
	// It returns the updated check, or an error.
	//
	// Deprecated: Use UpdateHeartbeatMonitor instead.
	UpdateHeartbeat(
		ctx context.Context,
		ID string,
		check HeartbeatCheck,
	) (*HeartbeatCheck, error)

	// UpdateHeartbeatMonitor updates an existing heartbeat monitor with the
	// specified details. It returns the updated monitor, or an error.
	UpdateHeartbeatMonitor(
		ctx context.Context,
		ID string,
		monitor HeartbeatMonitor,
	) (*HeartbeatMonitor, error)

	// UpdateTCPCheck updates an existing TCP check with the specified details.
	// It returns the updated check, or an error.
	UpdateTCPCheck(
		ctx context.Context,
		ID string,
		check TCPCheck,
	) (*TCPCheck, error)

	// UpdateTCPMonitor updates an existing TCP monitor with the specified
	// details. It returns the updated monitor, or an error.
	UpdateTCPMonitor(
		ctx context.Context,
		ID string,
		monitor TCPMonitor,
	) (*TCPMonitor, error)

	// UpdateURLMonitor updates an existing URL monitor with the specified details.
	// It returns the updated monitor, or an error.
	UpdateURLMonitor(
		ctx context.Context,
		ID string,
		monitor URLMonitor,
	) (*URLMonitor, error)

	// Delete deletes the check with the specified ID.
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

	// DeleteURLMonitor deletes the monitor with the specified ID.
	DeleteURLMonitor(
		ctx context.Context,
		ID string,
	) error

	// Get takes the ID of an existing check, and returns the check parameters,
	// or an error.
	GetCheck(
		ctx context.Context,
		ID string,
	) (*Check, error)

	// Get takes the ID of an existing TCP check, and returns the check
	// parameters, or an error.
	//
	// Deprecated: Use GetTCPMonitor instead.
	GetTCPCheck(
		ctx context.Context,
		ID string,
	) (*TCPCheck, error)

	// Get takes the ID of an existing TCP monitor, and returns the monitor
	// parameters, or an error.
	GetTCPMonitor(
		ctx context.Context,
		ID string,
	) (*TCPMonitor, error)

	// Get takes the ID of an existing URL monitor, and returns the monitor
	// parameters, or an error.
	GetURLMonitor(
		ctx context.Context,
		ID string,
	) (*URLMonitor, error)

	// CreateGroup creates a new check group with the specified details.
	// It returns the newly-created group, or an error.
	CreateGroup(
		ctx context.Context,
		group Group,
	) (*Group, error)

	// GetGroup takes the ID of an existing check group, and returns the
	// corresponding group, or an error.
	GetGroup(
		ctx context.Context,
		ID int64,
	) (*Group, error)

	// UpdateGroup takes the ID of an existing check group, and updates the
	// corresponding check group to match the supplied group. It returns the updated
	// group, or an error.
	UpdateGroup(
		ctx context.Context,
		ID int64,
		group Group,
	) (*Group, error)

	// DeleteGroup deletes the check group with the specified ID. It returns a
	DeleteGroup(
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

	// CreateTriggerCheck creates a new trigger with the specified details.
	CreateTriggerCheck(
		ctx context.Context,
		checkID string,
	) (*TriggerCheck, error)

	// GetTriggerCheck takes the ID of an existing trigger and returns it
	GetTriggerCheck(
		ctx context.Context,
		checkID string,
	) (*TriggerCheck, error)

	// DeleteTriggerCheck deletes the trigger with the specified ID.
	DeleteTriggerCheck(
		ctx context.Context,
		checkID string,
	) error

	// CreateTriggerGroup creates a new trigger with the specified details.
	CreateTriggerGroup(
		ctx context.Context,
		groupID int64,
	) (*TriggerGroup, error)

	// GetTriggerGroup takes the ID of an existing trigger and returns it
	GetTriggerGroup(
		ctx context.Context,
		groupID int64,
	) (*TriggerGroup, error)

	// DeleteTriggerGroup deletes the trigger with the specified ID.
	DeleteTriggerGroup(
		ctx context.Context,
		groupID int64,
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

	// SetAccountId sets ID on a client which is required when using User API keys.
	SetAccountId(ID string)

	// SetChecklySource sets the source of the check for analytics purposes.
	SetChecklySource(source string)

	// Get a specific runtime specs.
	GetRuntime(
		ctx context.Context,
		ID string,
	) (*Runtime, error)

	GetStaticIPs(ctx context.Context) ([]StaticIP, error)
}

// client represents a Checkly client. If the Debug field is set to an io.Writer
// (for example os.Stdout), then the client will dump API requests and responses
// to it.  To use a non-default HTTP client (for example, for testing, or to set
// a timeout), assign to the HTTPClient field. To set a non-default URL (for
// example, for testing), assign to the URL field.
type client struct {
	apiKey     string
	url        string
	accountId  string
	source     string
	httpClient *http.Client
	debug      io.Writer
}

// Check type constants
type CheckType string

// TypeBrowser is used to identify a browser check.
const TypeBrowser = "BROWSER"

// TypeAPI is used to identify an API check.
const TypeAPI = "API"

// TypeHeartbeat is used to identify a browser check.
const TypeHeartbeat = "HEARTBEAT"

// Escalation type constants

// RunBased identifies a run-based escalation type, for use with an AlertSettings.
const RunBased = "RUN_BASED"

// TimeBased identifies a time-based escalation type, for use with an AlertSettings.
const TimeBased = "TIME_BASED"

// Assertion source constants

// StatusCode identifies the HTTP status code as an assertion source.
const StatusCode = "STATUS_CODE"

// JSONBody identifies the JSON body data as an assertion source.
const JSONBody = "JSON_BODY"

// TextBody identifies the response body text as an assertion source.
const TextBody = "TEXT_BODY"

// Headers identifies the HTTP headers as an assertion source.
const Headers = "HEADERS"

// ResponseTime identifies the response time as an assertion source.
const ResponseTime = "RESPONSE_TIME"

// ResponseData identifies the response data of a TCP check as an assertion source.
const ResponseData = "RESPONSE_DATA"

// Assertion comparison constants

// Equals asserts that the source and target are equal.
const Equals = "EQUALS"

// NotEquals asserts that the source and target are not equal.
const NotEquals = "NOT_EQUALS"

// IsEmpty asserts that the source is empty.
const IsEmpty = "IS_EMPTY"

// NotEmpty asserts that the source is not empty.
const NotEmpty = "NOT_EMPTY"

// GreaterThan asserts that the source is greater than the target.
const GreaterThan = "GREATER_THAN"

// LessThan asserts that the source is less than the target.
const LessThan = "LESS_THAN"

// Contains asserts that the source contains a specified value.
const Contains = "CONTAINS"

// NotContains asserts that the source does not contain a specified value.
const NotContains = "NOT_CONTAINS"

// Check represents the parameters for an existing check.
type Check struct {
	ID                        string                     `json:"id"`
	Name                      string                     `json:"name"`
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
	PrivateLocations *[]string      `json:"privateLocations"`
	RuntimeID        *string        `json:"runtimeId"`
	RetryStrategy    *RetryStrategy `json:"retryStrategy,omitempty"`

	// Deprecated: this property will be removed in future versions.
	SSLCheck bool `json:"sslCheck"`
	// Deprecated: this property will be removed in future versions. Please use RetryStrategy instead.
	DoubleCheck bool `json:"doubleCheck"`
}

// Check represents the parameters for an existing check.
type MultiStepCheck struct {
	ID                        string                     `json:"id"`
	Name                      string                     `json:"name"`
	Type                      string                     `json:"checkType"`
	Frequency                 int                        `json:"frequency"`
	FrequencyOffset           int                        `json:"frequencyOffset,omitempty"`
	Activated                 bool                       `json:"activated"`
	Muted                     bool                       `json:"muted"`
	ShouldFail                bool                       `json:"shouldFail"`
	RunParallel               bool                       `json:"runParallel"`
	Locations                 []string                   `json:"locations"`
	Script                    string                     `json:"script,omitempty"`
	EnvironmentVariables      []EnvironmentVariable      `json:"environmentVariables"`
	Tags                      []string                   `json:"tags,omitempty"`
	AlertSettings             AlertSettings              `json:"alertSettings,omitempty"`
	UseGlobalAlertSettings    bool                       `json:"useGlobalAlertSettings"`
	GroupID                   int64                      `json:"groupId,omitempty"`
	GroupOrder                int                        `json:"groupOrder,omitempty"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	CreatedAt                 time.Time                  `json:"createdAt"`
	UpdatedAt                 time.Time                  `json:"updatedAt"`

	// Pointers
	PrivateLocations *[]string      `json:"privateLocations"`
	RuntimeID        *string        `json:"runtimeId"`
	RetryStrategy    *RetryStrategy `json:"retryStrategy,omitempty"`
}

type HeartbeatMonitor struct {
	ID                        string                     `json:"id"`
	Name                      string                     `json:"name"`
	Activated                 bool                       `json:"activated"`
	Muted                     bool                       `json:"muted"`
	Tags                      []string                   `json:"tags,omitempty"`
	AlertSettings             AlertSettings              `json:"alertSettings,omitempty"`
	UseGlobalAlertSettings    bool                       `json:"useGlobalAlertSettings"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	Heartbeat                 Heartbeat                  `json:"heartbeat"`
	CreatedAt                 time.Time                  `json:"createdAt"`
	UpdatedAt                 time.Time                  `json:"updatedAt"`
}

// HeartbeatCheck is an alias for HeartbeatMonitor for backwards compatibility
// purposes.
//
// Deprecated: Use HeartbeatMonitor instead.
type HeartbeatCheck = HeartbeatMonitor

// TCPMonitor represents a TCP monitor.
type TCPMonitor struct {
	ID                        string                     `json:"id,omitempty"`
	Name                      string                     `json:"name"`
	Frequency                 int                        `json:"frequency,omitempty"`
	FrequencyOffset           int                        `json:"frequencyOffset,omitempty"`
	Activated                 bool                       `json:"activated"`
	Muted                     bool                       `json:"muted"`
	ShouldFail                bool                       `json:"shouldFail"`
	RunParallel               bool                       `json:"runParallel"`
	Locations                 []string                   `json:"locations"`
	DegradedResponseTime      int                        `json:"degradedResponseTime,omitempty"`
	MaxResponseTime           int                        `json:"maxResponseTime,omitempty"`
	Tags                      []string                   `json:"tags,omitempty"`
	AlertSettings             *AlertSettings             `json:"alertSettings,omitempty"`
	UseGlobalAlertSettings    bool                       `json:"useGlobalAlertSettings"`
	Request                   TCPRequest                 `json:"request"`
	GroupID                   int64                      `json:"groupId,omitempty"`
	GroupOrder                int                        `json:"groupOrder,omitempty"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	PrivateLocations          *[]string                  `json:"privateLocations"`
	RuntimeID                 *string                    `json:"runtimeId"`
	RetryStrategy             *RetryStrategy             `json:"retryStrategy,omitempty"`
	CreatedAt                 time.Time                  `json:"created_at,omitempty"`
	UpdatedAt                 time.Time                  `json:"updated_at,omitempty"`
}

// TCPCheck is an alias for TCPMonitor for backwards compatibility purposes.
//
// Deprecated: Use TCPMonitor instead.
type TCPCheck = TCPMonitor

// URLMonitor represents a URL monitor.
type URLMonitor struct {
	ID                        string                     `json:"id,omitempty"`
	Name                      string                     `json:"name"`
	Frequency                 int                        `json:"frequency,omitempty"`
	FrequencyOffset           int                        `json:"frequencyOffset,omitempty"`
	Activated                 bool                       `json:"activated"`
	Muted                     bool                       `json:"muted"`
	ShouldFail                bool                       `json:"shouldFail"`
	RunParallel               bool                       `json:"runParallel"`
	Locations                 []string                   `json:"locations"`
	DegradedResponseTime      int                        `json:"degradedResponseTime,omitempty"`
	MaxResponseTime           int                        `json:"maxResponseTime,omitempty"`
	Tags                      []string                   `json:"tags,omitempty"`
	AlertSettings             *AlertSettings             `json:"alertSettings,omitempty"`
	UseGlobalAlertSettings    bool                       `json:"useGlobalAlertSettings"`
	Request                   URLRequest                 `json:"request"`
	GroupID                   int64                      `json:"groupId,omitempty"`
	GroupOrder                int                        `json:"groupOrder,omitempty"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	PrivateLocations          *[]string                  `json:"privateLocations"`
	RetryStrategy             *RetryStrategy             `json:"retryStrategy,omitempty"`
	CreatedAt                 time.Time                  `json:"created_at,omitempty"`
	UpdatedAt                 time.Time                  `json:"updated_at,omitempty"`
}

// URLRequest represents the parameters for the request made by a URL monitor.
type URLRequest struct {
	URL             string      `json:"url"`
	FollowRedirects bool        `json:"followRedirects"`
	SkipSSL         bool        `json:"skipSSL"`
	Assertions      []Assertion `json:"assertions"`
	IPFamily        string      `json:"ipFamily,omitempty"`
}

func (r *URLRequest) toRequest() Request {
	return Request{
		Method:          "GET",
		URL:             r.URL,
		FollowRedirects: r.FollowRedirects,
		SkipSSL:         r.SkipSSL,
		Assertions:      r.Assertions,
		IPFamily:        r.IPFamily,
	}
}

// Heartbeat represents the parameter for the heartbeat check.
type Heartbeat struct {
	Period     int    `json:"period"`
	PeriodUnit string `json:"periodUnit"`
	Grace      int    `json:"grace"`
	GraceUnit  string `json:"graceUnit"`
	PingToken  string `json:"pingToken"`
}

// Request represents the parameters for the request made by the check.
type Request struct {
	Method          string      `json:"method"`
	URL             string      `json:"url"`
	FollowRedirects bool        `json:"followRedirects"`
	SkipSSL         bool        `json:"skipSSL"`
	Body            string      `json:"body"`
	BodyType        string      `json:"bodyType,omitempty"`
	Headers         []KeyValue  `json:"headers"`
	QueryParameters []KeyValue  `json:"queryParameters"`
	Assertions      []Assertion `json:"assertions"`
	BasicAuth       *BasicAuth  `json:"basicAuth,omitempty"`
	IPFamily        string      `json:"ipFamily,omitempty"`
}

// Assertion represents an assertion about an API response, which will be
// verified as part of the check.
type Assertion struct {
	Edit          bool   `json:"edit"`
	Order         int    `json:"order"`
	ArrayIndex    int    `json:"arrayIndex"`
	ArraySelector int    `json:"arraySelector"`
	Source        string `json:"source"`
	Property      string `json:"property"`
	Comparison    string `json:"comparison"`
	Target        string `json:"target"`
}

// BasicAuth represents the HTTP basic authentication credentials for a request.
type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// KeyValue represents a key-value pair, for example a request header setting,
// or a query parameter.
type KeyValue struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Locked bool   `json:"locked"`
}

// TCPRequest represents the parameters for a TCP check's connection.
type TCPRequest struct {
	Hostname   string      `json:"hostname"`
	Port       uint16      `json:"port"`
	Data       string      `json:"data,omitempty"`
	Assertions []Assertion `json:"assertions,omitempty"`
	IPFamily   string      `json:"ipFamily,omitempty"`
}

// EnvironmentVariable represents a key-value pair for setting environment
// values during check execution.
type EnvironmentVariable struct {
	Key    string `json:"key,omitempty"`
	Value  string `json:"value"`
	Locked bool   `json:"locked"`
	Secret bool   `json:"secret"`
}

// PrivateLocationKey represents the keys that the private location has.
type PrivateLocationKey struct {
	Id        string `json:"id"`
	MaskedKey string `json:"maskedKey"`
	RawKey    string `json:"rawKey"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// DashboardKey represents the keys that the dashboard has.
type DashboardKey struct {
	Id        string `json:"id"`
	MaskedKey string `json:"maskedKey"`
	RawKey    string `json:"rawKey"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// AlertSettings represents an alert configuration.
type AlertSettings struct {
	EscalationType              string                      `json:"escalationType,omitempty"`
	RunBasedEscalation          RunBasedEscalation          `json:"runBasedEscalation,omitempty"`
	TimeBasedEscalation         TimeBasedEscalation         `json:"timeBasedEscalation,omitempty"`
	Reminders                   Reminders                   `json:"reminders,omitempty"`
	ParallelRunFailureThreshold ParallelRunFailureThreshold `json:"parallelRunFailureThreshold,omitempty"`
	// Deprecated: this property will be removed in future versions.
	SSLCertificates SSLCertificates `json:"sslCertificates,omitempty"`
}

// ParallelRunFailureThreshold represent an alert escalation based on the number
// of failing regions, only applicable for parallel checks
type ParallelRunFailureThreshold struct {
	Enabled    bool `json:"enabled,omitempty"`
	Percentage int  `json:"percentage,omitempty"`
}

// RunBasedEscalation represents an alert escalation based on a number of failed
// check runs.
type RunBasedEscalation struct {
	FailedRunThreshold int `json:"failedRunThreshold,omitempty"`
}

// TimeBasedEscalation represents an alert escalation based on the number of
// minutes after a check first starts failing.
type TimeBasedEscalation struct {
	MinutesFailingThreshold int `json:"minutesFailingThreshold,omitempty"`
}

// Reminders represents the number of reminders to send after an alert
// notification, and the time interval between them.
type Reminders struct {
	Amount   int `json:"amount,omitempty"`
	Interval int `json:"interval,omitempty"`
}

// Deprecated: this type will be removed in future versions.
// SSLCertificates represents alert settings for expiring SSL certificates.
type SSLCertificates struct {
	Enabled        bool `json:"enabled,omitempty"`
	AlertThreshold int  `json:"alertThreshold,omitempty"`
}

type RetryStrategy struct {
	Type               string `json:"type"`
	BaseBackoffSeconds int    `json:"baseBackoffSeconds"`
	MaxRetries         int    `json:"maxRetries"`
	MaxDurationSeconds int    `json:"maxDurationSeconds"`
	SameRegion         bool   `json:"sameRegion"`
}

// Group represents a check group.
type Group struct {
	ID                        int64                      `json:"id,omitempty"`
	Name                      string                     `json:"name"`
	Activated                 bool                       `json:"activated"`
	Muted                     bool                       `json:"muted"`
	RunParallel               bool                       `json:"runParallel"`
	Tags                      []string                   `json:"tags"`
	Locations                 []string                   `json:"locations"`
	Concurrency               int                        `json:"concurrency"`
	APICheckDefaults          APICheckDefaults           `json:"apiCheckDefaults"`
	EnvironmentVariables      []EnvironmentVariable      `json:"environmentVariables"`
	UseGlobalAlertSettings    bool                       `json:"useGlobalAlertSettings"`
	AlertSettings             AlertSettings              `json:"alertSettings,omitempty"`
	SetupSnippetID            int64                      `json:"setupSnippetId,omitempty"`
	TearDownSnippetID         int64                      `json:"tearDownSnippetId,omitempty"`
	LocalSetupScript          string                     `json:"localSetupScript,omitempty"`
	LocalTearDownScript       string                     `json:"localTearDownScript,omitempty"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	CreatedAt                 time.Time                  `json:"createdAt"`
	UpdatedAt                 time.Time                  `json:"updatedAt"`

	// Pointers
	RuntimeID        *string        `json:"runtimeId"`
	PrivateLocations *[]string      `json:"privateLocations"`
	RetryStrategy    *RetryStrategy `json:"retryStrategy,omitempty"`

	// Deprecated: this property will be removed in future versions. Please use RetryStrategy instead.
	DoubleCheck bool `json:"doubleCheck"`
}

// APICheckDefaults represents the default settings for API checks within a
// given group.
type APICheckDefaults struct {
	BaseURL         string      `json:"url"`
	Headers         []KeyValue  `json:"headers,omitempty"`
	QueryParameters []KeyValue  `json:"queryParameters,omitempty"`
	Assertions      []Assertion `json:"assertions,omitempty"`
	BasicAuth       BasicAuth   `json:"basicAuth,omitempty"`
}

// CheckResult represents a Check result
type CheckResult struct {
	ID                  string              `json:"id"`
	Name                string              `json:"name"`
	CheckID             string              `json:"checkId"`
	HasFailures         bool                `json:"hasFailures"`
	HasErrors           bool                `json:"hasErrors"`
	IsDegraded          bool                `json:"isDegraded"`
	OverMaxResponseTime bool                `json:"overMaxResponseTime"`
	RunLocation         string              `json:"runLocation"`
	ResponseTime        int64               `json:"responseTime"`
	ApiCheckResult      *ApiCheckResult     `json:"apiCheckResult"`
	BrowserCheckResult  *BrowserCheckResult `json:"browserCheckResult"`
	CheckRunID          int64               `json:"checkRunId"`
	Attempts            int64               `json:"attempts"`
	StartedAt           time.Time           `json:"startedAt"`
	StoppedAt           time.Time           `json:"stoppedAt"`
	CreatedAt           time.Time           `json:"created_at"`
}

// ApiCheckResult represents an API Check result
type ApiCheckResult map[string]interface{}

// BrowserCheckResult represents a Browser Check result
type BrowserCheckResult map[string]interface{}

// CheckResultsFilter represents the parameters that can be passed while
// getting Check Results
type CheckResultsFilter struct {
	Limit       int64
	Page        int64
	Location    string
	To          int64
	From        int64
	CheckType   CheckType
	HasFailures bool
}

// Snippet defines Snippet type
type Snippet struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Script    string    `json:"script"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

const (
	AlertTypeEmail     = "EMAIL"
	AlertTypeSlack     = "SLACK"
	AlertTypeWebhook   = "WEBHOOK"
	AlertTypeSMS       = "SMS"
	AlertTypePagerduty = "PAGERDUTY"
	AlertTypeOpsgenie  = "OPSGENIE"
	AlertTypeCall      = "CALL"
)

// AlertChannelSubscription represents a subscription to an alert channel.
type AlertChannelSubscription struct {
	ChannelID int64 `json:"alertChannelId"`
	Activated bool  `json:"activated"`
}

// AlertChannelEmail defines a type for an email alert channel
type AlertChannelEmail struct {
	Address string `json:"address"`
}

// AlertChannelSlack defines a type for a slack alert channel
type AlertChannelSlack struct {
	WebhookURL string `json:"url"`
	Channel    string `json:"channel"`
}

// AlertChannelSMS defines a type for a sms alert channel
type AlertChannelSMS struct {
	Name   string `json:"name"`
	Number string `json:"number"`
}

// AlertChannelCALL defines a type for a phone call alert channel
type AlertChannelCall struct {
	Name   string `json:"name"`
	Number string `json:"number"`
}

// AlertChannelOpsgenie defines a type for an opsgenie alert channel
type AlertChannelOpsgenie struct {
	Name     string `json:"name"`
	APIKey   string `json:"apiKey"`
	Region   string `json:"region"`
	Priority string `json:"priority"`
}

// AlertChannelPagerduty defines a type for an pager duty alert channel
type AlertChannelPagerduty struct {
	Account     string `json:"account,omitempty"`
	ServiceKey  string `json:"serviceKey"`
	ServiceName string `json:"serviceName,omitempty"`
}

// AlertChannelWebhook defines a type for a webhook alert channel
type AlertChannelWebhook struct {
	Name            string     `json:"name"`
	URL             string     `json:"url"`
	WebhookType     string     `json:"webhookType,omitempty"`
	Method          string     `json:"method,omitempty"`
	Template        string     `json:"template,omitempty"`
	WebhookSecret   string     `json:"webhookSecret,omitempty"`
	Headers         []KeyValue `json:"headers,omitempty"`
	QueryParameters []KeyValue `json:"queryParameters,omitempty"`
}

// AlertChannel represents an alert channel and its subscribed checks. The API
// defines this data as read-only.
type AlertChannel struct {
	ID                 int64                  `json:"id,omitempty"`
	Type               string                 `json:"type"`
	Email              *AlertChannelEmail     `json:"-"`
	Slack              *AlertChannelSlack     `json:"-"`
	SMS                *AlertChannelSMS       `json:"-"`
	CALL               *AlertChannelCall      `json:"-"`
	Opsgenie           *AlertChannelOpsgenie  `json:"-"`
	Webhook            *AlertChannelWebhook   `json:"-"`
	Pagerduty          *AlertChannelPagerduty `json:"-"`
	SendRecovery       *bool                  `json:"sendRecovery"`
	SendFailure        *bool                  `json:"sendFailure"`
	SendDegraded       *bool                  `json:"sendDegraded"`
	SSLExpiry          *bool                  `json:"sslExpiry"`
	SSLExpiryThreshold *int                   `json:"sslExpiryThreshold"`
	CreatedAt          string                 `json:"created_at"`
	UpdatedAt          string                 `json:"updated_at"`
}

// Dashboard defines a type for a dashboard.
type Dashboard struct {
	ID                 int64          `json:"id,omitempty"`
	DashboardID        string         `json:"dashboardId,omitempty"`
	CustomUrl          string         `json:"customUrl"`
	CustomDomain       string         `json:"customDomain,omitempty"`
	IsPrivate          bool           `json:"isPrivate,omitempty"`
	Logo               string         `json:"logo,omitempty"`
	Link               string         `json:"link,omitempty"`
	Description        string         `json:"description,omitempty"`
	Favicon            string         `json:"favicon,omitempty"`
	Header             string         `json:"header,omitempty"`
	Width              string         `json:"width,omitempty"`
	RefreshRate        int            `json:"refreshRate"`
	ChecksPerPage      int            `json:"checksPerPage,omitempty"`
	PaginationRate     int            `json:"paginationRate"`
	Paginate           bool           `json:"paginate"`
	Tags               []string       `json:"tags,omitempty"`
	HideTags           bool           `json:"hideTags,omitempty"`
	UseTagsAndOperator bool           `json:"useTagsAndOperator,omitempty"`
	EnableIncidents    bool           `json:"enableIncidents"`
	CreatedAt          string         `json:"created_at"`
	UpdatedAt          string         `json:"updated_at"`
	Keys               []DashboardKey `json:"keys,omitempty"`
}

// MaintenanceWindow defines a type for a maintenance window.
type MaintenanceWindow struct {
	ID             int64    `json:"id"`
	Name           string   `json:"name"`
	StartsAt       string   `json:"startsAt"`
	EndsAt         string   `json:"endsAt"`
	RepeatInterval int      `json:"repeatInterval,omitempty"`
	RepeatUnit     string   `json:"repeatUnit,omitempty"`
	RepeatEndsAt   string   `json:"repeatEndsAt,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
}

// PrivateLocation defines a type for a private location.
type PrivateLocation struct {
	ID         string               `json:"id"`
	Name       string               `json:"name"`
	SlugName   string               `json:"slugName"`
	Icon       string               `json:"icon,omitempty"`
	Keys       []PrivateLocationKey `json:"keys,omitempty"`
	LastSeen   string               `json:"lastSeen,omitempty"`
	AgentCount int                  `json:"agentCount,omitempty"`
	CreatedAt  string               `json:"created_at"`
	UpdatedAt  string               `json:"updated_at"`
}

// Trigger defines a type for a check trigger.
type TriggerCheck struct {
	ID        int64  `json:"id,omitempty"`
	CheckId   string `json:"checkId,omitempty"`
	Token     string `json:"token"`
	URL       string `json:"url"`
	CalledAt  string `json:"called_at"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Trigger defines a type for a group trigger.
type TriggerGroup struct {
	ID        int64  `json:"id,omitempty"`
	GroupId   int64  `json:"groupId,omitempty"`
	Token     string `json:"token"`
	URL       string `json:"url"`
	CalledAt  string `json:"called_at"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Location struct {
	Name   string `json:"name"`
	Region string `json:"region"`
}

type Runtime struct {
	Name             string `json:"name"`
	MultiStepSupport bool   `json:"multiStepSupport"`
	Stage            string `json:"stage"`
	RuntimeEndOfLife string `json:"runtimeEndOfLife"`
	Description      string `json:"description"`
}

// This type is used to describe Checkly's official
// public range of IP addresses checks are executed from
// see https://www.checklyhq.com/docs/monitoring/allowlisting/#ip-range-allowlisting
type StaticIP struct {
	Region  string
	Address netip.Prefix
}

// SetConfig sets config of alert channel based on it's type
func (a *AlertChannel) SetConfig(cfg interface{}) {
	switch v := cfg.(type) {
	case *AlertChannelEmail:
		a.Email = cfg.(*AlertChannelEmail)
	case *AlertChannelSMS:
		a.SMS = cfg.(*AlertChannelSMS)
	case *AlertChannelCall:
		a.CALL = cfg.(*AlertChannelCall)
	case *AlertChannelSlack:
		a.Slack = cfg.(*AlertChannelSlack)
	case *AlertChannelWebhook:
		a.Webhook = cfg.(*AlertChannelWebhook)
	case *AlertChannelOpsgenie:
		a.Opsgenie = cfg.(*AlertChannelOpsgenie)
	case *AlertChannelPagerduty:
		a.Pagerduty = cfg.(*AlertChannelPagerduty)
	default:
		log.Printf("Unknown config type %v", v)
	}
}

// GetConfig gets the config of the alert channel based on it's type
func (a *AlertChannel) GetConfig() (cfg map[string]interface{}) {
	byts := []byte{}
	var err error
	switch a.Type {
	case AlertTypeEmail:
		byts, err = json.Marshal(a.Email)
	case AlertTypeSMS:
		byts, err = json.Marshal(a.SMS)
	case AlertTypeCall:
		byts, err = json.Marshal(a.CALL)
	case AlertTypeSlack:
		byts, err = json.Marshal(a.Slack)
	case AlertTypeOpsgenie:
		byts, err = json.Marshal(a.Opsgenie)
	case AlertTypePagerduty:
		byts, err = json.Marshal(a.Pagerduty)
	case AlertTypeWebhook:
		byts, err = json.Marshal(a.Webhook)
	}

	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(byts, &cfg)
	return cfg
}

// AlertChannelConfigFromJSON gets AlertChannel.config from JSON
func AlertChannelConfigFromJSON(channelType string, cfgJSON []byte) (interface{}, error) {
	switch channelType {
	case AlertTypeEmail:
		r := AlertChannelEmail{}
		json.Unmarshal(cfgJSON, &r)
		return &r, nil
	case AlertTypeSMS:
		r := AlertChannelSMS{}
		json.Unmarshal(cfgJSON, &r)
		return &r, nil
	case AlertTypeCall:
		r := AlertChannelCall{}
		json.Unmarshal(cfgJSON, &r)
		return &r, nil
	case AlertTypeSlack:
		r := AlertChannelSlack{}
		json.Unmarshal(cfgJSON, &r)
		return &r, nil
	case AlertTypeOpsgenie:
		r := AlertChannelOpsgenie{}
		json.Unmarshal(cfgJSON, &r)
		return &r, nil
	case AlertTypePagerduty:
		r := AlertChannelPagerduty{}
		json.Unmarshal(cfgJSON, &r)
		return &r, nil
	case AlertTypeWebhook:
		r := AlertChannelWebhook{}
		json.Unmarshal(cfgJSON, &r)
		return &r, nil
	}
	return nil, fmt.Errorf("Unknown AlertChannel.config type")
}

type ClientCertificate struct {
	// ID is the Checkly identifier of the client certificate.
	ID string `json:"id,omitempty"`

	// Host is the host domain that the certificate should be used for.
	Host string `json:"host"`

	// Certificate is the client certificate in PEM format.
	Certificate string `json:"cert"`

	// PrivateKey is the private key for the certificate in PEM format.
	PrivateKey string `json:"key"`

	// Passphrase is an optional passphrase for the private key.
	Passphrase string `json:"passphrase,omitempty"`

	// TrustedCA is an optional PEM formatted bundle of CA certificates that
	// the client should trust. The bundle may contain many CA certificates.
	TrustedCA string `json:"ca,omitempty"`

	// CreatedAt is the time when the client certificate was created.
	CreatedAt time.Time `json:"created_at"`
}

type StatusPageTheme string

const (
	StatusPageThemeAuto  StatusPageTheme = "AUTO"
	StatusPageThemeDark  StatusPageTheme = "DARK"
	StatusPageThemeLight StatusPageTheme = "LIGHT"
)

type StatusPage struct {
	// ID is the Checkly identifier of the status page.
	ID string `json:"id,omitempty"`

	// Name is the name of the status page.
	Name string `json:"name"`

	// URL is the unique subdomain of the status page.
	URL string `json:"url"`

	// CustomDomain is an optional user-managed domain that hosts the status
	// page.
	CustomDomain string `json:"customDomain,omitempty"`

	// Logo is a URL to an image file to use as the logo for the status page.
	Logo string `json:"logo,omitempty"`

	// RedirectTo is the URL the user should be redirected to when clicking
	// the logo.
	RedirectTo string `json:"redirectTo,omitempty"`

	// Favicon is a URL to an image file to use as the favicon of the status
	// page.
	Favicon string `json:"favicon,omitempty"`

	// DefaultTheme is the default theme of the status page.
	DefaultTheme StatusPageTheme `json:"defaultTheme,omitempty"`

	// Cards is a list of cards to include on the status page.
	Cards []StatusPageCard `json:"cards"`
}

type StatusPageCard struct {
	// Name is the name of the card.
	Name string `json:"name"`

	// Services is the list of services to include in the card.
	Services []StatusPageService `json:"services"`
}

type StatusPageService struct {
	// ID is the Checkly identifier of the service.
	ID string `json:"id,omitempty"`

	// Name is the name of the service.
	Name string `json:"name"`
}
