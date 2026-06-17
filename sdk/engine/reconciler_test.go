package engine

import (
	"testing"
	"time"

	"github.com/meshguard/sdk/sdk/queue"
	"github.com/meshguard/sdk/sdk/types"
)

type mockStore struct {
	events map[string]*types.MeshGuardEvent
}

func newMockStore() *mockStore {
	return &mockStore{events: make(map[string]*types.MeshGuardEvent)}
}

func (m *mockStore) Save(event *types.MeshGuardEvent) error {
	m.events[event.ID] = event
	return nil
}

func (m *mockStore) Get(id string) (*types.MeshGuardEvent, error) {
	return m.events[id], nil
}

func (m *mockStore) GetByStatus(status types.EventStatus) ([]*types.MeshGuardEvent, error) {
	var result []*types.MeshGuardEvent
	for _, evt := range m.events {
		if evt.Status == status {
			result = append(result, evt)
		}
	}
	return result, nil
}

func (m *mockStore) Update(event *types.MeshGuardEvent) error {
	m.events[event.ID] = event
	return nil
}

func (m *mockStore) Delete(id string) error {
	delete(m.events, id)
	return nil
}

func (m *mockStore) Close() error {
	return nil
}

func TestReconcilerPauseResume(t *testing.T) {
	store := newMockStore()
	reconciler := NewReconciler(store)

	if reconciler.IsPaused() {
		t.Error("Reconciler should not be paused initially")
	}

	reconciler.Pause()
	if !reconciler.IsPaused() {
		t.Error("Reconciler should be paused after Pause()")
	}

	reconciler.Resume()
	if reconciler.IsPaused() {
		t.Error("Reconciler should not be paused after Resume()")
	}
}

func TestReconcileWithPendingEvents(t *testing.T) {
	store := newMockStore()
	reconciler := NewReconciler(store)

	// Add pending events
	events := []*types.MeshGuardEvent{
		{ID: "e1", Status: types.EventStatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "e2", Status: types.EventStatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, evt := range events {
		store.Save(evt)
	}

	summary, err := reconciler.Reconcile()
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}

	if summary.EventsChecked != 2 {
		t.Errorf("Expected 2 events checked, got %d", summary.EventsChecked)
	}

	if summary.EventsSettled != 2 {
		t.Errorf("Expected 2 events settled, got %d", summary.EventsSettled)
	}

	if summary.EventsFailed != 0 {
		t.Errorf("Expected 0 events failed, got %d", summary.EventsFailed)
	}

	// Verify events are completed
	for _, id := range []string{"e1", "e2"} {
		evt, _ := store.Get(id)
		if evt.Status != types.EventStatusCompleted {
			t.Errorf("Event %s status: got %s, want %s", id, evt.Status, types.EventStatusCompleted)
		}
	}
}

func TestReconcileWithFailedEvents(t *testing.T) {
	store := newMockStore()
	reconciler := NewReconciler(store)

	events := []*types.MeshGuardEvent{
		{ID: "f1", Status: types.EventStatusFailed, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "f2", Status: types.EventStatusFailed, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, evt := range events {
		store.Save(evt)
	}

	summary, err := reconciler.Reconcile()
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}

	if summary.EventsChecked != 2 {
		t.Errorf("Expected 2 events checked, got %d", summary.EventsChecked)
	}

	if summary.EventsSettled != 2 {
		t.Errorf("Expected 2 events settled, got %d", summary.EventsSettled)
	}
}

func TestReconcilePausedMidway(t *testing.T) {
	store := newMockStore()
	reconciler := NewReconciler(store)

	// Add events
	for i := 1; i <= 5; i++ {
		store.Save(&types.MeshGuardEvent{
			ID:        string(rune('a' + i)),
			Status:    types.EventStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	// Pause immediately
	reconciler.Pause()

	summary, err := reconciler.Reconcile()
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}

	// Should process 0 events when paused
	if summary.EventsSettled > 0 {
		t.Errorf("Expected 0 events settled when paused, got %d", summary.EventsSettled)
	}
}
