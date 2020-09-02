package checkly_test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	checkly "github.com/checkly/checkly-go-sdk"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var wantCheckID = "73d29e72-6540-4bb5-967e-e07fa2c9465e"

var wantCheck = checkly.Check{
	Name:        "test",
	Type:        checkly.TypeAPI,
	Frequency:   10,
	Activated:   true,
	Muted:       false,
	DoubleCheck: true,
	SSLCheck:    true,
	ShouldFail:  false,
	Locations:   []string{"eu-west-1"},
	Request: checkly.Request{
		Method: http.MethodGet,
		URL:    "https://example.com",
		Headers: []checkly.KeyValue{
			{
				Key:   "X-Test",
				Value: "foo",
			},
		},
		QueryParameters: []checkly.KeyValue{
			{
				Key:   "query",
				Value: "foo",
			},
		},
		Assertions: []checkly.Assertion{
			{
				Source:     checkly.StatusCode,
				Comparison: checkly.Equals,
				Target:     "200",
			},
		},
		Body:     "",
		BodyType: "NONE",
		BasicAuth: &checkly.BasicAuth{
			Username: "",
			Password: "",
		},
	},
	Script: "foo",
	EnvironmentVariables: []checkly.EnvironmentVariable{
		{
			Key:   "ENVTEST",
			Value: "Hello world",
		},
	},
	Tags: []string{
		"foo",
		"bar",
	},
	LocalSetupScript:    "setitup",
	LocalTearDownScript: "tearitdown",
	AlertSettings: checkly.AlertSettings{
		EscalationType: checkly.RunBased,
		RunBasedEscalation: checkly.RunBasedEscalation{
			FailedRunThreshold: 1,
		},
		TimeBasedEscalation: checkly.TimeBasedEscalation{
			MinutesFailingThreshold: 5,
		},
		Reminders: checkly.Reminders{
			Interval: 5,
		},
		SSLCertificates: checkly.SSLCertificates{
			Enabled:        false,
			AlertThreshold: 30,
		},
	},
	UseGlobalAlertSettings: false,
	DegradedResponseTime:   15000,
	MaxResponseTime:        30000,
	GroupID:                0,
	GroupOrder:             0,
}

func cannedResponseServer(t *testing.T, wantMethod string, wantURL string, validate func(*testing.T, []byte), status int, filename string) *httptest.Server {
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wantMethod != r.Method {
			t.Errorf("want %q request, got %q", wantMethod, r.Method)
		}
		if r.URL.String() != wantURL {
			t.Errorf("want %q, got %q", wantURL, r.URL.String())
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		validate(t, body)
		w.WriteHeader(status)
		data, err := os.Open(fmt.Sprintf("testdata/%s", filename))
		if err != nil {
			t.Fatal(err)
		}
		defer data.Close()
		io.Copy(w, data)
	}))
}

func validateCheck(t *testing.T, body []byte) {
	var gotCheck checkly.Check
	err := json.Unmarshal(body, &gotCheck)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(wantCheck, gotCheck) {
		t.Error(cmp.Diff(wantCheck, gotCheck))
	}
}

func validateEmptyBody(t *testing.T, body []byte) {
	if len(body) > 0 {
		t.Errorf("expected empty body, but got %q", body)
	}
}

func validateAnything(*testing.T, []byte) {
}

var ignoreCheckFields = cmpopts.IgnoreFields(checkly.Check{}, "ID")

func TestAPIError(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks",
		validateAnything,
		http.StatusBadRequest,
		"BadRequest.json",
	)
	defer ts.Close()
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	_, err := client.Create(checkly.Check{})
	if err == nil {
		t.Fatal("want error when API returns 'bad request' status, got nil")
	}
	if !strings.Contains(err.Error(), "frequency") {
		t.Errorf("want API error value to contain 'frequency', got %q", err.Error())
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/checks",
		validateCheck,
		http.StatusCreated,
		"CreateCheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	gotCheck, err := client.Create(wantCheck)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(wantCheck, gotCheck, ignoreCheckFields) {
		t.Error(cmp.Diff(wantCheck, gotCheck, ignoreCheckFields))
	}
}

