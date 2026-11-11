package checkly

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"syscall"
	"time"
)

// Connection retry policy. A request that gets no response because the
// connection broke, such as a connection reset by the API's edge, is re-sent
// when that cannot apply a change twice. HTTP/2 multiplexes the requests of a
// client over one connection, so a single reset fails every request in flight
// on it, and Go's HTTP/2 transport only re-sends a request that it knows the
// server did not process.
const (
	// connectionMaxRetries is how many times a request is re-sent after a
	// broken connection before the error is returned to the caller.
	connectionMaxRetries = 2
	// connectionBaseBackoff is the first backoff step. A broken connection
	// is usually replaced by the next request, so the waits are short.
	connectionBaseBackoff = 250 * time.Millisecond
	// connectionRequestMargin is the time that must be left before the
	// context deadline after a wait, so the re-sent request has a chance to
	// complete. Otherwise the caller gets the connection error at once.
	connectionRequestMargin = 2 * time.Second
)

// safeToResend reports whether re-sending a request with method to path
// cannot apply a change twice. The SDK's PUT requests replace the whole
// object, and a repeated DELETE fails with a 404 at worst. A POST may create
// something even though its response was lost, so it is not re-sent, except
// to peekCodeBundlePath, which only reads.
func safeToResend(method, path string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete:
		return true
	case http.MethodPost:
		return path == peekCodeBundlePath
	}
	return false
}

// brokenConnection reports whether err is a connection that broke while a
// request was sent or its response read: a reset or aborted connection, a
// broken pipe, a response cut short, or an HTTP/2 stream or connection error,
// such as GOAWAY or a connection the transport closed because another request
// on it broke. It does not match errors that a retry would only repeat or
// delay, such as a refused connection, a DNS or TLS failure, a timeout or an
// ended context.
func brokenConnection(err error) bool {
	// A timeout would only be repeated. Some timeouts, such as the response
	// header timeout of golang.org/x/net/http2, have HTTP/2 types that the
	// type name match below would accept.
	var timeout interface{ Timeout() bool }
	if errors.As(err, &timeout) && timeout.Timeout() {
		return false
	}
	if errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ECONNABORTED) ||
		errors.Is(err, syscall.EPIPE) ||
		errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, io.EOF) {
		return true
	}
	for _, target := range platformConnectionErrors {
		if errors.Is(err, target) {
			return true
		}
	}
	// HTTP/2 errors, such as stream errors and GOAWAY, can only be
	// recognized by type name: net/http bundles its HTTP/2 implementation
	// with unexported types (http.http2StreamError), and a client configured
	// with golang.org/x/net/http2 returns that package's types
	// (http2.StreamError). Only the bundled types are exercised by tests.
	// The errors for a lost or closed connection are plain errors.New values,
	// recognized by their message instead.
	for e := err; e != nil; e = errors.Unwrap(e) {
		name := fmt.Sprintf("%T", e)
		if strings.HasPrefix(name, "http.http2") || strings.HasPrefix(name, "http2.") || strings.HasPrefix(name, "*http2.") {
			return true
		}
		if _, ok := http2ConnectionErrors[e.Error()]; ok {
			return true
		}
	}
	return false
}

// http2ConnectionErrors are the messages of the HTTP/2 transport's errors for
// a connection that broke while requests were in flight on it, in both
// net/http and golang.org/x/net/http2. HTTP/2 multiplexes the requests of a
// client over one connection, and when one request breaks it, for example
// with a reset while sending its body, the transport closes the connection
// and fails every other request on it with "client connection force closed".
var http2ConnectionErrors = map[string]struct{}{
	"http2: client connection force closed via ClientConn.Close":  {},
	"http2: client connection lost":                               {},
	"http2: client conn is closed":                                {},
	"http2: Transport received Server's graceful shutdown GOAWAY": {},
}

// connectionRetry returns how long to wait before re-sending a request with
// method to path that failed with err after retries earlier connection
// retries, and false when it must not be re-sent: it is not safe to re-send,
// err is not a broken connection, the context has ended, the retries are
// spent, or the wait would leave too little of the context's deadline for the
// request. The waits back off exponentially with jitter, so that the requests
// that shared a broken connection don't all retry at once.
func connectionRetry(ctx context.Context, method, path string, err error, retries int) (time.Duration, bool) {
	if !safeToResend(method, path) || !brokenConnection(err) || ctx.Err() != nil || retries >= connectionMaxRetries {
		return 0, false
	}
	wait := equalJitter(connectionBaseBackoff << retries)
	if !leavesTime(ctx, time.Now(), wait+connectionRequestMargin) {
		return 0, false
	}
	return wait, true
}

// withRetries adds to err how many times the request was re-sent, after a 429
// or a broken connection, if it was.
func withRetries(err error, retries int) error {
	if retries == 0 {
		return err
	}
	return fmt.Errorf("%w; retried %d %s", err, retries, plural(retries, "time", "times"))
}
