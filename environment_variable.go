package checkly

import (
	"context"
	"fmt"
	"net/http"
)

// EnvironmentVariable represents a key-value pair for setting environment
// values during check execution.
type EnvironmentVariable struct {
	Key    string `json:"key,omitempty"`
	Value  string `json:"value"`
	Locked bool   `json:"locked"`
	Secret bool   `json:"secret"`
}

// CreateEnvironmentVariable creates a new environment variable with the
// specified details.  It returns the newly-created environment variable,
// or an error.
func (c *client) CreateEnvironmentVariable(
	ctx context.Context,
	envVar EnvironmentVariable,
) (*EnvironmentVariable, error) {
	var result EnvironmentVariable
	err := c.callJSON(ctx, http.MethodPost, "/v1/variables", envVar, &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetEnvironmentVariable takes the ID of an existing environment variable, and returns the
// corresponding environment variable, or an error.
func (c *client) GetEnvironmentVariable(
	ctx context.Context,
	key string,
) (*EnvironmentVariable, error) {
	result := EnvironmentVariable{}
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/variables/%s", key), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateEnvironmentVariable takes the ID of an existing environment variable, and updates the
// corresponding environment variable to match the supplied environment variable. It returns the updated
// environment variable, or an error.
func (c *client) UpdateEnvironmentVariable(
	ctx context.Context,
	key string,
	envVar EnvironmentVariable,
) (*EnvironmentVariable, error) {
	var result EnvironmentVariable
	err := c.callJSON(ctx, http.MethodPut, fmt.Sprintf("/v1/variables/%s", key), envVar, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteEnvironmentVariable deletes the environment variable with the specified ID.
func (c *client) DeleteEnvironmentVariable(
	ctx context.Context,
	key string,
) error {
	_, err := c.call(ctx, http.MethodDelete, fmt.Sprintf("/v1/variables/%s", key), nil, http.StatusNoContent)
	return err
}
