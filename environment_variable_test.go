package checkly_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/checkly/checkly-go-sdk/v2"
	"github.com/google/go-cmp/cmp"
)

var testEnvVariable = checkly.EnvironmentVariable{
	Key:   "k1",
	Value: "v1",
}

func validateEnvVariable(t *testing.T, body []byte) {
	var result checkly.EnvironmentVariable
	err := json.Unmarshal(body, &result)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testEnvVariable, result) {
		t.Error(cmp.Diff(testEnvVariable, result))
	}
}

func TestCreateEnvironmentVariable(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/variables",
		validateEnvVariable,
		http.StatusCreated,
		"CreateEnvironmentVariable.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	result, err := client.CreateEnvironmentVariable(context.Background(), testEnvVariable)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testEnvVariable, *result, nil) {
		t.Error(cmp.Diff(testEnvVariable, *result, nil))
	}
}

func TestGetEnvironmentVariable(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/variables/%s", testEnvVariable.Key),
		validateEmptyBody,
		http.StatusOK,
		"CreateEnvironmentVariable.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	result, err := client.GetEnvironmentVariable(context.Background(), testEnvVariable.Key)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testEnvVariable, *result, nil) {
		t.Error(cmp.Diff(testEnvVariable, *result, nil))
	}
}

func TestUpdateEnvironmentVariable(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/variables/%s", testEnvVariable.Key),
		validateEnvVariable,
		http.StatusOK,
		"UpdateEnvironmentVariable.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	result, err := client.UpdateEnvironmentVariable(
		context.Background(),
		testEnvVariable.Key,
		testEnvVariable,
	)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testEnvVariable, *result, nil) {
		t.Error(cmp.Diff(testEnvVariable, *result, nil))
	}
}

func TestDeleteEnvironmentVariable(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/variables/%s", testEnvVariable.Key),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	err := client.DeleteEnvironmentVariable(context.Background(), testEnvVariable.Key)
	if err != nil {
		t.Error(err)
	}
}
