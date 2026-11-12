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

// MaintenanceWindow is a scheduled, optionally recurring period during which
// checks are paused or their alerts silenced. It is the input and result of
// CreateMaintenanceWindow and UpdateMaintenanceWindow. A zero value means the
// API's default: no repeat, UTC, no description, empty tag lists, nothing
// paused or silenced, and a StatusPageVisibility that is hidden and unlinked.
type MaintenanceWindow struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	StartsAt string `json:"startsAt"`
	EndsAt   string `json:"endsAt"`

	// RepeatInterval, between 1 and 365, and RepeatUnit repeat the window
	// from its first occurrence. A zero RepeatInterval means no repeat.
	RepeatInterval int                         `json:"repeatInterval"`
	RepeatUnit     MaintenanceWindowRepeatUnit `json:"repeatUnit"`

	// RepeatEndsAt is the last calendar date, in Timezone, on which an
	// occurrence may start, preferably as YYYY-MM-DD.
	RepeatEndsAt string `json:"repeatEndsAt"`

	// Tags selects, by tag, the checks and groups the window applies to.
	Tags []string `json:"tags"`

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
	// there.
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
// Status pages with services (v1) are linked with ServiceIDs or
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
	// page with components, at most one entry per page. The API rejects an
	// update with nil ComponentConfigs and a non-empty StatusPageIDs unless
	// ServiceIDs or AffectAllServices is set; when one of them is, nil keeps
	// the stored entries of the status pages that stay linked. Send back the
	// entries a read returned to keep them.
	ComponentConfigs []MaintenanceWindowComponentConfig `json:"componentConfigs"`
}

// MaintenanceWindowComponentConfig lists the components of one status page
// that a maintenance window affects, in
// MaintenanceWindowStatusPageVisibility.ComponentConfigs.
type MaintenanceWindowComponentConfig struct {
	// StatusPageID is the ID of a status page listed in StatusPageIDs.
	StatusPageID string `json:"statusPageId"`
	// AffectAllComponents affects every component of the page, including
	// components added later, and requires Components to be empty.
	AffectAllComponents bool                               `json:"affectAllComponents"`
	Components          []MaintenanceWindowComponentImpact `json:"components,omitempty"`
}

// MaintenanceWindowComponentImpact is a status page component affected by a
// maintenance window, in MaintenanceWindowComponentConfig.Components, with
// the status it shows while the window is in progress.
type MaintenanceWindowComponentImpact struct {
	ComponentID string `json:"componentId"`
	// An empty Status means [StatusPageComponentStatusV3UnderMaintenance],
	// also on update.
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

func (c *client) GetMaintenanceWindow(
	ctx context.Context,
	id int64,
) (*MaintenanceWindow, error) {
	result := MaintenanceWindow{}
	err := c.callJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/maintenance-windows/%d", id), nil, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) UpdateMaintenanceWindow(
	ctx context.Context,
	id int64,
	mw MaintenanceWindow,
) (*MaintenanceWindow, error) {
	var result MaintenanceWindow
	err := c.callJSON(ctx, http.MethodPut, fmt.Sprintf("/v1/maintenance-windows/%d", id), mw, &result, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) DeleteMaintenanceWindow(
	ctx context.Context,
	id int64,
) error {
	_, err := c.call(ctx, http.MethodDelete, fmt.Sprintf("/v1/maintenance-windows/%d", id), nil, http.StatusNoContent)
	return err
}
