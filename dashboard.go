package checkly

import (
	"context"
	"net/http"
)

// DashboardKey is an access key of a dashboard, in Dashboard.Keys.
type DashboardKey struct {
	ID        string `json:"id"`
	MaskedKey string `json:"maskedKey"`
	RawKey    string `json:"rawKey"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// DashboardWidth is the page width of a dashboard, in [Dashboard].Width.
type DashboardWidth string

// Dashboard widths, the values of [Dashboard].Width. The API default is
// DashboardWidthFull.
const (
	DashboardWidthFull  DashboardWidth = "FULL"
	DashboardWidth960PX DashboardWidth = "960PX"
)

// Dashboard is a public or private web page that shows the status of the
// checks selected by Tags. It is the input and result of CreateDashboard and
// UpdateDashboard.
type Dashboard struct {
	ID int64 `json:"id,omitempty"`

	// DashboardID is the identifier that GetDashboard, UpdateDashboard and
	// DeleteDashboard take.
	DashboardID string `json:"dashboardId,omitempty"`

	// CustomURL is a subdomain of checklyhq.com, unique across all
	// accounts. CustomDomain is a domain of the account's own, such as
	// "status.example.com". CustomDomain is left out of the CreateDashboard
	// request when empty; UpdateDashboard removes it when empty.
	CustomURL    string `json:"customUrl"`
	CustomDomain string `json:"customDomain,omitempty"`

	// IsPrivate makes the dashboard viewable only with one of its Keys.
	// IsPrivate is left out of the request when false.
	IsPrivate bool `json:"isPrivate,omitempty"`

	// Logo and Favicon are http or https URLs of image files. Link is
	// where clicking the logo leads. Logo, Link, Description and Favicon are
	// left out of the CreateDashboard request when empty; UpdateDashboard
	// removes them when empty.
	Logo        string `json:"logo,omitempty"`
	Link        string `json:"link,omitempty"`
	Description string `json:"description,omitempty"`
	Favicon     string `json:"favicon,omitempty"`

	// Header is the text at the top of the dashboard. It is required on
	// create. Header is left out of the CreateDashboard request when empty;
	// UpdateDashboard removes it when empty.
	Header string `json:"header,omitempty"`
	// ShowHeader is left out of the request when nil; the API defaults it to
	// true.
	ShowHeader *bool `json:"showHeader,omitempty"`

	// Width is left out of the request when empty.
	Width DashboardWidth `json:"width,omitempty"`

	// RefreshRate is how often the dashboard refreshes, in seconds: 60, 300
	// or 600. RefreshRate is left out of the request when zero, so the API
	// applies its default, 60.
	RefreshRate int `json:"refreshRate,omitempty"`

	// ChecksPerPage is between 1 and 20. ChecksPerPage is left out of the
	// request when zero, so the API applies its default, 15.
	ChecksPerPage int `json:"checksPerPage,omitempty"`

	// PaginationRate is how often the dashboard turns the page, in
	// seconds: 30, 60 or 300. PaginationRate is left out of the request when
	// zero, so the API applies its default, 60.
	PaginationRate int `json:"paginationRate,omitempty"`

	// Paginate is left out of the request when nil; the API defaults it to
	// true.
	Paginate *bool `json:"paginate,omitempty"`

	// SortByStatus lists failing checks first. SortByStatus is left out of
	// the request when false.
	SortByStatus bool `json:"sortByStatus,omitempty"`
	// ShowGroupNames is left out of the request when nil; the API defaults
	// it to true.
	ShowGroupNames *bool `json:"showGroupNames,omitempty"`

	// Tags selects the checks to show. UseTagsAndOperator requires a check
	// to have all of them. Tags is left out of the request when empty.
	Tags []string `json:"tags,omitempty"`
	// HideTags and UseTagsAndOperator are left out of the request when false.
	HideTags           bool   `json:"hideTags,omitempty"`
	UseTagsAndOperator bool   `json:"useTagsAndOperator,omitempty"`
	EnableIncidents    bool   `json:"enableIncidents"`
	ExpandChecks       bool   `json:"expandChecks"`
	ShowCheckRunLinks  bool   `json:"showCheckRunLinks"`
	CustomCSS          string `json:"customCSS"`
	// ShowP95 and ShowP99 are left out of the request when nil; the API
	// defaults them to true.
	ShowP95   *bool  `json:"showP95,omitempty"`
	ShowP99   *bool  `json:"showP99,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`

	// Keys are the access keys of the dashboard. The response of
	// CreateDashboard includes a new key with its RawKey; keys sent in a
	// request are ignored.
	Keys []DashboardKey `json:"keys,omitempty"`
}

// dashboardUpdatePayload is the body of a dashboard update. The update
// endpoint keeps the stored value of an omitted key that has no default, so
// the fields below shadow the embedded fields that omit their empty value:
// they are always sent, as null (or "" for header, which does not accept
// null) when empty, so that emptying them clears the stored value. Omitted
// keys that have a default, such as refreshRate or paginate, get it on
// update as on create.
type dashboardUpdatePayload struct {
	Dashboard
	CustomDomain *string `json:"customDomain"`
	Logo         *string `json:"logo"`
	Link         *string `json:"link"`
	Description  *string `json:"description"`
	Favicon      *string `json:"favicon"`
	Header       string  `json:"header"`
}

func newDashboardUpdatePayload(dashboard Dashboard) dashboardUpdatePayload {
	return dashboardUpdatePayload{
		Dashboard:    dashboard,
		CustomDomain: nullIfZero(dashboard.CustomDomain),
		Logo:         nullIfZero(dashboard.Logo),
		Link:         nullIfZero(dashboard.Link),
		Description:  nullIfZero(dashboard.Description),
		Favicon:      nullIfZero(dashboard.Favicon),
		Header:       dashboard.Header,
	}
}

func (c *client) CreateDashboard(
	ctx context.Context,
	dashboard Dashboard,
) (*Dashboard, error) {
	var result Dashboard
	err := c.callJSON(ctx, http.MethodPost, "/v1/dashboards", dashboard, &result, http.StatusOK, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetDashboard(
	ctx context.Context,
	id string,
) (*Dashboard, error) {
	result := Dashboard{}
	err := c.callJSON(ctx, http.MethodGet, "/v1/dashboards/"+id, nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateDashboard(
	ctx context.Context,
	id string,
	dashboard Dashboard,
) (*Dashboard, error) {
	var result Dashboard
	err := c.callJSON(ctx, http.MethodPut, "/v1/dashboards/"+id, newDashboardUpdatePayload(dashboard), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteDashboard(
	ctx context.Context,
	id string,
) error {
	_, err := c.call(ctx, http.MethodDelete, "/v1/dashboards/"+id, nil, http.StatusNoContent)
	return err
}
