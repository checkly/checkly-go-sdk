package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// DNSMonitor sends a DNS query and checks the answer against its assertions.
// It is the input of [Client.CreateDNSMonitor] and [Client.UpdateDNSMonitor],
// and is returned by [Client.GetDNSMonitor].
type DNSMonitor struct {
	// ID is the monitor's UUID, assigned by the API.
	ID          string  `json:"id,omitempty"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	// Frequency is the number of minutes between runs: 0, 1, 2, 5, 10, 15,
	// 30, 60, 120, 180, 360, 720 or 1440.
	Frequency int `json:"frequency"`
	// FrequencyOffset is 10, 20 or 30 when Frequency is 0, for a run every
	// that many seconds. For a higher Frequency the API assigns the offset,
	// but rejects a value sent above Frequency * 10 for a Frequency up to 60,
	// or above Frequency / 5 rounded up for a higher one. FrequencyOffset is
	// left out of the request when zero.
	FrequencyOffset int  `json:"frequencyOffset,omitempty"`
	Activated       bool `json:"activated"`
	// Muted stops notifications when the monitor fails or recovers.
	Muted bool `json:"muted"`
	// RunParallel runs the monitor in all its locations in parallel.
	RunParallel bool `json:"runParallel"`
	// Locations are the public locations to run from, as region codes such
	// as "us-east-1". Nil is sent as an empty list.
	Locations []string `json:"locations"`
	// DegradedResponseTime and MaxResponseTime are the degraded and failing
	// response time limits in milliseconds, each at most 5000.
	DegradedResponseTime int      `json:"degradedResponseTime,omitempty"`
	MaxResponseTime      int      `json:"maxResponseTime,omitempty"`
	Tags                 []string `json:"tags,omitempty"`
	// AlertSettings is left out of the request when nil.
	AlertSettings *AlertSettings `json:"alertSettings,omitempty"`
	// UseGlobalAlertSettings applies the account's alert settings instead
	// of AlertSettings.
	UseGlobalAlertSettings bool       `json:"useGlobalAlertSettings"`
	Request                DNSRequest `json:"request"`
	// GroupID is the ID of the check group the monitor belongs to; 0 means
	// none.
	GroupID int64 `json:"groupId,omitempty"`
	// GroupOrder is the monitor's position in its group, which sets the run
	// order when the group is triggered from the API or CI/CD.
	GroupOrder int `json:"groupOrder,omitempty"`
	// AlertChannelSubscriptions are the alert channels that receive the
	// monitor's alerts. The SDK disables automatic assignment, so an
	// empty list leaves the monitor without any.
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	// PrivateLocations are the slug names of the account's private
	// locations to run from. Nil is sent as an empty list. The Checkly agent
	// at those locations must be version 7.0.0 or later.
	PrivateLocations *[]string `json:"privateLocations"`
	// RetryStrategy sets how failed runs are retried; nil disables retries.
	// In a group whose RetryStrategy is not [RetryStrategyTypeFallback], the
	// group's strategy applies instead.
	RetryStrategy *RetryStrategy `json:"retryStrategy"`
	// TriggerIncident makes the monitor open and resolve an incident on a
	// status page service, based on its alert configuration. Nil turns it
	// off.
	TriggerIncident *IncidentTrigger `json:"triggerIncident"`
	CreatedAt       time.Time        `json:"created_at,omitempty"`
	UpdatedAt       time.Time        `json:"updated_at,omitempty"`
}

func (*DNSMonitor) isCheck() {}

// DNSRecordType is the DNS record type a DNS monitor queries, in
// [DNSRequest].RecordType.
type DNSRecordType string

// DNS record types, the values of [DNSRequest].RecordType.
const (
	DNSRecordTypeA     DNSRecordType = "A"
	DNSRecordTypeAAAA  DNSRecordType = "AAAA"
	DNSRecordTypeCNAME DNSRecordType = "CNAME"
	DNSRecordTypeMX    DNSRecordType = "MX"
	DNSRecordTypeTXT   DNSRecordType = "TXT"
	DNSRecordTypeSOA   DNSRecordType = "SOA"
	DNSRecordTypeNS    DNSRecordType = "NS"
	DNSRecordTypeHTTPS DNSRecordType = "HTTPS"
)

// DNSProtocol is the transport protocol of a DNS monitor's query, in
// [DNSRequest].Protocol.
type DNSProtocol string

// DNS protocols, the values of [DNSRequest].Protocol. The API default is
// DNSProtocolUDP.
const (
	DNSProtocolUDP DNSProtocol = "UDP"
	DNSProtocolTCP DNSProtocol = "TCP"
)

// DNSRequest is the query of a [DNSMonitor].
type DNSRequest struct {
	RecordType DNSRecordType `json:"recordType"`
	Query      string        `json:"query"`
	// NameServer and Port select the name server to query. Set both or
	// neither. NameServer is left out of the request when empty. Port is
	// left out of the request when zero.
	NameServer string `json:"nameServer,omitempty"`
	Port       int    `json:"port,omitempty"`
	// Protocol is left out of the request when empty.
	Protocol   DNSProtocol `json:"protocol,omitempty"`
	Assertions []Assertion `json:"assertions"`
}

type dnsMonitorPayload struct {
	DNSMonitor
	Type        CheckType  `json:"checkType"`
	DoubleCheck bool       `json:"doubleCheck"`
	ShouldFail  bool       `json:"shouldFail"`
	GroupID     *int64     `json:"groupId"`
	ID          *string    `json:"id,omitempty"`         // Skip, can't be changed.
	CreatedAt   *time.Time `json:"created_at,omitempty"` // Skip, can't be changed.
	UpdatedAt   *time.Time `json:"updated_at,omitempty"` // Skip, can't be changed.
	// The fields below shadow embedded fields whose zero value would be
	// omitted so that they are always sent and can actually be cleared via
	// an update.
	Tags                      []string                   `json:"tags"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions"`
	DegradedResponseTime      int                        `json:"degradedResponseTime"`
	MaxResponseTime           int                        `json:"maxResponseTime"`
	GroupOrder                *int                       `json:"groupOrder"`
}

