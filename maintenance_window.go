package checkly

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// MaintenanceWindowRepeatUnit is the unit of a maintenance window's repeat
// interval, in [MaintenanceWindow].RepeatUnit.
type MaintenanceWindowRepeatUnit string

// Maintenance window repeat units, the values of
// [MaintenanceWindow].RepeatUnit.
const (
	MaintenanceWindowRepeatUnitDay   MaintenanceWindowRepeatUnit = "DAY"
	MaintenanceWindowRepeatUnitWeek  MaintenanceWindowRepeatUnit = "WEEK"
	MaintenanceWindowRepeatUnitMonth MaintenanceWindowRepeatUnit = "MONTH"
)

// MaintenanceWindow defines a type for a maintenance window.
//
// Create and update requests always send every field, so an update replaces
// the whole window: the API keeps a field's stored value only when the field
// is omitted, and this type omits only a nil
// StatusPageVisibility.ComponentConfigs. A zero value means the API's
// default: no repeat, UTC, no description, empty tag lists, nothing paused or
// silenced, and a StatusPageVisibility that is hidden and unlinked.
type MaintenanceWindow struct {
	ID             int64                       `json:"id"`
	Name           string                      `json:"name"`
	StartsAt       string                      `json:"startsAt"`
	EndsAt         string                      `json:"endsAt"`
	RepeatInterval int                         `json:"repeatInterval"`
	RepeatUnit     MaintenanceWindowRepeatUnit `json:"repeatUnit"`
	RepeatEndsAt   string                      `json:"repeatEndsAt"`
	Tags           []string                    `json:"tags"`

	// Timezone is the named IANA time zone used for recurring maintenance
	// scheduling, e.g. "America/New_York". UTC offset identifiers such as
	// "+05:00" are not accepted. Empty means UTC.
	Timezone string `json:"timezone"`

	// PauseAllChecks pauses every check in the account regardless of Tags.
	PauseAllChecks bool `json:"pauseAllChecks"`

	// SilenceAlertsTags selects which checks have their alerts silenced.
	// Ignored when SilenceAllAlerts is true.
	SilenceAlertsTags []string `json:"silenceAlertsTags"`

	// SilenceAllAlerts silences alerts for every check in the account,
	// overriding SilenceAlertsTags.
	SilenceAllAlerts bool `json:"silenceAllAlerts"`

	// Description is shown on linked status pages when the window is visible
	// there. An empty Description is sent as null; the API compares null and
	// an empty string as different values, so an update of a window stored
	// with an empty-string description counts as a description change.
	Description string `json:"description"`

	// StatusPageVisibility controls whether and how the window appears on
	// status pages.
	StatusPageVisibility MaintenanceWindowStatusPageVisibility `json:"statusPageVisibility"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// MarshalJSON sends every field. Zero values the API does not accept are
// sent as its equivalent default: null for an unset repeat, time zone or
// description, and an empty list for nil tag lists. Fields declared here
// shadow the embedded ones of the same JSON name.
func (mw MaintenanceWindow) MarshalJSON() ([]byte, error) {
	type alias MaintenanceWindow
	return json.Marshal(struct {
		alias
		RepeatInterval    *int                         `json:"repeatInterval"`
		RepeatUnit        *MaintenanceWindowRepeatUnit `json:"repeatUnit"`
		RepeatEndsAt      *string                      `json:"repeatEndsAt"`
		Tags              []string                     `json:"tags"`
		Timezone          *string                      `json:"timezone"`
		SilenceAlertsTags []string                     `json:"silenceAlertsTags"`
		Description       *string                      `json:"description"`
	}{
		alias:             alias(mw),
		RepeatInterval:    nullIfZero(mw.RepeatInterval),
		RepeatUnit:        nullIfZero(mw.RepeatUnit),
		RepeatEndsAt:      nullIfZero(mw.RepeatEndsAt),
		Tags:              emptyIfNil(mw.Tags),
		Timezone:          nullIfZero(mw.Timezone),
		SilenceAlertsTags: emptyIfNil(mw.SilenceAlertsTags),
		Description:       nullIfZero(mw.Description),
	})
}

// MaintenanceWindowStatusPageVisibility controls how a maintenance window
// appears on status pages. Every setting other than Enabled only takes effect
// when Enabled is true. The zero value is the API's default: hidden, with no
// linked status pages or services.
//
// An account without the status page maintenance windows entitlement can only
// change Enabled from true to false; any other change is rejected, so such an
// account must send the stored settings back unchanged apart from Enabled.
//
// Status pages with services (v2) are linked with ServiceIDs or
// AffectAllServices, and status pages with components (v3) with
// ComponentConfigs.
type MaintenanceWindowStatusPageVisibility struct {
	Enabled bool `json:"enabled"`

	// Severity is the severity shown on status pages. Empty means none.
	Severity              IncidentSeverity `json:"severity"`
	AffectAllServices     bool             `json:"affectAllServices"`
	NotifyOnStart         bool             `json:"notifyOnStart"`
	NotifyOnEnd           bool             `json:"notifyOnEnd"`
	SuppressAutoIncidents bool             `json:"suppressAutoIncidents"`
	ReminderMinutesBefore []int            `json:"reminderMinutesBefore"`

	// AutoStart, AutoEnd and ShowAffectedServices default to true when nil.
	AutoStart            *bool `json:"autoStart"`
	AutoEnd              *bool `json:"autoEnd"`
	ShowAffectedServices *bool `json:"showAffectedServices"`

	// StatusPageIDs requires ServiceIDs or AffectAllServices for a status
	// page with services, and an entry in ComponentConfigs for one with
	// components.
	StatusPageIDs []string `json:"statusPageIds"`
	ServiceIDs    []string `json:"serviceIds"`

	// ComponentConfigs lists the affected components of each linked status
	// page with components, at most one entry per page. Unlike the other
	// settings, nil is left out of the request, and an update then keeps
	// the stored component configs; a non-nil slice replaces them all, so an
	// empty one removes them. The API rejects an update that leaves them
	// out while StatusPageIDs is non-empty and neither ServiceIDs nor
	// AffectAllServices is set, so send back the ones a read returned.
	ComponentConfigs []MaintenanceWindowComponentConfig `json:"componentConfigs"`
}

// MaintenanceWindowComponentConfig lists the components of one status page
// that a maintenance window affects.
type MaintenanceWindowComponentConfig struct {
	// StatusPageID is the ID of a status page listed in StatusPageIDs.
	StatusPageID string `json:"statusPageId"`
	// AffectAllComponents affects every component of the page, including
	// components added later. It requires Components to be empty.
	AffectAllComponents bool                               `json:"affectAllComponents"`
	Components          []MaintenanceWindowComponentImpact `json:"components,omitempty"`
}

// MaintenanceWindowComponentImpact is a status page component affected by a
// maintenance window, and the status it shows while the window is in
// progress.
type MaintenanceWindowComponentImpact struct {
	ComponentID string `json:"componentId"`
	// Status is the status the component shows. Empty means
	// UNDER_MAINTENANCE.
	Status StatusPageComponentStatusV3 `json:"status,omitempty"`
}

// MarshalJSON sends every setting other than a nil ComponentConfigs, using
// the API's defaults for unset ones. Fields declared here shadow the embedded
// ones of the same JSON name.
func (v MaintenanceWindowStatusPageVisibility) MarshalJSON() ([]byte, error) {
	type alias MaintenanceWindowStatusPageVisibility
	// omitempty would also leave out an empty slice, which must be sent to
	// remove the stored component configs; only a nil pointer is left out.
	var componentConfigs *[]MaintenanceWindowComponentConfig
	if v.ComponentConfigs != nil {
		componentConfigs = &v.ComponentConfigs
	}
	return json.Marshal(struct {
		alias
		Severity              *IncidentSeverity                   `json:"severity"`
		ReminderMinutesBefore []int                               `json:"reminderMinutesBefore"`
		AutoStart             bool                                `json:"autoStart"`
		AutoEnd               bool                                `json:"autoEnd"`
		ShowAffectedServices  bool                                `json:"showAffectedServices"`
		StatusPageIDs         []string                            `json:"statusPageIds"`
		ServiceIDs            []string                            `json:"serviceIds"`
		ComponentConfigs      *[]MaintenanceWindowComponentConfig `json:"componentConfigs,omitempty"`
	}{
		alias:                 alias(v),
		Severity:              nullIfZero(v.Severity),
		ReminderMinutesBefore: emptyIfNil(v.ReminderMinutesBefore),
		AutoStart:             trueIfNil(v.AutoStart),
		AutoEnd:               trueIfNil(v.AutoEnd),
		ShowAffectedServices:  trueIfNil(v.ShowAffectedServices),
		StatusPageIDs:         emptyIfNil(v.StatusPageIDs),
		ServiceIDs:            emptyIfNil(v.ServiceIDs),
		ComponentConfigs:      componentConfigs,
	})
}

// CreateMaintenanceWindow creates a new window with the specified details.
func (c *client) CreateMaintenanceWindow(
	ctx context.Context,
	mw MaintenanceWindow,
) (*MaintenanceWindow, error) {
	var result MaintenanceWindow
	err := c.callJSON(ctx, http.MethodPost, "/v1/maintenance-windows", mw, &result, http.StatusOK, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetMaintenanceWindow takes the ID of an existing window, and returns the
// corresponding window.
func (c *client) GetMaintenanceWindow(
	ctx context.Context,
	ID int64,
) (*MaintenanceWindow, error) {
	result := MaintenanceWindow{}
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/maintenance-windows/%d", ID), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateMaintenanceWindow takes the ID of an existing window, and updates the
// corresponding window to match the supplied one.
func (c *client) UpdateMaintenanceWindow(
	ctx context.Context,
	ID int64,
	mw MaintenanceWindow,
) (*MaintenanceWindow, error) {
	var result MaintenanceWindow
	err := c.callJSON(ctx, http.MethodPut, fmt.Sprintf("/v1/maintenance-windows/%d", ID), mw, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteMaintenanceWindow deletes the window with the specified ID.
func (c *client) DeleteMaintenanceWindow(
	ctx context.Context,
	ID int64,
) error {
	_, err := c.call(ctx, http.MethodDelete, fmt.Sprintf("/v1/maintenance-windows/%d", ID), nil, http.StatusNoContent)
	return err
}
