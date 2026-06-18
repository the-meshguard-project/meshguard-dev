package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/meshguard/sdk/sdk/engine"
	"github.com/meshguard/sdk/sdk/queue"
	"github.com/meshguard/sdk/sdk/types"
)

// BitcoinRPC simple RPC client for Bitcoin Core
type BitcoinRPC struct {
	URL      string
	User     string
	Password string
}

// RPCRequest represents a Bitcoin RPC request
type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// RPCResponse represents a Bitcoin RPC response
type RPCResponse struct {
	Result interface{} `json:"result"`
	Error  *RPCError   `json:"error"`
	ID     string      `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Call makes an RPC call to Bitcoin Core
func (b *BitcoinRPC) Call(method string, params ...interface{}) (interface{}, error) {
	reqBody := RPCRequest{
		JSONRPC: "1.0",
		ID:      "meshguard",
		Method:  method,
		Params:  params,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", b.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(b.User, b.Password)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bitcoin core not reachable: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rpcResp RPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, err
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

func main() {
	fmt.Println("=== MeshGuard SDK - Bitcoin Core Regtest Integration ===\n")

	// Bitcoin Core connection (update these for your setup)
	btc := &BitcoinRPC{
		URL:      "http://127.0.0.1:18443", // regtest default port
		User:     "user",                    // your rpcuser
		Password: "password",                // your rpcpassword
	}

	// Test connection
	fmt.Println("Testing Bitcoin Core connection...")
	info, err := btc.Call("getblockchaininfo")
	if err != nil {
		fmt.Printf("⚠ Bitcoin Core not available: %v\n", err)
		fmt.Println("To run this demo:")
		fmt.Println("1. Start Bitcoin Core in regtest mode:")
		fmt.Println("   bitcoind -regtest -daemon")
		fmt.Println("2. Update RPC credentials in this file")
		fmt.Println("\nRunning offline simulation instead...\n")
		runOfflineSimulation()
		return
	}

	infoMap := info.(map[string]interface{})
	fmt.Printf("✓ Connected to Bitcoin Core (chain: %s, blocks: %.0f)\n\n",
		infoMap["chain"], infoMap["blocks"])

	// Initialize SDK
	dbPath := "meshguard_btc_regtest.db"
	store, err := queue.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()
	defer os.Remove(dbPath)
	defer os.Remove(dbPath + "-shm")
	defer os.Remove(dbPath + "-wal")

	clock := types.NewClock(0)
	reconciler := engine.NewReconciler(store)

	// Get or create a wallet address
	fmt.Println("Getting wallet address...")
	addr, err := btc.Call("getnewaddress")
	if err != nil {
		log.Printf("Warning: %v\n", err)
		addr = "tb1qexampleaddress"
	}
	fmt.Printf("Address: %s\n\n", addr)

	// Simulate network partition
	fmt.Println("--- Simulating Network Partition ---")
	fmt.Println("Queuing Bitcoin transactions...\n")

	// Create test events
	events := []*types.MeshGuardEvent{
		{
			ID:        "btc-payment-1",
			Sequence:  clock.Next(),
			Type:      "bitcoin_sendtoaddress",
			Status:    types.EventStatusPending,
			Payload:   []byte(fmt.Sprintf(`{"address": "%s", "amount": 0.001}`, addr)),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "btc-payment-2",
			Sequence:  clock.Next(),
			Type:      "bitcoin_sendtoaddress",
			Status:    types.EventStatusPending,
			Payload:   []byte(fmt.Sprintf(`{"address": "%s", "amount": 0.002}`, addr)),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, evt := range events {
		if err := store.Save(evt); err != nil {
			log.Fatalf("Failed to save event: %v", err)
		}
		fmt.Printf("  Queued: %s (seq: %d)\n", evt.ID, evt.Sequence)
	}

	fmt.Println("\n✓ Transactions queued during partition\n")

	// Simulate recovery
	fmt.Println("--- Network Recovered ---")
	fmt.Println("Running reconciliation...\n")

	summary, err := reconciler.Reconcile()
	if err != nil {
		log.Fatalf("Reconciliation failed: %v", err)
	}

	// Display results
	fmt.Println("--- Reconciliation Summary ---")
	fmt.Printf("Duration:        %s\n", summary.CompletedAt.Sub(summary.StartedAt))
	fmt.Printf("Events Checked:  %d\n", summary.EventsChecked)
	fmt.Printf("Events Settled:  %d\n", summary.EventsSettled)
	fmt.Printf("Events Failed:   %d\n\n", summary.EventsFailed)

	completed, _ := store.GetByStatus(types.EventStatusCompleted)
	fmt.Printf("✓ Successfully processed %d transactions\n", len(completed))

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("The SDK successfully queued and reconciled Bitcoin transactions")
}

func runOfflineSimulation() {
	fmt.Println("--- Running Offline Simulation ---\n")

	dbPath := "meshguard_offline.db"
	store, _ := queue.NewSQLiteStore(dbPath)
	defer store.Close()
	defer os.Remove(dbPath)
	defer os.Remove(dbPath + "-shm")
	defer os.Remove(dbPath + "-wal")

	clock := types.NewClock(0)
	reconciler := engine.NewReconciler(store)

	// Queue events
	for i := 1; i <= 3; i++ {
		evt := &types.MeshGuardEvent{
			ID:        fmt.Sprintf("offline-tx-%d", i),
			Sequence:  clock.Next(),
			Type:      "bitcoin_payment",
			Status:    types.EventStatusPending,
			Payload:   []byte(fmt.Sprintf(`{"txid": "sim%d", "amount": 0.00%d}`, i, i)),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		store.Save(evt)
		fmt.Printf("  Queued: %s\n", evt.ID)
	}

	fmt.Println("\n✓ Events queued in SQLite\n")

	// Reconcile
	summary, _ := reconciler.Reconcile()
	fmt.Printf("Reconciliation: %d events settled\n", summary.EventsSettled)
	fmt.Println("\n✓ Offline simulation complete")
}
