package checkly_test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	checkly "github.com/checkly/checkly-go-sdk"
)

// testAccountID is the account ID test clients authenticate with: account
// IDs are UUIDs, and an API key needs one.
const testAccountID = "a1b2c3d4-0000-4000-8000-000000000001"

// newTestClient returns the client NewClient builds from opts, failing the
// test if NewClient returns an error.
func newTestClient(t testing.TB, opts ...checkly.ClientOption) checkly.Client {
	t.Helper()
	c, err := checkly.NewClient(opts...)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func cannedResponseServer(
	t *testing.T,
	wantMethod string,
	wantURL string,
	validate func(*testing.T, []byte),
	status int,
	filename string,
) *httptest.Server {
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wantMethod != r.Method {
			t.Errorf("want %q request, got %q", wantMethod, r.Method)
		}
		if r.URL.String() != wantURL {
			t.Errorf("want %q, got %q", wantURL, r.URL.String())
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		validate(t, body)
		w.WriteHeader(status)
		data, err := os.Open(fmt.Sprintf("fixtures/%s", filename))
		if err != nil {
			t.Error(err)
		}
		defer data.Close()
		io.Copy(w, data)
	}))
}

func validateEmptyBody(t *testing.T, body []byte) {
	if len(body) > 0 {
		t.Errorf("expected empty body, but got %q", body)
	}
}

func validateAnything(*testing.T, []byte) {
}

var trueVal = true

// The v3 endpoints differ in how they treat a missing key: the page update
// only writes the keys present, while component and automation rule updates
// replace the resource wholesale. These tests pin the raw JSON, because
// decoding it back into the struct cannot tell an omitted key from an empty
// one.
func marshalToMap(t *testing.T, v interface{}) map[string]interface{} {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

// assertRawBodyField asserts that the raw JSON request body contains the
// given key with exactly the given serialized value. This catches fields
// added or normalized by the payload wrappers, which a round trip through
// the public struct cannot verify.
func assertRawBodyField(t *testing.T, body []byte, key, want string) {
	t.Helper()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	got, ok := raw[key]
	if !ok {
		t.Errorf("missing key %q in request body", key)
		return
	}
	if string(got) != want {
		t.Errorf("key %q in request body: want %s, got %s", key, want, got)
	}
}

// assertRawBodyKeyAbsent asserts that the raw JSON request body does not
// contain the given key.
func assertRawBodyKeyAbsent(t *testing.T, body []byte, key string) {
	t.Helper()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	if got, ok := raw[key]; ok {
		t.Errorf("expected key %q to be absent from request body, got %s", key, got)
	}
}

// assertRawNestedBodyField asserts that the JSON object at the given
// top-level key of the raw request body contains the given nested key with
// exactly the given serialized value.
func assertRawNestedBodyField(t *testing.T, body []byte, key, nestedKey, want string) {
	t.Helper()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("decoding error for data %q: %v", body, err)
	}
	outer, ok := raw[key]
	if !ok {
		t.Errorf("missing key %q in request body", key)
		return
	}
	var nested map[string]json.RawMessage
	if err := json.Unmarshal(outer, &nested); err != nil {
		t.Fatalf("decoding error for key %q data %q: %v", key, outer, err)
	}
	got, ok := nested[nestedKey]
	if !ok {
		t.Errorf("missing key %q.%q in request body", key, nestedKey)
		return
	}
	if string(got) != want {
		t.Errorf("key %q.%q in request body: want %s, got %s", key, nestedKey, want, got)
	}
}
