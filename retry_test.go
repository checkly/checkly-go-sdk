package checkly

import (
	"context"
	"errors"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"
)

// rateLimitServer answers each request with the next entry of responses
// (repeating the last one once exhausted) and records every request body.
type rateLimitServer struct {
	*httptest.Server
	mu     sync.Mutex
	bodies []string
}

type cannedReply struct {
	status  int
	headers map[string]string
	body    string
}

func newRateLimitServer(t *testing.T, responses ...cannedReply) *rateLimitServer {
	t.Helper()
	s := &rateLimitServer{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading request body: %v", err)
		}
		s.mu.Lock()
		s.bodies = append(s.bodies, string(body))
		n := len(s.bodies)
		s.mu.Unlock()

		reply := responses[len(responses)-1]
		if n <= len(responses) {
			reply = responses[n-1]
		}
		for k, v := range reply.headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(reply.status)
		io.WriteString(w, reply.body)
	}))
	t.Cleanup(s.Close)
	return s
}

func (s *rateLimitServer) requests() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.bodies...)
}

// newTestClient returns a client whose waits are recorded instead of slept.
func newTestClient(url string) (*client, *[]time.Duration) {
	c := NewClient(url, "test-key", nil, nil).(*client)
	var waits []time.Duration
	c.sleep = func(ctx context.Context, d time.Duration) error {
		waits = append(waits, d)
		return ctx.Err()
	}
	return c, &waits
}

var tooMany = cannedReply{status: http.StatusTooManyRequests, body: `{"message":"Rate limit exceeded"}`}

func TestRateLimitRetriesThenSucceeds(t *testing.T) {
	srv := newRateLimitServer(t, tooMany, tooMany, cannedReply{status: 200, body: `{"ok":true}`})
	c, waits := newTestClient(srv.URL)

	res, err := c.send(context.Background(), http.MethodGet, "/v1/things", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.status != 200 || string(res.body) != `{"ok":true}` {
		t.Fatalf("want 200 {\"ok\":true}, got %d %q", res.status, res.body)
	}
	if res.retries != 2 {
		t.Errorf("want 2 retries, got %d", res.retries)
	}
	if got := len(srv.requests()); got != 3 {
		t.Errorf("want 3 requests, got %d", got)
	}
	if len(*waits) != 2 {
		t.Fatalf("want 2 waits, got %v", *waits)
	}
	// Exponential backoff with equal jitter: step/2 <= wait <= step.
	for i, w := range *waits {
		step := rateLimitBaseBackoff << i
		if w < step/2 || w > step {
			t.Errorf("wait %d: want within [%v, %v], got %v", i, step/2, step, w)
		}
	}
}

func TestRateLimitHonoursRetryAfter(t *testing.T) {
	srv := newRateLimitServer(t,
		cannedReply{status: 429, headers: map[string]string{"Retry-After": "7", "X-RateLimit-Reset": "1"}},
		cannedReply{status: 200, body: "{}"},
	)
	c, waits := newTestClient(srv.URL)

	res, err := c.send(context.Background(), http.MethodGet, "/v1/things", nil)
	if err != nil || res.status != 200 {
		t.Fatalf("want 200, got %d, %v", res.status, err)
	}
	if len(*waits) != 1 {
		t.Fatalf("want 1 wait, got %v", *waits)
	}
	lo := 7*time.Second + rateLimitHintPadding
	if w := (*waits)[0]; w < lo || w >= lo+rateLimitHintJitter {
		t.Errorf("want wait in [%v, %v), got %v", lo, lo+rateLimitHintJitter, w)
	}
}

func TestRateLimitWaitHints(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	tests := []struct {
		name     string
		headers  map[string]string
		min, max time.Duration
	}{
		{"retry-after seconds", map[string]string{"Retry-After": "3"}, 3500 * time.Millisecond, 4500 * time.Millisecond},
		{"retry-after http date", map[string]string{"Retry-After": now.Add(5 * time.Second).Format(http.TimeFormat)}, 5500 * time.Millisecond, 6500 * time.Millisecond},
		{"retry-after in the past", map[string]string{"Retry-After": now.Add(-time.Hour).Format(http.TimeFormat)}, 500 * time.Millisecond, 1500 * time.Millisecond},
		{"reset header", map[string]string{"X-RateLimit-Reset": "1767323055"}, 10500 * time.Millisecond, 11500 * time.Millisecond},
		{"unparseable retry-after falls back to reset", map[string]string{"Retry-After": "soon", "X-RateLimit-Reset": "1767323047"}, 2500 * time.Millisecond, 3500 * time.Millisecond},
		{"hint capped", map[string]string{"Retry-After": "86400"}, rateLimitMaxWait + rateLimitHintPadding, rateLimitMaxWait + rateLimitHintPadding + rateLimitHintJitter},
		{"huge retry-after seconds capped", map[string]string{"Retry-After": "9223372036854775807"}, rateLimitMaxWait + rateLimitHintPadding, rateLimitMaxWait + rateLimitHintPadding + rateLimitHintJitter},
		{"far-future retry-after date capped", map[string]string{"Retry-After": "Fri, 31 Dec 9999 23:59:59 GMT"}, rateLimitMaxWait + rateLimitHintPadding, rateLimitMaxWait + rateLimitHintPadding + rateLimitHintJitter},
		{"far-future reset capped", map[string]string{"X-RateLimit-Reset": "99999999999999"}, rateLimitMaxWait + rateLimitHintPadding, rateLimitMaxWait + rateLimitHintPadding + rateLimitHintJitter},
		{"no hint", nil, rateLimitBaseBackoff / 2, rateLimitBaseBackoff},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := http.Header{}
			for k, v := range tt.headers {
				h.Set(k, v)
			}
			for i := 0; i < 20; i++ {
				got := rateLimitWait(h, 0, now)
				if got < tt.min || got > tt.max {
					t.Fatalf("want within [%v, %v], got %v", tt.min, tt.max, got)
				}
			}
		})
	}

	// Backoff without hints stays positive and never exceeds the cap.
	for retry := 0; retry < rateLimitMaxRetries; retry++ {
		if got := rateLimitWait(http.Header{}, retry, now); got > rateLimitMaxWait || got <= 0 {
			t.Fatalf("retry %d: want within (0, %v], got %v", retry, rateLimitMaxWait, got)
		}
	}
}

