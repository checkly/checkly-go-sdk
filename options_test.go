package checkly_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/checkly/checkly-go-sdk/v2"
)

func TestNewClientDefaults(t *testing.T) {
	t.Parallel()
	got := checkly.ClientSettings(newTestClient(t))
	want := checkly.Settings{
		URL:        checkly.DefaultBaseURL,
		Source:     "go-sdk",
		HTTPClient: http.DefaultClient,
	}
	if got != want {
		t.Errorf("want %+v, got %+v", want, got)
	}
}

func TestClientOptions(t *testing.T) {
	t.Parallel()
	custom := &http.Client{}
	defaults := checkly.Settings{
		URL:        checkly.DefaultBaseURL,
		Source:     "go-sdk",
		HTTPClient: http.DefaultClient,
	}
	with := func(change func(*checkly.Settings)) checkly.Settings {
		s := defaults
		change(&s)
		return s
	}
	tests := []struct {
		name string
		opts []checkly.ClientOption
		want checkly.Settings
	}{
		{
			name: "every option",
			opts: []checkly.ClientOption{
				checkly.WithBaseURL("https://api-test.checklyhq.com"),
				checkly.WithHTTPClient(custom),
				checkly.WithSource("tool"),
			},
			want: checkly.Settings{URL: "https://api-test.checklyhq.com", Source: "tool", HTTPClient: custom},
		},
		{
			name: "later option wins",
			opts: []checkly.ClientOption{
				checkly.WithBaseURL("https://one.example"), checkly.WithBaseURL("https://two.example"),
				checkly.WithHTTPClient(&http.Client{}), checkly.WithHTTPClient(custom),
				checkly.WithSource("one"), checkly.WithSource("two"),
			},
			want: with(func(s *checkly.Settings) {
				s.URL, s.HTTPClient, s.Source = "https://two.example", custom, "two"
			}),
		},
		{
			name: "empty and nil values restore the defaults",
			opts: []checkly.ClientOption{
				checkly.WithBaseURL("https://api-test.checklyhq.com"), checkly.WithBaseURL(""),
				checkly.WithHTTPClient(custom), checkly.WithHTTPClient(nil),
				checkly.WithSource("tool"), checkly.WithSource(""),
			},
			want: defaults,
		},
		{
			name: "nil options are skipped",
			opts: []checkly.ClientOption{nil, checkly.WithSource("tool"), nil},
			want: with(func(s *checkly.Settings) { s.Source = "tool" }),
		},
		{
			name: "base URL with a trailing slash",
			opts: []checkly.ClientOption{checkly.WithBaseURL("https://api-test.checklyhq.com/")},
			want: with(func(s *checkly.Settings) { s.URL = "https://api-test.checklyhq.com" }),
		},
		{
			name: "base URL with several trailing slashes",
			opts: []checkly.ClientOption{checkly.WithBaseURL("https://api-test.checklyhq.com//")},
			want: with(func(s *checkly.Settings) { s.URL = "https://api-test.checklyhq.com" }),
		},
		{
			name: "base URL with a path",
			opts: []checkly.ClientOption{checkly.WithBaseURL("http://localhost:3000/api/")},
			want: with(func(s *checkly.Settings) { s.URL = "http://localhost:3000/api" }),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := checkly.ClientSettings(newTestClient(t, tt.opts...)); got != tt.want {
				t.Errorf("want %+v, got %+v", tt.want, got)
			}
		})
	}
}

func TestWithAuth(t *testing.T) {
	t.Parallel()
	upper := strings.ToUpper(testAccountID)
	const otherAccountID = "b2c3d4e5-0000-4000-8000-000000000002"
	tests := []struct {
		name         string
		opts         []checkly.ClientOption
		key, account string
	}{
		{name: "sets both values", opts: []checkly.ClientOption{checkly.WithAuth("key", testAccountID)}, key: "key", account: testAccountID},
		{name: "uppercase account ID", opts: []checkly.ClientOption{checkly.WithAuth("key", upper)}, key: "key", account: upper},
		{name: "later WithAuth replaces an earlier one", opts: []checkly.ClientOption{checkly.WithAuth("one", testAccountID), checkly.WithAuth("two", otherAccountID)}, key: "two", account: otherAccountID},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := checkly.ClientSettings(newTestClient(t, tt.opts...))
			if got.APIKey != tt.key || got.AccountID != tt.account {
				t.Errorf("want API key %q and account %q, got %q and %q", tt.key, tt.account, got.APIKey, got.AccountID)
			}
		})
	}
}

