package checkly_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/checkly/checkly-go-sdk/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var ignoreAlertChannelFields = cmpopts.IgnoreFields(checkly.AlertChannel{}, "ID")

func getTestAlertChannelEmail() *checkly.AlertChannel {
	return &checkly.AlertChannel{
		ID:   1,
		Type: checkly.AlertChannelTypeEmail,
		Email: &checkly.AlertChannelEmail{
			Address: "test@example.com",
		},
	}
}

func getTestAlertChannelSlack() checkly.AlertChannel {
	ac := checkly.AlertChannel{
		ID:   1,
		Type: checkly.AlertChannelTypeSlack,
		Slack: &checkly.AlertChannelSlack{
			Channel:    "test",
			WebhookURL: "https://slack.com/test",
		},
	}
	return ac
}

// validateAlertChannel checks that body is the request body of the test alert
// channel of its type.
func validateAlertChannel(t *testing.T, body []byte) {
	var req struct {
		Type   checkly.AlertChannelType `json:"type"`
		Config json.RawMessage          `json:"config"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		t.Errorf("decoding error for data %q: %v", body, err)
		return
	}
	switch req.Type {
	case checkly.AlertChannelTypeEmail:
		var got checkly.AlertChannelEmail
		if err := json.Unmarshal(req.Config, &got); err != nil {
			t.Errorf("decoding config %s: %v", req.Config, err)
			return
		}
		if want := *getTestAlertChannelEmail().Email; got != want {
			t.Errorf("want config %+v, got %+v", want, got)
		}
	case checkly.AlertChannelTypeSlack:
		var got checkly.AlertChannelSlack
		if err := json.Unmarshal(req.Config, &got); err != nil {
			t.Errorf("decoding config %s: %v", req.Config, err)
			return
		}
		if want := *getTestAlertChannelSlack().Slack; got != want {
			t.Errorf("want config %+v, got %+v", want, got)
		}
	default:
		t.Errorf("unexpected alert channel type %q", req.Type)
	}
}

func TestCreateAlertChannel(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/alert-channels",
		validateAlertChannel,
		http.StatusOK,
		"CreateAlertChannelEmail.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	ta := getTestAlertChannelEmail()
	ac, err := client.CreateAlertChannel(context.Background(), *ta)
	if err != nil {
		t.Fatal(err)
	}
	ta.CreatedAt = "0001-01-11T00:00:00Z"
	ta.UpdatedAt = "0001-01-01T00:00:00Z"
	if !cmp.Equal(ta, ac, ignoreAlertChannelFields) {
		t.Error(cmp.Diff(ta, ac, ignoreAlertChannelFields))
	}
}

func TestGetAlertChannel(t *testing.T) {
	t.Parallel()
	ta := getTestAlertChannelEmail()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/alert-channels/%d", ta.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetAlertChannelEmail.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	ac, err := client.GetAlertChannel(context.Background(), ta.ID)
	if err != nil {
		t.Fatal(err)
	}
	ta.CreatedAt = "0001-01-13T00:00:00Z"
	ta.UpdatedAt = "0001-01-01T00:00:00Z"
	if !cmp.Equal(*ta, *ac, ignoreAlertChannelFields) {
		t.Error(cmp.Diff(*ta, *ac, ignoreAlertChannelFields))
	}
}

func TestUpdateAlertChannel(t *testing.T) {
	t.Parallel()
	ta := getTestAlertChannelEmail()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/alert-channels/%d", ta.ID),
		validateAlertChannel,
		http.StatusOK,
		"UpdateAlertChannel.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	ac, err := client.UpdateAlertChannel(context.Background(), ta.ID, *ta)
	if err != nil {
		t.Fatal(err)
	}
	ta.CreatedAt = "0001-01-14T00:00:00Z"
	ta.UpdatedAt = "0001-01-01T00:00:00Z"
	if !cmp.Equal(ta, ac, ignoreAlertChannelFields) {
		t.Error(cmp.Diff(ta, ac, ignoreAlertChannelFields))
	}
}

func TestDeleteAlertChannel(t *testing.T) {
	t.Parallel()
	ta := getTestAlertChannelEmail()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/alert-channels/%d", ta.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	err := client.DeleteAlertChannel(context.Background(), ta.ID)
	if err != nil {
		t.Error(err)
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
		_, _ = io.WriteString(w, `{"id":1}`)
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
			ac:   checkly.AlertChannel{Type: checkly.AlertChannelTypeEmail, Email: &checkly.AlertChannelEmail{Address: "foo@test.com"}},
			want: `{"address":"foo@test.com"}`,
		},
		{
			name: "slack",
			ac:   checkly.AlertChannel{Type: checkly.AlertChannelTypeSlack, Slack: &checkly.AlertChannelSlack{WebhookURL: "http://example.com/", Channel: "foochan"}},
			want: `{"url":"http://example.com/","channel":"foochan"}`,
		},
		{
			name: "slack app",
			ac:   checkly.AlertChannel{Type: checkly.AlertChannelTypeSlackApp, SlackApp: &checkly.AlertChannelSlackApp{SlackChannels: []string{"#ops", "@John"}}},
			want: `{"slackChannels":["#ops","@John"]}`,
		},
		{
			name: "sms",
			ac:   checkly.AlertChannel{Type: checkly.AlertChannelTypeSMS, SMS: &checkly.AlertChannelSMS{Name: "foo", Number: "0123456789"}},
			want: `{"name":"foo","number":"0123456789"}`,
		},
		{
			name: "call",
			ac:   checkly.AlertChannel{Type: checkly.AlertChannelTypeCall, CALL: &checkly.AlertChannelCall{Name: "foo", Number: "0123456789"}},
			want: `{"name":"foo","number":"0123456789"}`,
		},
		{
			name: "webhook",
			ac: checkly.AlertChannel{Type: checkly.AlertChannelTypeWebhook, Webhook: &checkly.AlertChannelWebhook{
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
			ac:   checkly.AlertChannel{Type: checkly.AlertChannelTypeOpsgenie, Opsgenie: &checkly.AlertChannelOpsgenie{Name: "foo", APIKey: "bar", Region: "regio-1", Priority: "highp"}},
			want: `{"name":"foo","apiKey":"bar","region":"regio-1","priority":"highp"}`,
		},
		{
			name: "pagerduty",
			ac:   checkly.AlertChannel{Type: checkly.AlertChannelTypePagerduty, Pagerduty: &checkly.AlertChannelPagerduty{Account: "foo", ServiceKey: "xxx", ServiceName: "bar"}},
			want: `{"account":"foo","serviceKey":"xxx","serviceName":"bar"}`,
		},
		{
			name: "nil config",
			ac:   checkly.AlertChannel{Type: checkly.AlertChannelTypeEmail, SMS: &checkly.AlertChannelSMS{Name: "foo"}},
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
func getAlertChannel(t *testing.T, channelType checkly.AlertChannelType, config string) (*checkly.AlertChannel, error) {
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
		channelType checkly.AlertChannelType
		config      string
		want        checkly.AlertChannel
	}{
		{checkly.AlertChannelTypeEmail, fixtureConfig(t, "CreateAlertChannelEmail.json"), checkly.AlertChannel{Email: &checkly.AlertChannelEmail{Address: "test@example.com"}}},
		// The fixture's webhookUrl key is not the struct's url key, so only
		// the channel decodes.
		{checkly.AlertChannelTypeSlack, fixtureConfig(t, "CreateAlertChannelSlack.json"), checkly.AlertChannel{Slack: &checkly.AlertChannelSlack{Channel: "test"}}},
		{checkly.AlertChannelTypeSlackApp, `{"slackChannels":["C123","C456"]}`, checkly.AlertChannel{SlackApp: &checkly.AlertChannelSlackApp{SlackChannels: []string{"C123", "C456"}}}},
		{checkly.AlertChannelTypeSMS, `{"name":"On call","number":"+15550100"}`, checkly.AlertChannel{SMS: &checkly.AlertChannelSMS{Name: "On call", Number: "+15550100"}}},
		{checkly.AlertChannelTypeCall, `{"name":"On call","number":"+15550100"}`, checkly.AlertChannel{CALL: &checkly.AlertChannelCall{Name: "On call", Number: "+15550100"}}},
		{checkly.AlertChannelTypeOpsgenie, `{"name":"Ops","apiKey":"fake-opsgenie-key","region":"EU","priority":"P3"}`, checkly.AlertChannel{Opsgenie: &checkly.AlertChannelOpsgenie{Name: "Ops", APIKey: "fake-opsgenie-key", Region: "EU", Priority: "P3"}}},
		{checkly.AlertChannelTypePagerduty, `{"account":"acme","serviceKey":"fake-service-key","serviceName":"API"}`, checkly.AlertChannel{Pagerduty: &checkly.AlertChannelPagerduty{Account: "acme", ServiceKey: "fake-service-key", ServiceName: "API"}}},
		{
			checkly.AlertChannelTypeWebhook,
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
		t.Run(string(tt.channelType), func(t *testing.T) {
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
			_, err := getAlertChannel(t, checkly.AlertChannelTypeWebhook, raw)
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
		got, err := getAlertChannel(t, checkly.AlertChannelTypeSMS, `{"name":"On call","number":15550100}`)
		if err != nil {
			t.Fatal(err)
		}
		if want := (&checkly.AlertChannelSMS{Name: "On call"}); !reflect.DeepEqual(got.SMS, want) {
			t.Errorf("want %#v, got %#v", want, got.SMS)
		}
	})
	t.Run("nested type mismatch tolerated", func(t *testing.T) {
		t.Parallel()
		got, err := getAlertChannel(t, checkly.AlertChannelTypeWebhook, `{"name":"Hook","headers":[{"key":"X-Env","value":7}]}`)
		if err != nil {
			t.Fatal(err)
		}
		want := &checkly.AlertChannelWebhook{Name: "Hook", Headers: []checkly.KeyValue{{Key: "X-Env"}}}
		if !reflect.DeepEqual(got.Webhook, want) {
			t.Errorf("want %#v, got %#v", want, got.Webhook)
		}
	})
}

// TestAlertChannelDecoding checks how alert channel responses decode: JSON
// null leaves a field unset, a field of the wrong type is an error naming the
// field but not a string value, and errors name the request that was made.
func TestAlertChannelDecoding(t *testing.T) {
	t.Parallel()
	t.Run("null fields", func(t *testing.T) {
		t.Parallel()
		c := newErrorServer(t, errorReply{status: 200, body: `{"id":7,"type":"EMAIL","sendRecovery":null,"sendFailure":true,"sendDegraded":null,"sslExpiry":null,"sslExpiryThreshold":null,"config":{"address":"a@example.com"}}`})
		ac, err := c.GetAlertChannel(context.Background(), 7)
		if err != nil {
			t.Fatal(err)
		}
		if ac.ID != 7 || ac.Type != "EMAIL" || ac.SendRecovery != nil || ac.SendDegraded != nil || ac.SSLExpiry != nil || ac.SSLExpiryThreshold != nil {
			t.Errorf("want null fields unset, got %+v", ac)
		}
		if ac.SendFailure == nil || !*ac.SendFailure {
			t.Errorf("want sendFailure true, got %v", ac.SendFailure)
		}
		if ac.Email == nil || ac.Email.Address != "a@example.com" {
			t.Errorf("want the email config decoded, got %+v", ac.Email)
		}
	})
	t.Run("timestamps", func(t *testing.T) {
		t.Parallel()
		c := newErrorServer(t, errorReply{status: 200, body: `{"id":7,"type":"EMAIL","created_at":"2026-01-02T03:04:05.678Z","updated_at":"2026-02-03T04:05:06.789Z"}`})
		ac, err := c.GetAlertChannel(context.Background(), 7)
		if err != nil {
			t.Fatal(err)
		}
		if ac.CreatedAt != "2026-01-02T03:04:05.678Z" || ac.UpdatedAt != "2026-02-03T04:05:06.789Z" {
			t.Errorf("want the timestamps decoded, got created_at %q, updated_at %q", ac.CreatedAt, ac.UpdatedAt)
		}
	})
	t.Run("null config", func(t *testing.T) {
		t.Parallel()
		c := newErrorServer(t, errorReply{status: 200, body: `{"id":7,"type":"EMAIL","config":null}`})
		ac, err := c.GetAlertChannel(context.Background(), 7)
		if err != nil {
			t.Fatal(err)
		}
		if ac.Email != nil {
			t.Errorf("want no email config for a null config, got %+v", ac.Email)
		}
	})
	t.Run("config not an object", func(t *testing.T) {
		t.Parallel()
		c := newErrorServer(t, errorReply{status: 200, body: `{"id":7,"type":"EMAIL","config":"SECRET-VALUE-MARKER"}`})
		_, err := c.GetAlertChannel(context.Background(), 7)
		want := "GET /v1/alert-channels/7: decoding 200 response: config: decoding EMAIL alert channel config: got a JSON string, want an object"
		if err == nil || err.Error() != want {
			t.Errorf("want %q, got %v", want, err)
		}
	})
	t.Run("unknown type", func(t *testing.T) {
		t.Parallel()
		c := newErrorServer(t, errorReply{status: 200, body: `{"id":7,"type":"TEAMS","config":{}}`})
		_, err := c.GetAlertChannel(context.Background(), 7)
		want := `GET /v1/alert-channels/7: decoding 200 response: config: unknown alert channel type "TEAMS"`
		if err == nil || err.Error() != want {
			t.Errorf("want %q, got %v", want, err)
		}
	})
	tests := []struct {
		name string
		body string
		// want holds substrings of the error. Go 1.24 added the embedded
		// struct's name to the field path of encoding/json type errors, so
		// the path between the two parts depends on the Go version.
		want []string
	}{
		{"wrong bool", `{"type":"EMAIL","sslExpiry":"SECRET-VALUE-MARKER"}`, []string{`cannot unmarshal string into Go struct field alertChannelResponse.`, `.sslExpiry of type bool`}},
		{"wrong type", `{"type":42}`, []string{`cannot unmarshal number into Go struct field alertChannelResponse.`, `.type of type checkly.AlertChannelType`}},
		{"wrong id", `{"id":"SECRET-VALUE-MARKER"}`, []string{`cannot unmarshal string into Go struct field alertChannelResponse.`, `.id of type int64`}},
		{"fractional id", `{"id":1.5}`, []string{`cannot unmarshal number 1.5 into Go struct field alertChannelResponse.`, `.id of type int64`}},
		{"wrong threshold", `{"sslExpiryThreshold":true}`, []string{`cannot unmarshal bool into Go struct field alertChannelResponse.`, `.sslExpiryThreshold of type int`}},
		{"fractional threshold", `{"sslExpiryThreshold":30.5}`, []string{`cannot unmarshal number 30.5 into Go struct field alertChannelResponse.`, `.sslExpiryThreshold of type int`}},
		{"wrong timestamp", `{"created_at":["SECRET-VALUE-MARKER"]}`, []string{`cannot unmarshal array into Go struct field alertChannelResponse.`, `.created_at of type string`}},
		{"not json", `SECRET-VALUE-MARKER`, []string{`decoding 201 response (19 bytes): invalid character`}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := newErrorServer(t, errorReply{status: 201, body: tt.body})
			_, err := c.CreateAlertChannel(context.Background(), checkly.AlertChannel{})
			if err == nil {
				t.Fatal("want error, got nil")
			}
			prefix := "POST /v1/alert-channels: "
			if !strings.HasPrefix(err.Error(), prefix) {
				t.Errorf("want error starting with %q, got %q", prefix, err)
			}
			for _, want := range tt.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("want error containing %q, got %q", want, err)
				}
			}
			if strings.Contains(err.Error(), "SECRET-VALUE-MARKER") {
				t.Errorf("want the field value left out of the error, got %q", err)
			}
		})
	}
}

// The alert channel update resets an omitted autoSubscribe to false, so the
// setting decodes from responses and is sent when set.
func TestAlertChannelAutoSubscribe(t *testing.T) {
	t.Parallel()
	c, lastBody := newRecordingServer(t, http.StatusOK, `{"id":7,"type":"EMAIL","autoSubscribe":true,"config":{"address":"a@example.com"}}`)
	on := true
	ac, err := c.UpdateAlertChannel(context.Background(), 7, checkly.AlertChannel{
		Type:          checkly.AlertChannelTypeEmail,
		Email:         &checkly.AlertChannelEmail{Address: "a@example.com"},
		AutoSubscribe: &on,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertRawBodyField(t, lastBody(), "autoSubscribe", `true`)
	if ac.AutoSubscribe == nil || !*ac.AutoSubscribe {
		t.Errorf("want autoSubscribe true decoded, got %v", ac.AutoSubscribe)
	}

	// The API rejects null, so nil is left out.
	if _, err := c.CreateAlertChannel(context.Background(), checkly.AlertChannel{Type: checkly.AlertChannelTypeEmail}); err != nil {
		t.Fatal(err)
	}
	assertRawBodyKeyAbsent(t, lastBody(), "autoSubscribe")
}
