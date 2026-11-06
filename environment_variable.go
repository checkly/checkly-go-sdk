package checkly

import (
	"context"
	"fmt"
	"net/http"
)

// EnvironmentVariable is an environment variable that checks can read when
// they run. It is the input and result of CreateEnvironmentVariable and
// UpdateEnvironmentVariable for account-level variables, and the element
// type of the EnvironmentVariables fields of checks and groups.
type EnvironmentVariable struct {
	// Key is the variable's name, such as "API_KEY". UpdateEnvironmentVariable
	// rejects a different key. Key is left out of the request when empty.
	Key string `json:"key,omitempty"`

	// Value must not be empty for an account-level variable. Reads return an
	// empty Value for a secret variable, and a secret variable of a check or
	// group sent with an empty Value keeps its stored value.
	Value string `json:"value"`

	// Locked hides the value in the Checkly UI, like a password.
	Locked bool `json:"locked"`

	// Secret stores the value as a secret, which reads never return. A
	// secret variable cannot be made non-secret again.
	Secret bool `json:"secret"`
}

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

func (c *client) DeleteEnvironmentVariable(
	ctx context.Context,
	key string,
) error {
	_, err := c.call(ctx, http.MethodDelete, fmt.Sprintf("/v1/variables/%s", key), nil, http.StatusNoContent)
	return err
}