// setEnv sets the variable for the test, or unsets it when value is empty,
// and restores it afterwards.
func setEnv(t *testing.T, key, value string) {
	t.Setenv(key, value)
	if value == "" {
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
}

// TestWithAuthFromEnv changes the environment, so it does not run in
// parallel.
func TestWithAuthFromEnv(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envAccount   string
		opts         []checkly.ClientOption
		key, account string
	}{
		{name: "both set", envKey: "key", envAccount: testAccountID, key: "key", account: testAccountID},
		{name: "a later WithAuth replaces it", envKey: "key", envAccount: testAccountID, opts: []checkly.ClientOption{checkly.WithAuth("other", testAccountID)}, key: "other", account: testAccountID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnv(t, "CHECKLY_API_KEY", tt.envKey)
			setEnv(t, "CHECKLY_ACCOUNT_ID", tt.envAccount)
			opts := append([]checkly.ClientOption{checkly.WithAuthFromEnv()}, tt.opts...)
			got := checkly.ClientSettings(newTestClient(t, opts...))
			if got.APIKey != tt.key || got.AccountID != tt.account {
				t.Errorf("want API key %q and account %q, got %q and %q", tt.key, tt.account, got.APIKey, got.AccountID)
			}
		})
	}
	t.Run("not read without the option", func(t *testing.T) {
		setEnv(t, "CHECKLY_API_KEY", "key")
		setEnv(t, "CHECKLY_ACCOUNT_ID", testAccountID)
		if got := checkly.ClientSettings(newTestClient(t)); got.APIKey != "" || got.AccountID != "" {
			t.Errorf("want no auth, got API key %q and account %q", got.APIKey, got.AccountID)
		}
	})
}

// newClientError is one construction failure: the options, the error text
// and, if any, the sentinel error it matches.
type newClientError struct {
	name     string
	opts     []checkly.ClientOption
	want     string
	sentinel error
}

func checkNewClientError(t *testing.T, tt newClientError) {
	t.Helper()
	c, err := checkly.NewClient(tt.opts...)
	if c != nil {
		t.Errorf("want a nil Client, got %v", c)
	}
	if err == nil || err.Error() != tt.want {
		t.Fatalf("want error %q, got %v", tt.want, err)
	}
	if tt.sentinel != nil && !errors.Is(err, tt.sentinel) {
		t.Errorf("want error matching %v, got %v", tt.sentinel, err)
	}
}

