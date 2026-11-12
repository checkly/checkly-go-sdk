package checkly

import (
	"encoding/json"
	"testing"
)

// TestURLMonitorRequestPayload pins the request object a URL monitor write
// sends: only the fields a URL monitor has. Above all there is no headers
// key, as an update that sends one replaces the headers stored for accounts
// that can set URL monitor headers.
func TestURLMonitorRequestPayload(t *testing.T) {
	for _, tc := range []struct {
		name    string
		request URLRequest
		want    string
	}{
		{
			name:    "unset",
			request: URLRequest{URL: "https://example.com"},
			want:    `{"method":"GET","url":"https://example.com","followRedirects":false,"skipSSL":false,"assertions":[]}`,
		},
		{
			name: "set",
			request: URLRequest{
				URL:             "https://example.com",
				FollowRedirects: true,
				SkipSSL:         true,
				Assertions:      []Assertion{{Source: AssertionSourceStatusCode, Comparison: AssertionComparisonEquals, Target: "200"}},
				IPFamily:        "IPv6",
			},
			want: `{"method":"GET","url":"https://example.com","followRedirects":true,"skipSSL":true,` +
				`"assertions":[{"source":"STATUS_CODE","property":"","comparison":"EQUALS","target":"200","regex":""}],"ipFamily":"IPv6"}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(createURLMonitorPayload(URLMonitor{Name: "url", Request: tc.request}))
			if err != nil {
				t.Fatal(err)
			}
			var got struct {
				Request json.RawMessage `json:"request"`
			}
			if err := json.Unmarshal(payload, &got); err != nil {
				t.Fatal(err)
			}
			if string(got.Request) != tc.want {
				t.Errorf("request = %s, want %s", got.Request, tc.want)
			}
		})
	}
}
