package checkly_test

import (
	"encoding/json"
	checkly "github.com/checkly/checkly-go-sdk"
	"testing"
)

func TestGRPCStreamingConfigurationRoundTrip(t *testing.T) {
	cfg := checkly.GRPCConfig{Mode: "BEHAVIOR", Method: "example.Chat/Exchange", Messages: []json.RawMessage{json.RawMessage(`{"name":"first"}`), json.RawMessage(`{"name":"second"}`)}, StreamTermination: &checkly.GRPCStreamTermination{DurationSeconds: 10.5, MaxMessages: 5}}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var got checkly.GRPCConfig
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Messages) != 2 || string(got.Messages[1]) != `{"name":"second"}` || got.StreamTermination == nil || got.StreamTermination.DurationSeconds != 10.5 || got.StreamTermination.MaxMessages != 5 {
		t.Fatalf("streaming configuration lost in round trip: %s", data)
	}
	var response checkly.GRPCCheckResponse
	if err := json.Unmarshal([]byte(`{"messageCount":0,"sentMessageCount":2,"grpcStatusCode":1,"streamTerminationReason":"duration"}`), &response); err != nil {
		t.Fatal(err)
	}
	if response.MessageCount == nil || *response.MessageCount != 0 || response.SentMessageCount == nil || *response.SentMessageCount != 2 || response.StreamTerminationReason != "duration" {
		t.Fatalf("streaming result lost: %+v", response)
	}
}
