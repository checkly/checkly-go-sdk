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

var ignoreSnippetFields = cmpopts.IgnoreFields(checkly.Snippet{}, "ID")

var testSnippet = checkly.Snippet{
	ID:     1,
	Name:   "snippet1",
	Script: "script1",
}

func validateSnippet(t *testing.T, body []byte) {
	var gotSnippet checkly.Snippet
	err := json.Unmarshal(body, &gotSnippet)
	if err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if !cmp.Equal(testSnippet, gotSnippet) {
		t.Error(cmp.Diff(testSnippet, gotSnippet))
	}
}

func TestCreateSnippet(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPost,
		"/v1/snippets",
		validateSnippet,
		http.StatusCreated,
		"CreateSnippet.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	gotSnippet, err := client.CreateSnippet(context.Background(), testSnippet)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testSnippet, *gotSnippet, ignoreSnippetFields) {
		t.Error(cmp.Diff(testSnippet, *gotSnippet, ignoreSnippetFields))
	}
}

func TestGetSnippet(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodGet,
		fmt.Sprintf("/v1/snippets/%d", testSnippet.ID),
		validateEmptyBody,
		http.StatusOK,
		"CreateSnippet.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	gotSnippet, err := client.GetSnippet(context.Background(), testSnippet.ID)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testSnippet, *gotSnippet, ignoreSnippetFields) {
		t.Error(cmp.Diff(testSnippet, *gotSnippet, ignoreSnippetFields))
	}
}

func TestUpdateSnippet(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodPut,
		fmt.Sprintf("/v1/snippets/%d", testSnippet.ID),
		validateSnippet,
		http.StatusOK,
		"UpdateSnippet.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	gotSnippet, err := client.UpdateSnippet(context.Background(), testSnippet.ID, testSnippet)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testSnippet, *gotSnippet, ignoreSnippetFields) {
		t.Error(cmp.Diff(testSnippet, *gotSnippet, ignoreSnippetFields))
	}
}

func TestDeleteSnippet(t *testing.T) {
	t.Parallel()
	ts := cannedResponseServer(t,
		http.MethodDelete,
		fmt.Sprintf("/v1/snippets/%d", testSnippet.ID),
		validateEmptyBody,
		http.StatusNoContent,
		"Empty.json",
	)
	defer ts.Close()
	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	err := client.DeleteSnippet(context.Background(), testSnippet.ID)
	if err != nil {
		t.Error(err)
	}
}
