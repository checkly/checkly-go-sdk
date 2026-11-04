package checkly_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	checkly "github.com/checkly/checkly-go-sdk"
)

func TestAlertChannelEmail(t *testing.T) {
	ac := checkly.AlertChannel{
		Type: checkly.AlertTypeEmail,
	}
	cfg := checkly.AlertChannelEmail{
		Address: "foo@test.com",
	}

	ac.SetConfig(&cfg)

	if ac.Email == nil {
		t.Error("Config shouldn't be nil")
		return
	}

	if ac.Email.Address != cfg.Address {
		t.Errorf(
			"Expected email to be: `%s`, got: `%s`",
			cfg.Address,
			ac.Email.Address,
		)
	}
}

func TestAlertChannelSlack(t *testing.T) {
	ac := checkly.AlertChannel{
		Type: checkly.AlertTypeSlack,
	}
	cfg := checkly.AlertChannelSlack{
		WebhookURL: "http://example.com/",
		Channel:    "foochan",
	}

	ac.SetConfig(&cfg)

	if ac.Slack == nil {
		t.Error("Config shouldn't be nil")
		return
	}

	if ac.Slack.WebhookURL != cfg.WebhookURL {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			cfg.WebhookURL,
			ac.Slack.WebhookURL,
		)
	}

	if ac.Slack.Channel != cfg.Channel {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			cfg.Channel,
			ac.Slack.Channel,
		)
	}
}

func TestAlertChannelSlackApp(t *testing.T) {
	ac := checkly.AlertChannel{
		Type: checkly.AlertTypeSlackApp,
	}
	cfg := checkly.AlertChannelSlackApp{
		SlackChannels: []string{"#ops", "@John"},
	}

	ac.SetConfig(&cfg)

	if ac.SlackApp == nil {
		t.Error("Config shouldn't be nil")
		return
	}

	if !reflect.DeepEqual(ac.SlackApp.SlackChannels, cfg.SlackChannels) {
		t.Errorf(
			"Expected: %v, got: %v",
			cfg.SlackChannels,
			ac.SlackApp.SlackChannels,
		)
	}

	got := ac.GetConfig()
	channels, ok := got["slackChannels"].([]interface{})
	if !ok {
		t.Fatalf("Expected slackChannels in config, got: %v", got)
	}
	if len(channels) != 2 || channels[0] != "#ops" || channels[1] != "@John" {
		t.Errorf("Unexpected slackChannels: %v", channels)
	}
}

func TestAlertChannelSlackAppConfigFromJSON(t *testing.T) {
	raw := []byte(`{"slackChannels":["C123","C456"]}`)
	v, err := checkly.AlertChannelConfigFromJSON(checkly.AlertTypeSlackApp, raw)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	got, ok := v.(*checkly.AlertChannelSlackApp)
	if !ok {
		t.Fatalf("Expected *AlertChannelSlackApp, got %T", v)
	}
	want := &checkly.AlertChannelSlackApp{
		SlackChannels: []string{"C123", "C456"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestAlertChannelSMS(t *testing.T) {
	ac := checkly.AlertChannel{
		Type: checkly.AlertTypeSMS,
	}
	cfg := checkly.AlertChannelSMS{
		Name:   "foo",
		Number: "0123456789",
	}

	ac.SetConfig(&cfg)

	if ac.SMS == nil {
		t.Error("Config shouldn't be nil")
		return
	}

	if ac.SMS.Name != cfg.Name {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			cfg.Name,
			ac.SMS.Name,
		)
	}

	if ac.SMS.Number != cfg.Number {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			cfg.Number,
			ac.SMS.Number,
		)
	}
}

func TestAlertChannelCALL(t *testing.T) {
	ac := checkly.AlertChannel{
		Type: checkly.AlertTypeCall,
	}
	cfg := checkly.AlertChannelCall{
		Name:   "foo",
		Number: "0123456789",
	}

	ac.SetConfig(&cfg)

	if ac.CALL == nil {
		t.Error("Config shouldn't be nil")
		return
	}

	if ac.CALL.Name != cfg.Name {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			cfg.Name,
			ac.CALL.Name,
		)
	}

	if ac.CALL.Number != cfg.Number {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			cfg.Number,
			ac.CALL.Number,
		)
	}
}

