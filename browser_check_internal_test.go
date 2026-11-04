package checkly

import "testing"

func TestScriptCheckPayloadsSendNullForEmptyScript(t *testing.T) {
	scriptPath := "__checks__/check.spec.js"
	for name, m := range map[string]map[string]any{
		"browser":    marshalToMap(t, createBrowserCheckPayload(BrowserCheck{ScriptPath: &scriptPath})),
		"multi-step": marshalToMap(t, createMultiStepCheckPayload(MultiStepCheck{ScriptPath: &scriptPath})),
	} {
		if value, ok := m["script"]; !ok || value != nil {
			t.Errorf("%s: expected script to be null, got %#v", name, value)
		}
	}
}