func TestNewClientErrors(t *testing.T) {
	t.Parallel()
	const invalidAccount = "checkly: WithAuth: the account ID is not a UUID; check that the API key and account ID are not swapped"
	tests := []newClientError{
		{name: "base URL without scheme", opts: []checkly.ClientOption{checkly.WithBaseURL("api.checklyhq.com")}, want: `checkly: WithBaseURL: not an http or https URL`},
		{name: "ftp base URL", opts: []checkly.ClientOption{checkly.WithBaseURL("ftp://x")}, want: `checkly: WithBaseURL: not an http or https URL`},
		{name: "base URL without host", opts: []checkly.ClientOption{checkly.WithBaseURL("https://")}, want: `checkly: WithBaseURL: "https:" has no host`},
		{name: "base URL with only a port", opts: []checkly.ClientOption{checkly.WithBaseURL("http://:8080")}, want: `checkly: WithBaseURL: "http://:8080" has no host`},
		{name: "base URL with a query", opts: []checkly.ClientOption{checkly.WithBaseURL("https://x.example?a=1")}, want: `checkly: WithBaseURL: "https://x.example?a=1" has a query`},
		{name: "base URL with an empty query", opts: []checkly.ClientOption{checkly.WithBaseURL("https://x.example?")}, want: `checkly: WithBaseURL: "https://x.example?" has a query`},
		{name: "base URL with a fragment", opts: []checkly.ClientOption{checkly.WithBaseURL("https://x.example#top")}, want: `checkly: WithBaseURL: "https://x.example#top" has a fragment`},
		{name: "base URL with an empty fragment", opts: []checkly.ClientOption{checkly.WithBaseURL("https://x.example#")}, want: `checkly: WithBaseURL: "https://x.example" has a fragment`},
		{name: "unparseable base URL", opts: []checkly.ClientOption{checkly.WithBaseURL("http://%zz")}, want: "checkly: WithBaseURL: not a valid URL"},
		{name: "base URL with a password and a query", opts: []checkly.ClientOption{checkly.WithBaseURL("https://user:s3cret@proxy.example?x=1")}, want: `checkly: WithBaseURL: "https://user:xxxxx@proxy.example?x=1" has a query`},
		{name: "opaque base URL", opts: []checkly.ClientOption{checkly.WithBaseURL("http:proxy.example")}, want: "checkly: WithBaseURL: not an absolute URL; write it as http://host or https://host"},
		{name: "unparseable base URL with a password", opts: []checkly.ClientOption{checkly.WithBaseURL("https://user:s3cret@%zz")}, want: "checkly: WithBaseURL: not a valid URL"},
		{name: "arguments swapped", opts: []checkly.ClientOption{checkly.WithAuth(testAccountID, "api-key")}, want: invalidAccount, sentinel: checkly.ErrInvalidAccountID},
		{name: "account ID abc", opts: []checkly.ClientOption{checkly.WithAuth("key", "abc")}, want: invalidAccount, sentinel: checkly.ErrInvalidAccountID},
		{name: "account ID missing a UUID group", opts: []checkly.ClientOption{checkly.WithAuth("key", "a1b2c3d4-0000-4000-000000000001")}, want: invalidAccount, sentinel: checkly.ErrInvalidAccountID},
		{name: "empty API key", opts: []checkly.ClientOption{checkly.WithAuth("", testAccountID)}, want: "checkly: WithAuth: the API key is empty", sentinel: checkly.ErrMissingAPIKey},
		{name: "API key with a newline", opts: []checkly.ClientOption{checkly.WithAuth("key\n", testAccountID)}, want: "checkly: WithAuth: the API key contains a character that HTTP headers do not allow, such as a newline", sentinel: checkly.ErrInvalidAPIKey},
		{name: "source with a newline", opts: []checkly.ClientOption{checkly.WithSource("tool\r\n")}, want: "checkly: WithSource: the source contains a character that HTTP headers do not allow, such as a newline"},
		{name: "empty account ID", opts: []checkly.ClientOption{checkly.WithAuth("key", "")}, want: "checkly: WithAuth: the account ID is empty", sentinel: checkly.ErrMissingAccountID},
		{name: "both empty", opts: []checkly.ClientOption{checkly.WithAuth("", "")}, want: "checkly: WithAuth: the API key is empty", sentinel: checkly.ErrMissingAPIKey},
		{name: "empty values after valid auth", opts: []checkly.ClientOption{checkly.WithAuth("key", testAccountID), checkly.WithAuth("", "")}, want: "checkly: WithAuth: the API key is empty", sentinel: checkly.ErrMissingAPIKey},
		{name: "first invalid option wins", opts: []checkly.ClientOption{checkly.WithAuth("key", ""), checkly.WithBaseURL("ftp://x")}, want: "checkly: WithAuth: the account ID is empty", sentinel: checkly.ErrMissingAccountID},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkNewClientError(t, tt)
		})
	}
	t.Run("no password in base URL errors", func(t *testing.T) {
		t.Parallel()
		for _, baseURL := range []string{"https://user:s3cret@proxy.example?x=1", "https://user:s3cret@%zz", "user:s3cret@proxy.example", "http:user:s3cret@proxy.example"} {
			_, err := checkly.NewClient(checkly.WithBaseURL(baseURL))
			if err == nil || strings.Contains(err.Error(), "s3cret") {
				t.Errorf("%s: want an error without the password, got %v", baseURL, err)
			}
		}
	})
	t.Run("parse error is wrapped", func(t *testing.T) {
		t.Parallel()
		var urlErr *url.Error
		if _, err := checkly.NewClient(checkly.WithBaseURL("http://%zz")); !errors.As(err, &urlErr) {
			t.Errorf("want a wrapped *url.Error, got %v", err)
		}
	})
	t.Run("first option error stops construction", func(t *testing.T) {
		t.Parallel()
		var applied bool
		if _, err := checkly.NewClient(checkly.WithBaseURL("ftp://x"), checkly.RecordApplied(&applied)); err == nil {
			t.Fatal("want error, got nil")
		}
		if applied {
			t.Error("want the option after the failing one not applied")
		}
	})
}