func TestAlertChannelWebhook(t *testing.T) {
	ac := checkly.AlertChannel{
		Type: checkly.AlertTypeWebhook,
	}
	cfg := checkly.AlertChannelWebhook{
		Name:          "foo",
		Method:        "GET",
		Template:      "bar",
		URL:           "http://foo.com",
		WebhookSecret: "scrt",
		Headers: []checkly.KeyValue{
			{
				Key:    "fookey",
				Value:  "fooval",
				Locked: false,
			},
			{
				Key:    "barkey",
				Value:  "barval",
				Locked: true,
			},
		},
		QueryParameters: []checkly.KeyValue{
			{
				Key:    "fookey",
				Value:  "fooval",
				Locked: true,
			},
			{
				Key:    "barkey",
				Value:  "barval",
				Locked: true,
			},
		},
	}

	ac.SetConfig(&cfg)

	if ac.Webhook == nil {
		t.Error("Config shouldn't be nil")
		return
	}

	if ac.Webhook.Name != cfg.Name {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			cfg.Name,
			ac.Webhook.Name,
		)
	}

	if ac.Webhook.Method != cfg.Method {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			cfg.Method,
			ac.Webhook.Method,
		)
	}

	if ac.Webhook.Template != cfg.Template {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			cfg.Template,
			ac.Webhook.Template,
		)
	}

	if ac.Webhook.URL != cfg.URL {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			cfg.URL,
			ac.Webhook.URL,
		)
	}

	if ac.Webhook.WebhookSecret != cfg.WebhookSecret {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			cfg.WebhookSecret,
			ac.Webhook.WebhookSecret,
		)
	}

	if len(ac.Webhook.Headers) != len(cfg.Headers) {
		t.Errorf(
			"Expected: %d headers, got: %d headers",
			len(cfg.Headers),
			len(ac.Webhook.Headers),
		)
	} else {
		if ac.Webhook.Headers[0].Key != cfg.Headers[0].Key {
			t.Errorf(
				"Expected Headers[0].Key to be: `%s`, got: `%s`",
				cfg.Headers[0].Key,
				ac.Webhook.Headers[0].Key,
			)
		}
		if ac.Webhook.Headers[0].Value != cfg.Headers[0].Value {
			t.Errorf(
				"Expected Headers[0].Value to be: `%s`, got: `%s`",
				cfg.Headers[0].Value,
				ac.Webhook.Headers[0].Value,
			)
		}
		if ac.Webhook.Headers[0].Locked != cfg.Headers[0].Locked {
			t.Errorf(
				"Expected Headers[0].Locked to be: `%t`, got `%t`",
				cfg.Headers[0].Locked,
				ac.Webhook.Headers[0].Locked,
			)
		}
	}

	if len(ac.Webhook.QueryParameters) != len(cfg.QueryParameters) {
		t.Errorf(
			"Expected: %d queryParameters, got: %d queryParameters",
			len(cfg.QueryParameters),
			len(ac.Webhook.QueryParameters),
		)
	} else {
		if ac.Webhook.QueryParameters[0].Key != cfg.QueryParameters[0].Key {
			t.Errorf(
				"Expected QueryParameters[0].Key to be: `%s`, got: `%s`",
				cfg.QueryParameters[0].Key,
				ac.Webhook.QueryParameters[0].Key,
			)
		}
		if ac.Webhook.QueryParameters[0].Value != cfg.QueryParameters[0].Value {
			t.Errorf(
				"Expected QueryParameters[0].Value to be: `%s`, got: `%s`",
				cfg.QueryParameters[0].Value,
				ac.Webhook.QueryParameters[0].Value,
			)
		}
		if ac.Webhook.QueryParameters[0].Locked != cfg.QueryParameters[0].Locked {
			t.Errorf(
				"Expected QueryParameters[0].Locked to be: `%t`, got `%t`",
				cfg.QueryParameters[0].Locked,
				ac.Webhook.QueryParameters[0].Locked,
			)
		}
	}
}

