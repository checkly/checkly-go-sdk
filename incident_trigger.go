package checkly

// IncidentSeverity is the severity of an incident, in
// [IncidentTrigger].Severity and [MaintenanceWindowStatusPageVisibility].Severity.
type IncidentSeverity string

// Incident severities, the values of [IncidentTrigger].Severity and
// [MaintenanceWindowStatusPageVisibility].Severity.
const (
	IncidentSeverityMinor    IncidentSeverity = "MINOR"
	IncidentSeverityMedium   IncidentSeverity = "MEDIUM"
	IncidentSeverityMajor    IncidentSeverity = "MAJOR"
	IncidentSeverityCritical IncidentSeverity = "CRITICAL"
)

// IncidentTrigger makes a check or monitor open and resolve an incident on a
// status page service, based on its alert configuration. The API requires
// every field.
type IncidentTrigger struct {
	// ServiceID is the UUID of the status page service the incident is
	// opened on.
	ServiceID   string           `json:"serviceId"`
	Severity    IncidentSeverity `json:"severity"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	// NotifySubscribers notifies subscribers when the incident is opened.
	NotifySubscribers bool `json:"notifySubscribers"`
}
