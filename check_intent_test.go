package checkly_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	checkly "github.com/checkly/checkly-go-sdk"
	"github.com/google/go-cmp/cmp"
)

// selfHealingCase calls the create, update and get methods of one check type
// with the given intent and auto-repair setting, and returns what each
// method decoded.
type selfHealingCase struct {
	checkType string
	create    func(c checkly.Client, intent *checkly.CheckIntent, repair *bool) (*checkly.CheckIntent, *bool, error)
	update    func(c checkly.Client, intent *checkly.CheckIntent, repair *bool) (*checkly.CheckIntent, *bool, error)
	get       func(c checkly.Client) (*checkly.CheckIntent, *bool, error)
}

var selfHealingCases = map[string]selfHealingCase{
	"APICheck": {
		checkType: "API",
		create: func(c checkly.Client, intent *checkly.CheckIntent, repair *bool) (*checkly.CheckIntent, *bool, error) {
			r, err := c.CreateAPICheck(context.Background(), checkly.APICheck{Name: "n", Intent: intent, AIAutoRepairEnabled: repair})
			if err != nil {
				return nil, nil, err
			}
			return r.Intent, r.AIAutoRepairEnabled, nil
		},
		update: func(c checkly.Client, intent *checkly.CheckIntent, repair *bool) (*checkly.CheckIntent, *bool, error) {
			r, err := c.UpdateAPICheck(context.Background(), "id", checkly.APICheck{Name: "n", Intent: intent, AIAutoRepairEnabled: repair})
			if err != nil {
				return nil, nil, err
			}
			return r.Intent, r.AIAutoRepairEnabled, nil
		},
		get: func(c checkly.Client) (*checkly.CheckIntent, *bool, error) {
			r, err := c.GetAPICheck(context.Background(), "id")
			if err != nil {
				return nil, nil, err
			}
			return r.Intent, r.AIAutoRepairEnabled, nil
		},
	},
	"BrowserCheck": {
		checkType: "BROWSER",
		create: func(c checkly.Client, intent *checkly.CheckIntent, repair *bool) (*checkly.CheckIntent, *bool, error) {
			r, err := c.CreateBrowserCheck(context.Background(), checkly.BrowserCheck{Name: "n", Intent: intent, AIAutoRepairEnabled: repair})
			if err != nil {
				return nil, nil, err
			}
			return r.Intent, r.AIAutoRepairEnabled, nil
		},
		update: func(c checkly.Client, intent *checkly.CheckIntent, repair *bool) (*checkly.CheckIntent, *bool, error) {
			r, err := c.UpdateBrowserCheck(context.Background(), "id", checkly.BrowserCheck{Name: "n", Intent: intent, AIAutoRepairEnabled: repair})
			if err != nil {
				return nil, nil, err
			}
			return r.Intent, r.AIAutoRepairEnabled, nil
		},
		get: func(c checkly.Client) (*checkly.CheckIntent, *bool, error) {
			r, err := c.GetBrowserCheck(context.Background(), "id")
			if err != nil {
				return nil, nil, err
			}
			return r.Intent, r.AIAutoRepairEnabled, nil
		},
	},
	"MultiStepCheck": {
		checkType: "MULTI_STEP",
		create: func(c checkly.Client, intent *checkly.CheckIntent, repair *bool) (*checkly.CheckIntent, *bool, error) {
			r, err := c.CreateMultiStepCheck(context.Background(), checkly.MultiStepCheck{Name: "n", Intent: intent, AIAutoRepairEnabled: repair})
			if err != nil {
				return nil, nil, err
			}
			return r.Intent, r.AIAutoRepairEnabled, nil
		},
		update: func(c checkly.Client, intent *checkly.CheckIntent, repair *bool) (*checkly.CheckIntent, *bool, error) {
			r, err := c.UpdateMultiStepCheck(context.Background(), "id", checkly.MultiStepCheck{Name: "n", Intent: intent, AIAutoRepairEnabled: repair})
			if err != nil {
				return nil, nil, err
			}
			return r.Intent, r.AIAutoRepairEnabled, nil
		},
		get: func(c checkly.Client) (*checkly.CheckIntent, *bool, error) {
			r, err := c.GetMultiStepCheck(context.Background(), "id")
			if err != nil {
				return nil, nil, err
			}
			return r.Intent, r.AIAutoRepairEnabled, nil
		},
	},
	"PlaywrightCheck": {
		checkType: "PLAYWRIGHT",
		create: func(c checkly.Client, intent *checkly.CheckIntent, repair *bool) (*checkly.CheckIntent, *bool, error) {
			r, err := c.CreatePlaywrightCheck(context.Background(), checkly.PlaywrightCheck{Name: "n", Intent: intent, AIAutoRepairEnabled: repair})
			if err != nil {
				return nil, nil, err
			}
			return r.Intent, r.AIAutoRepairEnabled, nil
		},
		update: func(c checkly.Client, intent *checkly.CheckIntent, repair *bool) (*checkly.CheckIntent, *bool, error) {
			r, err := c.UpdatePlaywrightCheck(context.Background(), "id", checkly.PlaywrightCheck{Name: "n", Intent: intent, AIAutoRepairEnabled: repair})
			if err != nil {
				return nil, nil, err
			}
			return r.Intent, r.AIAutoRepairEnabled, nil
		},
		get: func(c checkly.Client) (*checkly.CheckIntent, *bool, error) {
			r, err := c.GetPlaywrightCheck(context.Background(), "id")
			if err != nil {
				return nil, nil, err
			}
			return r.Intent, r.AIAutoRepairEnabled, nil
		},
	},
}

