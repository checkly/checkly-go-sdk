package checkly

type IncidentSeverity string

const (
	IncidentSeverityMinor    IncidentSeverity = "MINOR"
	IncidentSeverityMedium   IncidentSeverity = "MEDIUM"
	IncidentSeverityMajor    IncidentSeverity = "MAJOR"
	IncidentSeverityCritical IncidentSeverity = "CRITICAL"
)

type IncidentTrigger struct {
	// ServiceID is the identifier of the status page service that the
	// incident will be associated with.
	ServiceID string `json:"serviceId"`

	// Severity is the severity level of the incident.
	Severity IncidentSeverity `json:"severity"`

	// Name is the name of the incident.
	Name string `json:"name"`

	// Description is a detailed description of the incident.
	Description string `json:"description"`

	// NotifySubscribers determines whether to notify subscribers when the
	// incident is triggered.
	NotifySubscribers bool `json:"notifySubscribers"`
}