func TestAlertChannelOpsgenie(t *testing.T) {
	ac := checkly.AlertChannel{
		Type: checkly.AlertTypeOpsgenie,
	}

	cfg := checkly.AlertChannelOpsgenie{
		Name:     "foo",
		APIKey:   "bar",
		Region:   "regio-1",
		Priority: "highp",
	}
	ac.SetConfig(&cfg)

	if ac.Opsgenie == nil {
		t.Error("Config shouldn't be nil")
		return
	}

	if ac.Opsgenie.Name != cfg.Name {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			cfg.Name,
			ac.Opsgenie.Name,
		)
	}

	if ac.Opsgenie.APIKey != cfg.APIKey {
		t.Errorf(
			"Expected: `%s`, got:`%s`",
			cfg.APIKey,
			ac.Opsgenie.APIKey,
		)
	}

	if ac.Opsgenie.Region != cfg.Region {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			cfg.Region,
			ac.Opsgenie.Region,
		)
	}

	if ac.Opsgenie.Priority != cfg.Priority {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			cfg.Priority,
			ac.Opsgenie.Priority,
		)
	}
}

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

func TestAlertChannelPagerduty(t *testing.T) {
	ac := checkly.AlertChannel{
		Type: checkly.AlertTypePagerduty,
	}

	cfg := checkly.AlertChannelPagerduty{
		Account:     "foo",
		ServiceKey:  "xxx",
		ServiceName: "bar",
	}
	ac.SetConfig(&cfg)

	if ac.Pagerduty == nil {
		t.Error("Config shouldn't be nil")
		return
	}

	if ac.Pagerduty.Account != cfg.Account {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			cfg.Account,
			ac.Pagerduty.Account,
		)
	}

	if ac.Pagerduty.ServiceKey != cfg.ServiceKey {
		t.Errorf(
			"Expected: `%s`, got:`%s`",
			cfg.ServiceKey,
			ac.Pagerduty.ServiceKey,
		)
	}

	if ac.Pagerduty.ServiceName != cfg.ServiceName {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			cfg.ServiceName,
			ac.Pagerduty.ServiceName,
		)
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

