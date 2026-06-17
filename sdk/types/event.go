package types

import "time"

// EventStatus represents the lifecycle state of a MeshGuard event
type EventStatus string

const (
	EventStatusPending     EventStatus = "pending"
	EventStatusProcessing  EventStatus = "processing"
	EventStatusCompleted   EventStatus = "completed"
	EventStatusFailed      EventStatus = "failed"
	EventStatusReconciling EventStatus = "reconciling"
)

// MeshGuardEvent represents a transaction event in the queue
type MeshGuardEvent struct {
	ID          string      `json:"id"`
	Sequence    uint64      `json:"sequence"`
	Type        string      `json:"type"`
	Status      EventStatus `json:"status"`
	Payload     []byte      `json:"payload"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Retries     int         `json:"retries"`
	LastError   string      `json:"last_error,omitempty"`
}

// Transition updates the event status with state validation
func (e *MeshGuardEvent) Transition(newStatus EventStatus) bool {
	validTransitions := map[EventStatus][]EventStatus{
		EventStatusPending:     {EventStatusProcessing, EventStatusFailed},
		EventStatusProcessing:  {EventStatusCompleted, EventStatusFailed, EventStatusReconciling},
		EventStatusReconciling: {EventStatusCompleted, EventStatusFailed},
		EventStatusFailed:      {EventStatusPending, EventStatusReconciling},
	}

	allowed, exists := validTransitions[e.Status]
	if !exists {
		return false
	}

	for _, valid := range allowed {
		if valid == newStatus {
			e.Status = newStatus
			e.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}
