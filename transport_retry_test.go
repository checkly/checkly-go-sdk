package checkly

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// scriptedTransport answers each request with the next entry of replies,
// repeating the last one once exhausted: an error, or a response with body
// "{}" when the entry is nil. The response has the status at the same index
// of status, or 200.
type scriptedTransport struct {
	mu      sync.Mutex
	replies []error
	status  []int
	n       int
}

func (s *scriptedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.n++
	n := s.n
	reply := s.replies[len(s.replies)-1]
	if n <= len(s.replies) {
		reply = s.replies[n-1]
	}
	if reply != nil {
		return nil, reply
	}
	status := http.StatusOK
	if n <= len(s.status) && s.status[n-1] != 0 {
		status = s.status[n-1]
	}
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader("{}")),
		Request:    req,
	}, nil
}

func (s *scriptedTransport) requests() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.n
}

// newScriptedClient returns a client that sends its requests through
// transport and records its waits instead of sleeping.
func newScriptedClient(t *testing.T, transport http.RoundTripper) (*client, *[]time.Duration) {
	t.Helper()
	return newTestClient(t, DefaultBaseURL, WithHTTPClient(&http.Client{Transport: transport}))
}

var connectionReset = &net.OpError{Op: "read", Net: "tcp", Err: os.NewSyscallError("read", syscall.ECONNRESET)}

func TestConnectionRetryRetriesIdempotentMethods(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			transport := &scriptedTransport{replies: []error{connectionReset, connectionReset, nil}}
			c, waits := newScriptedClient(t, transport)

			res, err := c.send(context.Background(), true, method, "/v1/things", []byte("{}"))
			if err != nil {
				t.Fatal(err)
			}
			if res.status != http.StatusOK {
				t.Errorf("want 200, got %d", res.status)
			}
			if res.retries != 2 {
				t.Errorf("want 2 retries, got %d", res.retries)
			}
			if got := transport.requests(); got != 3 {
				t.Errorf("want 3 requests, got %d", got)
			}
			if len(*waits) != 2 {
				t.Fatalf("want 2 waits, got %v", *waits)
			}
			for i, w := range *waits {
				step := connectionBaseBackoff << i
				if w < step/2 || w > step {
					t.Errorf("wait %d: want within [%v, %v], got %v", i, step/2, step, w)
				}
			}
		})
	}
}

func TestConnectionRetryForPeekCodeBundle(t *testing.T) {
	transport := &scriptedTransport{replies: []error{connectionReset, nil}}
	c, _ := newScriptedClient(t, transport)

	if _, err := c.send(context.Background(), true, http.MethodPost, peekCodeBundlePath, []byte("{}")); err != nil {
		t.Fatal(err)
	}
	if got := transport.requests(); got != 2 {
		t.Errorf("want 2 requests, got %d", got)
	}
}

func TestConnectionRetryCountInAPIError(t *testing.T) {
	// A DELETE whose first attempt succeeded but whose response was lost
	// gets a 404 when it is re-sent; the error says it was retried.
	transport := &scriptedTransport{replies: []error{connectionReset, nil}, status: []int{0, http.StatusNotFound}}
	c, _ := newScriptedClient(t, transport)

	_, err := c.call(context.Background(), http.MethodDelete, "/v1/things/1", nil, http.StatusNoContent)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want a not found error, got %v", err)
	}
	if !strings.Contains(err.Error(), "retried 1 time") {
		t.Errorf("want the retry count in %q", err)
	}
}

func TestConnectionRetryCountIncludesRateLimitRetries(t *testing.T) {
	transport := &scriptedTransport{
		replies: []error{nil, connectionReset},
		status:  []int{http.StatusTooManyRequests},
	}
	c, _ := newScriptedClient(t, transport)

	_, err := c.send(context.Background(), true, http.MethodGet, "/v1/things", nil)
	if !errors.Is(err, syscall.ECONNRESET) {
		t.Fatalf("want the connection reset, got %v", err)
	}
	if want := fmt.Sprintf("retried %d times", 1+connectionMaxRetries); !strings.HasSuffix(err.Error(), want) {
		t.Errorf("want %q at the end of %q", want, err)
	}
}

