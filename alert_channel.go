package checkly

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// AlertChannelType is the type of an alert channel, in [AlertChannel].Type.
type AlertChannelType string

// Alert channel types, the values of [AlertChannel].Type. Each selects the
// config field of AlertChannel that writes use and reads set, such as Email
// for AlertChannelTypeEmail.
const (
	AlertChannelTypeEmail AlertChannelType = "EMAIL"

	// AlertChannelTypeSlack identifies legacy Slack webhook alert channels.
	//
	// Deprecated: Use [AlertChannelTypeSlackApp] instead.
	AlertChannelTypeSlack AlertChannelType = "SLACK"

	AlertChannelTypeSlackApp  AlertChannelType = "SLACK_APP"
	AlertChannelTypeWebhook   AlertChannelType = "WEBHOOK"
	AlertChannelTypeSMS       AlertChannelType = "SMS"
	AlertChannelTypePagerduty AlertChannelType = "PAGERDUTY"
	AlertChannelTypeOpsgenie  AlertChannelType = "OPSGENIE"
	AlertChannelTypeCall      AlertChannelType = "CALL"
)

// AlertChannelSubscription subscribes a check or group to the alert channel
// ChannelID, in the AlertChannelSubscriptions field of checks, monitors and
// groups.
type AlertChannelSubscription struct {
	ChannelID int64 `json:"alertChannelId"`
	Activated bool  `json:"activated"`
}

// AlertChannelEmail is the config of an [AlertChannelTypeEmail] alert
// channel.
type AlertChannelEmail struct {
	Address string `json:"address"`
}

// AlertChannelSlack is the config of an [AlertChannelTypeSlack] alert
// channel, which posts to a Slack incoming webhook.
//
// Deprecated: Use [AlertChannelSlackApp] instead. The legacy Slack webhook
// integration is being phased out in favor of the Checkly Slack App.
type AlertChannelSlack struct {
	// WebhookURL is a Slack incoming webhook URL, starting with
	// "https://hooks.slack.com/services/", or a reference to an environment
	// variable such as "{{SLACK_WEBHOOK_URL}}".
	WebhookURL string `json:"url"`
	Channel    string `json:"channel"`
}

// AlertChannelSlackApp is the config of an [AlertChannelTypeSlackApp] alert
// channel, which alerts through the Checkly Slack App.
type AlertChannelSlackApp struct {
	// SlackChannels lists at least one target: a channel such as
	// "#operations", or a user such as "@tim" for a direct message.
	SlackChannels []string `json:"slackChannels"`
}

// AlertChannelSMS is the config of an [AlertChannelTypeSMS] alert channel.
// The API rejects a Number in a country it does not support.
type AlertChannelSMS struct {
	Name   string `json:"name"`
	Number string `json:"number"`
}

// AlertChannelCall is the config of an [AlertChannelTypeCall] alert channel,
// which alerts by phone call. The API rejects a Number in a country it does
// not support.
type AlertChannelCall struct {
	Name   string `json:"name"`
	Number string `json:"number"`
}

// OpsgenieRegion is the Opsgenie region of an Opsgenie alert channel, in
// [AlertChannelOpsgenie].Region.
type OpsgenieRegion string

// Opsgenie regions, the values of [AlertChannelOpsgenie].Region.
const (
	OpsgenieRegionEU OpsgenieRegion = "EU"
	OpsgenieRegionUS OpsgenieRegion = "US"
)

// OpsgeniePriority is the priority of the alerts an Opsgenie alert channel
// creates, in [AlertChannelOpsgenie].Priority.
type OpsgeniePriority string

// Opsgenie priorities, the values of [AlertChannelOpsgenie].Priority.
const (
	OpsgeniePriorityP1 OpsgeniePriority = "P1"
	OpsgeniePriorityP2 OpsgeniePriority = "P2"
	OpsgeniePriorityP3 OpsgeniePriority = "P3"
	OpsgeniePriorityP4 OpsgeniePriority = "P4"
	OpsgeniePriorityP5 OpsgeniePriority = "P5"
)

// AlertChannelOpsgenie is the config of an [AlertChannelTypeOpsgenie] alert
// channel. Every field is required.
type AlertChannelOpsgenie struct {
	Name     string           `json:"name"`
	APIKey   string           `json:"apiKey"`
	Region   OpsgenieRegion   `json:"region"`
	Priority OpsgeniePriority `json:"priority"`
}

