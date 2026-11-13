package checkly_test

import (
	"context"
	"testing"

	"github.com/checkly/checkly-go-sdk/v2"
	"github.com/google/go-cmp/cmp"
)

// getAccount calls GetAccount on a client whose server answers with body,
// and fails the test unless the client requested the current account.
func getAccount(t *testing.T, body string) *checkly.Account {
	t.Helper()
	srv, sent := headerServer(t, body)
	c := newTestClient(t, checkly.WithBaseURL(srv.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(srv.Client()))
	got, err := c.GetAccount(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if requests := sent(); len(requests) != 1 || requests[0].path != "/v1/accounts/me" {
		t.Errorf("want one request to /v1/accounts/me, got %+v", requests)
	}
	return got
}

func TestGetAccount(t *testing.T) {
	t.Parallel()
	got := getAccount(t, readFixture(t, "GetAccount.json"))
	want := checkly.Account{
		ID:       testAccountID,
		Name:     "Example account",
		Features: []string{checkly.AccountFeatureExtendedResponseTimeLimits},
		Integrations: []checkly.AccountIntegration{
			{Type: checkly.AccountIntegrationSlack, Connected: false},
		},
	}
	if !cmp.Equal(want, *got) {
		t.Error(cmp.Diff(want, *got))
	}
	if !got.HasFeature(checkly.AccountFeatureExtendedResponseTimeLimits) {
		t.Errorf("want HasFeature(%q) true", checkly.AccountFeatureExtendedResponseTimeLimits)
	}
	if got.HasFeature("OTHER_FEATURE") {
		t.Error(`want HasFeature("OTHER_FEATURE") false`)
	}
}

// An API that reports no features, or predates reporting them, yields an
// account without features.
func TestGetAccountWithoutFeatures(t *testing.T) {
	t.Parallel()
	for name, body := range map[string]string{
		"absent": `{"id":"` + testAccountID + `","name":"Example account"}`,
		"empty":  `{"id":"` + testAccountID + `","name":"Example account","features":[]}`,
		"null":   `{"id":"` + testAccountID + `","name":"Example account","features":null}`,
	} {
		body := body
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := getAccount(t, body)
			if len(got.Features) != 0 {
				t.Errorf("want no features, got %q", got.Features)
			}
			if got.HasFeature(checkly.AccountFeatureExtendedResponseTimeLimits) {
				t.Errorf("want HasFeature(%q) false", checkly.AccountFeatureExtendedResponseTimeLimits)
			}
		})
	}
}

func TestAccountIntegration(t *testing.T) {
	t.Parallel()
	account := checkly.Account{Integrations: []checkly.AccountIntegration{
		{Type: "OTHER", Connected: false},
		{Type: checkly.AccountIntegrationSlack, Connected: true},
	}}
	got, ok := account.Integration(checkly.AccountIntegrationSlack)
	if want := (checkly.AccountIntegration{Type: checkly.AccountIntegrationSlack, Connected: true}); !ok || got != want {
		t.Errorf("Integration(%q) = %+v, %v; want %+v, true", checkly.AccountIntegrationSlack, got, ok, want)
	}
	// A type the API does not report is not found, rather than reported
	// as not connected.
	if got, ok := (&checkly.Account{}).Integration(checkly.AccountIntegrationSlack); ok {
		t.Errorf("Integration on an account without integrations = %+v, true; want false", got)
	}
}
