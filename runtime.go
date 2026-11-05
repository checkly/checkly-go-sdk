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

// Runtime describes a Checkly runtime, the versioned set of Node.js and
// libraries that checks run on. It is returned by GetRuntime.
type Runtime struct {
	// Name is the runtime's ID, such as "2025.04", as the RuntimeID fields
	// of checks and groups take it.
	Name             string `json:"name"`
	MultiStepSupport bool   `json:"multiStepSupport"`

	// Stage is the runtime's life stage.
	Stage RuntimeStage `json:"stage"`

	// RuntimeEndOfLife is the date on which the runtime will be removed,
	// if one is set.
	RuntimeEndOfLife string `json:"runtimeEndOfLife"`
	Description      string `json:"description"`
}

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
