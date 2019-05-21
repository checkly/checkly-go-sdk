package checkly

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
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

// Check represents the parameters for an existing check.
type Check struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"checkType"`
	Activated bool    `json:"activated"`
	Request   Request `json:"request"`
}

// Request represents the parameters for the request made by the check.
type Request struct {
	Method string `json:"method"`
	URL    string `json:"url"`
}

// Params represents a set of parameters sent with an API call.
type Params map[string]string

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
	p := Params{
		"name":      check.Name,
		"checkType": check.Type,
		"activated": fmt.Sprintf("%t", check.Activated),
	}
	status, res, err := c.MakeAPICall(http.MethodPost, "checks", p)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("unexpected response status %d", status)
	}
	m := make(map[string]interface{})
	if err = json.NewDecoder(strings.NewReader(res)).Decode(&m); err != nil {
		return "", fmt.Errorf("decoding error: %v", err)
	}
	rawID, ok := m["id"]
	if !ok {
		return "", errors.New("no ID field in response")
	}
	ID, ok := rawID.(string)
	if !ok {
		return "", fmt.Errorf("bad ID: %q", rawID)
	}
	return ID, nil
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
		return Check{}, fmt.Errorf("decoding error: %v", err)
	}
	return check, nil
}

// MakeAPICall calls the Checkly API with the specified verb and stores the
// returned data in the Response struct.
func (c *Client) MakeAPICall(method string, URL string, params Params) (statusCode int, response string, err error) {
	var body io.Reader
	if params != nil {
		form := url.Values{}
		for k, v := range params {
			form.Add(k, v)
		}
		body = strings.NewReader(form.Encode())
	}
	requestURL := c.URL + "/v1/" + URL
	req, err := http.NewRequest(method, requestURL, body)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.apiKey)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
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
	if c.Debug != nil {
		responseDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return resp.StatusCode, "", fmt.Errorf("error dumping HTTP response: %v", err)
		}
		fmt.Fprintln(c.Debug, string(responseDump))
		fmt.Fprintln(c.Debug)
	}
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", err
	}
	return resp.StatusCode, string(res), nil
}
