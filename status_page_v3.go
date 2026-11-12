package checkly

import (
	"context"
	"fmt"
	"net/http"
)

// StatusPageV3 is a v3 (components-based) status page. Unlike StatusPage, a
// v3 page has no cards or services: its structure is managed through
// StatusPageComponentV3 resources, and incidents can be automated with
// StatusPageAutomationRuleV3 resources.
type StatusPageV3 struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`

	// URL is the unique subdomain of the status page.
	URL string `json:"url"`

	// CustomDomain is an optional user-managed domain that hosts the status
	// page.
	CustomDomain string `json:"customDomain"`

	// Description is a short description shown on the status page.
	Description string `json:"description"`

	// Logo is a URL to an image file to use as the logo for the status page.
	Logo string `json:"logo"`

	// LogoDark is a URL to an image file to use as the logo in dark mode.
	LogoDark string `json:"logoDark"`

	// RedirectTo is the URL the user should be redirected to when clicking
	// the logo.
	RedirectTo string `json:"redirectTo"`

	// Favicon is a URL to an image file to use as the favicon of the status
	// page.
	Favicon string `json:"favicon"`

	// An empty DefaultTheme means [StatusPageThemeAuto], also on update.
	DefaultTheme StatusPageTheme `json:"defaultTheme,omitempty"`

	// PrivacyPolicyLink is a link to a privacy policy, shown in the page
	// footer.
	PrivacyPolicyLink string `json:"privacyPolicyLink"`

	// TermsOfServiceLink is a link to terms of service, shown in the page
	// footer.
	TermsOfServiceLink string `json:"termsOfServiceLink"`

	// SupportLink is a support contact link, shown in the page footer. It
	// accepts http, https and mailto URLs.
	SupportLink string `json:"supportLink"`

	// FooterText is free-form footer text.
	FooterText string `json:"footerText"`

	// GoogleAnalyticsTag is a Google Analytics tag ID (e.g. "G-XXXXXXXXXX")
	// to embed on the public page.
	GoogleAnalyticsTag string `json:"googleAnalyticsTag"`

	// AllowIndexing determines whether search engines may index the public
	// page.
	AllowIndexing bool `json:"allowIndexing"`

	// ThemeColors are custom colors for the light and dark theme of the
	// page. Setting them requires custom theme colors to be part of the
	// account's plan; the API rejects them otherwise. A nil value clears
	// the custom colors on writes. Reads always return a complete palette:
	// the stored colors, or the defaults when none are stored or the plan
	// does not include them. A page that was read therefore cannot be sent
	// back unchanged: set ThemeColors to nil, or to the colors that should
	// be stored, before updating it, or the defaults are stored as custom
	// colors (and rejected on plans without them).
	ThemeColors *StatusPageV3ThemeColors `json:"themeColors"`
}

// StatusPageV3ThemeColors holds the custom colors of a v3 status page, one
// group per theme. The API requires both groups, with every color set, when
// custom colors are given.
type StatusPageV3ThemeColors struct {
	// Light are the colors used when the page renders in light mode.
	Light StatusPageV3ThemeColorGroup `json:"light"`

	// Dark are the colors used when the page renders in dark mode.
	Dark StatusPageV3ThemeColorGroup `json:"dark"`
}

// StatusPageV3ThemeColorGroup holds the colors of one theme of a v3 status
// page. Each color is a hex string such as "#FF0000" or "#F00".
type StatusPageV3ThemeColorGroup struct {
	// BodyBackgroundColor is the background of the page.
	BodyBackgroundColor string `json:"bodyBackgroundColor"`

	// HeaderBackgroundColor is the background of the page header.
	HeaderBackgroundColor string `json:"headerBackgroundColor"`

	// HeaderFontColor is the color of text in the page header.
	HeaderFontColor string `json:"headerFontColor"`

	// TitleFontColor is the color of titles and headings.
	TitleFontColor string `json:"titleFontColor"`

	// BodyFontColor is the color of regular body text.
	BodyFontColor string `json:"bodyFontColor"`

	// BodyFontColorMuted is the color of de-emphasized body text, such as
	// timestamps.
	BodyFontColorMuted string `json:"bodyFontColorMuted"`

	// NavigationFontColor is the color of navigation links.
	NavigationFontColor string `json:"navigationFontColor"`

	// LinkFontColor is the color of links in the page content.
	LinkFontColor string `json:"linkFontColor"`

	// CardBackgroundColor is the background of component and incident
	// cards.
	CardBackgroundColor string `json:"cardBackgroundColor"`

	// BorderColor is the color of borders and dividers.
	BorderColor string `json:"borderColor"`

	// PrimaryButtonBackgroundColor is the background of primary buttons,
	// such as "Subscribe".
	PrimaryButtonBackgroundColor string `json:"primaryButtonBackgroundColor"`

	// PrimaryButtonFontColor is the color of text on primary buttons.
	PrimaryButtonFontColor string `json:"primaryButtonFontColor"`
}

// StatusPageComponentV3Type is the type of a v3 status page component, in
// [StatusPageComponentV3].Type.
type StatusPageComponentV3Type string

// Component types, the values of [StatusPageComponentV3].Type: a SERVICE has
// a status of its own, a GROUP contains other components.
const (
	StatusPageComponentV3TypeService StatusPageComponentV3Type = "SERVICE"
	StatusPageComponentV3TypeGroup   StatusPageComponentV3Type = "GROUP"
)

// StatusPageComponentV3Configuration holds a component's type-specific
// settings. A SERVICE component only supports ShowHistoricalData; a GROUP
// component supports both. Unset (nil) properties take the type's default
// on writes; the API rejects a property the component's type does not
// support. Reads always return every property the type supports.
type StatusPageComponentV3Configuration struct {
	// ShowHistoricalData shows the historical status (the uptime bar) of
	// the component on the status page. Defaults to true.
	ShowHistoricalData *bool `json:"showHistoricalData,omitempty"`

	// ExpandedByDefault renders the group expanded when the status page
	// loads. GROUP components only. Defaults to false.
	ExpandedByDefault *bool `json:"expandedByDefault,omitempty"`
}

// StatusPageComponentV3 represents a component of a v3 status page: either a
// SERVICE (a monitored thing with its own status) or a GROUP (a container
// for other components).
type StatusPageComponentV3 struct {
	ID string `json:"id,omitempty"`

	// StatusPageID is the identifier of the v3 status page the component
	// belongs to. It is read-only: on writes the page is addressed through
	// the URL path instead.
	StatusPageID string `json:"statusPageId,omitempty"`

	// An empty Type means [StatusPageComponentV3TypeService], also on update.
	Type StatusPageComponentV3Type `json:"type,omitempty"`

	// Name is the name shown on the status page.
	Name string `json:"name"`

	// Description is an optional description shown next to the name.
	Description string `json:"description,omitempty"`

	// DisplayOrder is the position among siblings; lower comes first.
	DisplayOrder int `json:"displayOrder"`

	// Hidden hides the component from the public page while keeping it
	// available for incidents and automation.
	Hidden bool `json:"hidden"`

	// Configuration holds the component's type-specific settings. Nil means
	// the defaults of the component's type, also on update.
	Configuration *StatusPageComponentV3Configuration `json:"configuration,omitempty"`

	// ParentID is the identifier of the GROUP component this component sits
	// under. It must belong to the same status page.
	ParentID string `json:"parentId,omitempty"`
}

// StatusPageComponentStatusV3 is the status a v3 status page component
// shows, in [MaintenanceWindowComponentImpact].Status.
type StatusPageComponentStatusV3 string

// Component statuses, the values of [MaintenanceWindowComponentImpact].Status.
const (
	StatusPageComponentStatusV3Operational         StatusPageComponentStatusV3 = "OPERATIONAL"
	StatusPageComponentStatusV3UnderMaintenance    StatusPageComponentStatusV3 = "UNDER_MAINTENANCE"
	StatusPageComponentStatusV3DegradedPerformance StatusPageComponentStatusV3 = "DEGRADED_PERFORMANCE"
	StatusPageComponentStatusV3PartialOutage       StatusPageComponentStatusV3 = "PARTIAL_OUTAGE"
	StatusPageComponentStatusV3MajorOutage         StatusPageComponentStatusV3 = "MAJOR_OUTAGE"
)

// StatusPageTargetImpactV3 is the impact an automated incident sets on a
// component: every component status except OPERATIONAL.
type StatusPageTargetImpactV3 string

// Target impacts, the values of
// [StatusPageAutomationRuleComponentV3].TargetImpact.
const (
	StatusPageTargetImpactV3UnderMaintenance    StatusPageTargetImpactV3 = "UNDER_MAINTENANCE"
	StatusPageTargetImpactV3DegradedPerformance StatusPageTargetImpactV3 = "DEGRADED_PERFORMANCE"
	StatusPageTargetImpactV3PartialOutage       StatusPageTargetImpactV3 = "PARTIAL_OUTAGE"
	StatusPageTargetImpactV3MajorOutage         StatusPageTargetImpactV3 = "MAJOR_OUTAGE"
)

// StatusPageAutomationRuleComponentV3 links an automation rule to a
// component of the same status page, with the impact an automated incident
// sets on it.
type StatusPageAutomationRuleComponentV3 struct {
	ComponentID string `json:"componentId"`

	// TargetImpact is the impact set on the component while the automated
	// incident is open.
	TargetImpact StatusPageTargetImpactV3 `json:"targetImpact"`
}

// StatusPageAutomationRuleV3 represents an automation rule of a v3 status
// page. When a check whose tags overlap with the rule's tags fails, Checkly
// opens one incident on the page impacting the listed components, and
// resolves it when the check recovers.
type StatusPageAutomationRuleV3 struct {
	ID string `json:"id,omitempty"`

	// StatusPageID is the identifier of the v3 status page the rule belongs
	// to. It is read-only: on writes the page is addressed through the URL
	// path instead.
	StatusPageID string `json:"statusPageId,omitempty"`

	Name string `json:"name"`

	// Enabled determines whether the rule may open incidents. A disabled
	// rule never opens incidents.
	Enabled bool `json:"enabled"`

	// FirstUpdate is the body of the status update that opens the incident.
	FirstUpdate string `json:"firstUpdate"`

	// LastUpdate is the body of the status update that resolves the
	// incident.
	LastUpdate string `json:"lastUpdate"`

	// NotifySubscribers determines whether subscribers are notified of the
	// automated updates.
	NotifySubscribers bool `json:"notifySubscribers"`

	// CoolDownWindowMinutes is the minimum number of minutes after an
	// automated incident before this rule may open the next one. 0 disables
	// the cool down.
	CoolDownWindowMinutes int `json:"coolDownWindowMinutes"`

	// Tags is the list of tags the rule matches on. A failing check matches
	// when it, or its group, carries ANY of these tags.
	Tags []string `json:"tags"`

	// Components is the list of components an automated incident impacts,
	// with the impact each gets.
	Components []StatusPageAutomationRuleComponentV3 `json:"components,omitempty"`
}

func (c *client) CreateStatusPageV3(
	ctx context.Context,
	page StatusPageV3,
) (*StatusPageV3, error) {
	var result StatusPageV3
	err := c.callJSON(ctx, http.MethodPost, "/v3/status-pages", page, &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetStatusPageV3(
	ctx context.Context,
	id string,
) (*StatusPageV3, error) {
	var result StatusPageV3
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v3/status-pages/%s", id), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateStatusPageV3(
	ctx context.Context,
	id string,
	page StatusPageV3,
) (*StatusPageV3, error) {
	var result StatusPageV3
	err := c.callJSON(ctx, http.MethodPut, fmt.Sprintf("/v3/status-pages/%s", id), page, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteStatusPageV3(
	ctx context.Context,
	id string,
) error {
	_, err := c.call(ctx, http.MethodDelete, fmt.Sprintf("/v3/status-pages/%s", id), nil, http.StatusNoContent)
	return err
}

func (c *client) CreateStatusPageComponentV3(
	ctx context.Context,
	statusPageID string,
	component StatusPageComponentV3,
) (*StatusPageComponentV3, error) {
	var result StatusPageComponentV3
	err := c.callJSON(ctx, http.MethodPost, fmt.Sprintf("/v3/status-pages/%s/components", statusPageID), component, &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetStatusPageComponentV3(
	ctx context.Context,
	statusPageID string,
	componentID string,
) (*StatusPageComponentV3, error) {
	var result StatusPageComponentV3
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v3/status-pages/%s/components/%s", statusPageID, componentID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateStatusPageComponentV3(
	ctx context.Context,
	statusPageID string,
	componentID string,
	component StatusPageComponentV3,
) (*StatusPageComponentV3, error) {
	var result StatusPageComponentV3
	err := c.callJSON(ctx, http.MethodPut, fmt.Sprintf("/v3/status-pages/%s/components/%s", statusPageID, componentID), component, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteStatusPageComponentV3(
	ctx context.Context,
	statusPageID string,
	componentID string,
) error {
	_, err := c.call(ctx, http.MethodDelete, fmt.Sprintf("/v3/status-pages/%s/components/%s", statusPageID, componentID), nil, http.StatusNoContent)
	return err
}

func (c *client) CreateStatusPageAutomationRuleV3(
	ctx context.Context,
	statusPageID string,
	rule StatusPageAutomationRuleV3,
) (*StatusPageAutomationRuleV3, error) {
	var result StatusPageAutomationRuleV3
	err := c.callJSON(ctx, http.MethodPost, fmt.Sprintf("/v3/status-pages/%s/automation-rules", statusPageID), rule, &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetStatusPageAutomationRuleV3(
	ctx context.Context,
	statusPageID string,
	ruleID string,
) (*StatusPageAutomationRuleV3, error) {
	var result StatusPageAutomationRuleV3
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v3/status-pages/%s/automation-rules/%s", statusPageID, ruleID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateStatusPageAutomationRuleV3(
	ctx context.Context,
	statusPageID string,
	ruleID string,
	rule StatusPageAutomationRuleV3,
) (*StatusPageAutomationRuleV3, error) {
	var result StatusPageAutomationRuleV3
	err := c.callJSON(ctx, http.MethodPut, fmt.Sprintf("/v3/status-pages/%s/automation-rules/%s", statusPageID, ruleID), rule, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteStatusPageAutomationRuleV3(
	ctx context.Context,
	statusPageID string,
	ruleID string,
) error {
	_, err := c.call(ctx, http.MethodDelete, fmt.Sprintf("/v3/status-pages/%s/automation-rules/%s", statusPageID, ruleID), nil, http.StatusNoContent)
	return err
}
