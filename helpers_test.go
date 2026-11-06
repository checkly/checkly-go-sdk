package checkly_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	checkly "github.com/checkly/checkly-go-sdk/v2"
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
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		validate(t, body)
		w.WriteHeader(status)
		data, err := os.Open(fmt.Sprintf("fixtures/%s", filename))
		if err != nil {
			t.Error(err)
		}
		defer func() { _ = data.Close() }()
		_, _ = io.Copy(w, data)
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
	assertRawBodyPath(t, body, want, key)
}

// assertRawBodyKeyAbsent asserts that the raw JSON request body does not
// contain the given key.
func assertRawBodyKeyAbsent(t *testing.T, body []byte, key string) {
	t.Helper()
	assertRawBodyPathAbsent(t, body, key)
}

// assertRawNestedBodyField asserts that the JSON object at the given
// top-level key of the raw request body contains the given nested key with
// exactly the given serialized value.
func assertRawNestedBodyField(t *testing.T, body []byte, key, nestedKey, want string) {
	t.Helper()
	assertRawBodyPath(t, body, want, key, nestedKey)
}

// lookupRawBodyPath returns the raw JSON value at the given path of nested
// object keys in body, and whether every key on the path is present.
func lookupRawBodyPath(t *testing.T, body []byte, path []string) (json.RawMessage, bool) {
	t.Helper()
	value := json.RawMessage(body)
	for _, key := range path {
		var object map[string]json.RawMessage
		if err := json.Unmarshal(value, &object); err != nil {
			t.Fatalf("decoding error at %v for data %q: %v", path, body, err)
		}
		next, ok := object[key]
		if !ok {
			return nil, false
		}
		value = next
	}
	return value, true
}

// assertRawBodyPath asserts that the raw JSON request body contains the
// value want, serialized exactly, at the given path of nested object keys.
func assertRawBodyPath(t *testing.T, body []byte, want string, path ...string) {
	t.Helper()
	got, ok := lookupRawBodyPath(t, body, path)
	if !ok {
		t.Errorf("missing key %v in request body", path)
		return
	}
	if string(got) != want {
		t.Errorf("key %v in request body: want %s, got %s", path, want, got)
	}
}

// assertRawBodyPathAbsent asserts that the raw JSON request body has no key
// at the given path of nested object keys.
func assertRawBodyPathAbsent(t *testing.T, body []byte, path ...string) {
	t.Helper()
	if got, ok := lookupRawBodyPath(t, body, path); ok {
		t.Errorf("expected key %v to be absent from request body, got %s", path, got)
	}
}

// payloadCase is a client call whose request body a test validates.
type payloadCase struct {
	name     string
	method   string
	url      string
	status   int
	fixture  string
	validate func(*testing.T, []byte)
	call     func(context.Context, checkly.Client) error
}

// runPayloadCases runs each case against a server that validates the
// request body and answers with the case's fixture.
func runPayloadCases(t *testing.T, cases []payloadCase) {
	t.Helper()
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ts := cannedResponseServer(t, tc.method, tc.url, tc.validate, tc.status, tc.fixture)
			defer ts.Close()
			client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
			if err := tc.call(context.Background(), client); err != nil {
				t.Error(err)
			}
		})
	}
}

// newRecordingServer answers every request with status and body, and returns
// a client for it and a function that returns the body of the last request.
func newRecordingServer(t *testing.T, status int, body string) (checkly.Client, func() []byte) {
	t.Helper()
	var (
		mu   sync.Mutex
		last []byte
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		mu.Lock()
		last = got
		mu.Unlock()
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	c := newTestClient(t, checkly.WithBaseURL(srv.URL), checkly.WithAuth("test-key", testAccountID), checkly.WithHTTPClient(srv.Client()))
	return c, func() []byte {
		mu.Lock()
		defer mu.Unlock()
		return last
	}
}
