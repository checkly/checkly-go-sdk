package checkly

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

// Connection retry policy. A request that gets no response, for example
// because its connection was reset by the API's edge, is re-sent when that
// cannot apply a change twice, unless a retry would only repeat the failure. HTTP/2 multiplexes the requests of a
// client over one connection, so a single reset fails every request in flight
// on it, and Go's HTTP/2 transport only re-sends a request that it knows the
// server did not process.
const (
	// connectionMaxRetries is how many times a request that got no response
	// is re-sent before the error is returned to the caller.
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

// retryableFailure reports whether err is a request that got no response for
// a reason that re-sending it may fix. That is any failure to send the request
// or read its response, such as a reset or aborted connection, a response cut
// short, or an HTTP/2 stream or connection error, except failures that a retry
// would only repeat or delay: a timeout, an ended context, a DNS failure, a
// TLS or certificate failure, and a failure to connect, such as a refused
// connection. Every exception is recognized by its type, so failures that
// Go's transport reports as plain errors, such as the HTTP/2 connection the
// transport closes when another request on it breaks it, are retried too.
// Errors the SDK finds before sending the request, such as a missing API key,
// are never retried. Failures that have no type to recognize are retried like
// any other, although a retry repeats them: those net/http finds without
// sending anything, such as an unsupported URL scheme, a proxy that refuses
// the CONNECT tunnel, and errors from a custom transport or redirect policy
// passed to WithHTTPClient.
func retryableFailure(err error) bool {
	var transport *transportError
	var read *responseReadError
	if !errors.As(err, &transport) && !errors.As(err, &read) {
		return false
	}
	var timeout interface{ Timeout() bool }
	if errors.As(err, &timeout) && timeout.Timeout() {
		return false
	}
	// url.Error.Timeout only looks at the error it wraps, and net/http wraps
	// an HTTP/1 read timeout once more, so a deadline is also checked by
	// identity.
	if errors.Is(err, os.ErrDeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return false
	}
	if tlsFailure(err) {
		return false
	}
	// net/http reports a failure to connect to the API, or to the proxy, as
	// a *net.OpError with Op "dial" or "proxyconnect". A connection that
	// broke after it was established has another Op, such as "read".
	var opErr *net.OpError
	if errors.As(err, &opErr) && (opErr.Op == "dial" || opErr.Op == "proxyconnect") {
		return false
	}
	return true
}

// tlsFailure reports whether err is a failed TLS handshake or certificate
// check, or a TLS connection to a server that does not speak TLS. Over TCP,
// crypto/tls reports an alert, such as a server that requires a client
// certificate or a TLS version the client does not offer, as a *net.OpError
// with Op "remote error" when the server sent it and "local error" when the
// client did; it uses those Ops for nothing else. An alert on an established
// connection, such as a corrupted record, is rare and is not retried either.
// Handshake failures that crypto/tls reports as plain errors are not
// recognized.
func tlsFailure(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) && (opErr.Op == "remote error" || opErr.Op == "local error") {
		return true
	}
	// http.Client reports an https request to a server that answers in
	// plain HTTP with ErrSchemeMismatch instead of the RecordHeaderError.
	if errors.Is(err, http.ErrSchemeMismatch) {
		return true
	}
	var (
		verification *tls.CertificateVerificationError
		recordHeader tls.RecordHeaderError
		unknownCA    x509.UnknownAuthorityError
		hostname     x509.HostnameError
		invalid      x509.CertificateInvalidError
	)
	return errors.As(err, &verification) ||
		errors.As(err, &recordHeader) ||
		errors.As(err, &unknownCA) ||
		errors.As(err, &hostname) ||
		errors.As(err, &invalid)
}

// connectionRetry returns how long to wait before re-sending a request with
// method to path that failed with err after retries earlier connection
// retries, and false when it must not be re-sent: it is not safe to re-send,
// err is not a retryable failure, the context has ended, the retries are
// spent, or the wait would leave too little of the context's deadline for the
// request. The waits back off exponentially with jitter, so that the requests
// that shared a broken connection don't all retry at once.
func connectionRetry(ctx context.Context, method, path string, err error, retries int) (time.Duration, bool) {
	if !safeToResend(method, path) || !retryableFailure(err) || ctx.Err() != nil || retries >= connectionMaxRetries {
		return 0, false
	}
	wait := equalJitter(connectionBaseBackoff << retries)
	if !leavesTime(ctx, time.Now(), wait+connectionRequestMargin) {
		return 0, false
	}
	return wait, true
}

// withRetries adds to err how many times the request was re-sent, after a 429
// or a failure that got no response, if it was.
func withRetries(err error, retries int) error {
	if retries == 0 {
		return err
	}
	return fmt.Errorf("%w; retried %d %s", err, retries, plural(retries, "time", "times"))
}
