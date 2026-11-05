package checkly

import (
	"context"
	"net/http"
)

// DashboardKey represents the keys that the dashboard has.
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

// Dashboard defines a type for a dashboard.
type Dashboard struct {
	ID                 int64          `json:"id,omitempty"`
	DashboardID        string         `json:"dashboardId,omitempty"`
	CustomURL          string         `json:"customUrl"`
	CustomDomain       string         `json:"customDomain,omitempty"`
	IsPrivate          bool           `json:"isPrivate,omitempty"`
	Logo               string         `json:"logo,omitempty"`
	Link               string         `json:"link,omitempty"`
	Description        string         `json:"description,omitempty"`
	Favicon            string         `json:"favicon,omitempty"`
	Header             string         `json:"header,omitempty"`
	ShowHeader         *bool          `json:"showHeader,omitempty"` // If nil, defaults to true.
	Width              DashboardWidth `json:"width,omitempty"`
	RefreshRate        int            `json:"refreshRate,omitempty"` // One of 60, 300 or 600 seconds. If 0, defaults to 60.
	ChecksPerPage      int            `json:"checksPerPage,omitempty"`
	PaginationRate     int            `json:"paginationRate,omitempty"` // One of 30, 60 or 300 seconds. If 0, defaults to 60.
	Paginate           *bool          `json:"paginate,omitempty"`       // If nil, defaults to true.
	SortByStatus       bool           `json:"sortByStatus,omitempty"`
	ShowGroupNames     *bool          `json:"showGroupNames,omitempty"` // If nil, defaults to true.
	Tags               []string       `json:"tags,omitempty"`
	HideTags           bool           `json:"hideTags,omitempty"`
	UseTagsAndOperator bool           `json:"useTagsAndOperator,omitempty"`
	EnableIncidents    bool           `json:"enableIncidents"`
	ExpandChecks       bool           `json:"expandChecks"`
	ShowCheckRunLinks  bool           `json:"showCheckRunLinks"`
	CustomCSS          string         `json:"customCSS"`
	ShowP95            *bool          `json:"showP95,omitempty"` // If nil, defaults to true.
	ShowP99            *bool          `json:"showP99,omitempty"` // If nil, defaults to true.
	CreatedAt          string         `json:"created_at"`
	UpdatedAt          string         `json:"updated_at"`
	Keys               []DashboardKey `json:"keys,omitempty"`
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
	ID string,
) (*Dashboard, error) {
	result := Dashboard{}
	err := c.callJSON(ctx, http.MethodGet, "/v1/dashboards/"+ID, nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateDashboard(
	ctx context.Context,
	ID string,
	dashboard Dashboard,
) (*Dashboard, error) {
	var result Dashboard
	err := c.callJSON(ctx, http.MethodPut, "/v1/dashboards/"+ID, newDashboardUpdatePayload(dashboard), &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteDashboard(
	ctx context.Context,
	ID string,
) error {
	_, err := c.call(ctx, http.MethodDelete, "/v1/dashboards/"+ID, nil, http.StatusNoContent)
	return err
}
