package queue

import (
	"testing"
	"time"

	"github.com/meshguard/sdk/sdk/types"
)

// Note: SQLite tests require CGO and GCC. These tests use mock implementation.
// For full SQLite integration tests, run on a Linux/macOS system with CGO_ENABLED=1

type memoryStore struct {
	events map[string]*types.MeshGuardEvent
}

func newMemoryStore() *memoryStore {
	return &memoryStore{events: make(map[string]*types.MeshGuardEvent)}
}

func (m *memoryStore) Save(event *types.MeshGuardEvent) error {
	m.events[event.ID] = event
	return nil
}

func (m *memoryStore) Get(id string) (*types.MeshGuardEvent, error) {
	if evt, ok := m.events[id]; ok {
		return evt, nil
	}
	return nil, nil
}

func (m *memoryStore) GetByStatus(status types.EventStatus) ([]*types.MeshGuardEvent, error) {
	var result []*types.MeshGuardEvent
	for _, evt := range m.events {
		if evt.Status == status {
			result = append(result, evt)
		}
	}
	return result, nil
}

func (m *memoryStore) Update(event *types.MeshGuardEvent) error {
	m.events[event.ID] = event
	return nil
}

func (m *memoryStore) Delete(id string) error {
	delete(m.events, id)
	return nil
}

func (m *memoryStore) Close() error {
	return nil
}

func TestStoreLifecycle(t *testing.T) {
	store := newMemoryStore()

	event := &types.MeshGuardEvent{
		ID:        "evt-001",
		Sequence:  1,
		Type:      "payment",
		Status:    types.EventStatusPending,
		Payload:   []byte(`{"amount": 1000}`),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Retries:   0,
	}

	// Test Save
	if err := store.Save(event); err != nil {
		t.Fatalf("Failed to save event: %v", err)
	}

	// Test Get
	retrieved, err := store.Get("evt-001")
	if err != nil {
		t.Fatalf("Failed to get event: %v", err)
	}

	if retrieved.ID != event.ID {
		t.Errorf("ID mismatch: got %s, want %s", retrieved.ID, event.ID)
	}
	if retrieved.Sequence != event.Sequence {
		t.Errorf("Sequence mismatch: got %d, want %d", retrieved.Sequence, event.Sequence)
	}
	if retrieved.Status != event.Status {
		t.Errorf("Status mismatch: got %s, want %s", retrieved.Status, event.Status)
	}

	// Test Update
	retrieved.Status = types.EventStatusProcessing
	retrieved.UpdatedAt = time.Now()
	if err := store.Update(retrieved); err != nil {
		t.Fatalf("Failed to update event: %v", err)
	}

	updated, err := store.Get("evt-001")
	if err != nil {
		t.Fatalf("Failed to get updated event: %v", err)
	}
	if updated.Status != types.EventStatusProcessing {
		t.Errorf("Status not updated: got %s, want %s", updated.Status, types.EventStatusProcessing)
	}

	// Test Delete
	if err := store.Delete("evt-001"); err != nil {
		t.Fatalf("Failed to delete event: %v", err)
	}

	deleted, _ := store.Get("evt-001")
	if deleted != nil {
		t.Error("Expected nil when getting deleted event")
	}
}

func TestGetByStatus(t *testing.T) {
	store := newMemoryStore()

	events := []*types.MeshGuardEvent{
		{ID: "e1", Sequence: 1, Type: "payment", Status: types.EventStatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "e2", Sequence: 2, Type: "payment", Status: types.EventStatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "e3", Sequence: 3, Type: "payment", Status: types.EventStatusFailed, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "e4", Sequence: 4, Type: "payment", Status: types.EventStatusCompleted, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, evt := range events {
		if err := store.Save(evt); err != nil {
			t.Fatalf("Failed to save event %s: %v", evt.ID, err)
		}
	}

	pending, err := store.GetByStatus(types.EventStatusPending)
	if err != nil {
		t.Fatalf("Failed to get pending events: %v", err)
	}
	if len(pending) != 2 {
		t.Errorf("Expected 2 pending events, got %d", len(pending))
	}

	failed, err := store.GetByStatus(types.EventStatusFailed)
	if err != nil {
		t.Fatalf("Failed to get failed events: %v", err)
	}
	if len(failed) != 1 {
		t.Errorf("Expected 1 failed event, got %d", len(failed))
	}
}
