package checkly_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/checkly/checkly-go-sdk/v2"
)

// The API rejects an empty browsers list, and its create and update schemas
// default an omitted one, so the payload leaves out a nil or empty list and
// sends any other list as given.
func TestPlaywrightCheckPayloadBrowsers(t *testing.T) {
	t.Parallel()
	methods := map[string]struct {
		method string
		path   string
		status int
		call   func(checkly.Client, checkly.PlaywrightCheck) error
	}{
		"create": {
			method: http.MethodPost,
			path:   "/v1/checks/playwright?autoAssignAlerts=false",
			status: http.StatusCreated,
			call: func(c checkly.Client, check checkly.PlaywrightCheck) error {
				_, err := c.CreatePlaywrightCheck(context.Background(), check)
				return err
			},
		},
		"update": {
			method: http.MethodPut,
			path:   "/v1/checks/playwright/id1?autoAssignAlerts=false",
			status: http.StatusOK,
			call: func(c checkly.Client, check checkly.PlaywrightCheck) error {
				_, err := c.UpdatePlaywrightCheck(context.Background(), "id1", check)
				return err
			},
		},
	}
	browsers := map[string]struct {
		browsers []string
		want     string // serialized value; empty means the key is absent
	}{
		"nil":   {browsers: nil},
		"empty": {browsers: []string{}},
		"set":   {browsers: []string{"chromium", "firefox"}, want: `["chromium","firefox"]`},
	}
	for methodName, m := range methods {
		for browsersName, b := range browsers {
			m, b := m, b
			t.Run(methodName+"/"+browsersName, func(t *testing.T) {
				t.Parallel()
				validate := func(t *testing.T, body []byte) {
					if b.want == "" {
						assertRawBodyKeyAbsent(t, body, "browsers")
					} else {
						assertRawBodyField(t, body, "browsers", b.want)
					}
				}
				ts := cannedResponseServer(t, m.method, m.path, validate, m.status, "PlaywrightCheck.json")
				defer ts.Close()
				client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
				if err := m.call(client, checkly.PlaywrightCheck{Name: "Playwright check", Browsers: b.browsers}); err != nil {
					t.Error(err)
				}
			})
		}
	}
}
