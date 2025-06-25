package checkly_test

import (
	"bytes"
	"encoding/json"
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

func TestRetryStrategyUnmarshalJSON(t *testing.T) {
	t.Run("null value", func(t *testing.T) {
		raw := []byte(`null`)

		var retryStrategy *checkly.RetryStrategy

		err := json.Unmarshal(raw, &retryStrategy)
		if err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}

		if retryStrategy != nil {
			t.Fatalf("expected nil, got: %v", retryStrategy)
		}
	})

	t.Run("fallback value", func(t *testing.T) {
		raw := []byte(`"FALLBACK"`)

		var retryStrategy *checkly.RetryStrategy

		err := json.Unmarshal(raw, &retryStrategy)
		if err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}

		if retryStrategy.Type != "FALLBACK" {
			t.Fatalf("expected .Type == FALLBACK, got: %v", retryStrategy.Type)
		}
	})

	t.Run("normal value", func(t *testing.T) {
		raw := []byte(`{
			"type": "LINEAR",
			"baseBackoffSeconds": 60,
			"maxRetries": 2,
			"maxDurationSeconds": 600,
			"sameRegion": true
		}`)

		var retryStrategy *checkly.RetryStrategy

		err := json.Unmarshal(raw, &retryStrategy)
		if err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}

		if retryStrategy.Type != "LINEAR" {
			t.Fatalf("expected .Type == LINEAR, got: %v", retryStrategy.Type)
		}

		if retryStrategy.BaseBackoffSeconds != 60 {
			t.Fatalf("expected .BaseBackoffSeconds == 60, got: %v", retryStrategy.BaseBackoffSeconds)
		}

		if retryStrategy.MaxRetries != 2 {
			t.Fatalf("expected .MaxRetries == 2, got: %v", retryStrategy.MaxRetries)
		}

		if retryStrategy.MaxDurationSeconds != 600 {
			t.Fatalf("expected .MaxDurationSeconds == 600, got: %v", retryStrategy.MaxDurationSeconds)
		}

		if retryStrategy.SameRegion != true {
			t.Fatalf("expected .SameRegion == true, got: %v", retryStrategy.SameRegion)
		}
	})
}

func TestRetryStrategyMarshalJSON(t *testing.T) {
	testBytesEqual := func(t *testing.T, got, want []byte) {
		if !bytes.Equal(got, want) {
			t.Fatalf("expected: %s, got: %s", want, got)
		}
	}

	t.Run("null value", func(t *testing.T) {
		var retryStrategy *checkly.RetryStrategy

		raw, err := json.Marshal(&retryStrategy)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}

		testBytesEqual(t, raw, []byte(`null`))
	})

	t.Run("fallback value", func(t *testing.T) {
		retryStrategy := checkly.RetryStrategy{
			Type: "FALLBACK",
		}

		raw, err := json.Marshal(&retryStrategy)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}

		testBytesEqual(t, raw, []byte(`"FALLBACK"`))
	})

	t.Run("normal value", func(t *testing.T) {
		retryStrategy := checkly.RetryStrategy{
			Type:               "LINEAR",
			BaseBackoffSeconds: 60,
			MaxRetries:         2,
			MaxDurationSeconds: 600,
			SameRegion:         true,
		}

		raw, err := json.Marshal(&retryStrategy)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}

		testBytesEqual(t, raw, []byte(`{"type":"LINEAR","baseBackoffSeconds":60,"maxRetries":2,"maxDurationSeconds":600,"sameRegion":true}`))
	})
}
