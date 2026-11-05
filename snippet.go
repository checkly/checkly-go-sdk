package checkly

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Snippet defines Snippet type
type Snippet struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Script    string    `json:"script"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CreateSnippet creates a new snippet with the specified details. It returns
// the newly-created snippet, or an error.
func (c *client) CreateSnippet(
	ctx context.Context,
	snippet Snippet,
) (*Snippet, error) {
	var result Snippet
	err := c.callJSON(ctx, http.MethodPost, "/v1/snippets", snippet, &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetSnippet takes the ID of an existing snippet, and returns the
// corresponding snippet, or an error.
func (c *client) GetSnippet(
	ctx context.Context,
	ID int64,
) (*Snippet, error) {
	result := Snippet{}
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/snippets/%d", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateSnippet takes the ID of an existing snippet, and updates the
// corresponding snippet to match the supplied snippet. It returns the updated
// snippet, or an error.
func (c *client) UpdateSnippet(
	ctx context.Context,
	ID int64,
	snippet Snippet,
) (*Snippet, error) {
	var result Snippet
	err := c.callJSON(ctx, http.MethodPut, fmt.Sprintf("/v1/snippets/%d", ID), snippet, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteSnippet deletes the snippet with the specified ID. It returns a
func (c *client) DeleteSnippet(
	ctx context.Context,
	ID int64,
) error {
	_, err := c.call(ctx, http.MethodDelete, fmt.Sprintf("/v1/snippets/%d", ID), nil, http.StatusNoContent)
	return err
}
