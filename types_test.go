package checkly_test

import (
	"testing"

	checkly "github.com/checkly/checkly-go-sdk"
)

func TestAlertChannelEmail(t *testing.T) {
	ac := checkly.AlertChannel{
		Type: checkly.AlertTypeEmail,
	}
	email := "foo@test.com"
	cfg := map[string]interface{}{
		"address": email,
	}

	ac.SetConfig(cfg)

	if ac.Email == nil {
		t.Error("Config shouldn't be nil")
		return
	}

	if ac.Email.Address != email {
		t.Errorf(
			"Expected email to be: `%s`, got: `%s`",
			email,
			ac.Email.Address,
		)
	}
}

func TestAlertChannelSlack(t *testing.T) {
	ac := checkly.AlertChannel{
		Type: checkly.AlertTypeSlack,
	}
	webhookURL := "http://example.com/"
	channel := "foochan"
	cfg := map[string]interface{}{
		"url":     webhookURL,
		"channel": channel,
	}

	ac.SetConfig(cfg)

	if ac.Slack == nil {
		t.Error("Config shouldn't be nil")
		return
	}

	if ac.Slack.WebhookURL != webhookURL {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			webhookURL,
			ac.Slack.WebhookURL,
		)
	}

	if ac.Slack.Channel != channel {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			channel,
			ac.Slack.Channel,
		)
	}
}

func TestAlertChannelSMS(t *testing.T) {
	ac := checkly.AlertChannel{
		Type: checkly.AlertTypeSMS,
	}
	name := "foo"
	number := "0123456789"
	cfg := map[string]interface{}{
		"name":   name,
		"number": number,
	}

	ac.SetConfig(cfg)

	if ac.SMS == nil {
		t.Error("Config shouldn't be nil")
		return
	}

	if ac.SMS.Name != name {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			name,
			ac.SMS.Name,
		)
	}

	if ac.SMS.Number != number {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			number,
			ac.SMS.Number,
		)
	}
}

func TestAlertChannelWebhook(t *testing.T) {
	ac := checkly.AlertChannel{
		Type: checkly.AlertTypeWebhook,
	}

	name := "foo"
	method := "GET"
	headers := []checkly.KeyValue{
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
	}
	queryParameters := []checkly.KeyValue{
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
	}
	template := "footemp"
	url := "http://foo.com"
	webhookSecret := "foosecret"
	cfg := map[string]interface{}{
		"name":            name,
		"method":          method,
		"template":        template,
		"url":             url,
		"webhookSecret":   webhookSecret,
		"headers":         headers,
		"queryParameters": queryParameters,
	}

	ac.SetConfig(cfg)

	if ac.Webhook == nil {
		t.Error("Config shouldn't be nil")
		return
	}

	if ac.Webhook.Name != name {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			name,
			ac.Webhook.Name,
		)
	}

	if ac.Webhook.Method != method {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			method,
			ac.Webhook.Method,
		)
	}

	if ac.Webhook.Template != template {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			template,
			ac.Webhook.Template,
		)
	}

	if ac.Webhook.URL != url {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			url,
			ac.Webhook.URL,
		)
	}

	if ac.Webhook.WebhookSecret != webhookSecret {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			webhookSecret,
			ac.Webhook.WebhookSecret,
		)
	}

	if len(ac.Webhook.Headers) != len(headers) {
		t.Errorf(
			"Expected: %d headers, got: %d headers",
			len(headers),
			len(ac.Webhook.Headers),
		)
	} else {
		if ac.Webhook.Headers[0].Key != headers[0].Key {
			t.Errorf(
				"Expected Headers[0].Key to be: `%s`, got: `%s`",
				headers[0].Key,
				ac.Webhook.Headers[0].Key,
			)
		}
		if ac.Webhook.Headers[0].Value != headers[0].Value {
			t.Errorf(
				"Expected Headers[0].Value to be: `%s`, got: `%s`",
				headers[0].Value,
				ac.Webhook.Headers[0].Value,
			)
		}
		if ac.Webhook.Headers[0].Locked != headers[0].Locked {
			t.Errorf(
				"Expected Headers[0].Locked to be: `%t`, got `%t`",
				headers[0].Locked,
				ac.Webhook.Headers[0].Locked,
			)
		}
	}

	if len(ac.Webhook.QueryParameters) != len(queryParameters) {
		t.Errorf(
			"Expected: %d queryParameters, got: %d queryParameters",
			len(queryParameters),
			len(ac.Webhook.QueryParameters),
		)
	} else {
		if ac.Webhook.QueryParameters[0].Key != queryParameters[0].Key {
			t.Errorf(
				"Expected QueryParameters[0].Key to be: `%s`, got: `%s`",
				queryParameters[0].Key,
				ac.Webhook.QueryParameters[0].Key,
			)
		}
		if ac.Webhook.QueryParameters[0].Value != queryParameters[0].Value {
			t.Errorf(
				"Expected QueryParameters[0].Value to be: `%s`, got: `%s`",
				queryParameters[0].Value,
				ac.Webhook.QueryParameters[0].Value,
			)
		}
		if ac.Webhook.QueryParameters[0].Locked != queryParameters[0].Locked {
			t.Errorf(
				"Expected QueryParameters[0].Locked to be: `%t`, got `%t`",
				queryParameters[0].Locked,
				ac.Webhook.QueryParameters[0].Locked,
			)
		}
	}
}

func TestAlertChannelOpsgenie(t *testing.T) {
	ac := checkly.AlertChannel{
		Type: checkly.AlertTypeOpsgenie,
	}

	name := "foo"
	apiKey := "dkdkjdkd34"
	region := "fooregion"
	priority := "prio1"
	cfg := map[string]interface{}{
		"name":     name,
		"apiKey":   apiKey,
		"region":   region,
		"priority": priority,
	}

	ac.SetConfig(cfg)

	if ac.Opsgenie == nil {
		t.Error("Config shouldn't be nil")
		return
	}

	if ac.Opsgenie.Name != name {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			name,
			ac.Opsgenie.Name,
		)
	}

	if ac.Opsgenie.APIKey != apiKey {
		t.Errorf(
			"Expected: `%s`, got:`%s`",
			apiKey,
			ac.Opsgenie.APIKey,
		)
	}

	if ac.Opsgenie.Region != region {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			region,
			ac.Opsgenie.Region,
		)
	}

	if ac.Opsgenie.Priority != priority {
		t.Errorf(
			"Expected: `%s`, got: `%s`",
			priority,
			ac.Opsgenie.Priority,
		)
	}

}
