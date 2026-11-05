package checkly

import (
	"context"
	"fmt"
	"net/http"
)

// PrivateLocationKey is an API key of a private location, in
// PrivateLocation.Keys.
type PrivateLocationKey struct {
	ID        string `json:"id"`
	MaskedKey string `json:"maskedKey"`

	// RawKey is the full key. Only the response of CreatePrivateLocation
	// includes it, for the key that the new location gets.
	RawKey    string `json:"rawKey"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// PrivateLocation is a location that runs checks on the account's own
// infrastructure, through Checkly agents. It is the input and result of
// CreatePrivateLocation and UpdatePrivateLocation.
type PrivateLocation struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// SlugName identifies the location in the PrivateLocations fields of
	// checks and groups. It is required on create and cannot be changed:
	// the API ignores it on update.
	SlugName string `json:"slugName"`

	// Icon defaults to "location" on create.
	// Icon is left out of the request when empty.
	Icon       string               `json:"icon,omitempty"`
	Keys       []PrivateLocationKey `json:"keys,omitempty"`
	LastSeen   string               `json:"lastSeen,omitempty"`
	AgentCount int                  `json:"agentCount,omitempty"`
	CreatedAt  string               `json:"created_at"`
	UpdatedAt  string               `json:"updated_at"`

	// ProxyURL is the proxy that API checks running in the location send
	// their HTTP requests through, such as https://user:password@host:3128.
	// It can embed credentials, so treat it as sensitive. Nil and an empty
	// string are sent as null, which removes the proxy on update.
	ProxyURL *string `json:"proxyUrl"`
}

// createPrivateLocationPayload returns the request body of
// CreatePrivateLocation and UpdatePrivateLocation. The API rejects an empty
// proxyUrl, so an empty ProxyURL is sent as null instead.
func createPrivateLocationPayload(pl PrivateLocation) PrivateLocation {
	if pl.ProxyURL != nil && *pl.ProxyURL == "" {
		pl.ProxyURL = nil
	}
	return pl
}

func (c *client) CreatePrivateLocation(
	ctx context.Context,
	pl PrivateLocation,
) (*PrivateLocation, error) {
	var result PrivateLocation
	err := c.callJSON(ctx, http.MethodPost, "/v1/private-locations", createPrivateLocationPayload(pl), &result, http.StatusOK, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetPrivateLocation(
	ctx context.Context,
	ID string,
) (*PrivateLocation, error) {
	result := PrivateLocation{}
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/private-locations/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdatePrivateLocation(
	ctx context.Context,
	ID string,
	pl PrivateLocation,
) (*PrivateLocation, error) {
	var result PrivateLocation
	err := c.callJSON(ctx, http.MethodPut, fmt.Sprintf("/v1/private-locations/%s", ID), createPrivateLocationPayload(pl), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeletePrivateLocation(
	ctx context.Context,
	ID string,
) error {
	_, err := c.call(ctx, http.MethodDelete, fmt.Sprintf("/v1/private-locations/%s", ID), nil, http.StatusNoContent)
	return err
}
