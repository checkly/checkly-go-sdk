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

func validateStatusPageService(t *testing.T, body []byte) {
	var payload checkly.StatusPageService
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testStatusPageService, payload) {
		t.Error(cmp.Diff(testStatusPageService, payload))
	}
}

var testStatusPageService = checkly.StatusPageService{
	ID:   "8a894b49-467f-4af6-9230-cd4d3bd452f4",
	Name: "Foo service",
}

var ignoreStatusPageServiceFields = cmpopts.IgnoreFields(checkly.StatusPageService{}, "ID")

func TestCreateStatusPageService(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/status-pages/services",
		validateStatusPageService,
		http.StatusCreated,
		"CreateStatusPageService.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.CreateStatusPageService(context.Background(), testStatusPageService)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testStatusPageService, *response, ignoreStatusPageServiceFields) {
		t.Error(cmp.Diff(testStatusPageService, *response, ignoreStatusPageServiceFields))
	}
}

func TestDeleteStatusPageService(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/status-pages/services/%s", testStatusPageService.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	err := client.DeleteStatusPageService(context.Background(), testStatusPageService.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateStatusPageService(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/status-pages/services/%s", testStatusPageService.ID),
		validateStatusPageService,
		http.StatusOK,
		"UpdateStatusPageService.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	_, err := client.UpdateStatusPageService(context.Background(), testStatusPageService.ID, testStatusPageService)
	if err != nil {
		t.Error(err)
	}
}

func TestGetStatusPageService(t *testing.T) {
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/status-pages/services/%s", testStatusPageService.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetStatusPageService.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.GetStatusPageService(context.Background(), testStatusPageService.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testStatusPageService, *response, ignoreStatusPageServiceFields) {
		t.Error(cmp.Diff(testStatusPageService, *response, ignoreStatusPageServiceFields))
	}
}

func validateStatusPage(t *testing.T, body []byte) {
	var payload checkly.StatusPage
	err := json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testStatusPage, payload) {
		t.Error(cmp.Diff(testStatusPage, payload))
	}
}

var testStatusPage = checkly.StatusPage{
	ID:           "cd8d05a4-c292-4dc4-a78f-1dea65e5457e",
	Name:         "Foo status page",
	URL:          "foo-status-page",
	DefaultTheme: checkly.StatusPageThemeAuto,
	Cards: []checkly.StatusPageCard{
		{
			Name: "Foo card",
			Services: []checkly.StatusPageService{
				{
					ID:   "8a894b49-467f-4af6-9230-cd4d3bd452f4",
					Name: "Foo service",
				},
			},
		},
	},
}

var ignoreStatusPageFields = cmpopts.IgnoreFields(checkly.StatusPage{}, "ID")

func TestCreateStatusPage(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/status-pages",
		validateStatusPage,
		http.StatusCreated,
		"CreateStatusPage.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.CreateStatusPage(context.Background(), testStatusPage)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testStatusPage, *response, ignoreStatusPageFields) {
		t.Error(cmp.Diff(testStatusPage, *response, ignoreStatusPageFields))
	}
}

func TestDeleteStatusPage(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/status-pages/%s", testStatusPage.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	err := client.DeleteStatusPage(context.Background(), testStatusPage.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateStatusPage(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/status-pages/%s", testStatusPage.ID),
		validateStatusPage,
		http.StatusOK,
		"UpdateStatusPage.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	_, err := client.UpdateStatusPage(context.Background(), testStatusPage.ID, testStatusPage)
	if err != nil {
		t.Error(err)
	}
}

func TestGetStatusPage(t *testing.T) {
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/status-pages/%s", testStatusPage.ID),
		validateEmptyBody,
		http.StatusOK,
		"GetStatusPage.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	response, err := client.GetStatusPage(context.Background(), testStatusPage.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testStatusPage, *response, ignoreStatusPageFields) {
		t.Error(cmp.Diff(testStatusPage, *response, ignoreStatusPageFields))
	}
}

// validateStatusPageEmptyPayload asserts that the optional keys of an empty
// status page are sent, as the update endpoint keeps the stored value of an
// omitted key: the custom domain as null, which removes it, and the others as
// empty strings.
func validateStatusPageEmptyPayload(t *testing.T, body []byte) {
	assertRawBodyField(t, body, "customDomain", `null`)
	assertRawBodyField(t, body, "logo", `""`)
	assertRawBodyField(t, body, "redirectTo", `""`)
	assertRawBodyField(t, body, "favicon", `""`)
}

func TestStatusPageSerializesEmptyValues(t *testing.T) {
	t.Parallel()
	page := checkly.StatusPage{Name: "Foo status page", URL: "foo-status-page"}
	runPayloadCases(t, []payloadCase{
		{
			name:     "CreateStatusPage",
			method:   http.MethodPost,
			url:      "/v1/status-pages",
			status:   http.StatusCreated,
			fixture:  "CreateStatusPage.json",
			validate: validateStatusPageEmptyPayload,
			call: func(ctx context.Context, c checkly.Client) error {
				_, err := c.CreateStatusPage(ctx, page)
				return err
			},
		},
		{
			name:     "UpdateStatusPage",
			method:   http.MethodPut,
			url:      fmt.Sprintf("/v1/status-pages/%s", testStatusPage.ID),
			status:   http.StatusOK,
			fixture:  "UpdateStatusPage.json",
			validate: validateStatusPageEmptyPayload,
			call: func(ctx context.Context, c checkly.Client) error {
				_, err := c.UpdateStatusPage(ctx, testStatusPage.ID, page)
				return err
			},
		},
	})
}

func TestStatusPageSerializesSetValues(t *testing.T) {
	t.Parallel()
	runPayloadCases(t, []payloadCase{{
		name:    "UpdateStatusPage",
		method:  http.MethodPut,
		url:     fmt.Sprintf("/v1/status-pages/%s", testStatusPage.ID),
		status:  http.StatusOK,
		fixture: "UpdateStatusPage.json",
		validate: func(t *testing.T, body []byte) {
			assertRawBodyField(t, body, "customDomain", `"status.example.com"`)
			assertRawBodyField(t, body, "logo", `"https://example.com/logo.png"`)
			assertRawBodyField(t, body, "redirectTo", `"https://example.com"`)
			assertRawBodyField(t, body, "favicon", `"https://example.com/favicon.ico"`)
		},
		call: func(ctx context.Context, c checkly.Client) error {
			_, err := c.UpdateStatusPage(ctx, testStatusPage.ID, checkly.StatusPage{
				Name:         "Foo status page",
				URL:          "foo-status-page",
				CustomDomain: "status.example.com",
				Logo:         "https://example.com/logo.png",
				RedirectTo:   "https://example.com",
				Favicon:      "https://example.com/favicon.ico",
			})
			return err
		},
	}})
}
