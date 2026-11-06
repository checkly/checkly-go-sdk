package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ClientCertificate is a TLS client certificate that checks present to the
// hosts it matches. It is the input and result of CreateClientCertificate and
// the result of GetClientCertificate; a certificate cannot be updated.
type ClientCertificate struct {
	ID string `json:"id,omitempty"`

	// Host is the host domain that the certificate should be used for,
	// without a scheme. Wildcards are supported, e.g. "*.acme.com".
	Host string `json:"host"`

	// Path is an optional URL path prefix that limits the certificate to
	// requests under that path, e.g. "/partner/api". Matching is on whole
	// path segments: "/partner" applies to "/partner" and "/partner/orders"
	// but not to "/partnership". API checks and Multistep checks match on
	// path; gRPC, SSL and TCP monitors only use certificates without one.
	// Leave it empty to use the certificate for every path on the host.
	// Path is left out of the request when empty.
	Path string `json:"path,omitempty"`

	// Certificate is the client certificate in PEM format.
	Certificate string `json:"cert"`

	// PrivateKey is the private key for the certificate in PEM format. It is
	// only sent: reads never return it.
	PrivateKey string `json:"key"`

	// Passphrase is an optional passphrase for the private key. Like
	// PrivateKey, reads never return it.
	// Passphrase is left out of the request when empty.
	Passphrase string `json:"passphrase,omitempty"`

	// TrustedCA is an optional PEM formatted bundle of CA certificates that
	// the client should trust. The bundle may contain many CA certificates.
	// TrustedCA is left out of the request when empty.
	TrustedCA string `json:"ca,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

func (c *client) CreateClientCertificate(
	ctx context.Context,
	cs ClientCertificate,
) (*ClientCertificate, error) {
	var result ClientCertificate
	err := c.callJSON(ctx, http.MethodPost, "/v1/client-certificates", cs, &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetClientCertificate(
	ctx context.Context,
	ID string,
) (*ClientCertificate, error) {
	var result ClientCertificate
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/client-certificates/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteClientCertificate(
	ctx context.Context,
	ID string,
) error {
	_, err := c.call(ctx, http.MethodDelete, fmt.Sprintf("/v1/client-certificates/%s", ID), nil, http.StatusNoContent)
	return err
}