// TestRateLimitHintUncapped checks that the server's hint is returned as it
// is, in each header form: only rateLimitWait caps it.
func TestRateLimitHintUncapped(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	tests := []struct {
		name    string
		headers map[string]string
		want    time.Duration
	}{
		{"retry-after seconds", map[string]string{"Retry-After": "86400"}, 24 * time.Hour},
		{"retry-after http date", map[string]string{"Retry-After": now.Add(2 * time.Hour).Format(http.TimeFormat)}, 2 * time.Hour},
		{"reset header", map[string]string{"X-RateLimit-Reset": strconv.FormatInt(now.Add(3*time.Hour).Unix(), 10)}, 3 * time.Hour},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := http.Header{}
			for k, v := range tt.headers {
				h.Set(k, v)
			}
			got, ok := rateLimitHint(h, now)
			if !ok || got != tt.want {
				t.Errorf("want %v, got %v (ok %v)", tt.want, got, ok)
			}
		})
	}
}

// TestCeilSeconds checks the rounding of the wait that a rate-limit error
// shows: up to whole seconds, so a short wait never shows as "0s", except
// where rounding up would overflow.
func TestCeilSeconds(t *testing.T) {
	tests := []struct {
		in, want time.Duration
	}{
		{time.Nanosecond, time.Second},
		{300 * time.Millisecond, time.Second},
		{time.Second, time.Second},
		{1500 * time.Millisecond, 2 * time.Second},
		{12 * time.Second, 12 * time.Second},
		{math.MaxInt64, math.MaxInt64},
	}
	for _, tt := range tests {
		if got := ceilSeconds(tt.in); got != tt.want {
			t.Errorf("ceilSeconds(%v): want %v, got %v", tt.in, tt.want, got)
		}
	}
}

