package checkly

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type UploadCodeBundleOptions struct {
	ChecksumSha256 string
}

type CodeBundle struct {
	Key            string    `json:"key"`
	VersionID      string    `json:"versionID,omitempty"`
	LastModified   time.Time `json:"lastModified,omitempty"`
	ContentLength  int64     `json:"contentLength"`
	ChecksumSha256 string    `json:"checksumSha256"`
}

// The code bundle endpoints belong to the /next API. An upload streams its
// body, so it cannot be re-sent and is not retried on a 429; a peek is.
const (
	uploadCodeBundlePath = "/next/checkly-storage/upload-code-bundle"
	peekCodeBundlePath   = "/next/checkly-storage/peek-code-bundle"
)

func (c *client) UploadCodeBundle(
	ctx context.Context,
	data io.Reader,
	size int64,
	options UploadCodeBundleOptions,
) (*CodeBundle, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.url+uploadCodeBundlePath,
		io.LimitReader(data, size),
	)
	if err != nil {
		return nil, fmt.Errorf("%s %s: building request: %w", http.MethodPost, uploadCodeBundlePath, err)
	}

	req.Header.Set("x-checkly-source", c.source)
	if err := c.authenticate(req, http.MethodPost, uploadCodeBundlePath); err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/octet-stream")
	req.ContentLength = size

	if options.ChecksumSha256 != "" {
		req.Header.Add("x-bundle-checksum-sha256", options.ChecksumSha256)
	}

	res, err := c.roundTrip(req, uploadCodeBundlePath)
	if err != nil {
		return nil, err
	}
	if res.status != http.StatusOK {
		return nil, newAPIError(res, []int{http.StatusOK}, c.accountID)
	}
	var output CodeBundle
	if err := decode(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

type peekCodeBundlePayload struct {
	Key string `json:"key"`
}

// ErrCodeBundleNotFound is returned, wrapping the *APIError, when
// PeekCodeBundle finds no code bundle for the key.
var ErrCodeBundleNotFound = errors.New("code bundle not found")

func (c *client) PeekCodeBundle(
	ctx context.Context,
	key string,
) (*CodeBundle, error) {
	res, err := c.call(ctx, http.MethodPost, peekCodeBundlePath, peekCodeBundlePayload{Key: key}, http.StatusOK)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && codeBundleNotFound(apiErr) {
			return nil, fmt.Errorf("%w: %w", ErrCodeBundleNotFound, apiErr)
		}
		return nil, err
	}
	var output CodeBundle
	if err := decode(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

// invalidStorageKeyMessage starts the backend's message for a code bundle key
// that belongs to another account: "Invalid Checkly Storage key <key>". It
// has no trailing space, because an empty key leaves the message ending in
// one, which APIError.Message trims.
const invalidStorageKeyMessage = "Invalid Checkly Storage key"

// codeBundleNotFound reports whether a peek failed because there is no code
// bundle for the key: a 404, or the 403 the backend answers for a key that
// belongs to another account. Every other 403, such as a refused account
// access, a role without permission or a plan without the entitlement, is a
// failure the caller must see, not a missing bundle.
func codeBundleNotFound(e *APIError) bool {
	switch e.StatusCode {
	case http.StatusNotFound:
		return true
	case http.StatusForbidden:
		return e.Code == "" && strings.HasPrefix(e.Message, invalidStorageKeyMessage)
	}
	return false
}