func TestConnectionRetryNotForPost(t *testing.T) {
	transport := &scriptedTransport{replies: []error{connectionReset, nil}}
	c, waits := newScriptedClient(t, transport)

	_, err := c.send(context.Background(), true, http.MethodPost, "/v1/things", []byte("{}"))
	if !errors.Is(err, syscall.ECONNRESET) {
		t.Fatalf("want the connection reset, got %v", err)
	}
	if strings.Contains(err.Error(), "retried") {
		t.Errorf("want no retry count in %q", err)
	}
	if got := transport.requests(); got != 1 {
		t.Errorf("want 1 request, got %d", got)
	}
	if len(*waits) != 0 {
		t.Errorf("want no waits, got %v", *waits)
	}
}

func TestConnectionRetryGivesUpAfterMaxRetries(t *testing.T) {
	transport := &scriptedTransport{replies: []error{connectionReset}}
	c, waits := newScriptedClient(t, transport)

	_, err := c.send(context.Background(), true, http.MethodGet, "/v1/things", nil)
	if !errors.Is(err, syscall.ECONNRESET) {
		t.Fatalf("want the connection reset, got %v", err)
	}
	var transportErr *transportError
	if !errors.As(err, &transportErr) {
		t.Errorf("want a *transportError in %v", err)
	}
	want := fmt.Sprintf("GET /v1/things: request to api.checklyhq.com failed: read tcp: read: connection reset by peer; retried %d times", connectionMaxRetries)
	if err.Error() != want {
		t.Errorf("want %q, got %q", want, err)
	}
	if got := transport.requests(); got != connectionMaxRetries+1 {
		t.Errorf("want %d requests, got %d", connectionMaxRetries+1, got)
	}
	if len(*waits) != connectionMaxRetries {
		t.Errorf("want %d waits, got %v", connectionMaxRetries, *waits)
	}
}

func TestConnectionRetryNotForOtherErrors(t *testing.T) {
	for name, err := range map[string]error{
		"connection refused": &net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connect", syscall.ECONNREFUSED)},
		"DNS failure":        &net.DNSError{Err: "no such host", Name: "api.checklyhq.com", IsNotFound: true},
		"certificate":        x509.UnknownAuthorityError{},
		"timeout":            context.DeadlineExceeded,
		"HTTP/2 timeout":     http2TimeoutError{},
	} {
		t.Run(name, func(t *testing.T) {
			transport := &scriptedTransport{replies: []error{err, nil}}
			c, waits := newScriptedClient(t, transport)

			if _, err := c.send(context.Background(), true, http.MethodGet, "/v1/things", nil); err == nil {
				t.Fatal("want an error")
			}
			if got := transport.requests(); got != 1 {
				t.Errorf("want 1 request, got %d", got)
			}
			if len(*waits) != 0 {
				t.Errorf("want no waits, got %v", *waits)
			}
		})
	}
}

func TestConnectionRetryRecognizesBrokenConnections(t *testing.T) {
	for name, err := range map[string]error{
		"reset":          connectionReset,
		"aborted":        &net.OpError{Op: "read", Net: "tcp", Err: os.NewSyscallError("read", syscall.ECONNABORTED)},
		"broken pipe":    &net.OpError{Op: "write", Net: "tcp", Err: os.NewSyscallError("write", syscall.EPIPE)},
		"EOF":            io.EOF,
		"unexpected EOF": fmt.Errorf("reading body: %w", io.ErrUnexpectedEOF),
		// net/http returns these as plain errors, wrapped in a *url.Error.
		"HTTP/2 force closed": &url.Error{Op: "Delete", URL: "https://api.checklyhq.com/v1/checks/x", Err: errors.New("http2: client connection force closed via ClientConn.Close")},
		"HTTP/2 lost":         errors.New("http2: client connection lost"),
		"HTTP/2 closed":       errors.New("http2: client conn is closed"),
		"HTTP/2 GOAWAY":       errors.New("http2: Transport received Server's graceful shutdown GOAWAY"),
	} {
		if !brokenConnection(err) {
			t.Errorf("%s: want a broken connection", name)
		}
	}
}

