// +build integration

package checkly

import (
	"net/http"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func getAPIKey(t *testing.T) string {
	key := os.Getenv("CHECKLY_API_KEY")
	if key == "" {
		t.Fatal("'CHECKLY_API_KEY' must be set for integration tests")
	}
	return key
}

func testCheck(name string) Check {
	return Check{
		Name:                 name,
		Type:                 TypeAPI,
		Frequency:            1,
		Activated:            true,
		Muted:                false,
		ShouldFail:           false,
		Locations:            []string{"eu-west-1"},
		Script:               "foo",
		DegradedResponseTime: 15000,
		MaxResponseTime:      30000,
		EnvironmentVariables: []EnvironmentVariable{
			{
				Key:   "ENVTEST",
				Value: "Hello world",
			},
		},
		DoubleCheck: false,
		Tags: []string{
			"foo",
			"bar",
		},
		SSLCheck:            true,
		SSLCheckDomain:      "example.com",
		LocalSetupScript:    "bogus",
		LocalTearDownScript: "bogus",
		AlertSettings: AlertSettings{
			EscalationType: RunBased,
			RunBasedEscalation: RunBasedEscalation{
				FailedRunThreshold: 1,
			},
			TimeBasedEscalation: TimeBasedEscalation{
				MinutesFailingThreshold: 5,
			},
			Reminders: Reminders{
				Interval: 5,
			},
			SSLCertificates: SSLCertificates{
				Enabled:        false,
				AlertThreshold: 3,
			},
		},
		AlertChannelSubscriptions: []Subscription{
			{
				AlertChannelID: 2996,
				Activated:      true,
			},
		},
		UseGlobalAlertSettings: false,
		Request: Request{
			Method: http.MethodGet,
			URL:    "http://example.com",
			Headers: []KeyValue{
				{
					Key:   "X-Test",
					Value: "foo",
				},
			},
			QueryParameters: []KeyValue{
				{
					Key:   "query",
					Value: "foo",
				},
			},
			Assertions: []Assertion{
				{
					Source:     StatusCode,
					Comparison: Equals,
					Target:     "200",
				},
			},
			Body:     "",
			BodyType: "NONE",
		},
	}
}

func TestCreateGetIntegration(t *testing.T) {
	t.Parallel()
	client := NewClient(getAPIKey(t))
	checkCreate := testCheck("integrationTestCreate")
	// client.Debug = os.Stdout
	ID, err := client.Create(checkCreate)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Delete(ID)
	check, err := client.Get(ID)
	if err != nil {
		t.Error(err)
	}
	checkCreate.ID = ID
	if !cmp.Equal(checkCreate, check, cmpopts.IgnoreFields(Check{}, "CreatedAt", "UpdatedAt")) {
		t.Error(cmp.Diff(checkCreate, check))
	}
}

func TestUpdateIntegration(t *testing.T) {
	t.Parallel()
	client := NewClient(getAPIKey(t))
	checkUpdate := testCheck("integrationTestUpdate")
	// client.Debug = os.Stdout
	ID, err := client.Create(checkUpdate)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Delete(ID)
	checkUpdate.Name = "integrationTestUpdate2"
	err = client.Update(ID, checkUpdate)
	if err != nil {
		t.Error(err)
	}
	check, err := client.Get(ID)
	if err != nil {
		t.Error(err)
	}
	checkUpdate.ID = ID
	if !cmp.Equal(checkUpdate, check, cmpopts.IgnoreFields(Check{}, "CreatedAt", "UpdatedAt")) {
		t.Error(cmp.Diff(checkUpdate, check))
	}
}

func TestDeleteIntegration(t *testing.T) {
	t.Parallel()
	client := NewClient(getAPIKey(t))
	checkDelete := testCheck("integrationTestDelete")
	ID, err := client.Create(checkDelete)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Delete(ID); err != nil {
		t.Error(err)
	}
	_, err = client.Get(ID)
	if err == nil {
		t.Error("want error getting deleted check, but got nil")
	}
}
