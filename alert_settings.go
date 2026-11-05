package checkly

// EscalationType is when an alert escalates, in
// [AlertSettings].EscalationType.
type EscalationType string

// Escalation types, the values of [AlertSettings].EscalationType.
const (
	// EscalationTypeRunBased alerts after a number of consecutive failed
	// runs, set in RunBasedEscalation. It is the API default on create.
	EscalationTypeRunBased EscalationType = "RUN_BASED"
	// EscalationTypeTimeBased alerts once a check has been failing for a
	// number of minutes, set in TimeBasedEscalation.
	EscalationTypeTimeBased EscalationType = "TIME_BASED"
)

// AlertSettings sets when a check, monitor or group sends alerts and
// reminders.
type AlertSettings struct {
	// EscalationType is left out of the request when empty.
	EscalationType              EscalationType              `json:"escalationType,omitempty"`
	RunBasedEscalation          RunBasedEscalation          `json:"runBasedEscalation,omitempty"`
	TimeBasedEscalation         TimeBasedEscalation         `json:"timeBasedEscalation,omitempty"`
	Reminders                   Reminders                   `json:"reminders,omitempty"`
	ParallelRunFailureThreshold ParallelRunFailureThreshold `json:"parallelRunFailureThreshold,omitempty"`
}

// ParallelRunFailureThreshold, when enabled, makes a check that runs in
// parallel alert only when a percentage of its parallel runs fail.
type ParallelRunFailureThreshold struct {
	Enabled bool `json:"enabled"`
	// Percentage is 10, 20, 30, 40, 50, 60, 70, 80, 90 or 100.
	// Percentage is left out of the request when zero.
	Percentage int `json:"percentage,omitempty"`
}

// RunBasedEscalation holds the threshold of the [EscalationTypeRunBased]
// escalation type.
type RunBasedEscalation struct {
	// FailedRunThreshold is the number of consecutive failed runs after
	// which an alert is sent, 1 to 5.
	// FailedRunThreshold is left out of the request when zero.
	FailedRunThreshold int `json:"failedRunThreshold,omitempty"`
}

// TimeBasedEscalation holds the threshold of the [EscalationTypeTimeBased]
// escalation type.
type TimeBasedEscalation struct {
	// MinutesFailingThreshold is the number of minutes after a check starts
	// failing that an alert is sent: 5, 10, 15 or 30.
	// MinutesFailingThreshold is left out of the request when zero.
	MinutesFailingThreshold int `json:"minutesFailingThreshold,omitempty"`
}

// Reminders sets the reminders sent after the first alert notification.
type Reminders struct {
	// Amount is the number of reminders: 0 to 5, or 100000.
	Amount int `json:"amount"`
	// Interval is the interval between reminders: 1, 2, 3, 4, 5, 10, 15 or
	// 30. Interval is left out of the request when zero.
	Interval int `json:"interval,omitempty"`
}