func createDNSMonitorPayload(monitor DNSMonitor) dnsMonitorPayload {
	payload := dnsMonitorPayload{
		DNSMonitor: monitor,
		// Unfortunately `checkType` is required for this endpoint.
		Type: CheckTypeDNS,
		// Unfortunately, this will default to true if not set.
		DoubleCheck: false,
		// The monitor type has no ShouldFail field, so it is always sent as
		// false, which is also the API default.
		ShouldFail: false,
	}

	// The API rejects null for the assertion list.
	payload.Request.Assertions = emptyIfNil(monitor.Request.Assertions)

	// GroupID must be null if empty or the group will not get unset on update.
	if monitor.GroupID != 0 {
		payload.GroupID = &monitor.GroupID
	}

	// A nil value for a list will cause the backend to not update the value.
	// We must send empty lists instead.
	if monitor.Locations == nil {
		payload.Locations = []string{}
	}

	if monitor.PrivateLocations == nil {
		payload.PrivateLocations = &[]string{}
	}

	payload.Tags = emptyIfNil(monitor.Tags)

	payload.AlertChannelSubscriptions = emptyIfNil(monitor.AlertChannelSubscriptions)

	// The response time limits must always be sent: 0 is a valid value, and
	// omitting the keys would make the backend default them on create and
	// keep the previous values on update.
	payload.DegradedResponseTime = monitor.DegradedResponseTime
	payload.MaxResponseTime = monitor.MaxResponseTime

	// GroupOrder must be null if empty or the order will not get unset on
	// update. Note that the check and TCP monitor endpoints differ: they
	// accept and reset on a literal 0.
	payload.GroupOrder = nullIfZero(monitor.GroupOrder)

	return payload
}

func (c *client) CreateDNSMonitor(
	ctx context.Context,
	monitor DNSMonitor,
) (*DNSMonitor, error) {
	var result DNSMonitor
	err := c.callJSON(ctx, http.MethodPost, withAutoAssignAlertsFlag("/v1/checks/dns"), createDNSMonitorPayload(monitor), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetDNSMonitor(
	ctx context.Context,
	id string,
) (*DNSMonitor, error) {
	var result DNSMonitor
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s", id), CheckTypeDNS, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateDNSMonitor(
	ctx context.Context,
	id string,
	monitor DNSMonitor,
) (*DNSMonitor, error) {
	var result DNSMonitor
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/dns/%s", id)), createDNSMonitorPayload(monitor), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteDNSMonitor(
	ctx context.Context,
	id string,
) error {
	return c.DeleteCheck(ctx, id)
}