// The check updates replace the stored intent and auto-repair setting, so
// create and update send both keys even when unset: as null, which clears an
// intent and restores the account's auto-repair setting.
func TestCheckSelfHealingFieldsPayload(t *testing.T) {
	t.Parallel()
	f := false
	for name, tc := range selfHealingCases {
		tc := tc
		for _, method := range []string{"create", "update"} {
			method := method
			t.Run(name+"/"+method, func(t *testing.T) {
				t.Parallel()
				status, call := http.StatusCreated, tc.create
				if method == "update" {
					status, call = http.StatusOK, tc.update
				}
				c, lastBody := newRecordingServer(t, status, `{"checkType":"`+tc.checkType+`"}`)

				if _, _, err := call(c, nil, nil); err != nil {
					t.Fatal(err)
				}
				assertRawBodyField(t, lastBody(), "intent", `null`)
				assertRawBodyField(t, lastBody(), "aiAutoRepairEnabled", `null`)

				// Nil statement lists are sent as [], because the API
				// rejects null for them.
				if _, _, err := call(c, &checkly.CheckIntent{Goal: "Checkout works"}, &f); err != nil {
					t.Fatal(err)
				}
				assertRawBodyField(t, lastBody(), "intent", `{"goal":"Checkout works","requiredOutcomes":[],"mustPreserve":[]}`)
				assertRawBodyField(t, lastBody(), "aiAutoRepairEnabled", `false`)

				intent := &checkly.CheckIntent{Goal: "g", RequiredOutcomes: []string{"r"}, MustPreserve: []string{"m"}}
				if _, _, err := call(c, intent, nil); err != nil {
					t.Fatal(err)
				}
				assertRawBodyField(t, lastBody(), "intent", `{"goal":"g","requiredOutcomes":["r"],"mustPreserve":["m"]}`)
			})
		}
	}
}

func TestCheckSelfHealingFieldsDecoding(t *testing.T) {
	t.Parallel()
	for name, tc := range selfHealingCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			set, err := json.Marshal(map[string]any{
				"checkType":           tc.checkType,
				"intent":              map[string]any{"goal": "g", "requiredOutcomes": []string{"r"}, "mustPreserve": []string{}},
				"aiAutoRepairEnabled": true,
			})
			if err != nil {
				t.Fatal(err)
			}
			c, _ := newRecordingServer(t, http.StatusOK, string(set))
			intent, repair, err := tc.get(c)
			if err != nil {
				t.Fatal(err)
			}
			want := &checkly.CheckIntent{Goal: "g", RequiredOutcomes: []string{"r"}, MustPreserve: []string{}}
			if !cmp.Equal(want, intent) {
				t.Error(cmp.Diff(want, intent))
			}
			if repair == nil || !*repair {
				t.Errorf("want aiAutoRepairEnabled true, got %v", repair)
			}

			c, _ = newRecordingServer(t, http.StatusOK, `{"checkType":"`+tc.checkType+`","intent":null,"aiAutoRepairEnabled":null}`)
			intent, repair, err = tc.get(c)
			if err != nil {
				t.Fatal(err)
			}
			if intent != nil || repair != nil {
				t.Errorf("want null fields unset, got intent %v, aiAutoRepairEnabled %v", intent, repair)
			}
		})
	}
}
