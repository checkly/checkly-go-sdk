package checkly_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	checkly "github.com/checkly/checkly-go-sdk"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func validateStatusPageV3(t *testing.T, body []byte) {
	var payload checkly.StatusPageV3
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testStatusPageV3, payload) {
		t.Error(cmp.Diff(testStatusPageV3, payload))
	}
}

var testStatusPageV3 = checkly.StatusPageV3{
	ID:            "e35f7e14-91b2-4d24-b7b6-e0f9e2f8e51c",
	Name:          "Foo v3 status page",
	URL:           "foo-v3-status-page",
	Description:   "All Foo systems",
	DefaultTheme:  checkly.StatusPageThemeAuto,
	SupportLink:   "mailto:support@example.org",
	FooterText:    "Foo Inc.",
	AllowIndexing: true,
	ThemeColors: &checkly.StatusPageV3ThemeColors{
		Light: checkly.StatusPageV3ThemeColorGroup{
			BodyBackgroundColor:          "#F9FAFB",
			HeaderBackgroundColor:        "#FFFFFF",
			HeaderFontColor:              "#151A1E",
			TitleFontColor:               "#212930",
			BodyFontColor:                "#475766",
			BodyFontColorMuted:           "#60758A",
			NavigationFontColor:          "#151A1E",
			LinkFontColor:                "#FF0000",
			CardBackgroundColor:          "#FFFFFF",
			BorderColor:                  "#E0E5EB",
			PrimaryButtonBackgroundColor: "#151A1E",
			PrimaryButtonFontColor:       "#FFFFFF",
		},
		Dark: checkly.StatusPageV3ThemeColorGroup{
			BodyBackgroundColor:          "#14171C",
			HeaderBackgroundColor:        "#171B21",
			HeaderFontColor:              "#FFFFFF",
			TitleFontColor:               "#ECEEF2",
			BodyFontColor:                "#C6CDD7",
			BodyFontColorMuted:           "#A3B3C2",
			NavigationFontColor:          "#FFFFFF",
			LinkFontColor:                "#248AFF",
			CardBackgroundColor:          "#171B21",
			BorderColor:                  "#242B34",
			PrimaryButtonBackgroundColor: "#242B34",
			PrimaryButtonFontColor:       "#FFFFFF",
		},
	},
}

var ignoreStatusPageV3Fields = cmpopts.IgnoreFields(checkly.StatusPageV3{}, "ID")

func TestCreateStatusPageV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v3/status-pages",
		validateStatusPageV3,
		http.StatusCreated,
		"CreateStatusPageV3.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.CreateStatusPageV3(context.Background(), testStatusPageV3)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testStatusPageV3, *response, ignoreStatusPageV3Fields) {
		t.Error(cmp.Diff(testStatusPageV3, *response, ignoreStatusPageV3Fields))
	}
}

func TestDeleteStatusPageV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v3/status-pages/%s", testStatusPageV3.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	err := client.DeleteStatusPageV3(context.Background(), testStatusPageV3.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateStatusPageV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v3/status-pages/%s", testStatusPageV3.ID),
		validateStatusPageV3,
		http.StatusOK,
		"UpdateStatusPageV3.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	_, err := client.UpdateStatusPageV3(context.Background(), testStatusPageV3.ID, testStatusPageV3)
	if err != nil {
		t.Error(err)
	}
}

func TestGetStatusPageV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v3/status-pages/%s", testStatusPageV3.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetStatusPageV3.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.GetStatusPageV3(context.Background(), testStatusPageV3.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testStatusPageV3, *response, ignoreStatusPageV3Fields) {
		t.Error(cmp.Diff(testStatusPageV3, *response, ignoreStatusPageV3Fields))
	}
}

func validateStatusPageComponentV3(t *testing.T, body []byte) {
	var payload checkly.StatusPageComponentV3
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testStatusPageComponentV3, payload) {
		t.Error(cmp.Diff(testStatusPageComponentV3, payload))
	}
}

var testStatusPageComponentV3ShowHistoricalData = false

