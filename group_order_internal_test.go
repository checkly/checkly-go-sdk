package checkly

import (
	"encoding/json"
	"testing"
)

// groupOrderType builds the create and update request payloads of one check
// or monitor type with a given group order, and decodes a response into the
// type to read its group order back.
type groupOrderType struct {
	name   string
	encode map[string]func(groupOrder *int) any
	decode func(data []byte) (*int, error)
}

func decodeGroupOrder[T any](groupOrder func(*T) *int) func([]byte) (*int, error) {
	return func(data []byte) (*int, error) {
		var v T
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return groupOrder(&v), nil
	}
}

func sameForCreateAndUpdate(build func(*int) any) map[string]func(*int) any {
	return map[string]func(*int) any{"create": build, "update": build}
}

var groupOrderTypes = []groupOrderType{
	{
		name: "APICheck",
		encode: sameForCreateAndUpdate(func(o *int) any {
			return createAPICheckPayload(APICheck{GroupOrder: o})
		}),
		decode: decodeGroupOrder(func(v *APICheck) *int { return v.GroupOrder }),
	},
	{
		name: "BrowserCheck",
		encode: sameForCreateAndUpdate(func(o *int) any {
			return createBrowserCheckPayload(BrowserCheck{GroupOrder: o})
		}),
		decode: decodeGroupOrder(func(v *BrowserCheck) *int { return v.GroupOrder }),
	},
	{
		name: "MultiStepCheck",
		encode: sameForCreateAndUpdate(func(o *int) any {
			return createMultiStepCheckPayload(MultiStepCheck{GroupOrder: o})
		}),
		decode: decodeGroupOrder(func(v *MultiStepCheck) *int { return v.GroupOrder }),
	},
	{
		name: "PlaywrightCheck",
		encode: sameForCreateAndUpdate(func(o *int) any {
			return createPlaywrightCheckPayload(PlaywrightCheck{GroupOrder: o})
		}),
		decode: decodeGroupOrder(func(v *PlaywrightCheck) *int { return v.GroupOrder }),
	},
	{
		name: "URLMonitor",
		encode: sameForCreateAndUpdate(func(o *int) any {
			return createURLMonitorPayload(URLMonitor{GroupOrder: o})
		}),
		decode: decodeGroupOrder(func(v *URLMonitor) *int { return v.GroupOrder }),
	},
	{
		name: "DNSMonitor",
		encode: sameForCreateAndUpdate(func(o *int) any {
			return createDNSMonitorPayload(DNSMonitor{GroupOrder: o})
		}),
		decode: decodeGroupOrder(func(v *DNSMonitor) *int { return v.GroupOrder }),
	},
	{
		name: "ICMPMonitor",
		encode: sameForCreateAndUpdate(func(o *int) any {
			return createICMPMonitorPayload(ICMPMonitor{GroupOrder: o})
		}),
		decode: decodeGroupOrder(func(v *ICMPMonitor) *int { return v.GroupOrder }),
	},
	{
		name: "GRPCMonitor",
		encode: map[string]func(*int) any{
			"create": func(o *int) any { return createGRPCMonitorPayload(GRPCMonitor{GroupOrder: o}) },
			"update": func(o *int) any { return updateGRPCMonitorPayload(GRPCMonitor{GroupOrder: o}) },
		},
		decode: decodeGroupOrder(func(v *GRPCMonitor) *int { return v.GroupOrder }),
	},
	{
		name: "SSLMonitor",
		encode: sameForCreateAndUpdate(func(o *int) any {
			return createSSLMonitorPayload(SSLMonitor{GroupOrder: o})
		}),
		decode: decodeGroupOrder(func(v *SSLMonitor) *int { return v.GroupOrder }),
	},
	{
		name: "TCPMonitor",
		encode: sameForCreateAndUpdate(func(o *int) any {
			return createTCPMonitorPayload(TCPMonitor{GroupOrder: o})
		}),
		decode: decodeGroupOrder(func(v *TCPMonitor) *int { return v.GroupOrder }),
	},
	{
		name: "TracerouteMonitor",
		encode: map[string]func(*int) any{
			"create": func(o *int) any { return createTracerouteMonitorPayload(TracerouteMonitor{GroupOrder: o}) },
			"update": func(o *int) any { return updateTracerouteMonitorPayload(TracerouteMonitor{GroupOrder: o}) },
		},
		decode: decodeGroupOrder(func(v *TracerouteMonitor) *int { return v.GroupOrder }),
	},
}

func intPtr(v int) *int { return &v }

// TestGroupOrderEncoding asserts that every create and update request sends
// the group order as given, 0 included, and a nil group order as null: 0 is
// the first position in a group, and null clears the position, so neither
// may be dropped or confused with the other.
func TestGroupOrderEncoding(t *testing.T) {
	cases := []struct {
		groupOrder *int
		want       string
	}{
		{nil, `null`},
		{intPtr(0), `0`},
		{intPtr(5), `5`},
	}
	for _, typ := range groupOrderTypes {
		for _, request := range []string{"create", "update"} {
			for _, c := range cases {
				t.Run(typ.name+"/"+request+"/"+c.want, func(t *testing.T) {
					body, err := json.Marshal(typ.encode[request](c.groupOrder))
					if err != nil {
						t.Fatal(err)
					}
					var fields map[string]json.RawMessage
					if err := json.Unmarshal(body, &fields); err != nil {
						t.Fatal(err)
					}
					got, ok := fields["groupOrder"]
					if !ok {
						t.Fatalf("groupOrder missing from request %s", body)
					}
					if string(got) != c.want {
						t.Errorf("groupOrder = %s, want %s", got, c.want)
					}
				})
			}
		}
	}
}

// TestGroupOrderDecoding asserts that a null group order in a response
// decodes to nil and any number, 0 included, to a pointer to it.
func TestGroupOrderDecoding(t *testing.T) {
	cases := []struct {
		json string
		want *int
	}{
		{`null`, nil},
		{`0`, intPtr(0)},
		{`5`, intPtr(5)},
	}
	for _, typ := range groupOrderTypes {
		for _, c := range cases {
			t.Run(typ.name+"/"+c.json, func(t *testing.T) {
				got, err := typ.decode([]byte(`{"groupOrder":` + c.json + `}`))
				if err != nil {
					t.Fatal(err)
				}
				switch {
				case c.want == nil && got != nil:
					t.Errorf("groupOrder = %d, want nil", *got)
				case c.want != nil && got == nil:
					t.Errorf("groupOrder = nil, want %d", *c.want)
				case c.want != nil && *got != *c.want:
					t.Errorf("groupOrder = %d, want %d", *got, *c.want)
				}
			})
		}
	}
}
