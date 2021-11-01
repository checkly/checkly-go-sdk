package checkly

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Client is an interface that implements Checkly's API
type Client interface {
	// Create creates a new check with the specified details.
	// It returns the newly-created check, or an error.
	Create(
		ctx context.Context,
		check Check,
	) (*Check, error)

	// Update updates an existing check with the specified details.
	// It returns the updated check, or an error.
	Update(
		ctx context.Context,
		ID string,
		check Check,
	) (*Check, error)

	// Delete deletes the check with the specified ID.
	// It returns a non-nil error if the request failed.
	Delete(
		ctx context.Context,
		ID string,
	) error

	// Get takes the ID of an existing check, and returns the check parameters,
	// or an error.
	Get(
		ctx context.Context,
		ID string,
	) (*Check, error)

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
	// non-nil error if the request failed.
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
	// non-nil error if the request failed.
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
	// non-nil error if the request failed.
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

	// DeleteAlertChannel deletes the alert channel with the specified ID. It returns a
	// non-nil error if the request failed.
	DeleteAlertChannel(
		ctx context.Context,
		ID int64,
	) error

	CreateDashboard(
		ctx context.Context,
		dashboard Dashboard,
	) (*Dashboard, error)

	GetDashboard(
		ctx context.Context,
		ID string,
	) (*Dashboard, error)

	UpdateDashboard(
		ctx context.Context,
		ID string,
		dashboard Dashboard,
	) (*Dashboard, error)

	DeleteDashboard(
		ctx context.Context,
		ID string,
	) error
}

// client represents a Checkly client. If the Debug field is set to an io.Writer
// (for example os.Stdout), then the client will dump API requests and responses
// to it.  To use a non-default HTTP client (for example, for testing, or to set
// a timeout), assign to the HTTPClient field. To set a non-default URL (for
// example, for testing), assign to the URL field.
type client struct {
	apiKey     string
	url        string
	httpClient *http.Client
	debug      io.Writer
}

// Check type constants
type CheckType string

// TypeBrowser is used to identify a browser check.
const TypeBrowser = "BROWSER"

// TypeAPI is used to identify an API check.
const TypeAPI = "API"

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
	Locations                 []string                   `json:"locations"`
	DegradedResponseTime      int                        `json:"degradedResponseTime"`
	MaxResponseTime           int                        `json:"maxResponseTime"`
	Script                    string                     `json:"script,omitempty"`
	EnvironmentVariables      []EnvironmentVariable      `json:"environmentVariables"`
	DoubleCheck               bool                       `json:"doubleCheck"`
	Tags                      []string                   `json:"tags,omitempty"`
	SSLCheck                  bool                       `json:"sslCheck"`
	SetupSnippetID            int64                      `json:"setupSnippetId,omitempty"`
	TearDownSnippetID         int64                      `json:"tearDownSnippetId,omitempty"`
	LocalSetupScript          string                     `json:"localSetupScript,omitempty"`
	LocalTearDownScript       string                     `json:"localTearDownScript,omitempty"`
	AlertSettings             AlertSettings              `json:"alertSettings,omitempty"`
	UseGlobalAlertSettings    bool                       `json:"useGlobalAlertSettings"`
	Request                   Request                    `json:"request"`
	GroupID                   int64                      `json:"groupId,omitempty"`
	GroupOrder                int                        `json:"groupOrder,omitempty"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	RuntimeID                 *string                    `json:"runtimeId"`
}

// Request represents the parameters for the request made by the check.
type Request struct {
	Method          string      `json:"method"`
	URL             string      `json:"url"`
	FollowRedirects bool        `json:"followRedirects"`
	Body            string      `json:"body"`
	BodyType        string      `json:"bodyType,omitempty"`
	Headers         []KeyValue  `json:"headers"`
	QueryParameters []KeyValue  `json:"queryParameters"`
	Assertions      []Assertion `json:"assertions"`
	BasicAuth       *BasicAuth  `json:"basicAuth,omitempty"`
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

// EnvironmentVariable represents a key-value pair for setting environment
// values during check execution.
type EnvironmentVariable struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Locked bool   `json:"locked"`
}

// AlertSettings represents an alert configuration.
type AlertSettings struct {
	EscalationType      string              `json:"escalationType,omitempty"`
	RunBasedEscalation  RunBasedEscalation  `json:"runBasedEscalation,omitempty"`
	TimeBasedEscalation TimeBasedEscalation `json:"timeBasedEscalation,omitempty"`
	Reminders           Reminders           `json:"reminders,omitempty"`
	SSLCertificates     SSLCertificates     `json:"sslCertificates,omitempty"`
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

// SSLCertificates represents alert settings for expiring SSL certificates.
type SSLCertificates struct {
	Enabled        bool `json:"enabled"`
	AlertThreshold int  `json:"alertThreshold"`
}

// Group represents a check group.
type Group struct {
	ID                        int64                      `json:"id,omitempty"`
	Name                      string                     `json:"name"`
	Activated                 bool                       `json:"activated"`
	Muted                     bool                       `json:"muted"`
	Tags                      []string                   `json:"tags"`
	Locations                 []string                   `json:"locations"`
	Concurrency               int                        `json:"concurrency"`
	APICheckDefaults          APICheckDefaults           `json:"apiCheckDefaults"`
	EnvironmentVariables      []EnvironmentVariable      `json:"environmentVariables"`
	DoubleCheck               bool                       `json:"doubleCheck"`
	UseGlobalAlertSettings    bool                       `json:"useGlobalAlertSettings"`
	AlertSettings             AlertSettings              `json:"alertSettings,omitempty"`
	SetupSnippetID            int64                      `json:"setupSnippetId,omitempty"`
	TearDownSnippetID         int64                      `json:"tearDownSnippetId,omitempty"`
	LocalSetupScript          string                     `json:"localSetupScript,omitempty"`
	LocalTearDownScript       string                     `json:"localTearDownScript,omitempty"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	RuntimeID                 *string                    `json:"runtimeId"`
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
	StartedAt           time.Time           `json:"startedAt"`
	StoppedAt           time.Time           `json:"stoppedAt"`
	CreatedAt           time.Time           `json:"created_at"`
	ResponseTime        int64               `json:"responseTime"`
	ApiCheckResult      *ApiCheckResult     `json:"apiCheckResult"`
	BrowserCheckResult  *BrowserCheckResult `json:"browserCheckResult"`
	CheckRunID          int64               `json:"checkRunId"`
	Attempts            int64               `json:"attempts"`
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