var testStatusPageComponentV3 = checkly.StatusPageComponentV3{
	ID:           "0e4f5a72-6a5c-42a1-9a8a-5f7d38c8a9d1",
	StatusPageID: testStatusPageV3.ID,
	Type:         checkly.StatusPageComponentV3TypeService,
	Name:         "Foo API",
	Description:  "The Foo public API",
	DisplayOrder: 1,
	Hidden:       false,
	Configuration: &checkly.StatusPageComponentV3Configuration{
		ShowHistoricalData: &testStatusPageComponentV3ShowHistoricalData,
	},
	ParentID: "0a2f26fb-47cc-42b7-91c6-40de3ec91a52",
}

var ignoreStatusPageComponentV3Fields = cmpopts.IgnoreFields(checkly.StatusPageComponentV3{}, "ID")

func TestCreateStatusPageComponentV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		fmt.Sprintf("/v3/status-pages/%s/components", testStatusPageV3.ID),
		validateStatusPageComponentV3,
		http.StatusCreated,
		"CreateStatusPageComponentV3.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.CreateStatusPageComponentV3(context.Background(), testStatusPageV3.ID, testStatusPageComponentV3)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testStatusPageComponentV3, *response, ignoreStatusPageComponentV3Fields) {
		t.Error(cmp.Diff(testStatusPageComponentV3, *response, ignoreStatusPageComponentV3Fields))
	}
}

func TestDeleteStatusPageComponentV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v3/status-pages/%s/components/%s", testStatusPageV3.ID, testStatusPageComponentV3.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	err := client.DeleteStatusPageComponentV3(context.Background(), testStatusPageV3.ID, testStatusPageComponentV3.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateStatusPageComponentV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v3/status-pages/%s/components/%s", testStatusPageV3.ID, testStatusPageComponentV3.ID),
		validateStatusPageComponentV3,
		http.StatusOK,
		"UpdateStatusPageComponentV3.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	_, err := client.UpdateStatusPageComponentV3(context.Background(), testStatusPageV3.ID, testStatusPageComponentV3.ID, testStatusPageComponentV3)
	if err != nil {
		t.Error(err)
	}
}

func TestGetStatusPageComponentV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v3/status-pages/%s/components/%s", testStatusPageV3.ID, testStatusPageComponentV3.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetStatusPageComponentV3.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.GetStatusPageComponentV3(context.Background(), testStatusPageV3.ID, testStatusPageComponentV3.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testStatusPageComponentV3, *response, ignoreStatusPageComponentV3Fields) {
		t.Error(cmp.Diff(testStatusPageComponentV3, *response, ignoreStatusPageComponentV3Fields))
	}
}

func validateStatusPageAutomationRuleV3(t *testing.T, body []byte) {
	var payload checkly.StatusPageAutomationRuleV3
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testStatusPageAutomationRuleV3, payload) {
		t.Error(cmp.Diff(testStatusPageAutomationRuleV3, payload))
	}
}

var testStatusPageAutomationRuleV3 = checkly.StatusPageAutomationRuleV3{
	ID:                    "5b0a9a1e-8e29-4b71-92e1-0db9b3b41f0a",
	StatusPageID:          testStatusPageV3.ID,
	Name:                  "Foo API outage",
	Enabled:               true,
	FirstUpdate:           "We are investigating an issue with the Foo API.",
	LastUpdate:            "The issue has been resolved.",
	NotifySubscribers:     true,
	CoolDownWindowMinutes: 5,
	Tags:                  []string{"foo-api"},
	Components: []checkly.StatusPageAutomationRuleComponentV3{
		{
			ComponentID:  testStatusPageComponentV3.ID,
			TargetImpact: checkly.StatusPageTargetImpactV3MajorOutage,
		},
	},
}

var ignoreStatusPageAutomationRuleV3Fields = cmpopts.IgnoreFields(checkly.StatusPageAutomationRuleV3{}, "ID")

func TestCreateStatusPageAutomationRuleV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		fmt.Sprintf("/v3/status-pages/%s/automation-rules", testStatusPageV3.ID),
		validateStatusPageAutomationRuleV3,
		http.StatusCreated,
		"CreateStatusPageAutomationRuleV3.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.CreateStatusPageAutomationRuleV3(context.Background(), testStatusPageV3.ID, testStatusPageAutomationRuleV3)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testStatusPageAutomationRuleV3, *response, ignoreStatusPageAutomationRuleV3Fields) {
		t.Error(cmp.Diff(testStatusPageAutomationRuleV3, *response, ignoreStatusPageAutomationRuleV3Fields))
	}
}

