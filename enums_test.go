package checkly_test

import (
	"encoding/json"
	"testing"

	checkly "github.com/checkly/checkly-go-sdk/v2"
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

		{string(checkly.BodyTypeJSON), "JSON"},
		{string(checkly.BodyTypeForm), "FORM"},
		{string(checkly.BodyTypeRaw), "RAW"},
		{string(checkly.BodyTypeGraphQL), "GRAPHQL"},
		{string(checkly.BodyTypeNone), "NONE"},

		{string(checkly.IPFamilyIPv4), "IPv4"},
		{string(checkly.IPFamilyIPv6), "IPv6"},

		{string(checkly.DNSRecordTypeA), "A"},
		{string(checkly.DNSRecordTypeAAAA), "AAAA"},
		{string(checkly.DNSRecordTypeCNAME), "CNAME"},
		{string(checkly.DNSRecordTypeMX), "MX"},
		{string(checkly.DNSRecordTypeTXT), "TXT"},
		{string(checkly.DNSRecordTypeSOA), "SOA"},
		{string(checkly.DNSRecordTypeNS), "NS"},
		{string(checkly.DNSRecordTypeHTTPS), "HTTPS"},

		{string(checkly.DNSProtocolUDP), "UDP"},
		{string(checkly.DNSProtocolTCP), "TCP"},

		{string(checkly.TracerouteProtocolTCP), "TCP"},
		{string(checkly.TracerouteProtocolICMP), "ICMP"},
		{string(checkly.TracerouteProtocolUDP), "UDP"},
		{string(checkly.TracerouteProtocolSCTP), "SCTP"},

		{string(checkly.GRPCModeBehavior), "BEHAVIOR"},
		{string(checkly.GRPCModeHealth), "HEALTH"},

		{string(checkly.GRPCEncodingProtobuf), "PROTOBUF"},
		{string(checkly.GRPCEncodingFlatbuffers), "FLATBUFFERS"},

		{string(checkly.GRPCServiceDefinitionReflection), "REFLECTION"},
		{string(checkly.GRPCServiceDefinitionProtoFile), "PROTO_FILE"},

		{string(checkly.HeartbeatUnitSeconds), "seconds"},
		{string(checkly.HeartbeatUnitMinutes), "minutes"},
		{string(checkly.HeartbeatUnitHours), "hours"},
		{string(checkly.HeartbeatUnitDays), "days"},

		{string(checkly.SSLClientCertificateModeAuto), "auto"},
		{string(checkly.SSLClientCertificateModeExplicit), "explicit"},

		{string(checkly.TLSVersion12), "TLS1.2"},
		{string(checkly.TLSVersion13), "TLS1.3"},

		{string(checkly.SSLBaselineSeverityFail), "fail"},
		{string(checkly.SSLBaselineSeverityDegrade), "degrade"},
		{string(checkly.SSLBaselineSeverityIgnore), "ignore"},

		{string(checkly.PlaywrightEngineNode), "node"},
		{string(checkly.PlaywrightEngineBun), "bun"},

		{string(checkly.OpsgenieRegionEU), "EU"},
		{string(checkly.OpsgenieRegionUS), "US"},

		{string(checkly.OpsgeniePriorityP1), "P1"},
		{string(checkly.OpsgeniePriorityP2), "P2"},
		{string(checkly.OpsgeniePriorityP3), "P3"},
		{string(checkly.OpsgeniePriorityP4), "P4"},
		{string(checkly.OpsgeniePriorityP5), "P5"},

		{string(checkly.RetryStrategyTypeFixed), "FIXED"},
		{string(checkly.RetryStrategyTypeLinear), "LINEAR"},
		{string(checkly.RetryStrategyTypeExponential), "EXPONENTIAL"},
		{string(checkly.RetryStrategyTypeSingleRetry), "SINGLE_RETRY"},
		{string(checkly.RetryStrategyTypeNoRetries), "NO_RETRIES"},
		{string(checkly.RetryStrategyTypeFallback), "FALLBACK"},

		{string(checkly.RetryOnlyOnNetworkError), "NETWORK_ERROR"},

		{string(checkly.MaintenanceWindowRepeatUnitDay), "DAY"},
		{string(checkly.MaintenanceWindowRepeatUnitWeek), "WEEK"},
		{string(checkly.MaintenanceWindowRepeatUnitMonth), "MONTH"},

		{string(checkly.DashboardWidthFull), "FULL"},
		{string(checkly.DashboardWidth960PX), "960PX"},

		{string(checkly.RuntimeStageAlpha), "ALPHA"},
		{string(checkly.RuntimeStageBeta), "BETA"},
		{string(checkly.RuntimeStageCurrent), "CURRENT"},
		{string(checkly.RuntimeStageStable), "STABLE"},
		{string(checkly.RuntimeStageDeprecated), "DEPRECATED"},
		{string(checkly.RuntimeStageRemoved), "REMOVED"},

		{string(checkly.StatusPageComponentStatusV3Operational), "OPERATIONAL"},
		{string(checkly.StatusPageComponentStatusV3UnderMaintenance), "UNDER_MAINTENANCE"},
		{string(checkly.StatusPageComponentStatusV3DegradedPerformance), "DEGRADED_PERFORMANCE"},
		{string(checkly.StatusPageComponentStatusV3PartialOutage), "PARTIAL_OUTAGE"},
		{string(checkly.StatusPageComponentStatusV3MajorOutage), "MAJOR_OUTAGE"},

		{string(checkly.IncidentSeverityMinor), "MINOR"},
		{string(checkly.IncidentSeverityMedium), "MEDIUM"},
		{string(checkly.IncidentSeverityMajor), "MAJOR"},
		{string(checkly.IncidentSeverityCritical), "CRITICAL"},

		{string(checkly.StatusPageThemeAuto), "AUTO"},
		{string(checkly.StatusPageThemeDark), "DARK"},
		{string(checkly.StatusPageThemeLight), "LIGHT"},

		{string(checkly.StatusPageComponentV3TypeService), "SERVICE"},
		{string(checkly.StatusPageComponentV3TypeGroup), "GROUP"},

		{string(checkly.StatusPageTargetImpactV3UnderMaintenance), "UNDER_MAINTENANCE"},
		{string(checkly.StatusPageTargetImpactV3DegradedPerformance), "DEGRADED_PERFORMANCE"},
		{string(checkly.StatusPageTargetImpactV3PartialOutage), "PARTIAL_OUTAGE"},
		{string(checkly.StatusPageTargetImpactV3MajorOutage), "MAJOR_OUTAGE"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("want %q, got %q", tt.want, tt.got)
		}
	}
}

// TestPlaywrightEngineWire asserts a Playwright check's engine is sent as its
// plain string and decodes back to the same value.
func TestPlaywrightEngineWire(t *testing.T) {
	t.Parallel()
	engine := checkly.PlaywrightEngineBun
	data, err := json.Marshal(checkly.PlaywrightCheck{Engine: &engine})
	if err != nil {
		t.Fatal(err)
	}
	var sent map[string]any
	if err := json.Unmarshal(data, &sent); err != nil {
		t.Fatal(err)
	}
	if sent["engine"] != "bun" {
		t.Errorf(`want engine "bun", got %v`, sent["engine"])
	}
	var got checkly.PlaywrightCheck
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Engine == nil || *got.Engine != engine {
		t.Errorf("want engine %q, got %v", engine, got.Engine)
	}
}