// TestWithAuthFromEnvErrors changes the environment, so it does not run in
// parallel.
func TestWithAuthFromEnvErrors(t *testing.T) {
	const keyNotSet = "checkly: WithAuthFromEnv: CHECKLY_API_KEY is not set"
	const accountNotSet = "checkly: WithAuthFromEnv: CHECKLY_ACCOUNT_ID is not set"
	tests := []struct {
		newClientError
		envKey, envAccount string
		// keySetEmpty sets CHECKLY_API_KEY to "" instead of unsetting it.
		keySetEmpty bool
		// after are options that follow WithAuthFromEnv.
		after []checkly.ClientOption
	}{
		{
			newClientError: newClientError{name: "neither set", want: keyNotSet, sentinel: checkly.ErrMissingAPIKey},
		},
		{
			newClientError: newClientError{name: "only the account ID set", want: keyNotSet, sentinel: checkly.ErrMissingAPIKey},
			envAccount:     testAccountID,
		},
		{
			newClientError: newClientError{name: "empty key", want: keyNotSet, sentinel: checkly.ErrMissingAPIKey},
			envAccount:     testAccountID,
			keySetEmpty:    true,
		},
		{
			newClientError: newClientError{name: "key with a newline", want: "checkly: WithAuthFromEnv: CHECKLY_API_KEY contains a character that HTTP headers do not allow, such as a newline", sentinel: checkly.ErrInvalidAPIKey},
			envKey:         "key\n",
			envAccount:     testAccountID,
		},
		{
			newClientError: newClientError{name: "only the key set", want: accountNotSet, sentinel: checkly.ErrMissingAccountID},
			envKey:         "key",
		},
		{
			newClientError: newClientError{name: "account ID not a UUID", want: "checkly: WithAuthFromEnv: CHECKLY_ACCOUNT_ID is not a UUID", sentinel: checkly.ErrInvalidAccountID},
			envKey:         "key",
			envAccount:     "abc",
		},
		{
			// The first invalid option stops NewClient, so a later valid
			// WithAuth does not rescue an invalid account ID.
			newClientError: newClientError{name: "account ID not a UUID before a valid WithAuth", want: "checkly: WithAuthFromEnv: CHECKLY_ACCOUNT_ID is not a UUID", sentinel: checkly.ErrInvalidAccountID},
			envKey:         "key",
			envAccount:     "abc",
			after:          []checkly.ClientOption{checkly.WithAuth("key", testAccountID)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnv(t, "CHECKLY_API_KEY", tt.envKey)
			if tt.keySetEmpty {
				t.Setenv("CHECKLY_API_KEY", "")
			}
			setEnv(t, "CHECKLY_ACCOUNT_ID", tt.envAccount)
			tt.opts = append([]checkly.ClientOption{checkly.WithAuthFromEnv()}, tt.after...)
			checkNewClientError(t, tt.newClientError)
		})
	}
}

// sentRequest is what headerServer records of a request.
type sentRequest struct {
	path   string
	header http.Header
}

// headerServer answers every request with 200 and body, and returns the
// requests it received so far.
func headerServer(t *testing.T, body string) (*httptest.Server, func() []sentRequest) {
	t.Helper()
	var (
		mu       sync.Mutex
		requests []sentRequest
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		mu.Lock()
		requests = append(requests, sentRequest{path: r.URL.Path, header: r.Header.Clone()})
		mu.Unlock()
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	return srv, func() []sentRequest {
		mu.Lock()
		defer mu.Unlock()
		return append([]sentRequest(nil), requests...)
	}
}

// uploadBundle calls UploadCodeBundle, which builds its request without the
// shared request helpers.
func uploadBundle(ctx context.Context, c checkly.Client) error {
	data := "bundle"
	_, err := c.UploadCodeBundle(ctx, strings.NewReader(data), int64(len(data)), checkly.UploadCodeBundleOptions{})
	return err
}

func TestMissingAPIKey(t *testing.T) {
	t.Parallel()
	calls := map[string]struct {
		call func(context.Context, checkly.Client) error
		want string
	}{
		"GetCheck": {
			call: func(ctx context.Context, c checkly.Client) error {
				_, err := c.GetCheck(ctx, "abc")
				return err
			},
			want: "GET /v1/checks/abc: no API key configured; set one with WithAuth",
		},
		"UploadCodeBundle": {
			call: uploadBundle,
			want: "POST /next/checkly-storage/upload-code-bundle: no API key configured; set one with WithAuth",
		},
	}
	for name, tt := range calls {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			srv, sent := headerServer(t, "{}")
			err := tt.call(context.Background(), newTestClient(t, checkly.WithBaseURL(srv.URL), checkly.WithHTTPClient(srv.Client())))
			if err == nil || err.Error() != tt.want {
				t.Errorf("want error %q, got %v", tt.want, err)
			}
			if !errors.Is(err, checkly.ErrMissingAPIKey) {
				t.Errorf("want error matching ErrMissingAPIKey, got %v", err)
			}
			var apiErr *checkly.APIError
			if errors.As(err, &apiErr) || errors.Is(err, checkly.ErrUnauthorized) {
				t.Errorf("want neither *APIError nor ErrUnauthorized, got %v", err)
			}
			if n := len(sent()); n != 0 {
				t.Errorf("want no requests, got %d", n)
			}
		})
	}
}

