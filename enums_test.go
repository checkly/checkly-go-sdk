package checkly_test

import (
	"testing"

	checkly "github.com/checkly/checkly-go-sdk"
)

// TestEnumWireValues pins every enum constant to the string the API uses, so
// renaming a constant never changes what is sent.
func TestEnumWireValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		got  string
		want string
	}{
		{string(checkly.AssertionSourceStatusCode), "STATUS_CODE"},
		{string(checkly.AssertionSourceJSONBody), "JSON_BODY"},
		{string(checkly.AssertionSourceTextBody), "TEXT_BODY"},
		{string(checkly.AssertionSourceHeaders), "HEADERS"},
		{string(checkly.AssertionSourceResponseTime), "RESPONSE_TIME"},
		{string(checkly.AssertionSourceResponseData), "RESPONSE_DATA"},
		{string(checkly.AssertionSourceCertificate), "CERTIFICATE"},
		{string(checkly.AssertionSourceConnection), "CONNECTION"},
		{string(checkly.AssertionSourceJSONResponse), "JSON_RESPONSE"},
		{string(checkly.AssertionSourceTextResponse), "TEXT_RESPONSE"},
		{string(checkly.AssertionSourceResponseCode), "RESPONSE_CODE"},
		{string(checkly.AssertionSourceTextAnswer), "TEXT_ANSWER"},
		{string(checkly.AssertionSourceJSONAnswer), "JSON_ANSWER"},
		{string(checkly.AssertionSourceLatency), "LATENCY"},
		{string(checkly.AssertionSourceHopCount), "HOP_COUNT"},
		{string(checkly.AssertionSourcePacketLoss), "PACKET_LOSS"},
		{string(checkly.AssertionSourceGRPCResponse), "GRPC_RESPONSE"},
		{string(checkly.AssertionSourceGRPCMetadata), "GRPC_METADATA"},
		{string(checkly.AssertionSourceGRPCHealthcheckStatus), "GRPC_HEALTHCHECK_STATUS"},
		{string(checkly.AssertionSourceGRPCStatusCode), "GRPC_STATUS_CODE"},

		{string(checkly.AssertionComparisonEquals), "EQUALS"},
		{string(checkly.AssertionComparisonNotEquals), "NOT_EQUALS"},
		{string(checkly.AssertionComparisonHasKey), "HAS_KEY"},
		{string(checkly.AssertionComparisonNotHasKey), "NOT_HAS_KEY"},
		{string(checkly.AssertionComparisonHasValue), "HAS_VALUE"},
		{string(checkly.AssertionComparisonNotHasValue), "NOT_HAS_VALUE"},
		{string(checkly.AssertionComparisonIsEmpty), "IS_EMPTY"},
		{string(checkly.AssertionComparisonNotEmpty), "NOT_EMPTY"},
		{string(checkly.AssertionComparisonGreaterThan), "GREATER_THAN"},
		{string(checkly.AssertionComparisonLessThan), "LESS_THAN"},
		{string(checkly.AssertionComparisonContains), "CONTAINS"},
		{string(checkly.AssertionComparisonNotContains), "NOT_CONTAINS"},
		{string(checkly.AssertionComparisonIsNull), "IS_NULL"},
		{string(checkly.AssertionComparisonNotNull), "NOT_NULL"},

		{string(checkly.EscalationTypeRunBased), "RUN_BASED"},
		{string(checkly.EscalationTypeTimeBased), "TIME_BASED"},

		{string(checkly.AlertChannelTypeEmail), "EMAIL"},
		{string(checkly.AlertChannelTypeSlack), "SLACK"},
		{string(checkly.AlertChannelTypeSlackApp), "SLACK_APP"},
		{string(checkly.AlertChannelTypeWebhook), "WEBHOOK"},
		{string(checkly.AlertChannelTypeSMS), "SMS"},
		{string(checkly.AlertChannelTypePagerduty), "PAGERDUTY"},
		{string(checkly.AlertChannelTypeOpsgenie), "OPSGENIE"},
		{string(checkly.AlertChannelTypeCall), "CALL"},

		{string(checkly.CheckTypeAgentic), "AGENTIC"},
		{string(checkly.CheckTypeAPI), "API"},
		{string(checkly.CheckTypeBrowser), "BROWSER"},
		{string(checkly.CheckTypeHeartbeat), "HEARTBEAT"},
		{string(checkly.CheckTypeICMP), "ICMP"},
		{string(checkly.CheckTypeMultiStep), "MULTI_STEP"},
		{string(checkly.CheckTypeTCP), "TCP"},
		{string(checkly.CheckTypePlaywright), "PLAYWRIGHT"},
		{string(checkly.CheckTypeTraceroute), "TRACEROUTE"},
		{string(checkly.CheckTypeURL), "URL"},
		{string(checkly.CheckTypeDNS), "DNS"},
		{string(checkly.CheckTypeSSL), "SSL"},
		{string(checkly.CheckTypeGRPC), "GRPC"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("want %q, got %q", tt.want, tt.got)
		}
	}
}
