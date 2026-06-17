package queue

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/meshguard/sdk/types"
)

// SQLiteStore implements EventStore using SQLite with WAL mode
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite-backed event store
func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.initialize(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *SQLiteStore) initialize() error {
	schema := `
	CREATE TABLE IF NOT EXISTS events (
		id TEXT PRIMARY KEY,
		sequence INTEGER NOT NULL,
		type TEXT NOT NULL,
		status TEXT NOT NULL,
		payload BLOB,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		retries INTEGER DEFAULT 0,
		last_error TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_status ON events(status);
	CREATE INDEX IF NOT EXISTS idx_sequence ON events(sequence);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *SQLiteStore) Save(event *types.MeshGuardEvent) error {
	query := `INSERT INTO events (id, sequence, type, status, payload, created_at, updated_at, retries, last_error)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, event.ID, event.Sequence, event.Type, event.Status,
		event.Payload, event.CreatedAt, event.UpdatedAt, event.Retries, event.LastError)
	return err
}

func (s *SQLiteStore) Get(id string) (*types.MeshGuardEvent, error) {
	query := `SELECT id, sequence, type, status, payload, created_at, updated_at, retries, last_error
	          FROM events WHERE id = ?`
	event := &types.MeshGuardEvent{}
	err := s.db.QueryRow(query, id).Scan(&event.ID, &event.Sequence, &event.Type, &event.Status,
		&event.Payload, &event.CreatedAt, &event.UpdatedAt, &event.Retries, &event.LastError)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("event not found: %s", id)
	}
	return event, err
}

func (s *SQLiteStore) GetByStatus(status types.EventStatus) ([]*types.MeshGuardEvent, error) {
	query := `SELECT id, sequence, type, status, payload, created_at, updated_at, retries, last_error
	          FROM events WHERE status = ? ORDER BY sequence ASC`
	rows, err := s.db.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*types.MeshGuardEvent
	for rows.Next() {
		event := &types.MeshGuardEvent{}
		if err := rows.Scan(&event.ID, &event.Sequence, &event.Type, &event.Status,
			&event.Payload, &event.CreatedAt, &event.UpdatedAt, &event.Retries, &event.LastError); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *SQLiteStore) Update(event *types.MeshGuardEvent) error {
	query := `UPDATE events SET status = ?, updated_at = ?, retries = ?, last_error = ? WHERE id = ?`
	_, err := s.db.Exec(query, event.Status, event.UpdatedAt, event.Retries, event.LastError, event.ID)
	return err
}

func (s *SQLiteStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM events WHERE id = ?", id)
	return err
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