func TestConnectionRetryNotOnCancelledContext(t *testing.T) {
	transport := &scriptedTransport{replies: []error{connectionReset, nil}}
	c, waits := newScriptedClient(t, transport)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.send(ctx, true, http.MethodGet, "/v1/things", nil); err == nil {
		t.Fatal("want an error")
	}
	if got := transport.requests(); got != 1 {
		t.Errorf("want 1 request, got %d", got)
	}
	if len(*waits) != 0 {
		t.Errorf("want no waits, got %v", *waits)
	}
}

func TestConnectionRetryWaitAbortsOnCancel(t *testing.T) {
	transport := &scriptedTransport{replies: []error{connectionReset, nil}}
	ctx, cancel := context.WithCancel(context.Background())
	c, _ := newTestClient(t, DefaultBaseURL,
		WithHTTPClient(&http.Client{Transport: transport}),
		withSleep(func(ctx context.Context, _ time.Duration) error {
			cancel()
			return ctx.Err()
		}),
	)

	_, err := c.send(ctx, true, http.MethodGet, "/v1/things", nil)
	if !errors.Is(err, syscall.ECONNRESET) || !errors.Is(err, context.Canceled) {
		t.Fatalf("want the connection reset and the cancellation, got %v", err)
	}
	if got := transport.requests(); got != 1 {
		t.Errorf("want 1 request, got %d", got)
	}
}

func TestConnectionRetryLeavesTimeForTheRetriedRequest(t *testing.T) {
	transport := &scriptedTransport{replies: []error{connectionReset, nil}}
	c, waits := newScriptedClient(t, transport)

	// The first wait is at most connectionBaseBackoff, which fits before the
	// deadline, but would leave less than connectionRequestMargin.
	ctx, cancel := context.WithTimeout(context.Background(), connectionRequestMargin)
	defer cancel()
	if _, err := c.send(ctx, true, http.MethodGet, "/v1/things", nil); !errors.Is(err, syscall.ECONNRESET) {
		t.Fatalf("want the connection reset, got %v", err)
	}
	if got := transport.requests(); got != 1 {
		t.Errorf("want 1 request, got %d", got)
	}
	if len(*waits) != 0 {
		t.Errorf("want no waits, got %v", *waits)
	}
}

func TestConnectionRetryCountsSeparatelyFromRateLimits(t *testing.T) {
	// A broken connection between two 429s neither uses up a rate-limit
	// retry nor resets the rate-limit backoff.
	transport := &scriptedTransport{
		replies: []error{nil, connectionReset, nil, nil},
		status:  []int{http.StatusTooManyRequests, 0, http.StatusTooManyRequests, http.StatusOK},
	}
	c, waits := newScriptedClient(t, transport)

	res, err := c.send(context.Background(), true, http.MethodGet, "/v1/things", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.status != http.StatusOK || res.retries != 3 {
		t.Errorf("want 200 after 3 retries, got %d after %d", res.status, res.retries)
	}
	if len(*waits) != 3 {
		t.Fatalf("want 3 waits, got %v", *waits)
	}
	if w := (*waits)[1]; w < connectionBaseBackoff/2 || w > connectionBaseBackoff {
		t.Errorf("want the connection wait within [%v, %v], got %v", connectionBaseBackoff/2, connectionBaseBackoff, w)
	}
	if w := (*waits)[2]; w < rateLimitBaseBackoff || w > 2*rateLimitBaseBackoff {
		t.Errorf("want the second rate-limit wait within [%v, %v], got %v", rateLimitBaseBackoff, 2*rateLimitBaseBackoff, w)
	}
}

// TestConnectionRetryOnRealReset checks that a reset from a real server, as
// net/http reports it, is recognized.
func TestConnectionRetryOnRealReset(t *testing.T) {
	var calls int
	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		n := calls
		mu.Unlock()
		if n == 1 {
			resetConnection(t, w)
			return
		}
		_, _ = io.WriteString(w, "{}")
	}))
	t.Cleanup(srv.Close)
	c, waits := newTestClient(t, srv.URL)

	res, err := c.send(context.Background(), true, http.MethodGet, "/v1/things", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.status != http.StatusOK || calls != 2 || len(*waits) != 1 {
		t.Errorf("want 200 after one retry, got %d after %d requests and waits %v", res.status, calls, *waits)
	}
}

