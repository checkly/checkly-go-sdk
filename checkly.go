package checkly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
)

// Client represents a Checkly client. If the Debug field is set to an io.Writer
// (for example os.Stdout), then the client will dump API requests and responses
// to it.  To use a non-default HTTP client (for example, for testing, or to set
// a timeout), assign to the HTTPClient field. To set a non-default URL (for
// example, for testing), assign to the URL field.
type Client struct {
	apiKey     string
	URL        string
	HTTPClient *http.Client
	Debug      io.Writer
}

// TypeBrowser is used to identify a browser check.
const TypeBrowser = "BROWSER"

// TypeAPI is used to identify an API check.
const TypeAPI = "API"

// RunBased identifies a run-based escalation type, for use with an AlertSettings.
const RunBased = "RUN_BASED"

// TimeBased identifies a time-based escalation type, for use with an AlertSettings.
const TimeBased = "TIME_BASED"

// Check represents the parameters for an existing check.
type Check struct {
	ID                     string                `json:"id"`
	Name                   string                `json:"name"`
	Type                   string                `json:"checkType"`
	Frequency              int                   `json:"frequency"`
	Activated              bool                  `json:"activated"`
	Muted                  bool                  `json:"muted"`
	ShouldFail             bool                  `json:"shouldFail"`
	Locations              []string              `json:"locations"`
	Script                 string                `json:"script,omitempty"`
	CreatedAt              time.Time             `json:"created_at,omitempty"`
	UpdatedAt              time.Time             `json:"updated_at,omitempty"`
	EnvironmentVariables   []EnvironmentVariable `json:"environment_variables"`
	DoubleCheck            bool                  `json:"doubleCheck"`
	Tags                   []string              `json:"tags,omitempty"`
	SSLCheck               bool                  `json:"sslCheck,omitempty"`
	SSLCheckDomain         string                `json:"sslCheckDomain,omitempty"`
	SetupSnippetID         int64                 `json:"setupSnippetId,omitempty"`
	TearDownSnippetID      int64                 `json:"tearDownSnippetId,omitempty"`
	LocalSetupScript       string                `json:"localSetupScript,omitempty"`
	LocalTearDownScript    string                `json:"localTearDownScript,omitempty"`
	AlertSettings          AlertSettings         `json:"alertSettings,omitempty"`
	UseGlobalAlertSettings bool                  `jons:"useGlobalAlertSettings"`
	Request                Request               `json:"request"`
}

// Request represents the parameters for the request made by the check.
type Request struct {
	Method string `json:"method"`
	URL    string `json:"url"`
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
	Enabled        bool `json:"enabled,omitempty"`
	AlertThreshold int  `json:"alertThreshold,omitempty"`
}

// NewClient takes a Checkly API key, and returns a Client ready to use.
func NewClient(apiKey string) Client {
	return Client{
		apiKey:     apiKey,
		URL:        "https://api.checklyhq.com",
		HTTPClient: http.DefaultClient,
	}
}

// Create creates a new check with the specified details. It returns the
// check ID of the newly-created check, or an error.
func (c *Client) Create(check Check) (string, error) {
	data, err := json.Marshal(check)
	if err != nil {
		return "", err
	}
	status, res, err := c.MakeAPICall(http.MethodPost, "checks", data)
	if err != nil {
		return "", err
	}
	if status != http.StatusCreated {
		return "", fmt.Errorf("unexpected response status %d", status)
	}
	var result Check
	if err = json.NewDecoder(strings.NewReader(res)).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding error for data %s: %v", res, err)
	}
	return result.ID, nil
}

// Delete deletes the check with the specified ID. It returns a non-nil
// error if the request failed.
func (c *Client) Delete(ID string) error {
	status, _, err := c.MakeAPICall(http.MethodDelete, "checks/"+ID, nil)
	if err != nil {
		return err
	}
	if status != http.StatusNoContent {
		return fmt.Errorf("unexpected response status %d", status)
	}
	return nil
}

// Get takes the ID of an existing check, and returns the check parameters, or
// an error.
func (c *Client) Get(ID string) (Check, error) {
	status, res, err := c.MakeAPICall(http.MethodGet, "checks/"+ID, nil)
	if err != nil {
		return Check{}, err
	}
	if status != http.StatusOK {
		return Check{}, fmt.Errorf("unexpected response status %d", status)
	}
	check := Check{}
	if err = json.NewDecoder(strings.NewReader(res)).Decode(&check); err != nil {
		return Check{}, fmt.Errorf("decoding error for data %s: %v", res, err)
	}
	return check, nil
}

// MakeAPICall calls the Checkly API with the specified verb and stores the
// returned data in the Response struct.
func (c *Client) MakeAPICall(method string, URL string, data []byte) (statusCode int, response string, err error) {
	requestURL := c.URL + "/v1/" + URL
	req, err := http.NewRequest(method, requestURL, bytes.NewBuffer(data))
	if err != nil {
		return 0, "", fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.apiKey)
	req.Header.Add("content-type", "application/json")
	if c.Debug != nil {
		requestDump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return 0, "", fmt.Errorf("error dumping HTTP request: %v", err)
		}
		fmt.Fprintln(c.Debug, string(requestDump))
		fmt.Fprintln(c.Debug)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()
	if c.Debug != nil || resp.StatusCode == http.StatusBadRequest {
		c.dumpResponse(resp)
	}
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", err
	}
	return resp.StatusCode, string(res), nil
}

// dumpResponse writes the raw response data to the debug output, if set, or
// standard error otherwise.
func (c *Client) dumpResponse(resp *http.Response) {
	// ignore errors dumping response - no recovery from this
	responseDump, _ := httputil.DumpResponse(resp, true)
	out := c.Debug
	if out == nil {
		out = os.Stderr
	}
	fmt.Fprintln(out, string(responseDump))
	fmt.Fprintln(out)
}
