package checkly

import (
	"context"
	"net/http"
	"time"
)

// SkipRetryWaits makes c re-send a rate-limited request without waiting, so
// that tests in package checkly_test can exercise the retry loop quickly. Even
// "Retry-After: 0" otherwise waits for the hint padding and jitter.
func SkipRetryWaits(c Client) {
	c.(*client).sleep = func(ctx context.Context, _ time.Duration) error {
		return ctx.Err()
	}
}

// Settings is the configuration NewClient gave a client, for tests in package
// checkly_test.
type Settings struct {
	URL        string
	APIKey     string
	AccountID  string
	Source     string
	HTTPClient *http.Client
}

// ClientSettings returns the configuration of c.
func ClientSettings(c Client) Settings {
	cl := c.(*client)
	return Settings{
		URL:        cl.url,
		APIKey:     cl.apiKey,
		AccountID:  cl.accountID,
		Source:     cl.source,
		HTTPClient: cl.httpClient,
	}
}

// RecordApplied returns an option that sets *applied when NewClient applies
// it, so that tests can tell whether construction reached it.
func RecordApplied(applied *bool) ClientOption {
	return func(*client) error {
		*applied = true
		return nil
	}
}