func TestGet(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/checks/%s", wantCheckID),
		validateEmptyBody,
		http.StatusOK,
		"GetCheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	gotCheck, err := client.Get(wantCheckID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(wantCheck, gotCheck, ignoreCheckFields) {
		t.Error(cmp.Diff(wantCheck, gotCheck, ignoreCheckFields))
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/checks/%s", wantCheckID),
		validateCheck,
		http.StatusOK,
		"UpdateCheck.json",
	)
	defer ts.Close()
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	gotCheck, err := client.Update(wantCheckID, wantCheck)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(wantCheck, gotCheck, ignoreCheckFields) {
		t.Error(cmp.Diff(wantCheck, gotCheck, ignoreCheckFields))
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/checks/%s", wantCheckID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	err := client.Delete(wantCheckID)
	if err != nil {
		t.Fatal(err)
	}
}

var wantGroupID int64 = 135

var wantGroup = checkly.Group{
	Name:        "test",
	Activated:   true,
	Muted:       false,
	Tags:        []string{"auto"},
	Locations:   []string{"eu-west-1"},
	Concurrency: 3,
	APICheckDefaults: checkly.APICheckDefaults{
		BaseURL: "example.com/api/test",
		Headers: []checkly.KeyValue{
			{
				Key:   "X-Test",
				Value: "foo",
			},
		},
		QueryParameters: []checkly.KeyValue{
			{
				Key:   "query",
				Value: "foo",
			},
		},
		Assertions: []checkly.Assertion{
			{
				Source:     checkly.StatusCode,
				Comparison: checkly.Equals,
				Target:     "200",
			},
		},
		BasicAuth: checkly.BasicAuth{
			Username: "user",
			Password: "pass",
		},
	},
	EnvironmentVariables: []checkly.EnvironmentVariable{
		{
			Key:   "ENVTEST",
			Value: "Hello world",
		},
	},
	DoubleCheck:            true,
	UseGlobalAlertSettings: false,
	AlertSettings: checkly.AlertSettings{
		EscalationType: checkly.RunBased,
		RunBasedEscalation: checkly.RunBasedEscalation{
			FailedRunThreshold: 1,
		},
		TimeBasedEscalation: checkly.TimeBasedEscalation{
			MinutesFailingThreshold: 5,
		},
		Reminders: checkly.Reminders{
			Amount:   0,
			Interval: 5,
		},
		SSLCertificates: checkly.SSLCertificates{
			Enabled:        true,
			AlertThreshold: 30,
		},
	},
	AlertChannelSubscriptions: []checkly.Subscription{
		{
			Activated:      true,
			AlertChannelID: 2996,
		},
	},
	LocalSetupScript:    "setup-test",
	LocalTearDownScript: "teardown-test",
}

func validateGroup(t *testing.T, body []byte) {
	var gotGroup checkly.Group
	err := json.Unmarshal(body, &gotGroup)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(wantGroup, gotGroup) {
		t.Error(cmp.Diff(wantGroup, gotGroup))
	}
}

var ignoreGroupFields = cmpopts.IgnoreFields(checkly.Group{}, "ID")

func TestCreateGroup(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/check-groups",
		validateGroup,
		http.StatusCreated,
		"CreateGroup.json",
	)
	defer ts.Close()
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	gotGroup, err := client.CreateGroup(wantGroup)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(wantGroup, gotGroup, ignoreGroupFields) {
		t.Error(cmp.Diff(wantGroup, gotGroup, ignoreGroupFields))
	}
}

func TestGetGroup(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/check-groups/%d", wantGroupID),
		validateEmptyBody,
		http.StatusOK,
		"CreateGroup.json",
	)
	defer ts.Close()
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	gotGroup, err := client.GetGroup(wantGroupID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(wantGroup, gotGroup, ignoreGroupFields) {
		t.Error(cmp.Diff(wantGroup, gotGroup, ignoreGroupFields))
	}
}

func TestUpdateGroup(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/check-groups/%d", wantGroupID),
		validateGroup,
		http.StatusOK,
		"UpdateGroup.json",
	)
	defer ts.Close()
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	gotGroup, err := client.UpdateGroup(wantGroupID, wantGroup)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(wantGroup, gotGroup, ignoreGroupFields) {
		t.Error(cmp.Diff(wantGroup, gotGroup, ignoreGroupFields))
	}
}

func TestDeleteGroup(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/check-groups/%d", wantGroupID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := checkly.NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	err := client.DeleteGroup(wantGroupID)
	if err != nil {
		t.Fatal(err)
	}
}
