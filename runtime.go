package checkly

import (
	"context"
	"fmt"
	"net/http"
)

// RuntimeStage is the release stage of a runtime, in [Runtime].Stage.
type RuntimeStage string

// Runtime stages, the values of [Runtime].Stage.
const (
	RuntimeStageAlpha      RuntimeStage = "ALPHA"
	RuntimeStageBeta       RuntimeStage = "BETA"
	RuntimeStageCurrent    RuntimeStage = "CURRENT"
	RuntimeStageStable     RuntimeStage = "STABLE"
	RuntimeStageDeprecated RuntimeStage = "DEPRECATED"
	RuntimeStageRemoved    RuntimeStage = "REMOVED"
)

type Runtime struct {
	Name             string       `json:"name"`
	MultiStepSupport bool         `json:"multiStepSupport"`
	Stage            RuntimeStage `json:"stage"`
	RuntimeEndOfLife string       `json:"runtimeEndOfLife"`
	Description      string       `json:"description"`
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