func TestRateLimitGivesUpAfterMaxRetries(t *testing.T) {
	srv := newRateLimitServer(t, tooMany)
	c, waits := newTestClient(srv.URL)

	res, err := c.send(context.Background(), http.MethodGet, "/v1/things", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.status != http.StatusTooManyRequests || string(res.body) != tooMany.body {
		t.Errorf("want final 429 with its body, got %d %q", res.status, res.body)
	}
	if res.retries != rateLimitMaxRetries {
		t.Errorf("want %d retries, got %d", rateLimitMaxRetries, res.retries)
	}
	if got := len(srv.requests()); got != rateLimitMaxRetries+1 {
		t.Errorf("want %d requests, got %d", rateLimitMaxRetries+1, got)
	}
	if len(*waits) != rateLimitMaxRetries {
		t.Errorf("want %d waits, got %v", rateLimitMaxRetries, *waits)
	}
}

func TestRateLimitDoesNotWaitPastDeadline(t *testing.T) {
	srv := newRateLimitServer(t,
		cannedReply{status: 429, headers: map[string]string{"Retry-After": "30"}, body: "limited"},
		cannedReply{status: 200},
	)
	c, waits := newTestClient(srv.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := c.send(ctx, http.MethodGet, "/v1/things", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.status != 429 || string(res.body) != "limited" {
		t.Errorf("want immediate 429, got %d %q", res.status, res.body)
	}
	if len(*waits) != 0 {
		t.Errorf("want no waits, got %v", *waits)
	}
	if got := len(srv.requests()); got != 1 {
		t.Errorf("want 1 request, got %d", got)
	}
}

func TestRateLimitLeavesTimeForTheRetriedRequest(t *testing.T) {
	srv := newRateLimitServer(t,
		cannedReply{status: 429, headers: map[string]string{"Retry-After": "10"}, body: "limited"},
		cannedReply{status: 200},
	)
	c, waits := newTestClient(srv.URL)

	// The wait (10s plus padding and jitter) fits before the deadline, but
	// would leave less than rateLimitRequestMargin for the re-sent request.
	ctx, cancel := context.WithTimeout(context.Background(), 14*time.Second)
	defer cancel()
	res, err := c.send(ctx, http.MethodGet, "/v1/things", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.status != 429 || string(res.body) != "limited" {
		t.Errorf("want immediate 429, got %d %q", res.status, res.body)
	}
	if len(*waits) != 0 {
		t.Errorf("want no waits, got %v", *waits)
	}
}

func TestRateLimitWaitAbortsOnCancel(t *testing.T) {
	srv := newRateLimitServer(t,
		cannedReply{status: 429, headers: map[string]string{"Retry-After": "30"}, body: "limited"},
		cannedReply{status: 200},
	)
	c := NewClient(srv.URL, "test-key", nil, nil).(*client)

	// Cancel once the client starts waiting, then let the real sleep run:
	// the 30s wait must end at once.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.sleep = func(ctx context.Context, d time.Duration) error {
		cancel()
		return sleepContext(ctx, d)
	}
	start := time.Now()
	_, err := c.send(ctx, http.MethodGet, "/v1/things", nil)
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("cancel did not abort the wait, took %v", elapsed)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("want context.Canceled, got %v", err)
	}
	// The error also describes the 429 the client gave up retrying.
	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("want ErrRateLimited, got %v", err)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 429 || apiErr.Message != "limited" || apiErr.RetryAfter != 30*time.Second {
		t.Errorf("want the 429 as an *APIError, got %#v", apiErr)
	}
	want := "GET /v1/things: 429 Too Many Requests: limited (retry after 30s); gave up waiting to retry: context canceled"
	if err.Error() != want {
		t.Errorf("want error %q, got %q", want, err)
	}
	if got := len(srv.requests()); got != 1 {
		t.Errorf("want 1 request, got %d", got)
	}
}

func TestRateLimitResendsPostBody(t *testing.T) {
	srv := newRateLimitServer(t, tooMany, tooMany, cannedReply{status: 201, body: "{}"})
	c, _ := newTestClient(srv.URL)

	payload := `{"name":"status page","components":[1,2,3]}`
	res, err := c.send(context.Background(), http.MethodPost, "/v1/things", []byte(payload))
	if err != nil || res.status != 201 {
		t.Fatalf("want 201, got %d, %v", res.status, err)
	}
	got := srv.requests()
	if len(got) != 3 {
		t.Fatalf("want 3 requests, got %d", len(got))
	}
	for i, b := range got {
		if b != payload {
			t.Errorf("request %d: want body %q, got %q", i, payload, b)
		}
	}
}
