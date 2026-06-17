package types

import "time"

// ReconciliationSummary captures the result of a reconciliation pass
type ReconciliationSummary struct {
	StartedAt     time.Time `json:"started_at"`
	CompletedAt   time.Time `json:"completed_at"`
	EventsChecked int       `json:"events_checked"`
	EventsSettled int       `json:"events_settled"`
	EventsFailed  int       `json:"events_failed"`
	Errors        []string  `json:"errors,omitempty"`
}
