package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/meshguard/sdk/sdk/engine"
	"github.com/meshguard/sdk/sdk/queue"
	"github.com/meshguard/sdk/sdk/types"
)

func main() {
	fmt.Println("=== MeshGuard SDK Regtest Demo ===\n")

	// Initialize SQLite store
	dbPath := "meshguard_regtest.db"
	store, err := queue.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()
	defer os.Remove(dbPath)
	defer os.Remove(dbPath + "-shm")
	defer os.Remove(dbPath + "-wal")

	// Initialize clock and reconciler
	clock := types.NewClock(0)
	reconciler := engine.NewReconciler(store)

	fmt.Println("✓ Initialized SQLite store and reconciler\n")

	// Simulate network partition scenario
	fmt.Println("--- Scenario: Network Partition Detected ---")
	fmt.Println("Simulating 5 Bitcoin transactions queued during partition...\n")

	// Queue events during "network partition"
	events := []*types.MeshGuardEvent{
		{
			ID:        "btc-tx-001",
			Sequence:  clock.Next(),
			Type:      "bitcoin_payment",
			Status:    types.EventStatusPending,
			Payload:   []byte(`{"txid": "abc123", "amount": 0.001, "to": "tb1qtest1"}`),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "btc-tx-002",
			Sequence:  clock.Next(),
			Type:      "bitcoin_payment",
			Status:    types.EventStatusPending,
			Payload:   []byte(`{"txid": "def456", "amount": 0.002, "to": "tb1qtest2"}`),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "btc-tx-003",
			Sequence:  clock.Next(),
			Type:      "bitcoin_payment",
			Status:    types.EventStatusPending,
			Payload:   []byte(`{"txid": "ghi789", "amount": 0.003, "to": "tb1qtest3"}`),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "btc-tx-004",
			Sequence:  clock.Next(),
			Type:      "bitcoin_payment",
			Status:    types.EventStatusFailed,
			Payload:   []byte(`{"txid": "jkl012", "amount": 0.004, "to": "tb1qtest4"}`),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			LastError: "insufficient funds",
			Retries:   1,
		},
		{
			ID:        "btc-tx-005",
			Sequence:  clock.Next(),
			Type:      "bitcoin_payment",
			Status:    types.EventStatusPending,
			Payload:   []byte(`{"txid": "mno345", "amount": 0.005, "to": "tb1qtest5"}`),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, evt := range events {
		if err := store.Save(evt); err != nil {
			log.Fatalf("Failed to queue event: %v", err)
		}
		fmt.Printf("  Queued: %s (seq: %d, status: %s)\n", evt.ID, evt.Sequence, evt.Status)
	}

	fmt.Println("\n✓ All transactions queued in local SQLite WAL\n")

	// Show queue state
	pending, _ := store.GetByStatus(types.EventStatusPending)
	failed, _ := store.GetByStatus(types.EventStatusFailed)
	fmt.Printf("Queue Status:\n")
	fmt.Printf("  Pending: %d events\n", len(pending))
	fmt.Printf("  Failed: %d events\n\n", len(failed))

	// Simulate network recovery
	fmt.Println("--- Network Recovered ---")
	fmt.Println("Starting reconciliation engine...\n")
	time.Sleep(500 * time.Millisecond)

	// Run reconciliation
	summary, err := reconciler.Reconcile()
	if err != nil {
		log.Fatalf("Reconciliation failed: %v", err)
	}

	// Display results
	fmt.Println("--- Reconciliation Summary ---")
	fmt.Printf("Started:  %s\n", summary.StartedAt.Format("15:04:05"))
	fmt.Printf("Finished: %s\n", summary.CompletedAt.Format("15:04:05"))
	fmt.Printf("Duration: %s\n\n", summary.CompletedAt.Sub(summary.StartedAt))
	fmt.Printf("Events Checked:  %d\n", summary.EventsChecked)
	fmt.Printf("Events Settled:  %d ✓\n", summary.EventsSettled)
	fmt.Printf("Events Failed:   %d ✗\n\n", summary.EventsFailed)

	if len(summary.Errors) > 0 {
		fmt.Println("Errors:")
		for _, err := range summary.Errors {
			fmt.Printf("  - %s\n", err)
		}
		fmt.Println()
	}

	// Verify final state
	completed, _ := store.GetByStatus(types.EventStatusCompleted)
	stillPending, _ := store.GetByStatus(types.EventStatusPending)
	stillFailed, _ := store.GetByStatus(types.EventStatusFailed)

	fmt.Println("--- Final Queue State ---")
	fmt.Printf("Completed: %d\n", len(completed))
	fmt.Printf("Pending:   %d\n", len(stillPending))
	fmt.Printf("Failed:    %d\n\n", len(stillFailed))

	// Show completed events
	if len(completed) > 0 {
		fmt.Println("Completed Transactions:")
		for _, evt := range completed {
			fmt.Printf("  %s (seq: %d) - %s\n", evt.ID, evt.Sequence, evt.Status)
		}
	}

	fmt.Println("\n✓ SDK Demo Complete")
	fmt.Println("\nNote: In production, this would integrate with Bitcoin Core RPC")
	fmt.Println("      to broadcast queued transactions after network recovery.")
}
