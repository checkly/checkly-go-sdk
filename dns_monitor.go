package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// DNSMonitor represents a DNS monitor.
type DNSMonitor struct {
	ID                        string                     `json:"id,omitempty"`
	Name                      string                     `json:"name"`
	Description               *string                    `json:"description"`
	Frequency                 int                        `json:"frequency"`
	FrequencyOffset           int                        `json:"frequencyOffset,omitempty"`
	Activated                 bool                       `json:"activated"`
	Muted                     bool                       `json:"muted"`
	RunParallel               bool                       `json:"runParallel"`
	Locations                 []string                   `json:"locations"`
	DegradedResponseTime      int                        `json:"degradedResponseTime,omitempty"`
	MaxResponseTime           int                        `json:"maxResponseTime,omitempty"`
	Tags                      []string                   `json:"tags,omitempty"`
	AlertSettings             *AlertSettings             `json:"alertSettings,omitempty"`
	UseGlobalAlertSettings    bool                       `json:"useGlobalAlertSettings"`
	Request                   DNSRequest                 `json:"request"`
	GroupID                   int64                      `json:"groupId,omitempty"`
	GroupOrder                int                        `json:"groupOrder,omitempty"`
	AlertChannelSubscriptions []AlertChannelSubscription `json:"alertChannelSubscriptions,omitempty"`
	// PrivateLocations are the slugs of the private locations the monitor runs
	// from. The Checkly agent at those locations must be version 7.0.0 or later.
	PrivateLocations *[]string        `json:"privateLocations"`
	RetryStrategy    *RetryStrategy   `json:"retryStrategy"`
	TriggerIncident  *IncidentTrigger `json:"triggerIncident"`
	CreatedAt        time.Time        `json:"created_at,omitempty"`
	UpdatedAt        time.Time        `json:"updated_at,omitempty"`
}

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

// DNSRequest represents the parameters for the request made by a DNS monitor.
type DNSRequest struct {
	RecordType DNSRecordType `json:"recordType"`
	Query      string        `json:"query"`
	NameServer string        `json:"nameServer,omitempty"`
	Port       int           `json:"port,omitempty"`
	Protocol   DNSProtocol   `json:"protocol,omitempty"`
	Assertions []Assertion   `json:"assertions"`
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
	// The update endpoint is PATCH-style: keys omitted from the payload keep
	// their server-side value. The fields below shadow embedded fields whose
	// zero value would be omitted so that they are always sent and can
	// actually be cleared via an update.
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
		// Unfortunately, this will default to true if not set.
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

// GetDNSMonitor takes the ID of an existing DNS monitor, and returns the
// monitor parameters, or an error. It fails with an error matching
// ErrCheckTypeMismatch when the ID belongs to a check of another type.
func (c *client) GetDNSMonitor(
	ctx context.Context,
	ID string,
) (*DNSMonitor, error) {
	var result DNSMonitor
	if err := c.getCheckOfType(ctx, fmt.Sprintf("/v1/checks/%s", ID), CheckTypeDNS, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateDNSMonitor(
	ctx context.Context,
	ID string,
	monitor DNSMonitor,
) (*DNSMonitor, error) {
	var result DNSMonitor
	err := c.callJSON(ctx, http.MethodPut, withAutoAssignAlertsFlag(fmt.Sprintf("/v1/checks/dns/%s", ID)), createDNSMonitorPayload(monitor), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteDNSMonitor(
	ctx context.Context,
	ID string,
) error {
	return c.DeleteCheck(ctx, ID)
}