// TestConnectionRetryOnHTTP2StreamError checks that a stream reset by a
// real HTTP/2 server, as net/http reports it, is recognized.
func TestConnectionRetryOnHTTP2StreamError(t *testing.T) {
	var calls, http2Calls int
	var mu sync.Mutex
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		n := calls
		if r.ProtoMajor == 2 {
			http2Calls++
		}
		mu.Unlock()
		if n == 1 {
			// Aborting the handler resets the stream with INTERNAL_ERROR.
			panic(http.ErrAbortHandler)
		}
		_, _ = io.WriteString(w, "{}")
	}))
	srv.EnableHTTP2 = true
	srv.StartTLS()
	t.Cleanup(srv.Close)
	c, waits := newTestClient(t, srv.URL, WithHTTPClient(srv.Client()))

	res, err := c.send(context.Background(), true, http.MethodGet, "/v1/things", nil)
	if err != nil {
		t.Fatal(err)
	}
	if http2Calls != calls {
		t.Fatalf("want HTTP/2 requests only, got %d of %d", http2Calls, calls)
	}
	if res.status != http.StatusOK || calls != 2 || len(*waits) != 1 {
		t.Errorf("want 200 after one retry, got %d after %d requests and waits %v", res.status, calls, *waits)
	}
}

// TestConnectionRetryOnHTTP2ConnectionClosed checks that a request whose
// HTTP/2 connection the transport closed while it was in flight, as net/http
// reports it, is recognized. The transport closes a connection that way when
// another request on it breaks it.
func TestConnectionRetryOnHTTP2ConnectionClosed(t *testing.T) {
	var calls int
	var mu sync.Mutex
	received := make(chan struct{})
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		n := calls
		mu.Unlock()
		if n == 1 {
			close(received)
			<-r.Context().Done()
			return
		}
		_, _ = io.WriteString(w, "{}")
	}))
	srv.EnableHTTP2 = true
	srv.StartTLS()
	t.Cleanup(srv.Close)

	transport := srv.Client().Transport.(*http.Transport).Clone()
	transport.Protocols = new(http.Protocols)
	transport.Protocols.SetHTTP2(true)
	var attempts int
	roundTrip := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		conn, err := transport.NewClientConn(req.Context(), "https", srv.Listener.Addr().String())
		if err != nil {
			return nil, err
		}
		attempts++
		if attempts == 1 {
			go func() {
				<-received
				_ = conn.Close()
			}()
		}
		return conn.RoundTrip(req)
	})
	c, waits := newTestClient(t, srv.URL, WithHTTPClient(&http.Client{Transport: roundTrip}))

	res, err := c.send(context.Background(), true, http.MethodDelete, "/v1/things/x", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.status != http.StatusOK || calls != 2 || len(*waits) != 1 {
		t.Errorf("want 200 after one retry, got %d after %d requests and waits %v", res.status, calls, *waits)
	}
}

// roundTripFunc is an http.RoundTripper implemented by a function.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

// http2TimeoutError is a timeout with the message of the response header
// timeout of golang.org/x/net/http2, which brokenConnection must reject
// whatever its type.
type http2TimeoutError struct{}

func (http2TimeoutError) Error() string { return "http2: timeout awaiting response headers" }
func (http2TimeoutError) Timeout() bool { return true }

// resetConnection closes the connection of w so that the client sees a
// reset instead of a response.
func resetConnection(t *testing.T, w http.ResponseWriter) {
	t.Helper()
	conn, _, err := http.NewResponseController(w).Hijack()
	if err != nil {
		t.Errorf("hijacking the connection: %v", err)
		return
	}
	if tcp, ok := conn.(*net.TCPConn); ok {
		_ = tcp.SetLinger(0)
	}
	_ = conn.Close()
}
