package checkly

import (
	"context"
	"fmt"
	"net/http"
)

type StatusPageTheme string

const (
	StatusPageThemeAuto  StatusPageTheme = "AUTO"
	StatusPageThemeDark  StatusPageTheme = "DARK"
	StatusPageThemeLight StatusPageTheme = "LIGHT"
)

type StatusPage struct {
	// ID is the Checkly identifier of the status page.
	ID string `json:"id,omitempty"`

	// Name is the name of the status page.
	Name string `json:"name"`

	// URL is the unique subdomain of the status page.
	URL string `json:"url"`

	// CustomDomain is an optional user-managed domain that hosts the status
	// page.
	CustomDomain string `json:"customDomain,omitempty"`

	// Logo is a URL to an image file to use as the logo for the status page.
	Logo string `json:"logo"`

	// RedirectTo is the URL the user should be redirected to when clicking
	// the logo.
	RedirectTo string `json:"redirectTo"`

	// Favicon is a URL to an image file to use as the favicon of the status
	// page.
	Favicon string `json:"favicon"`

	// DefaultTheme is the default theme of the status page.
	DefaultTheme StatusPageTheme `json:"defaultTheme,omitempty"`

	// Cards is a list of cards to include on the status page.
	Cards []StatusPageCard `json:"cards"`
}

type StatusPageCard struct {
	// Name is the name of the card.
	Name string `json:"name"`

	// Services is the list of services to include in the card.
	Services []StatusPageService `json:"services"`
}

type StatusPageService struct {
	// ID is the Checkly identifier of the service.
	ID string `json:"id,omitempty"`

	// Name is the name of the service.
	Name string `json:"name"`
}

// statusPagePayload sends an empty custom domain as null, which removes it,
// rather than omitting it: the update endpoint keeps the stored value of an
// omitted key.
type statusPagePayload struct {
	StatusPage
	CustomDomain *string `json:"customDomain"`
}

func createStatusPagePayload(page StatusPage) statusPagePayload {
	return statusPagePayload{
		StatusPage:   page,
		CustomDomain: nullIfZero(page.CustomDomain),
	}
}

func (c *client) CreateStatusPage(
	ctx context.Context,
	page StatusPage,
) (*StatusPage, error) {
	var result StatusPage
	err := c.callJSON(ctx, http.MethodPost, "/v1/status-pages", createStatusPagePayload(page), &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetStatusPage(
	ctx context.Context,
	ID string,
) (*StatusPage, error) {
	var result StatusPage
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/status-pages/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateStatusPage(
	ctx context.Context,
	ID string,
	page StatusPage,
) (*StatusPage, error) {
	var result StatusPage
	err := c.callJSON(ctx, http.MethodPut, fmt.Sprintf("/v1/status-pages/%s", ID), createStatusPagePayload(page), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteStatusPage(
	ctx context.Context,
	ID string,
) error {
	_, err := c.call(ctx, http.MethodDelete, fmt.Sprintf("/v1/status-pages/%s", ID), nil, http.StatusNoContent)
	return err
}

func (c *client) CreateStatusPageService(
	ctx context.Context,
	service StatusPageService,
) (*StatusPageService, error) {
	var result StatusPageService
	err := c.callJSON(ctx, http.MethodPost, "/v1/status-pages/services", service, &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetStatusPageService(
	ctx context.Context,
	ID string,
) (*StatusPageService, error) {
	var result StatusPageService
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/status-pages/services/%s", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateStatusPageService(
	ctx context.Context,
	ID string,
	service StatusPageService,
) (*StatusPageService, error) {
	var result StatusPageService
	err := c.callJSON(ctx, http.MethodPut, fmt.Sprintf("/v1/status-pages/services/%s", ID), service, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteStatusPageService(
	ctx context.Context,
	ID string,
) error {
	_, err := c.call(ctx, http.MethodDelete, fmt.Sprintf("/v1/status-pages/services/%s", ID), nil, http.StatusNoContent)
	return err
}
