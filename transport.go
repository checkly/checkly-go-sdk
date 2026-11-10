package checkly

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// authenticate adds the credential headers to req, an authenticated request:
// the API key and the account ID. When the client has no auth, it returns
// ErrMissingAPIKey, prefixed with method and path, and nothing is sent.
// WithAuth and WithAuthFromEnv always set both, so a client with an API key
// also has an account ID.
func (c *client) authenticate(req *http.Request, method string, path string) error {
	if c.apiKey == "" {
		return fmt.Errorf("%s %s: %w", method, path, ErrMissingAPIKey)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("x-checkly-account", c.accountID)
	return nil
}

// rawResponse is a response as send returns it, before its status is
// checked, together with the request it answers.
type rawResponse struct {
	// method and path name the request in error messages; path is as
	// errorPath returns it.
	method string
	path   string
	status int
	body   []byte
	header http.Header
	// retries is how many times the request was re-sent, after a 429 or a
	// broken connection.
	retries int
}

// errorPath returns path without its query string, as error messages show it.
func errorPath(path string) string {
	path, _, _ = strings.Cut(path, "?")
	return path
}

// send sends a request and returns the response, whatever its status.
// path is relative to the API root, such as "/v1/checks/abc?x=y" or
// "/next/checkly-storage/peek-code-bundle". A 429 response is retried as
// described in rate_limit.go; once the retries are spent, or the caller's deadline
// does not leave time for another one, the 429 is returned like any other
// response. When the context ends during a wait, the error matches both
// ErrRateLimited and the context error. An idempotent request whose
// connection broke is re-sent as described in transport_retry.go. auth tells
// whether the request carries credentials; only the public endpoints are sent
// without.
func (c *client) send(
	ctx context.Context,
	auth bool,
	method string,
	path string,
	data []byte,
) (rawResponse, error) {
	requestURL := c.url + path
	rateLimitRetries, connectionRetries := 0, 0
	for {
		res, err := c.doRequest(ctx, auth, method, requestURL, errorPath(path), data)
		res.retries = rateLimitRetries + connectionRetries
		if err != nil {
			wait, ok := connectionRetry(ctx, method, errorPath(path), err, connectionRetries)
			if !ok {
				return res, withRetries(err, res.retries)
			}
			if sleepErr := c.sleep(ctx, wait); sleepErr != nil {
				return res, fmt.Errorf("%w; gave up waiting to retry: %w", withRetries(err, res.retries), sleepErr)
			}
			connectionRetries++
			continue
		}
		if res.status != http.StatusTooManyRequests || rateLimitRetries >= rateLimitMaxRetries {
			return res, nil
		}

		// Rate limited: wait and re-send, unless the wait would leave too
		// little of the caller's deadline for the re-sent request, in which
		// case the caller sees the 429 itself.
		now := time.Now()
		wait := rateLimitWait(res.header, rateLimitRetries, now)
		if !leavesTime(ctx, now, wait+rateLimitRequestMargin) {
			return res, nil
		}
		if err := c.sleep(ctx, wait); err != nil {
			return res, fmt.Errorf("%w; gave up waiting to retry: %w", newAPIError(res, nil, c.accountID), err)
		}
		rateLimitRetries++
	}
}

// call marshals in as the request body, unless it is nil, sends the
// authenticated request to path, which is relative to the API root as send
// takes it, and returns the response. A response status not among want is
// returned as an *APIError.
func (c *client) call(
	ctx context.Context,
	method string,
	path string,
	in any,
	want ...int,
) (rawResponse, error) {
	return c.exchange(ctx, true, method, path, in, want...)
}

// exchange is call with auth, as send takes it, chosen by the caller.
func (c *client) exchange(
	ctx context.Context,
	auth bool,
	method string,
	path string,
	in any,
	want ...int,
) (rawResponse, error) {
	var data []byte
	if in != nil {
		var err error
		if data, err = json.Marshal(in); err != nil {
			return rawResponse{}, fmt.Errorf("%s %s: encoding request: %w", method, errorPath(path), err)
		}
	}
	res, err := c.send(ctx, auth, method, path, data)
	if err != nil {
		return rawResponse{}, err
	}
	for _, status := range want {
		if res.status == status {
			return res, nil
		}
	}
	return rawResponse{}, newAPIError(res, want, c.accountID)
}

// callJSON is call followed by decoding the response body into out.
func (c *client) callJSON(
	ctx context.Context,
	method string,
	path string,
	in any,
	out any,
	want ...int,
) error {
	res, err := c.call(ctx, method, path, in, want...)
	if err != nil {
		return err
	}
	return decode(res, out)
}

// callPublicJSON is callJSON without a request body for the public endpoints,
// which the API serves without authentication. It sends no credentials, so it
// works without an API key and never fails for a missing account ID.
func (c *client) callPublicJSON(
	ctx context.Context,
	method string,
	path string,
	out any,
	want ...int,
) error {
	res, err := c.exchange(ctx, false, method, path, nil, want...)
	if err != nil {
		return err
	}
	return decode(res, out)
}

// decode decodes the body of res, a successful response, into out. The error
// names the request, the status and the body's size, but never contains the
// body, which may hold secrets. Like the decoding it replaces, it reads the
// first JSON value and ignores the rest.
func decode(res rawResponse, out any) error {
	if err := json.NewDecoder(bytes.NewReader(res.body)).Decode(out); err != nil {
		return fmt.Errorf("%s %s: decoding %d response (%d bytes): %w", res.method, res.path, res.status, len(res.body), err)
	}
	return nil
}

// doRequest sends a single API request, with credentials when auth is set.
// The request is rebuilt from data on every call so that it can be re-sent
// after a rate-limited response or a broken connection. path is the request path for error
// messages, as errorPath returns it.
func (c *client) doRequest(
	ctx context.Context,
	auth bool,
	method string,
	requestURL string,
	path string,
	data []byte,
) (rawResponse, error) {
	req, err := http.NewRequestWithContext(ctx, method, requestURL, bytes.NewReader(data))
	if err != nil {
		return rawResponse{}, fmt.Errorf("%s %s: building request: %w", method, path, err)
	}

	req.Header.Set("x-checkly-source", c.source)
	if auth {
		if err := c.authenticate(req, method, path); err != nil {
			return rawResponse{}, err
		}
	}

	req.Header.Add("content-type", "application/json")

	return c.roundTrip(req, path)
}

// roundTrip sends req and reads the whole response. path is the request path
// for error messages.
func (c *client) roundTrip(req *http.Request, path string) (rawResponse, error) {
	method := req.Method
	resp, err := c.httpClient.Do(req)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			return rawResponse{}, &transportError{method: method, path: path, host: req.URL.Host, err: urlErr}
		}
		return rawResponse{}, fmt.Errorf("%s %s: request to %s failed: %w", method, path, req.URL.Host, err)
	}

	defer func() { _ = resp.Body.Close() }()

	res := rawResponse{method: method, path: path, status: resp.StatusCode, header: resp.Header}
	res.body, err = io.ReadAll(resp.Body)
	if err != nil {
		return res, fmt.Errorf("%s %s: reading %d response: %w", method, path, resp.StatusCode, err)
	}
	return res, nil
}
