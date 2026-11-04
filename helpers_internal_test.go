package checkly

import (
	"encoding/json"
	"testing"
)

// marshalToMap marshals a payload and decodes it into a generic map, so tests
// can tell an empty field apart from an omitted one.
func marshalToMap(t *testing.T, payload any) map[string]any {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return m
}
