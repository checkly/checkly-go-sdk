package checkly

import (
	"encoding/json"
	"testing"
)

// TestRequestMarshalsUnsetListsAsEmpty pins the request object of every
// payload that embeds a Request: the API rejects null headers, query
// parameters and assertions.
func TestRequestMarshalsUnsetListsAsEmpty(t *testing.T) {
	cases := []struct {
		name    string
		payload any
	}{
		{"APICheck", createAPICheckPayload(APICheck{Name: "api"})},
		{"Check", createCheckPayload(Check{Name: "check", Type: TypeAPI})},
		{"URLMonitor", createURLMonitorPayload(URLMonitor{Name: "url", Request: URLRequest{URL: "https://example.com"}})},
	}
	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			request, ok := marshalToMap(t, tt.payload)["request"].(map[string]any)
			if !ok {
				t.Fatal("missing request object")
			}
			for _, key := range []string{"headers", "queryParameters", "assertions"} {
				if list, ok := request[key].([]any); !ok || len(list) != 0 {
					t.Errorf("request.%s: want [], got %#v", key, request[key])
				}
			}
			if v, ok := request["basicAuth"]; !ok || v != nil {
				t.Errorf("request.basicAuth: want null, got %#v (present: %v)", v, ok)
			}
		})
	}
}

func TestRequestMarshalsSetValuesUnchanged(t *testing.T) {
	data, err := json.Marshal(Request{
		Method:          "POST",
		URL:             "https://example.com",
		FollowRedirects: true,
		SkipSSL:         true,
		Body:            "{}",
		BodyType:        "JSON",
		Headers:         []KeyValue{{Key: "X-Test", Value: "1"}},
		QueryParameters: []KeyValue{{Key: "q", Value: "2"}},
		Assertions:      []Assertion{{Source: StatusCode, Comparison: Equals, Target: "200"}},
		BasicAuth:       &BasicAuth{Username: "u", Password: "p"},
		IPFamily:        "IPv6",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := `{"method":"POST","url":"https://example.com","followRedirects":true,"skipSSL":true,"body":"{}","bodyType":"JSON",` +
		`"headers":[{"key":"X-Test","value":"1","locked":false}],"queryParameters":[{"key":"q","value":"2","locked":false}],` +
		`"assertions":[{"edit":false,"order":0,"arrayIndex":0,"arraySelector":0,"source":"STATUS_CODE","property":"","comparison":"EQUALS","target":"200","regex":""}],` +
		`"basicAuth":{"username":"u","password":"p"},"ipFamily":"IPv6"}`
	if string(data) != want {
		t.Errorf("want %s, got %s", want, data)
	}
}
