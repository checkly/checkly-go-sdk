package checkly_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	checkly "github.com/checkly/checkly-go-sdk"
)

func TestRetryStrategyRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input checkly.RetryStrategy
		want  checkly.RetryStrategy
	}{
		{
			name:  "FIXED strategy round-trips",
			input: checkly.RetryStrategy{Type: "FIXED", MaxRetries: 2, MaxDurationSeconds: 600, BaseBackoffSeconds: 60, SameRegion: true},
			want:  checkly.RetryStrategy{Type: "FIXED", MaxRetries: 2, MaxDurationSeconds: 600, BaseBackoffSeconds: 60, SameRegion: true},
		},
		{
			name:  "LINEAR strategy round-trips",
			input: checkly.RetryStrategy{Type: "LINEAR", MaxRetries: 3, MaxDurationSeconds: 300, BaseBackoffSeconds: 30, SameRegion: false},
			want:  checkly.RetryStrategy{Type: "LINEAR", MaxRetries: 3, MaxDurationSeconds: 300, BaseBackoffSeconds: 30, SameRegion: false},
		},
		{
			name:  "SINGLE_RETRY omits MaxRetries and MaxDurationSeconds",
			input: checkly.RetryStrategy{Type: "SINGLE_RETRY", BaseBackoffSeconds: 10, SameRegion: true},
			want:  checkly.RetryStrategy{Type: "SINGLE_RETRY", BaseBackoffSeconds: 10, SameRegion: true},
		},
		{
			name:  "NO_RETRIES marshals as null and back",
			input: checkly.RetryStrategy{Type: "NO_RETRIES"},
			want:  checkly.RetryStrategy{Type: "NO_RETRIES"},
		},
		{
			name:  "FALLBACK marshals as string and back",
			input: checkly.RetryStrategy{Type: "FALLBACK"},
			want:  checkly.RetryStrategy{Type: "FALLBACK"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}
			var got checkly.RetryStrategy
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v (JSON was: %s)", got, tt.want, data)
			}
		})
	}
}

func TestRetryStrategyUnmarshalClearsFields(t *testing.T) {
	// Unmarshal into a struct that already has values to verify fields are cleared.
	s := checkly.RetryStrategy{
		Type:               "FIXED",
		MaxRetries:         5,
		MaxDurationSeconds: 600,
		BaseBackoffSeconds: 60,
		SameRegion:         true,
	}
	if err := json.Unmarshal([]byte("null"), &s); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	want := checkly.RetryStrategy{Type: "NO_RETRIES"}
	if !reflect.DeepEqual(s, want) {
		t.Errorf("got %+v, want %+v", s, want)
	}
}

func TestRetryStrategyUnmarshalRejectsUnknownString(t *testing.T) {
	var s checkly.RetryStrategy
	err := json.Unmarshal([]byte(`"UNKNOWN"`), &s)
	if err == nil {
		t.Fatal("expected error for unknown string value, got nil")
	}
}

func TestRetryStrategyMarshalNoRetries(t *testing.T) {
	s := checkly.RetryStrategy{Type: "NO_RETRIES"}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	if string(data) != "null" {
		t.Errorf("got %s, want null", data)
	}
}

func TestRetryStrategyMarshalFallback(t *testing.T) {
	s := checkly.RetryStrategy{Type: "FALLBACK"}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	if string(data) != `"FALLBACK"` {
		t.Errorf("got %s, want %q", data, "FALLBACK")
	}
}

