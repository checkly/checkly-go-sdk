package checkly

// EscalationType is when an alert escalates, in
// [AlertSettings].EscalationType.
type EscalationType string

// Escalation types, the values of [AlertSettings].EscalationType.
const (
	// EscalationTypeRunBased alerts after a number of consecutive failed
	// runs, set in RunBasedEscalation. It is the API default.
	EscalationTypeRunBased EscalationType = "RUN_BASED"
	// EscalationTypeTimeBased alerts once a check has been failing for a
	// number of minutes, set in TimeBasedEscalation.
	EscalationTypeTimeBased EscalationType = "TIME_BASED"
)

// AlertSettings represents an alert configuration.
type AlertSettings struct {
	EscalationType              EscalationType              `json:"escalationType,omitempty"`
	RunBasedEscalation          RunBasedEscalation          `json:"runBasedEscalation,omitempty"`
	TimeBasedEscalation         TimeBasedEscalation         `json:"timeBasedEscalation,omitempty"`
	Reminders                   Reminders                   `json:"reminders,omitempty"`
	ParallelRunFailureThreshold ParallelRunFailureThreshold `json:"parallelRunFailureThreshold,omitempty"`
}

// ParallelRunFailureThreshold represent an alert escalation based on the number
// of failing regions, only applicable for parallel checks
type ParallelRunFailureThreshold struct {
	Enabled    bool `json:"enabled"`
	Percentage int  `json:"percentage,omitempty"`
}

// RunBasedEscalation represents an alert escalation based on a number of failed
// check runs.
type RunBasedEscalation struct {
	FailedRunThreshold int `json:"failedRunThreshold,omitempty"`
}

// TimeBasedEscalation represents an alert escalation based on the number of
// minutes after a check first starts failing.
type TimeBasedEscalation struct {
	MinutesFailingThreshold int `json:"minutesFailingThreshold,omitempty"`
}

// Reminders represents the number of reminders to send after an alert
// notification, and the time interval between them.
type Reminders struct {
	Amount   int `json:"amount"`
	Interval int `json:"interval,omitempty"`
}
