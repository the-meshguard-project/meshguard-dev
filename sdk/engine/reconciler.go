package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/meshguard/sdk/queue"
	"github.com/meshguard/sdk/types"
)

// Reconciler handles event replay and settlement after network recovery
type Reconciler struct {
	store   queue.EventStore
	paused  bool
	mu      sync.RWMutex
}

// NewReconciler creates a reconciliation engine
func NewReconciler(store queue.EventStore) *Reconciler {
	return &Reconciler{
		store:  store,
		paused: false,
	}
}

// Pause stops event processing
func (r *Reconciler) Pause() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.paused = true
}

// Resume restarts event processing
func (r *Reconciler) Resume() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.paused = false
}

// IsPaused returns the current pause state
func (r *Reconciler) IsPaused() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.paused
}

// Reconcile processes all pending and failed events
func (r *Reconciler) Reconcile() (*types.ReconciliationSummary, error) {
	summary := &types.ReconciliationSummary{
		StartedAt: time.Now(),
	}

	// Get pending events
	pending, err := r.store.GetByStatus(types.EventStatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending events: %w", err)
	}

	// Get failed events
	failed, err := r.store.GetByStatus(types.EventStatusFailed)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch failed events: %w", err)
	}

	events := append(pending, failed...)
	summary.EventsChecked = len(events)

	for _, event := range events {
		if r.IsPaused() {
			break
		}

		if err := r.processEvent(event); err != nil {
			summary.EventsFailed++
			summary.Errors = append(summary.Errors, fmt.Sprintf("Event %s: %v", event.ID, err))
		} else {
			summary.EventsSettled++
		}
	}

	summary.CompletedAt = time.Now()
	return summary, nil
}

func (r *Reconciler) processEvent(event *types.MeshGuardEvent) error {
	// Transition to reconciling state
	if !event.Transition(types.EventStatusReconciling) {
		return fmt.Errorf("invalid state transition from %s", event.Status)
	}

	if err := r.store.Update(event); err != nil {
		return err
	}

	// TODO: Actual settlement logic here
	// For now, mark as completed
	if event.Transition(types.EventStatusCompleted) {
		return r.store.Update(event)
	}

	return fmt.Errorf("failed to complete event")
}