// AlertChannelPagerduty is the config of an [AlertChannelTypePagerduty]
// alert channel. Only ServiceKey is required.
type AlertChannelPagerduty struct {
	Account     string `json:"account,omitempty"`
	ServiceKey  string `json:"serviceKey"`
	ServiceName string `json:"serviceName,omitempty"`
}

// AlertChannelWebhook is the config of an [AlertChannelTypeWebhook] alert
// channel, which sends an HTTP request for each alert. Name and URL are
// required.
type AlertChannelWebhook struct {
	Name string `json:"name"`
	URL  string `json:"url"`

	// WebhookType is the kind of webhook, such as "WEBHOOK_TELEGRAM".
	WebhookType string `json:"webhookType,omitempty"`

	// Method is GET, POST, PUT, HEAD, DELETE or PATCH, as the net/http
	// constants such as http.MethodPost name them. Empty means POST.
	Method string `json:"method,omitempty"`

	// Template is the request body, with placeholders such as
	// "{{ALERT_TITLE}}" and "{{RESULT_LINK}}".
	Template        string     `json:"template,omitempty"`
	WebhookSecret   string     `json:"webhookSecret,omitempty"`
	Headers         []KeyValue `json:"headers,omitempty"`
	QueryParameters []KeyValue `json:"queryParameters,omitempty"`
}

// AlertChannel is a destination for alerts, such as an email address or a
// Slack channel. It is the input of CreateAlertChannel and
// UpdateAlertChannel, which use the config field that matches Type, and
// their result.
type AlertChannel struct {
	ID   int64            `json:"id,omitempty"`
	Type AlertChannelType `json:"type"`

	// Email, Slack, SlackApp, SMS, CALL, Opsgenie, Webhook and Pagerduty
	// hold the config of the channel. Only the one that matches Type is
	// used, and a read sets only that one.
	Email     *AlertChannelEmail     `json:"-"`
	Slack     *AlertChannelSlack     `json:"-"`
	SlackApp  *AlertChannelSlackApp  `json:"-"`
	SMS       *AlertChannelSMS       `json:"-"`
	CALL      *AlertChannelCall      `json:"-"`
	Opsgenie  *AlertChannelOpsgenie  `json:"-"`
	Webhook   *AlertChannelWebhook   `json:"-"`
	Pagerduty *AlertChannelPagerduty `json:"-"`

	// SendRecovery, SendFailure and SendDegraded select which alerts the
	// channel sends: on recovery, on failure, and on degraded results. Nil
	// means the API's default on create, and an update with nil keeps the
	// stored value.
	SendRecovery *bool `json:"sendRecovery"`
	SendFailure  *bool `json:"sendFailure"`
	SendDegraded *bool `json:"sendDegraded"`

	// SSLExpiry enables alerts for expiring SSL certificates, and
	// SSLExpiryThreshold, between 1 and 30, sets when they start. Nil means
	// the API's defaults, false and 30, also on update.
	SSLExpiry          *bool  `json:"sslExpiry"`
	SSLExpiryThreshold *int   `json:"sslExpiryThreshold"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`

	// AutoSubscribe subscribes checks created later to the channel. Nil
	// means the API's default, false, also on update.
	AutoSubscribe *bool `json:"autoSubscribe"`
}

// configField returns a pointer to the config field that matches a.Type, such
// as &a.Email, or nil for an unknown type. Marshaling it yields the config, or
// null when the field is nil; unmarshaling into it sets the field, to nil for
// JSON null.
func (a *AlertChannel) configField() any {
	switch a.Type {
	case AlertChannelTypeEmail:
		return &a.Email
	case AlertChannelTypeSMS:
		return &a.SMS
	case AlertChannelTypeCall:
		return &a.CALL
	case AlertChannelTypeSlack:
		return &a.Slack
	case AlertChannelTypeSlackApp:
		return &a.SlackApp
	case AlertChannelTypeOpsgenie:
		return &a.Opsgenie
	case AlertChannelTypePagerduty:
		return &a.Pagerduty
	case AlertChannelTypeWebhook:
		return &a.Webhook
	}
	return nil
}

