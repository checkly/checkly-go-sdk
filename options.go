package checkly

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// ClientOption configures a client built by NewClient. Each option validates
// its value when NewClient applies it, and the first invalid option stops
// NewClient with its error. Among valid options, which apply in order, a
// later one overrides an earlier one; for WithBaseURL, WithHTTPClient and
// WithSource an empty or nil value restores the default.
type ClientOption func(*client) error

// The environment variables WithAuthFromEnv reads.
const (
	envAPIKey    = "CHECKLY_API_KEY"
	envAccountID = "CHECKLY_ACCOUNT_ID"
)

// isUUID reports whether s is a UUID in the canonical 8-4-4-4-12 hexadecimal
// form, in either case. Account IDs are UUIDs of any version and variant.
func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, r := range s {
		switch i {
		case 8, 13, 18, 23:
			if r != '-' {
				return false
			}
		default:
			if !('0' <= r && r <= '9' || 'a' <= r && r <= 'f' || 'A' <= r && r <= 'F') {
				return false
			}
		}
	}
	return true
}

// WithBaseURL sets the API root, such as "https://api.checklyhq.com". It must
// be an http or https URL with a host and without a query or fragment; a
// path is allowed. Trailing slashes are trimmed. An empty string restores
// DefaultBaseURL.
//
// The error for an invalid value quotes the URL with any userinfo password
// redacted. When the value does not parse, the error says only that it is not
// a valid URL: it wraps the *url.Error, for errors.As, but leaves its text,
// which quotes the input, out of the message.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *client) error {
		if baseURL == "" {
			c.url = DefaultBaseURL
			return nil
		}
		u, err := url.Parse(baseURL)
		if err != nil {
			return &messageError{msg: "checkly: WithBaseURL: not a valid URL", wrapped: err}
		}
		// The first two errors do not quote the value: url.Parse may read
		// userinfo such as "user:password@host" as the scheme or as the opaque
		// part of "http:user:password@host", and Redacted masks neither.
		switch {
		case u.Scheme != "http" && u.Scheme != "https":
			return errors.New("checkly: WithBaseURL: not an http or https URL")
		case u.Opaque != "":
			return errors.New("checkly: WithBaseURL: not an absolute URL; write it as http://host or https://host")
		case u.Hostname() == "":
			return fmt.Errorf("checkly: WithBaseURL: %q has no host", u.Redacted())
		case u.RawQuery != "" || u.ForceQuery:
			return fmt.Errorf("checkly: WithBaseURL: %q has a query", u.Redacted())
		case strings.Contains(baseURL, "#"):
			return fmt.Errorf("checkly: WithBaseURL: %q has a fragment", u.Redacted())
		}
		c.url = strings.TrimRight(baseURL, "/")
		return nil
	}
}

// WithHTTPClient sets the HTTP client that sends requests, for example to set
// a timeout or a proxy. Nil restores http.DefaultClient.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *client) error {
		if httpClient == nil {
			c.httpClient = http.DefaultClient
			return nil
		}
		c.httpClient = httpClient
		return nil
	}
}

// WithSource sets the x-checkly-source header, which identifies the calling
// tool to the API. An empty string restores the default, "go-sdk".
func WithSource(source string) ClientOption {
	return func(c *client) error {
		if source == "" {
			c.source = defaultSource
			return nil
		}
		c.source = source
		return nil
	}
}

// WithAuth sets the credentials that authenticate requests: the API key and
// the ID of the account it acts on (a UUID), sent as the x-checkly-account
// header. Both are required; for a client without auth, pass no auth option.
// A later WithAuth or WithAuthFromEnv replaces the credentials completely.
//
// NewClient fails with "checkly: WithAuth: the API key is empty" (matching
// ErrMissingAPIKey), "checkly: WithAuth: the account ID is empty" (matching
// ErrMissingAccountID), or, for an account ID that is not a UUID, such as an
// API key passed in its place, "checkly: WithAuth: the account ID is not a
// UUID; check that the API key and account ID are not swapped" (matching
// ErrInvalidAccountID). The errors never include the values, because a
// misplaced value may be the API key.
func WithAuth(apiKey, accountID string) ClientOption {
	return func(c *client) error {
		switch {
		case apiKey == "":
			return &messageError{msg: "checkly: WithAuth: the API key is empty", wrapped: ErrMissingAPIKey}
		case accountID == "":
			return &messageError{msg: "checkly: WithAuth: the account ID is empty", wrapped: ErrMissingAccountID}
		case !isUUID(accountID):
			return &messageError{
				msg:     "checkly: WithAuth: the account ID is not a UUID; check that the API key and account ID are not swapped",
				wrapped: ErrInvalidAccountID,
			}
		}
		c.apiKey = apiKey
		c.accountID = accountID
		return nil
	}
}

// WithAuthFromEnv sets the credentials from the CHECKLY_API_KEY and
// CHECKLY_ACCOUNT_ID environment variables, which the Checkly CLI and the
// Terraform provider use too, reading them when NewClient applies the option.
// Both are required: if one is unset or empty, NewClient fails with an error
// naming it, "checkly: WithAuthFromEnv: CHECKLY_API_KEY is not set" (matching
// ErrMissingAPIKey) or "checkly: WithAuthFromEnv: CHECKLY_ACCOUNT_ID is not
// set" (matching ErrMissingAccountID). The account ID must be a UUID, or
// NewClient fails with "checkly: WithAuthFromEnv: CHECKLY_ACCOUNT_ID is not a
// UUID" (matching ErrInvalidAccountID).
//
// NewClient never reads the environment without this option.
func WithAuthFromEnv() ClientOption {
	return func(c *client) error {
		apiKey := os.Getenv(envAPIKey)
		if apiKey == "" {
			return &messageError{
				msg:     fmt.Sprintf("checkly: WithAuthFromEnv: %s is not set", envAPIKey),
				wrapped: ErrMissingAPIKey,
			}
		}
		accountID := os.Getenv(envAccountID)
		if accountID == "" {
			return &messageError{
				msg:     fmt.Sprintf("checkly: WithAuthFromEnv: %s is not set", envAccountID),
				wrapped: ErrMissingAccountID,
			}
		}
		if !isUUID(accountID) {
			return &messageError{
				msg:     fmt.Sprintf("checkly: WithAuthFromEnv: %s is not a UUID", envAccountID),
				wrapped: ErrInvalidAccountID,
			}
		}
		c.apiKey = apiKey
		c.accountID = accountID
		return nil
	}
}