// TestAlertChannelConfigFromJSONTypes decodes a realistic config of every
// alert channel type. The email and Slack configs come from the fixtures.
func TestAlertChannelConfigFromJSONTypes(t *testing.T) {
	tests := []struct {
		channelType string
		config      string
		want        interface{}
	}{
		{checkly.AlertTypeEmail, fixtureConfig(t, "CreateAlertChannelEmail.json"), &checkly.AlertChannelEmail{Address: "test@example.com"}},
		// The fixture's webhookUrl key is not the struct's url key, so only
		// the channel decodes.
		{checkly.AlertTypeSlack, fixtureConfig(t, "CreateAlertChannelSlack.json"), &checkly.AlertChannelSlack{Channel: "test"}},
		{checkly.AlertTypeSlackApp, `{"slackChannels":["C123","C456"]}`, &checkly.AlertChannelSlackApp{SlackChannels: []string{"C123", "C456"}}},
		{checkly.AlertTypeSMS, `{"name":"On call","number":"+15550100"}`, &checkly.AlertChannelSMS{Name: "On call", Number: "+15550100"}},
		{checkly.AlertTypeCall, `{"name":"On call","number":"+15550100"}`, &checkly.AlertChannelCall{Name: "On call", Number: "+15550100"}},
		{checkly.AlertTypeOpsgenie, `{"name":"Ops","apiKey":"fake-opsgenie-key","region":"EU","priority":"P3"}`, &checkly.AlertChannelOpsgenie{Name: "Ops", APIKey: "fake-opsgenie-key", Region: "EU", Priority: "P3"}},
		{checkly.AlertTypePagerduty, `{"account":"acme","serviceKey":"fake-service-key","serviceName":"API"}`, &checkly.AlertChannelPagerduty{Account: "acme", ServiceKey: "fake-service-key", ServiceName: "API"}},
		{
			checkly.AlertTypeWebhook,
			`{"name":"Hook","url":"https://example.com/hook","webhookType":"WEBHOOK_GENERIC","method":"POST","template":"{}","webhookSecret":"fake-secret","headers":[{"key":"X-Env","value":"prod","locked":false}],"queryParameters":[]}`,
			&checkly.AlertChannelWebhook{
				Name:            "Hook",
				URL:             "https://example.com/hook",
				WebhookType:     "WEBHOOK_GENERIC",
				Method:          "POST",
				Template:        "{}",
				WebhookSecret:   "fake-secret",
				Headers:         []checkly.KeyValue{{Key: "X-Env", Value: "prod"}},
				QueryParameters: []checkly.KeyValue{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.channelType, func(t *testing.T) {
			got, err := checkly.AlertChannelConfigFromJSON(tt.channelType, []byte(tt.config))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("want %#v, got %#v", tt.want, got)
			}
			// A null config is no config.
			got, err = checkly.AlertChannelConfigFromJSON(tt.channelType, []byte(" null "))
			if err != nil || got != nil {
				t.Errorf("want nil for a null config, got %#v, %v", got, err)
			}
		})
	}
}

func TestAlertChannelConfigFromJSONErrors(t *testing.T) {
	t.Run("unknown type", func(t *testing.T) {
		_, err := checkly.AlertChannelConfigFromJSON("TEAMS", []byte(`{}`))
		if want := `unknown alert channel type "TEAMS"`; err == nil || err.Error() != want {
			t.Errorf("want %q, got %v", want, err)
		}
	})
	t.Run("syntax error", func(t *testing.T) {
		_, err := checkly.AlertChannelConfigFromJSON(checkly.AlertTypeOpsgenie, []byte(`{"apiKey":"SECRET-VALUE-MARKER"`))
		want := "decoding OPSGENIE alert channel config: "
		if err == nil || !strings.HasPrefix(err.Error(), want) {
			t.Fatalf("want error starting with %q, got %v", want, err)
		}
		if strings.Contains(err.Error(), "SECRET-VALUE-MARKER") {
			t.Errorf("want the config left out of the error, got %q", err)
		}
	})
	t.Run("not an object", func(t *testing.T) {
		for raw, kind := range map[string]string{
			`"SECRET-VALUE-MARKER"`:   "string",
			`["SECRET-VALUE-MARKER"]`: "array",
			`42`:                      "number",
			`true`:                    "bool",
		} {
			_, err := checkly.AlertChannelConfigFromJSON(checkly.AlertTypeWebhook, []byte(raw))
			want := "decoding WEBHOOK alert channel config: got a JSON " + kind + ", want an object"
			if err == nil || err.Error() != want {
				t.Errorf("%s: want %q, got %v", raw, want, err)
			}
		}
	})
	// A field of an unexpected type is skipped, so that a channel whose
	// config the API changed stays readable.
	t.Run("type mismatch tolerated", func(t *testing.T) {
		got, err := checkly.AlertChannelConfigFromJSON(checkly.AlertTypeSMS, []byte(`{"name":"On call","number":15550100}`))
		if err != nil {
			t.Fatal(err)
		}
		if want := (&checkly.AlertChannelSMS{Name: "On call"}); !reflect.DeepEqual(got, want) {
			t.Errorf("want %#v, got %#v", want, got)
		}
	})
	t.Run("nested type mismatch tolerated", func(t *testing.T) {
		got, err := checkly.AlertChannelConfigFromJSON(checkly.AlertTypeWebhook, []byte(`{"name":"Hook","headers":[{"key":"X-Env","value":7}]}`))
		if err != nil {
			t.Fatal(err)
		}
		want := &checkly.AlertChannelWebhook{Name: "Hook", Headers: []checkly.KeyValue{{Key: "X-Env"}}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("want %#v, got %#v", want, got)
		}
	})
}
