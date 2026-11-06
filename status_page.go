package checkly

import (
	"context"
	"fmt"
	"net/http"
)

// StatusPageTheme is the default color theme of a status page, in
// [StatusPage].DefaultTheme and [StatusPageV3].DefaultTheme.
type StatusPageTheme string

// Status page themes, the values of [StatusPage].DefaultTheme and
// [StatusPageV3].DefaultTheme.
const (
	StatusPageThemeAuto  StatusPageTheme = "AUTO"
	StatusPageThemeDark  StatusPageTheme = "DARK"
	StatusPageThemeLight StatusPageTheme = "LIGHT"
)

// StatusPage is a v1 status page, which shows StatusPageService entries
// grouped into cards. It is the input and result of CreateStatusPage and
// UpdateStatusPage, which both require Name, URL and Cards.
//
// Deprecated: Use [StatusPageV3] instead. The public API cannot migrate a v1
// status page to v3, so this type remains for existing v1 pages.
type StatusPage struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`

	// URL is the unique subdomain of the status page. The API stores it in
	// lowercase.
	URL string `json:"url"`

	// CustomDomain is an optional user-managed domain that hosts the status
	// page; empty removes it.
	CustomDomain string `json:"customDomain,omitempty"`

	// Logo is a URL to an image file to use as the logo for the status page.
	Logo string `json:"logo"`

	// RedirectTo is the URL the user should be redirected to when clicking
	// the logo.
	RedirectTo string `json:"redirectTo"`

	// Favicon is a URL to an image file to use as the favicon of the status
	// page.
	Favicon string `json:"favicon"`

	// DefaultTheme is left out of the request when empty, so the API
	// applies [StatusPageThemeAuto], also on update.
	DefaultTheme StatusPageTheme `json:"defaultTheme,omitempty"`

	Cards []StatusPageCard `json:"cards"`
}

// StatusPageCard is a card of a v1 status page, in StatusPage.Cards: a named
// group of services.
//
// Deprecated: Use a [StatusPageComponentV3] of type
// [StatusPageComponentV3TypeGroup] instead. The public API cannot migrate a v1
// status page to v3, so this type remains for existing v1 pages.
type StatusPageCard struct {
	Name string `json:"name"`

	// Services lists the services shown in the card. Create and update
	// requests refer to existing services by ID; see
	// CreateStatusPageService.
	Services []StatusPageService `json:"services"`
}

// StatusPageService is a service of the v1 status pages: something whose
// status a card shows and that incidents affect, as IncidentTrigger.ServiceID
// refers to it. It is the input and result of CreateStatusPageService and
// UpdateStatusPageService, which both require only Name.
//
// Deprecated: Use [StatusPageComponentV3] instead. The public API cannot
// migrate a v1 status page to v3, so this type remains for existing v1 pages.
// [IncidentTrigger].ServiceID takes only v1 services, not component IDs: v3
// pages automate incidents with [StatusPageAutomationRuleV3].
type StatusPageService struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}

// statusPagePayload sends an empty custom domain as null, which removes it,
// rather than omitting it: the update endpoint keeps the stored value of an
// omitted key. Unset card and service lists are sent as [], as the API
// rejects null for them.
type statusPagePayload struct {
	StatusPage
	CustomDomain *string `json:"customDomain"`
}

func createStatusPagePayload(page StatusPage) statusPagePayload {
	// Copy the cards so that filling in their services leaves the caller's
	// slice unchanged.
	cards := make([]StatusPageCard, len(page.Cards))
	for i, card := range page.Cards {
		card.Services = emptyIfNil(card.Services)
		cards[i] = card
	}
	page.Cards = cards
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