func (c *client) CreateAlertChannel(
	ctx context.Context,
	ac AlertChannel,
) (*AlertChannel, error) {
	res, err := c.call(ctx, http.MethodPost, "/v1/alert-channels", createAlertChannelPayload(ac), http.StatusOK, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return alertChannelFromJSON(res)
}

func (c *client) GetAlertChannel(
	ctx context.Context,
	id int64,
) (*AlertChannel, error) {
	res, err := c.call(ctx, http.MethodGet, fmt.Sprintf("/v1/alert-channels/%d", id), nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return alertChannelFromJSON(res)
}

func (c *client) UpdateAlertChannel(
	ctx context.Context,
	id int64,
	ac AlertChannel,
) (*AlertChannel, error) {
	res, err := c.call(ctx, http.MethodPut, fmt.Sprintf("/v1/alert-channels/%d", id), createAlertChannelPayload(ac), http.StatusOK)
	if err != nil {
		return nil, err
	}
	return alertChannelFromJSON(res)
}

func (c *client) DeleteAlertChannel(
	ctx context.Context,
	id int64,
) error {
	_, err := c.call(ctx, http.MethodDelete, fmt.Sprintf("/v1/alert-channels/%d", id), nil, http.StatusNoContent)
	return err
}

// alertChannelPayload is the request body of CreateAlertChannel and
// UpdateAlertChannel. Unset settings are omitted, because the API rejects JSON
// null for them. The ID is not sent, because the API strips it from the body.
type alertChannelPayload struct {
	Type               AlertChannelType `json:"type"`
	Config             any              `json:"config"`
	SendRecovery       *bool            `json:"sendRecovery,omitempty"`
	SendFailure        *bool            `json:"sendFailure,omitempty"`
	SendDegraded       *bool            `json:"sendDegraded,omitempty"`
	SSLExpiry          *bool            `json:"sslExpiry,omitempty"`
	SSLExpiryThreshold *int             `json:"sslExpiryThreshold,omitempty"`
	AutoSubscribe      *bool            `json:"autoSubscribe,omitempty"`
}

// createAlertChannelPayload returns the request body for ac. The config is the
// config field that matches ac.Type, sent as null when that field is nil or the
// type is unknown.
func createAlertChannelPayload(ac AlertChannel) alertChannelPayload {
	return alertChannelPayload{
		Type:               ac.Type,
		Config:             ac.configField(),
		SendRecovery:       ac.SendRecovery,
		SendFailure:        ac.SendFailure,
		SendDegraded:       ac.SendDegraded,
		SSLExpiry:          ac.SSLExpiry,
		SSLExpiryThreshold: ac.SSLExpiryThreshold,
		AutoSubscribe:      ac.AutoSubscribe,
	}
}

// alertChannelResponse is the response body of the alert channel methods. The
// config is kept raw, because its type depends on the channel's type; the
// config fields of AlertChannel are not decoded from JSON.
type alertChannelResponse struct {
	AlertChannel
	Config json.RawMessage `json:"config"`
}

// alertChannelFromJSON decodes res, a response holding an alert channel. A
// field that is JSON null is left unset; one of another type than expected is
// an error that names the field but never quotes a string value, because
// alert channel configs hold secrets. A config, including a null one, needs a
// known channel type; it is decoded by decodeAlertChannelConfig.
func alertChannelFromJSON(res rawResponse) (*AlertChannel, error) {
	var r alertChannelResponse
	if err := decode(res, &r); err != nil {
		return nil, err
	}
	ac := &r.AlertChannel
	if r.Config != nil {
		if err := decodeAlertChannelConfig(ac, r.Config); err != nil {
			return nil, fmt.Errorf("%s %s: decoding %d response: config: %w", res.method, res.path, res.status, err)
		}
	}
	return ac, nil
}

// decodeAlertChannelConfig decodes cfgJSON into the config field of ac that
// matches ac.Type. A field of the config whose JSON type does not match is
// skipped, so that an unexpected value does not make an existing channel
// unreadable. The errors never contain the config's values.
func decodeAlertChannelConfig(ac *AlertChannel, cfgJSON json.RawMessage) error {
	field := ac.configField()
	if field == nil {
		return fmt.Errorf("unknown alert channel type %q", ac.Type)
	}
	// json.Unmarshal keeps decoding after a type mismatch and reports the
	// first one as an *json.UnmarshalTypeError, so the config holds every
	// field that does match. A mismatch without a field is the config itself
	// being no object.
	if err := json.Unmarshal(cfgJSON, field); err != nil {
		var typeErr *json.UnmarshalTypeError
		if !errors.As(err, &typeErr) {
			return fmt.Errorf("decoding %s alert channel config: %w", ac.Type, err)
		}
		if typeErr.Field == "" {
			return fmt.Errorf("decoding %s alert channel config: got a JSON %s, want an object", ac.Type, typeErr.Value)
		}
	}
	return nil
}
