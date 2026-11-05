package checkly

import (
	"context"
	"fmt"
	"net/http"
)

// PrivateLocationKey represents the keys that the private location has.
type PrivateLocationKey struct {
	ID        string `json:"id"`
	MaskedKey string `json:"maskedKey"`
	RawKey    string `json:"rawKey"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// PrivateLocation defines a type for a private location.
type PrivateLocation struct {
	ID         string               `json:"id"`
	Name       string               `json:"name"`
	SlugName   string               `json:"slugName"`
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

// CreatePrivateLocation creates a new private location with the specified details.
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

// GetPrivateLocation takes the ID of an existing location, and returns the
// corresponding one.
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

// UpdatePrivateLocation takes the ID of an existing private location and updates it
// to match the new one.
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

// DeletePrivateLocation deletes the private location with the specified ID.
func (c *client) DeletePrivateLocation(
	ctx context.Context,
	ID string,
) error {
	_, err := c.call(ctx, http.MethodDelete, fmt.Sprintf("/v1/private-locations/%s", ID), nil, http.StatusNoContent)
	return err
}
