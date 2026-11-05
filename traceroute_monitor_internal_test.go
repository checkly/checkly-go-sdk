package checkly

import (
	"reflect"
	"strings"
	"testing"
)

// jsonFieldNames returns the JSON key of every exported field of the struct
// type t.
func jsonFieldNames(t reflect.Type) map[string]bool {
	names := map[string]bool{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "-" {
			continue
		}
		if name == "" {
			name = field.Name
		}
		names[name] = true
	}
	return names
}

// TestTracerouteUpdateRequestPayloadSendsEveryField asserts that the update
// request object, which is a separate struct, has a key for every key of
// TracerouteRequest, so that a field added to TracerouteRequest is not
// silently dropped from updates.
func TestTracerouteUpdateRequestPayloadSendsEveryField(t *testing.T) {
	t.Parallel()
	sent := jsonFieldNames(reflect.TypeOf(tracerouteUpdateRequestPayload{}))
	for name := range jsonFieldNames(reflect.TypeOf(TracerouteRequest{})) {
		if !sent[name] {
			t.Errorf("tracerouteUpdateRequestPayload has no field for TracerouteRequest key %q", name)
		}
	}
}
