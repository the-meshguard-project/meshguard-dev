package types

import (
	"testing"
	"time"
)

func TestEventTransition(t *testing.T) {
	tests := []struct {
		name        string
		from        EventStatus
		to          EventStatus
		shouldAllow bool
	}{
		{"pending to processing", EventStatusPending, EventStatusProcessing, true},
		{"pending to failed", EventStatusPending, EventStatusFailed, true},
		{"pending to completed", EventStatusPending, EventStatusCompleted, false},
		{"processing to completed", EventStatusProcessing, EventStatusCompleted, true},
		{"processing to failed", EventStatusProcessing, EventStatusFailed, true},
		{"processing to reconciling", EventStatusProcessing, EventStatusReconciling, true},
		{"failed to pending", EventStatusFailed, EventStatusPending, true},
		{"failed to reconciling", EventStatusFailed, EventStatusReconciling, true},
		{"reconciling to completed", EventStatusReconciling, EventStatusCompleted, true},
		{"reconciling to failed", EventStatusReconciling, EventStatusFailed, true},
		{"completed to any", EventStatusCompleted, EventStatusPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &MeshGuardEvent{
				ID:        "test-1",
				Status:    tt.from,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			oldUpdateTime := event.UpdatedAt
			time.Sleep(1 * time.Millisecond)

			result := event.Transition(tt.to)

			if result != tt.shouldAllow {
				t.Errorf("Transition from %s to %s: expected %v, got %v",
					tt.from, tt.to, tt.shouldAllow, result)
			}

			if result && event.Status != tt.to {
				t.Errorf("Status not updated: expected %s, got %s", tt.to, event.Status)
			}

			if result && !event.UpdatedAt.After(oldUpdateTime) {
				t.Error("UpdatedAt was not refreshed")
			}

			if !result && event.Status != tt.from {
				t.Errorf("Status should remain %s but changed to %s", tt.from, event.Status)
			}
		})
	}
}