// requestConfig sends ac with CreateAlertChannel and returns the config of
// the request body.
func requestConfig(t *testing.T, ac checkly.AlertChannel) json.RawMessage {
	t.Helper()
	configs := make(chan json.RawMessage, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Config json.RawMessage `json:"config"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decoding request body: %v", err)
		}
		configs <- body.Config
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, `{"id":1}`)
	}))
	t.Cleanup(srv.Close)
	c := newTestClient(t, checkly.WithBaseURL(srv.URL), checkly.WithAuth("test-key", testAccountID), checkly.WithHTTPClient(srv.Client()))
	if _, err := c.CreateAlertChannel(context.Background(), ac); err != nil {
		t.Fatal(err)
	}
	return <-configs
}

// TestAlertChannelConfigEncoding checks that a request sends the config field
// that matches the channel's type, and null when that field is nil or the
// type is unknown.
func TestAlertChannelConfigEncoding(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ac   checkly.AlertChannel
		want string
	}{
		{
			name: "email",
			ac:   checkly.AlertChannel{Type: checkly.AlertTypeEmail, Email: &checkly.AlertChannelEmail{Address: "foo@test.com"}},
			want: `{"address":"foo@test.com"}`,
		},
		{
			name: "slack",
			ac:   checkly.AlertChannel{Type: checkly.AlertTypeSlack, Slack: &checkly.AlertChannelSlack{WebhookURL: "http://example.com/", Channel: "foochan"}},
			want: `{"url":"http://example.com/","channel":"foochan"}`,
		},
		{
			name: "slack app",
			ac:   checkly.AlertChannel{Type: checkly.AlertTypeSlackApp, SlackApp: &checkly.AlertChannelSlackApp{SlackChannels: []string{"#ops", "@John"}}},
			want: `{"slackChannels":["#ops","@John"]}`,
		},
		{
			name: "sms",
			ac:   checkly.AlertChannel{Type: checkly.AlertTypeSMS, SMS: &checkly.AlertChannelSMS{Name: "foo", Number: "0123456789"}},
			want: `{"name":"foo","number":"0123456789"}`,
		},
		{
			name: "call",
			ac:   checkly.AlertChannel{Type: checkly.AlertTypeCall, CALL: &checkly.AlertChannelCall{Name: "foo", Number: "0123456789"}},
			want: `{"name":"foo","number":"0123456789"}`,
		},
		{
			name: "webhook",
			ac: checkly.AlertChannel{Type: checkly.AlertTypeWebhook, Webhook: &checkly.AlertChannelWebhook{
				Name:            "foo",
				Method:          "GET",
				Template:        "bar",
				URL:             "http://foo.com",
				WebhookSecret:   "scrt",
				Headers:         []checkly.KeyValue{{Key: "fookey", Value: "fooval"}, {Key: "barkey", Value: "barval", Locked: true}},
				QueryParameters: []checkly.KeyValue{{Key: "fookey", Value: "fooval", Locked: true}},
			}},
			want: `{"name":"foo","url":"http://foo.com","method":"GET","template":"bar","webhookSecret":"scrt",` +
				`"headers":[{"key":"fookey","value":"fooval","locked":false},{"key":"barkey","value":"barval","locked":true}],` +
				`"queryParameters":[{"key":"fookey","value":"fooval","locked":true}]}`,
		},
		{
			name: "opsgenie",
			ac:   checkly.AlertChannel{Type: checkly.AlertTypeOpsgenie, Opsgenie: &checkly.AlertChannelOpsgenie{Name: "foo", APIKey: "bar", Region: "regio-1", Priority: "highp"}},
			want: `{"name":"foo","apiKey":"bar","region":"regio-1","priority":"highp"}`,
		},
		{
			name: "pagerduty",
			ac:   checkly.AlertChannel{Type: checkly.AlertTypePagerduty, Pagerduty: &checkly.AlertChannelPagerduty{Account: "foo", ServiceKey: "xxx", ServiceName: "bar"}},
			want: `{"account":"foo","serviceKey":"xxx","serviceName":"bar"}`,
		},
		{
			name: "nil config",
			ac:   checkly.AlertChannel{Type: checkly.AlertTypeEmail, SMS: &checkly.AlertChannelSMS{Name: "foo"}},
			want: `null`,
		},
		{
			name: "unknown type",
			ac:   checkly.AlertChannel{Type: "TEAMS", Email: &checkly.AlertChannelEmail{Address: "foo@test.com"}},
			want: `null`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := requestConfig(t, tt.ac)
			var gotValue, wantValue any
			if err := json.Unmarshal(got, &gotValue); err != nil {
				t.Fatalf("decoding config %s: %v", got, err)
			}
			if err := json.Unmarshal([]byte(tt.want), &wantValue); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(gotValue, wantValue) {
				t.Errorf("want config %s, got %s", tt.want, got)
			}
		})
	}
}

// fixtureConfig returns the config object of the alert channel in a fixture.
func fixtureConfig(t *testing.T, name string) string {
	t.Helper()
	var ac struct {
		Config json.RawMessage `json:"config"`
	}
	if err := json.Unmarshal([]byte(readFixture(t, name)), &ac); err != nil {
		t.Fatal(err)
	}
	return string(ac.Config)
}

// getAlertChannel returns the alert channel GetAlertChannel decodes from a
// response of type channelType with config, a JSON value.
func getAlertChannel(t *testing.T, channelType, config string) (*checkly.AlertChannel, error) {
	t.Helper()
	body := fmt.Sprintf(`{"id":7,"type":%q,"config":%s}`, channelType, config)
	c := newErrorServer(t, errorReply{status: http.StatusOK, body: body})
	return c.GetAlertChannel(context.Background(), 7)
}

// TestAlertChannelConfigDecoding decodes a realistic config of every alert
// channel type into the config field of that type. The email and Slack
// configs come from the fixtures.
func TestAlertChannelConfigDecoding(t *testing.T) {
	t.Parallel()
	tests := []struct {
		channelType string
		config      string
		want        checkly.AlertChannel
	}{
		{checkly.AlertTypeEmail, fixtureConfig(t, "CreateAlertChannelEmail.json"), checkly.AlertChannel{Email: &checkly.AlertChannelEmail{Address: "test@example.com"}}},
		// The fixture's webhookUrl key is not the struct's url key, so only
		// the channel decodes.
		{checkly.AlertTypeSlack, fixtureConfig(t, "CreateAlertChannelSlack.json"), checkly.AlertChannel{Slack: &checkly.AlertChannelSlack{Channel: "test"}}},
		{checkly.AlertTypeSlackApp, `{"slackChannels":["C123","C456"]}`, checkly.AlertChannel{SlackApp: &checkly.AlertChannelSlackApp{SlackChannels: []string{"C123", "C456"}}}},
		{checkly.AlertTypeSMS, `{"name":"On call","number":"+15550100"}`, checkly.AlertChannel{SMS: &checkly.AlertChannelSMS{Name: "On call", Number: "+15550100"}}},
		{checkly.AlertTypeCall, `{"name":"On call","number":"+15550100"}`, checkly.AlertChannel{CALL: &checkly.AlertChannelCall{Name: "On call", Number: "+15550100"}}},
		{checkly.AlertTypeOpsgenie, `{"name":"Ops","apiKey":"fake-opsgenie-key","region":"EU","priority":"P3"}`, checkly.AlertChannel{Opsgenie: &checkly.AlertChannelOpsgenie{Name: "Ops", APIKey: "fake-opsgenie-key", Region: "EU", Priority: "P3"}}},
		{checkly.AlertTypePagerduty, `{"account":"acme","serviceKey":"fake-service-key","serviceName":"API"}`, checkly.AlertChannel{Pagerduty: &checkly.AlertChannelPagerduty{Account: "acme", ServiceKey: "fake-service-key", ServiceName: "API"}}},
		{
			checkly.AlertTypeWebhook,
			`{"name":"Hook","url":"https://example.com/hook","webhookType":"WEBHOOK_GENERIC","method":"POST","template":"{}","webhookSecret":"fake-secret","headers":[{"key":"X-Env","value":"prod","locked":false}],"queryParameters":[]}`,
			checkly.AlertChannel{Webhook: &checkly.AlertChannelWebhook{
				Name:            "Hook",
				URL:             "https://example.com/hook",
				WebhookType:     "WEBHOOK_GENERIC",
				Method:          "POST",
				Template:        "{}",
				WebhookSecret:   "fake-secret",
				Headers:         []checkly.KeyValue{{Key: "X-Env", Value: "prod"}},
				QueryParameters: []checkly.KeyValue{},
			}},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.channelType, func(t *testing.T) {
			t.Parallel()
			tt.want.ID = 7
			tt.want.Type = tt.channelType
			got, err := getAlertChannel(t, tt.channelType, tt.config)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("want %#v, got %#v", tt.want, *got)
			}
			// A null config is no config.
			got, err = getAlertChannel(t, tt.channelType, " null ")
			if err != nil {
				t.Fatal(err)
			}
			if want := (checkly.AlertChannel{ID: 7, Type: tt.channelType}); !reflect.DeepEqual(*got, want) {
				t.Errorf("want no config for a null config, got %#v", *got)
			}
		})
	}
}

func TestAlertChannelConfigDecodingErrors(t *testing.T) {
	t.Parallel()
	// A config needs a known type even when it is null.
	t.Run("unknown type", func(t *testing.T) {
		t.Parallel()
		_, err := getAlertChannel(t, "TEAMS", `null`)
		if want := `GET /v1/alert-channels/7: decoding 200 response: config: unknown alert channel type "TEAMS"`; err == nil || err.Error() != want {
			t.Errorf("want %q, got %v", want, err)
		}
	})
	t.Run("not an object", func(t *testing.T) {
		t.Parallel()
		for raw, kind := range map[string]string{
			`"SECRET-VALUE-MARKER"`:   "string",
			`["SECRET-VALUE-MARKER"]`: "array",
			`42`:                      "number",
			`true`:                    "bool",
		} {
			_, err := getAlertChannel(t, checkly.AlertTypeWebhook, raw)
			want := "GET /v1/alert-channels/7: decoding 200 response: config: decoding WEBHOOK alert channel config: got a JSON " + kind + ", want an object"
			if err == nil || err.Error() != want {
				t.Errorf("%s: want %q, got %v", raw, want, err)
			}
		}
	})
	// A field of an unexpected type is skipped, so that a channel whose
	// config the API changed stays readable.
	t.Run("type mismatch tolerated", func(t *testing.T) {
		t.Parallel()
		got, err := getAlertChannel(t, checkly.AlertTypeSMS, `{"name":"On call","number":15550100}`)
		if err != nil {
			t.Fatal(err)
		}
		if want := (&checkly.AlertChannelSMS{Name: "On call"}); !reflect.DeepEqual(got.SMS, want) {
			t.Errorf("want %#v, got %#v", want, got.SMS)
		}
	})
	t.Run("nested type mismatch tolerated", func(t *testing.T) {
		t.Parallel()
		got, err := getAlertChannel(t, checkly.AlertTypeWebhook, `{"name":"Hook","headers":[{"key":"X-Env","value":7}]}`)
		if err != nil {
			t.Fatal(err)
		}
		want := &checkly.AlertChannelWebhook{Name: "Hook", Headers: []checkly.KeyValue{{Key: "X-Env"}}}
		if !reflect.DeepEqual(got.Webhook, want) {
			t.Errorf("want %#v, got %#v", want, got.Webhook)
		}
	})
}