//Snippet defines Snippet type
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
)

// AlertChannelSubscription represents a subscription to an alert channel.
type AlertChannelSubscription struct {
	ChannelID int64 `json:"alertChannelId"`
	Activated bool  `json:"activated"`
}

//AlertChannelEmail defines a type for an email alert channel
type AlertChannelEmail struct {
	Address string `json:"address"`
}

//AlertChannelSlack defines a type for a slack alert channel
type AlertChannelSlack struct {
	WebhookURL string `json:"url"`
	Channel    string `json:"channel"`
}

//AlertChannelSMS defines a type for a sms alert channel
type AlertChannelSMS struct {
	Name   string `json:"name"`
	Number string `json:"number"`
}

//AlertChannelOpsgenie defines a type for an opsgenie alert channel
type AlertChannelOpsgenie struct {
	Name     string `json:"name"`
	APIKey   string `json:"apiKey"`
	Region   string `json:"region"`
	Priority string `json:"priority"`
}

//AlertChannelPagerduty defines a type for an pager duty alert channel
type AlertChannelPagerduty struct {
	Account     string `json:"account,omitempty"`
	ServiceKey  string `json:"serviceKey"`
	ServiceName string `json:"serviceName,omitempty"`
}

//AlertChannelWebhook defines a type for a webhook alert channel
type AlertChannelWebhook struct {
	Name            string     `json:"name"`
	URL             string     `json:"url"`
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
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	Email              *AlertChannelEmail     `json:"-"`
	Slack              *AlertChannelSlack     `json:"-"`
	SMS                *AlertChannelSMS       `json:"-"`
	Opsgenie           *AlertChannelOpsgenie  `json:"-"`
	Webhook            *AlertChannelWebhook   `json:"-"`
	Pagerduty          *AlertChannelPagerduty `json:"-"`
	SendRecovery       *bool                  `json:"sendRecovery"`
	SendFailure        *bool                  `json:"sendFailure"`
	SendDegraded       *bool                  `json:"sendDegraded"`
	SSLExpiry          *bool                  `json:"sslExpiry"`
	SSLExpiryThreshold *int                   `json:"sslExpiryThreshold"`
}

// Dashboard defines a type for a dashboard.
type Dashboard struct {
	DashboardID    string   `json:"dashboardId"`
	CustomUrl      string   `json:"customUrl"`
	CustomDomain   string   `json:"customDomain"`
	Logo           string   `json:"logo"`
	Header         string   `json:"header"`
	Width          string   `json:"width,omitempty"`
	RefreshRate    int      `json:"refreshRate"`
	Paginate       bool     `json:"paginate"`
	PaginationRate int      `json:"paginationRate"`
	Tags           []string `json:"tags,omitempty"`
	HideTags       bool     `json:"hideTags"`
}

//SetConfig sets config of alert channel based on it's type
func (a *AlertChannel) SetConfig(cfg interface{}) {
	switch v := cfg.(type) {
	case *AlertChannelEmail:
		a.Email = cfg.(*AlertChannelEmail)
	case *AlertChannelSMS:
		a.SMS = cfg.(*AlertChannelSMS)
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

//GetConfig gets the config of the alert channel based on it's type
func (a *AlertChannel) GetConfig() (cfg map[string]interface{}) {
	byts := []byte{}
	var err error
	switch a.Type {
	case AlertTypeEmail:
		byts, err = json.Marshal(a.Email)
	case AlertTypeSMS:
		byts, err = json.Marshal(a.SMS)
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

//AlertChannelConfigFromJSON gets AlertChannel.config from JSON
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
