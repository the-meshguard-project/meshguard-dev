package queue

import "github.com/meshguard/sdk/types"

// EventStore defines the interface for event persistence
type EventStore interface {
	// Save persists an event to the queue
	Save(event *types.MeshGuardEvent) error
	
	// Get retrieves an event by ID
	Get(id string) (*types.MeshGuardEvent, error)
	
	// GetByStatus retrieves all events with a specific status
	GetByStatus(status types.EventStatus) ([]*types.MeshGuardEvent, error)
	
	// Update modifies an existing event
	Update(event *types.MeshGuardEvent) error
	
	// Delete removes an event from the queue
	Delete(id string) error
	
	// Close releases any resources
	Close() error
}
