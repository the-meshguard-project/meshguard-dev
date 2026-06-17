package queue

import (
	"os"
	"testing"
	"time"

	"github.com/meshguard/sdk/sdk/types"
)

func TestSQLiteStoreIntegration(t *testing.T) {
	dbPath := "test_integration.db"
	defer os.Remove(dbPath)
	defer os.Remove(dbPath + "-shm")
	defer os.Remove(dbPath + "-wal")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	event := &types.MeshGuardEvent{
		ID:        "evt-integration-001",
		Sequence:  1,
		Type:      "payment",
		Status:    types.EventStatusPending,
		Payload:   []byte(`{"amount": 1000, "currency": "BTC"}`),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Retries:   0,
	}

	// Test Save
	if err := store.Save(event); err != nil {
		t.Fatalf("Failed to save event: %v", err)
	}

	// Test Get
	retrieved, err := store.Get("evt-integration-001")
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
	if string(retrieved.Payload) != string(event.Payload) {
		t.Errorf("Payload mismatch: got %s, want %s", retrieved.Payload, event.Payload)
	}

	// Test Update
	retrieved.Status = types.EventStatusProcessing
	retrieved.UpdatedAt = time.Now()
	retrieved.Retries = 1
	if err := store.Update(retrieved); err != nil {
		t.Fatalf("Failed to update event: %v", err)
	}

	updated, err := store.Get("evt-integration-001")
	if err != nil {
		t.Fatalf("Failed to get updated event: %v", err)
	}
	if updated.Status != types.EventStatusProcessing {
		t.Errorf("Status not updated: got %s, want %s", updated.Status, types.EventStatusProcessing)
	}
	if updated.Retries != 1 {
		t.Errorf("Retries not updated: got %d, want 1", updated.Retries)
	}

	// Test Delete
	if err := store.Delete("evt-integration-001"); err != nil {
		t.Fatalf("Failed to delete event: %v", err)
	}

	_, err = store.Get("evt-integration-001")
	if err == nil {
		t.Error("Expected error when getting deleted event")
	}
}

func TestSQLiteGetByStatusIntegration(t *testing.T) {
	dbPath := "test_status_integration.db"
	defer os.Remove(dbPath)
	defer os.Remove(dbPath + "-shm")
	defer os.Remove(dbPath + "-wal")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

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

	// Verify ordering by sequence
	if len(pending) == 2 && pending[0].Sequence > pending[1].Sequence {
		t.Error("Events not ordered by sequence")
	}
}
