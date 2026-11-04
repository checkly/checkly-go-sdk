package checkly_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	checkly "github.com/checkly/checkly-go-sdk"
)

// TestCodeBundleErrors checks the errors of the code bundle methods. The
// upload is sent outside the shared retry loop; the peek goes through it.
func TestCodeBundleErrors(t *testing.T) {
	t.Parallel()
	peek := func(c checkly.Client) error {
		_, err := c.PeekCodeBundle(context.Background(), "bundle-key")
		return err
	}
	upload := func(c checkly.Client) error {
		data := "bundle"
		_, err := c.UploadCodeBundle(context.Background(), strings.NewReader(data), int64(len(data)), checkly.UploadCodeBundleOptions{})
		return err
	}
	tests := []struct {
		name     string
		call     func(checkly.Client) error
		reply    errorReply
		want     string
		notFound bool
		is       []error
	}{
		{
			name:     "peek of a missing bundle",
			call:     peek,
			reply:    errorReply{status: 404, body: `{"statusCode":404,"error":"Not Found","message":"Not Found"}`},
			want:     "code bundle not found: POST /next/checkly-storage/peek-code-bundle: 404 Not Found",
			notFound: true,
			is:       []error{checkly.ErrNotFound},
		},
		{
			name:     "peek of another account's key",
			call:     peek,
			reply:    errorReply{status: 403, body: `{"statusCode":403,"error":"Forbidden","message":"Invalid Checkly Storage key accounts/other/bundle.tgz"}`},
			want:     "code bundle not found: POST /next/checkly-storage/peek-code-bundle: 403 Forbidden: Invalid Checkly Storage key accounts/other/bundle.tgz",
			notFound: true,
			is:       []error{checkly.ErrForbidden},
		},
		{
			// The backend's message for an empty key ends in a space,
			// which the client trims.
			name:     "peek of an empty key",
			call:     peek,
			reply:    errorReply{status: 403, body: `{"statusCode":403,"error":"Forbidden","message":"Invalid Checkly Storage key "}`},
			want:     "code bundle not found: POST /next/checkly-storage/peek-code-bundle: 403 Forbidden: Invalid Checkly Storage key",
			notFound: true,
			is:       []error{checkly.ErrForbidden},
		},
		{
			// A plan without the entitlement is refused with a plain 403,
			// which must not pass for a missing bundle.
			name:  "peek without the entitlement",
			call:  peek,
			reply: errorReply{status: 403, body: `{"statusCode":403,"error":"Forbidden","message":"Playwright Native Checks are not part of your plan. Upgrade or increase your limits to add more checks. Get in touch with support for any further questions."}`},
			want:  "POST /next/checkly-storage/peek-code-bundle: 403 Forbidden: Playwright Native Checks are not part of your plan. Upgrade or increase your limits to add more checks. Get in touch with support for any further questions.",
			is:    []error{checkly.ErrForbidden},
		},
		{
			name:  "peek with a bare 403",
			call:  peek,
			reply: errorReply{status: 403, body: `{"statusCode":403,"error":"Forbidden","message":"Forbidden"}`},
			want:  "POST /next/checkly-storage/peek-code-bundle: 403 Forbidden",
			is:    []error{checkly.ErrForbidden},
		},
		{
			name: "peek refused for the account",
			call: peek,
			reply: errorReply{
				status: 403,
				header: http.Header{"X-Checkly-Refused-By": {"permission"}},
				body:   `{"statusCode":403,"error":"Forbidden","message":"The selected account is not accessible with the provided credentials.","errorCode":"ERR_ACCOUNT_NOT_ACCESSIBLE"}`,
			},
			want: "POST /next/checkly-storage/peek-code-bundle: 403 Forbidden [ERR_ACCOUNT_NOT_ACCESSIBLE]: The selected account is not accessible with the provided credentials. Check the account ID \"" + testAccountID + "\".",
			is:   []error{checkly.ErrAccountNotAccessible},
		},
		{
			name: "peek refused for the role",
			call: peek,
			reply: errorReply{
				status: 403,
				header: http.Header{"X-Checkly-Refused-By": {"permission"}},
				body:   `{"statusCode":403,"error":"Forbidden","message":"Forbidden"}`,
			},
			want: "POST /next/checkly-storage/peek-code-bundle: 403 Forbidden. The API key's role does not permit this action.",
			is:   []error{checkly.ErrForbidden},
		},
		{
			name:  "upload too large",
			call:  upload,
			reply: errorReply{status: 413, header: http.Header{"Content-Type": {"text/plain"}}, body: "Payload content length greater than maximum allowed"},
			want:  "POST /next/checkly-storage/upload-code-bundle: 413 Request Entity Too Large: Payload content length greater than maximum allowed",
		},
		{
			// The code bundle requests are not retried.
			name:  "upload rate limited",
			call:  upload,
			reply: errorReply{status: 429, header: http.Header{"Retry-After": {"3"}}, body: `{"statusCode":429,"error":"Too Many Requests","message":"Rate limit exceeded"}`},
			want:  "POST /next/checkly-storage/upload-code-bundle: 429 Too Many Requests: Rate limit exceeded (retry after 3s)",
			is:    []error{checkly.ErrRateLimited},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.call(newErrorServer(t, tt.reply))
			if err == nil || err.Error() != tt.want {
				t.Fatalf("want error %q, got %v", tt.want, err)
			}
			var apiErr *checkly.APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != tt.reply.status {
				t.Errorf("want *APIError with status %d, got %#v", tt.reply.status, err)
			}
			if got := errors.Is(err, checkly.ErrCodeBundleNotFound); got != tt.notFound {
				t.Errorf("want errors.Is(err, ErrCodeBundleNotFound) %v, got %v", tt.notFound, got)
			}
			for _, target := range tt.is {
				if !errors.Is(err, target) {
					t.Errorf("want errors.Is(err, %v)", target)
				}
			}
		})
	}

	t.Run("decode error", func(t *testing.T) {
		t.Parallel()
		err := peek(newErrorServer(t, errorReply{status: 200, body: `{"key":`}))
		want := "POST /next/checkly-storage/peek-code-bundle: decoding 200 response (7 bytes): unexpected EOF"
		if err == nil || err.Error() != want {
			t.Errorf("want %q, got %v", want, err)
		}
	})
	t.Run("closed server", func(t *testing.T) {
		t.Parallel()
		srv := httptest.NewServer(http.NotFoundHandler())
		host := srv.Listener.Addr().String()
		srv.Close()
		err := upload(newTestClient(t, checkly.WithBaseURL(srv.URL), checkly.WithAuth("test-key", testAccountID)))
		want := "POST /next/checkly-storage/upload-code-bundle: request to " + host + " failed: "
		if err == nil || !strings.HasPrefix(err.Error(), want) {
			t.Fatalf("want error starting with %q, got %v", want, err)
		}
		var urlErr *url.Error
		if !errors.As(err, &urlErr) {
			t.Errorf("want a *url.Error in the chain, got %v", err)
		}
	})
}