func TestDeleteStatusPageAutomationRuleV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v3/status-pages/%s/automation-rules/%s", testStatusPageV3.ID, testStatusPageAutomationRuleV3.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	err := client.DeleteStatusPageAutomationRuleV3(context.Background(), testStatusPageV3.ID, testStatusPageAutomationRuleV3.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateStatusPageAutomationRuleV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v3/status-pages/%s/automation-rules/%s", testStatusPageV3.ID, testStatusPageAutomationRuleV3.ID),
		validateStatusPageAutomationRuleV3,
		http.StatusOK,
		"UpdateStatusPageAutomationRuleV3.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	_, err := client.UpdateStatusPageAutomationRuleV3(context.Background(), testStatusPageV3.ID, testStatusPageAutomationRuleV3.ID, testStatusPageAutomationRuleV3)
	if err != nil {
		t.Error(err)
	}
}

func TestGetStatusPageAutomationRuleV3(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v3/status-pages/%s/automation-rules/%s", testStatusPageV3.ID, testStatusPageAutomationRuleV3.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetStatusPageAutomationRuleV3.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.GetStatusPageAutomationRuleV3(context.Background(), testStatusPageV3.ID, testStatusPageAutomationRuleV3.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testStatusPageAutomationRuleV3, *response, ignoreStatusPageAutomationRuleV3Fields) {
		t.Error(cmp.Diff(testStatusPageAutomationRuleV3, *response, ignoreStatusPageAutomationRuleV3Fields))
	}
}

func TestStatusPageV3SerializesClearableFields(t *testing.T) {
	m := marshalToMap(t, checkly.StatusPageV3{Name: "foo", URL: "foo"})
	for _, key := range []string{
		"customDomain", "description", "logo", "logoDark", "redirectTo",
		"favicon", "privacyPolicyLink", "termsOfServiceLink", "supportLink",
		"footerText", "googleAnalyticsTag",
	} {
		value, ok := m[key]
		if !ok || value != "" {
			t.Errorf("expected %q to be sent as an empty string so the update clears it, got %#v (present: %v)", key, value, ok)
		}
	}
	if value, ok := m["themeColors"]; !ok || value != nil {
		t.Errorf("expected themeColors to be sent as null so the update clears it, got %#v (present: %v)", value, ok)
	}
	if value, ok := m["allowIndexing"]; !ok || value != false {
		t.Errorf("expected allowIndexing to always be sent, got %#v (present: %v)", value, ok)
	}
	// The API rejects an empty defaultTheme; omitted, it resets to AUTO.
	if value, ok := m["defaultTheme"]; ok {
		t.Errorf("expected an unset defaultTheme to be omitted, got %#v", value)
	}
}

func TestStatusPageComponentV3OmitsUnsetFields(t *testing.T) {
	m := marshalToMap(t, checkly.StatusPageComponentV3{Name: "foo"})
	// The update replaces the component wholesale, so an omitted key clears
	// it; configuration in particular must not be null, which the API
	// rejects.
	for _, key := range []string{"id", "statusPageId", "type", "description", "configuration", "parentId"} {
		if value, ok := m[key]; ok {
			t.Errorf("expected unset %q to be omitted, got %#v", key, value)
		}
	}
}

func TestStatusPageAutomationRuleV3OmitsEmptyComponents(t *testing.T) {
	for _, components := range [][]checkly.StatusPageAutomationRuleComponentV3{nil, {}} {
		m := marshalToMap(t, checkly.StatusPageAutomationRuleV3{
			Name:        "foo",
			FirstUpdate: "Investigating.",
			LastUpdate:  "Resolved.",
			Tags:        []string{"foo"},
			Components:  components,
		})
		// The API rejects a null list; omitted, it defaults to none.
		if value, ok := m["components"]; ok {
			t.Errorf("expected an empty component list to be omitted, got %#v", value)
		}
	}
}
