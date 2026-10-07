package checkly

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Rate-limit retry policy. The Checkly API answers HTTP 429 when a route's
// request budget is spent, typically with a per-minute window. A 429 response
// carries Retry-After (whole seconds until the budget refills, rounded to the
// nearest second) and X-RateLimit-Reset (Unix time, in seconds, at which it
// refills).
const (
	// rateLimitMaxRetries is how many times a request is re-sent after a 429
	// before the 429 is returned to the caller.
	rateLimitMaxRetries = 4
	// rateLimitRequestMargin is the time that must be left before the
	// context deadline after a wait, so the re-sent request has a chance to
	// complete. Otherwise the caller gets the 429 instead of a request that
	// fails on the deadline after a long wait.
	rateLimitRequestMargin = 5 * time.Second
	// rateLimitMaxWait caps a server hint and a backoff step. Rate-limit
	// windows are at most a minute long, so a longer hint is clamped rather
	// than obeyed. Padding and jitter are added on top of a clamped hint.
	rateLimitMaxWait = 60 * time.Second
	// rateLimitBaseBackoff is the first exponential backoff step, used when
	// the response carries no usable rate-limit header.
	rateLimitBaseBackoff = time.Second
	// A server-provided wait is padded by rateLimitHintPadding plus up to
	// rateLimitHintJitter. Retry-After may be rounded down by up to half a
	// second, and many concurrent clients told the same wait would otherwise
	// retry in lockstep.
	rateLimitHintPadding = 500 * time.Millisecond
	rateLimitHintJitter  = time.Second
)

// sleepContext waits for d, returning ctx.Err() early if ctx is done first.
func sleepContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
	// Both cases may be ready at once; never report a finished wait on a
	// cancelled context.
	return ctx.Err()
}

// rateLimitWait returns how long to wait before re-sending a request that got
// a 429. retry is zero for the first retry. It prefers the server's hint
// (Retry-After, then X-RateLimit-Reset) and falls back to exponential backoff
// with jitter. The result never exceeds rateLimitMaxWait plus the hint padding
// and jitter.
func rateLimitWait(h http.Header, retry int, now time.Time) time.Duration {
	wait, ok := retryAfterHint(h, now)
	if !ok {
		wait, ok = resetHint(h, now)
	}
	if ok {
		// Clamp before padding so that a hint far in the future cannot
		// overflow into a negative duration.
		return min(wait, rateLimitMaxWait) + rateLimitHintPadding +
			time.Duration(rand.Int63n(int64(rateLimitHintJitter)))
	}
	// "Equal jitter": half the exponential step is fixed, the other half
	// random, so retries back off but concurrent clients spread out.
	step := min(rateLimitBaseBackoff<<retry, rateLimitMaxWait)
	return step/2 + time.Duration(rand.Int63n(int64(step/2)+1))
}

// retryAfterHint parses Retry-After as either delay-seconds or an HTTP date.
func retryAfterHint(h http.Header, now time.Time) (time.Duration, bool) {
	v := strings.TrimSpace(h.Get("Retry-After"))
	if v == "" {
		return 0, false
	}
	if secs, err := strconv.ParseInt(v, 10, 64); err == nil {
		if secs < 0 {
			return 0, false
		}
		// Clamp in seconds first: converting a huge value would overflow.
		return time.Duration(min(secs, int64(rateLimitMaxWait/time.Second))) * time.Second, true
	}
	if t, err := http.ParseTime(v); err == nil {
		return max(t.Sub(now), 0), true
	}
	return 0, false
}

// resetHint parses X-RateLimit-Reset as a Unix timestamp in seconds.
func resetHint(h http.Header, now time.Time) (time.Duration, bool) {
	v := strings.TrimSpace(h.Get("X-RateLimit-Reset"))
	if v == "" {
		return 0, false
	}
	secs, err := strconv.ParseInt(v, 10, 64)
	if err != nil || secs <= 0 {
		return 0, false
	}
	return max(time.Unix(secs, 0).Sub(now), 0), true
}
