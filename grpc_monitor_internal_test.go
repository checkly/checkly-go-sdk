package checkly

import "testing"

// TestGRPCPayloadSendsEmptyValues asserts the gRPC payload sends empty values
// that clear the stored ones: the update endpoint keeps the stored value of an
// omitted field. It always sends the free-text field of the configured mode,
// never the one of the other mode, and sends assertions and metadata as empty
// lists.
func TestGRPCPayloadSendsEmptyValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		mode    GRPCMode
		send    string
		omitted string
	}{
		{mode: "BEHAVIOR", send: "message", omitted: "service"},
		{mode: "", send: "message", omitted: "service"},
		{mode: "HEALTH", send: "service", omitted: "message"},
	}
	for _, test := range tests {
		t.Run("mode "+string(test.mode), func(t *testing.T) {
			request := marshalToMap(t, createGRPCMonitorPayload(GRPCMonitor{
				Request: GRPCRequest{GRPCConfig: GRPCConfig{Mode: test.mode}},
			}))["request"].(map[string]any)
			config := request["grpcConfig"].(map[string]any)
			if v, ok := config[test.send]; !ok || v != "" {
				t.Errorf("%s = %v (present %v), want empty string", test.send, v, ok)
			}
			if _, ok := config[test.omitted]; ok {
				t.Errorf("%s must be omitted, got %v", test.omitted, config)
			}
			if assertions, ok := request["assertions"].([]any); !ok || len(assertions) != 0 {
				t.Errorf("assertions = %v, want an empty list", request["assertions"])
			}
			if metadata, ok := config["metadata"].([]any); !ok || len(metadata) != 0 {
				t.Errorf("metadata = %v, want an empty list", config["metadata"])
			}
		})
	}
}

// TestGRPCPayloadPreservesRequest asserts the payload wrappers keep the
// request fields they do not override.
func TestGRPCPayloadPreservesRequest(t *testing.T) {
	t.Parallel()
	request := marshalToMap(t, createGRPCMonitorPayload(GRPCMonitor{
		Request: GRPCRequest{
			URL:        "grpc.example.com",
			Assertions: []Assertion{{Source: "GRPC_STATUS_CODE"}},
			GRPCConfig: GRPCConfig{
				Mode:     "BEHAVIOR",
				Method:   "a.B/C",
				Message:  `{"name":"Checkly"}`,
				Service:  "ignored",
				Metadata: []GRPCMetadata{{Key: "x-key", Value: "value"}},
			},
		},
	}))["request"].(map[string]any)
	config := request["grpcConfig"].(map[string]any)
	if request["url"] != "grpc.example.com" || config["method"] != "a.B/C" || config["message"] != `{"name":"Checkly"}` {
		t.Errorf("request fields were not preserved: %v", request)
	}
	if len(request["assertions"].([]any)) != 1 || len(config["metadata"].([]any)) != 1 {
		t.Errorf("assertions or metadata were not preserved: %v", request)
	}
	if _, ok := config["service"]; ok {
		t.Errorf("BEHAVIOR payload must omit service, got %v", config)
	}
}

// TestGRPCConfigPayloadModeAndProtoContent asserts that an empty mode is sent
// as BEHAVIOR, and that an empty proto content is sent only with an explicit
// REFLECTION service definition: with an omitted definition the API would
// keep a stored PROTO_FILE definition and store an empty proto file for it.
func TestGRPCConfigPayloadModeAndProtoContent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		config       GRPCConfig
		wantMode     string
		protoContent any // nil means the key is absent
	}{
		{name: "empty", config: GRPCConfig{}, wantMode: "BEHAVIOR"},
		{name: "reflection", config: GRPCConfig{ServiceDefinition: "REFLECTION"}, wantMode: "BEHAVIOR", protoContent: ""},
		{name: "proto file", config: GRPCConfig{Mode: "BEHAVIOR", ServiceDefinition: "PROTO_FILE"}, wantMode: "BEHAVIOR"},
		{name: "proto content set", config: GRPCConfig{ProtoContent: "syntax"}, wantMode: "BEHAVIOR", protoContent: "syntax"},
		{name: "health", config: GRPCConfig{Mode: "HEALTH", ServiceDefinition: "REFLECTION"}, wantMode: "HEALTH"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := marshalToMap(t, createGRPCConfigPayload(test.config))
			if config["mode"] != test.wantMode {
				t.Errorf("mode = %v, want %s", config["mode"], test.wantMode)
			}
			got, ok := config["protoContent"]
			if test.protoContent == nil {
				if ok {
					t.Errorf("protoContent must be omitted, got %v", got)
				}
			} else if !ok || got != test.protoContent {
				t.Errorf("protoContent = %v (present %v), want %v", got, ok, test.protoContent)
			}
		})
	}
}
