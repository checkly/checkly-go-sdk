package checkly

import "testing"

func TestICMPMonitorPayloadSendsNonZeroThresholds(t *testing.T) {
	m := marshalToMap(t, createICMPMonitorPayload(ICMPMonitor{
		DegradedPacketLossThreshold: 5,
		MaxPacketLossThreshold:      50,
	}))
	if m["degradedPacketLossThreshold"] != float64(5) || m["maxPacketLossThreshold"] != float64(50) {
		t.Errorf("expected thresholds 5 and 50, got %v and %v", m["degradedPacketLossThreshold"], m["maxPacketLossThreshold"])
	}
}
