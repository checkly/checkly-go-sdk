package checkly

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"

	"strings"
)

// Client represents a Checkly client. If the Debug field is set to an
// io.Writer, then the client will dump API requests to it instead of calling
// the real API.  To use a non-default HTTP client (for example, for testing, or
// to set a timeout), assign to the HTTPClient field. To set a non-default URL
// (for example, for testing), assign to the URL field.
type Client struct {
	apiKey     string
	URL        string
	HTTPClient *http.Client
	Debug      io.Writer
}

// Error represents an API error.
type Error map[string]interface{}

// Params holds optional parameters for API calls.
type Params map[string]string

// Check represents a Checkly check.
type Check struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// NewClient takes a Checkly API key, and returns a Client ready to use.
func NewClient(apiKey string) Client {
	return Client{
		apiKey:     apiKey,
		URL:        "https://api.checklyhq.com",
		HTTPClient: http.DefaultClient,
	}
}

// CreateCheck creates a new check with the specified details. It returns the newly
// created check, or an error.
func (c *Client) CreateCheck(p Params) (Check, error) {
	res, err := c.MakeAPICall("checks", p)
	if err != nil {
		return Check{}, err
	}
	var m Check
	if err = json.NewDecoder(strings.NewReader(res)).Decode(&m); err != nil {
		return Check{}, fmt.Errorf("decoding error: %v", err)
	}
	return m, nil
}

// MakeAPICall calls the checkly API with the specified verb and stores the
// returned data in the Response struct.
func (c *Client) MakeAPICall(verb string, params Params) (string, error) {
	form := url.Values{}
	form.Add("api_key", c.apiKey)
	form.Add("format", "json")
	for k, v := range params {
		form.Add(k, v)
	}
	requestURL := c.URL + "/v1/" + verb
	req, err := http.NewRequest("POST", requestURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	if c.Debug != nil {
		requestDump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return "", fmt.Errorf("error dumping HTTP request: %v", err)
		}
		fmt.Fprintln(c.Debug, string(requestDump))
		fmt.Fprintln(c.Debug)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()
	if c.Debug != nil {
		responseDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return "", fmt.Errorf("error dumping HTTP response: %v", err)
		}
		fmt.Fprintln(c.Debug, string(responseDump))
		fmt.Fprintln(c.Debug)
	}
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(res), nil
}
