package checkly

import (
	"context"
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