func TestPublicMethodsSendNoCredentials(t *testing.T) {
	t.Parallel()
	calls := map[string]func(context.Context, checkly.Client) error{
		"GetStaticIPs": func(ctx context.Context, c checkly.Client) error {
			_, err := c.GetStaticIPs(ctx)
			return err
		},
		"GetRuntime": func(ctx context.Context, c checkly.Client) error {
			_, err := c.GetRuntime(ctx, "2024.02")
			return err
		},
	}
	clients := map[string][]checkly.ClientOption{
		"no auth":             nil,
		"API key and account": {checkly.WithAuth("key", testAccountID)},
	}
	for callName, call := range calls {
		for clientName, clientOpts := range clients {
			call, clientOpts := call, clientOpts
			t.Run(callName+"/"+clientName, func(t *testing.T) {
				t.Parallel()
				srv, sent := headerServer(t, "{}")
				opts := append([]checkly.ClientOption{checkly.WithBaseURL(srv.URL), checkly.WithHTTPClient(srv.Client())}, clientOpts...)
				if err := call(context.Background(), newTestClient(t, opts...)); err != nil {
					t.Fatalf("want success, got %v", err)
				}
				requests := sent()
				if len(requests) == 0 {
					t.Fatal("want a request, got none")
				}
				for _, r := range requests {
					for _, h := range []string{"Authorization", "X-Checkly-Account"} {
						if v := r.header.Get(h); v != "" {
							t.Errorf("%s: want no %s header, got %q", r.path, h, v)
						}
					}
					if got := r.header.Get("X-Checkly-Source"); got != "go-sdk" {
						t.Errorf("%s: want x-checkly-source go-sdk, got %q", r.path, got)
					}
				}
			})
		}
	}
}

func TestRequestHeaders(t *testing.T) {
	t.Parallel()
	getCheck := func(ctx context.Context, c checkly.Client) error {
		_, err := c.GetCheck(ctx, "abc")
		return err
	}
	tests := []struct {
		name string
		opts []checkly.ClientOption
		want map[string]string
	}{
		{
			name: "API key, account and source",
			opts: []checkly.ClientOption{checkly.WithAuth("key", testAccountID), checkly.WithSource("tool")},
			want: map[string]string{"Authorization": "Bearer key", "X-Checkly-Account": testAccountID, "X-Checkly-Source": "tool"},
		},
		{
			name: "API key and account with the default source restored",
			opts: []checkly.ClientOption{checkly.WithAuth("key", testAccountID), checkly.WithSource("tool"), checkly.WithSource("")},
			want: map[string]string{"Authorization": "Bearer key", "X-Checkly-Account": testAccountID, "X-Checkly-Source": "go-sdk"},
		},
	}
	calls := map[string]struct {
		path string
		call func(context.Context, checkly.Client) error
	}{
		"GetCheck":         {"/v1/checks/abc", getCheck},
		"UploadCodeBundle": {"/next/checkly-storage/upload-code-bundle", uploadBundle},
	}
	for _, tt := range tests {
		for callName, call := range calls {
			tt, call := tt, call
			t.Run(tt.name+"/"+callName, func(t *testing.T) {
				t.Parallel()
				srv, sent := headerServer(t, `{"checkType":"API"}`)
				// The trailing slash checks that the base URL is joined
				// without a double slash.
				opts := append([]checkly.ClientOption{checkly.WithBaseURL(srv.URL + "/"), checkly.WithHTTPClient(srv.Client())}, tt.opts...)
				if err := call.call(context.Background(), newTestClient(t, opts...)); err != nil {
					t.Fatalf("want success, got %v", err)
				}
				requests := sent()
				if len(requests) != 1 {
					t.Fatalf("want 1 request, got %d", len(requests))
				}
				if requests[0].path != call.path {
					t.Errorf("want path %q, got %q", call.path, requests[0].path)
				}
				got := map[string]string{}
				for h := range tt.want {
					got[h] = requests[0].header.Get(h)
				}
				if !cmp.Equal(tt.want, got) {
					t.Error(cmp.Diff(tt.want, got))
				}
			})
		}
	}
}
