package checkly

import (
	"context"
	"fmt"
	"net/http"
)

type Runtime struct {
	Name             string `json:"name"`
	MultiStepSupport bool   `json:"multiStepSupport"`
	Stage            string `json:"stage"`
	RuntimeEndOfLife string `json:"runtimeEndOfLife"`
	Description      string `json:"description"`
}

// GetRuntime returns the specs of a runtime. The endpoint is public, so the
// request is sent without credentials.
func (c *client) GetRuntime(
	ctx context.Context,
	ID string,
) (*Runtime, error) {
	result := Runtime{}
	err := c.callPublicJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/runtimes/%s", ID), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
