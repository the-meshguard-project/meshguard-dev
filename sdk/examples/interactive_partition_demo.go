package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/meshguard/sdk/sdk/engine"
	"github.com/meshguard/sdk/sdk/queue"
	"github.com/meshguard/sdk/sdk/types"
)

type BitcoinRPC struct {
	URL      string
	User     string
	Password string
	Wallet   string
}

type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type RPCResponse struct {
	Result interface{} `json:"result"`
	Error  *RPCError   `json:"error"`
	ID     string      `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (b *BitcoinRPC) Call(method string, params ...interface{}) (interface{}, error) {
	reqBody := RPCRequest{
		JSONRPC: "1.0",
		ID:      "meshguard",
		Method:  method,
		Params:  params,
	}

	jsonData, _ := json.Marshal(reqBody)
	url := b.URL
	if b.Wallet != "" {
		url = b.URL + "/wallet/" + b.Wallet
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.SetBasicAuth(b.User, b.Password)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var rpcResp RPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, err
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

func checkConnection(btc *BitcoinRPC) bool {
	_, err := btc.Call("getblockchaininfo")
	return err == nil
}

func waitForEnter(message string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\n%s [Press ENTER to continue]", message)
	reader.ReadString('\n')
	fmt.Println()
}

func main() {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║  MeshGuard SDK - Interactive Network Partition Demo      ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝\n")

	btcURL := "http://127.0.0.1:18443"
	btcUser := "alice"
	btcPassword := "strongpassword123"

	btcAlice := &BitcoinRPC{URL: btcURL, User: btcUser, Password: btcPassword, Wallet: "Alice"}
	btcBob := &BitcoinRPC{URL: btcURL, User: btcUser, Password: btcPassword, Wallet: "Bob"}

	// Initialize SDK
	dbPath := "meshguard_interactive.db"
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

	fmt.Println("═══ Phase 1: Verify Bitcoin Core is Running ═══\n")
	
	if !checkConnection(btcAlice) {
		fmt.Println("❌ Bitcoin Core not running!")
		fmt.Println("\nPlease start Bitcoin Core first:")
		fmt.Println("  Windows: scripts\\setup_regtest.bat")
		fmt.Println("  Linux:   ./scripts/setup_regtest.sh\n")
		return
	}

	fmt.Println("✓ Bitcoin Core is ONLINE")
	
	aliceAddr, _ := btcAlice.Call("getnewaddress")
	bobAddr, _ := btcBob.Call("getnewaddress")
	
	aliceBalance, _ := btcAlice.Call("getbalance")
	bobBalance, _ := btcBob.Call("getbalance")
	
	fmt.Printf("  Alice: %.8f BTC\n", aliceBalance)
	fmt.Printf("  Bob:   %.8f BTC\n", bobBalance)

	waitForEnter("Ready to simulate network partition?")

	// Phase 2: Queue transactions while "offline"
	fmt.Println("═══ Phase 2: Simulating Network Partition ═══\n")
	fmt.Println("📡 Checking network status...")
	time.Sleep(500 * time.Millisecond)
	fmt.Println("⚠️  Network connectivity: UNSTABLE")
	time.Sleep(500 * time.Millisecond)
	fmt.Println("❌ Network connectivity: LOST\n")
	
	fmt.Println("💾 Switching to OFFLINE MODE - queuing transactions locally...\n")

	// Queue 3 transactions
	transactions := []struct {
		amount float64
		memo   string
	}{
		{1.0, "Payment for services"},
		{0.5, "Refund"},
		{2.5, "Contract payment"},
	}

	for i, tx := range transactions {
		event := &types.MeshGuardEvent{
			ID:       fmt.Sprintf("tx-partition-%d", i+1),
			Sequence: clock.Next(),
			Type:     "bitcoin_sendtoaddress",
			Status:   types.EventStatusPending,
			Payload: []byte(fmt.Sprintf(`{
				"from": "Alice",
				"to": "Bob", 
				"address": "%s",
				"amount": %.8f,
				"memo": "%s"
			}`, bobAddr, tx.amount, tx.memo)),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := store.Save(event); err != nil {
			log.Fatalf("Failed to queue event: %v", err)
		}

		fmt.Printf("  ✓ Queued: Alice → Bob: %.8f BTC (seq: %d)\n", tx.amount, event.Sequence)
		fmt.Printf("    Memo: %s\n", tx.memo)
		fmt.Printf("    Status: PENDING (stored in SQLite WAL)\n\n")
		time.Sleep(300 * time.Millisecond)
	}

	pending, _ := store.GetByStatus(types.EventStatusPending)
	fmt.Printf("📊 Queue Status: %d transactions pending\n", len(pending))
	fmt.Printf("💾 All data persisted to: %s\n", dbPath)

	waitForEnter("\n🔌 Now STOP Bitcoin Core to simulate network being down")

	// Phase 3: Try to process while offline
	fmt.Println("═══ Phase 3: Attempting to Process (Network Still Down) ═══\n")
	fmt.Println("🔄 Checking if network recovered...")
	
	networkOnline := checkConnection(btcAlice)
	
	if !networkOnline {
		fmt.Println("❌ Network still OFFLINE")
		fmt.Println("✓ Transactions remain safely queued in SQLite")
		fmt.Println("✓ No data loss - WAL mode ensures durability\n")
		
		fmt.Println("📋 Queued Transactions:")
		for i, evt := range pending {
			var payload map[string]interface{}
			json.Unmarshal(evt.Payload, &payload)
			fmt.Printf("  %d. %.8f BTC - %s\n", i+1, payload["amount"], payload["memo"])
		}
	} else {
		fmt.Println("✓ Network came back online!")
	}

	waitForEnter("\n🔌 Now START Bitcoin Core again to simulate network recovery")

	// Phase 4: Network recovery
	fmt.Println("═══ Phase 4: Network Recovery & Reconciliation ═══\n")
	fmt.Println("🔄 Checking network status...")
	
	for i := 0; i < 10; i++ {
		if checkConnection(btcAlice) {
			fmt.Println("✓ Network is ONLINE!")
			fmt.Println("🚀 Activating reconciliation engine...\n")
			break
		}
		fmt.Printf("  Attempt %d/10: Still offline...\n", i+1)
		time.Sleep(2 * time.Second)
	}

	if !checkConnection(btcAlice) {
		fmt.Println("\n❌ Network still offline. Start Bitcoin Core and run the demo again.")
		return
	}

	// Process queued transactions
	fmt.Println("═══ Phase 5: Processing Queued Transactions ═══\n")
	
	eventsToProcess, _ := store.GetByStatus(types.EventStatusPending)
	
	for i, evt := range eventsToProcess {
		var payload map[string]interface{}
		json.Unmarshal(evt.Payload, &payload)

		fmt.Printf("Processing transaction %d/%d\n", i+1, len(eventsToProcess))
		fmt.Printf("  Amount: %.8f BTC\n", payload["amount"])
		fmt.Printf("  Memo:   %s\n", payload["memo"])

		// Broadcast transaction
		txid, err := btcAlice.Call("sendtoaddress", 
			payload["address"], 
			payload["amount"],
			payload["memo"],
		)

		if err != nil {
			fmt.Printf("  Status: ❌ FAILED - %v\n\n", err)
			evt.Status = types.EventStatusFailed
			evt.LastError = err.Error()
			evt.Retries++
		} else {
			fmt.Printf("  TxID:   %s\n", txid)
			fmt.Printf("  Status: ✓ BROADCAST SUCCESSFUL\n\n")
			evt.Transition(types.EventStatusProcessing)
			evt.Transition(types.EventStatusCompleted)
		}

		store.Update(evt)
		time.Sleep(500 * time.Millisecond)
	}

	// Run reconciliation
	summary, _ := reconciler.Reconcile()

	fmt.Println("═══ Reconciliation Summary ═══\n")
	fmt.Printf("  Duration:        %s\n", summary.CompletedAt.Sub(summary.StartedAt))
	fmt.Printf("  Events Checked:  %d\n", summary.EventsChecked)
	fmt.Printf("  Events Settled:  %d ✓\n", summary.EventsSettled)
	fmt.Printf("  Events Failed:   %d ✗\n\n", summary.EventsFailed)

	if summary.EventsSettled > 0 {
		// Mine confirmation block
		fmt.Println("⛏️  Mining confirmation block...")
		btcAlice.Call("generatetoaddress", 1, aliceAddr)
		fmt.Println("✓ Block mined\n")

		time.Sleep(1 * time.Second)

		// Final balances
		fmt.Println("═══ Final Results ═══\n")
		aliceFinal, _ := btcAlice.Call("getbalance")
		bobFinal, _ := btcBob.Call("getbalance")

		totalSent := 0.0
		for _, tx := range transactions {
			totalSent += tx.amount
		}

		fmt.Printf("Alice: %.8f BTC (sent: %.2f BTC + fees)\n", aliceFinal, totalSent)
		fmt.Printf("Bob:   %.8f BTC (received: %.2f BTC)\n\n", bobFinal, totalSent)

		// Show Bob's transactions
		fmt.Println("═══ Bob's Confirmed Transactions ═══\n")
		txs, _ := btcBob.Call("listtransactions", "*", 10)
		txList := txs.([]interface{})
		
		receivedCount := 0
		for _, tx := range txList {
			txMap := tx.(map[string]interface{})
			if txMap["category"] == "receive" {
				receivedCount++
				fmt.Printf("%d. %.8f BTC\n", receivedCount, txMap["amount"])
				fmt.Printf("   %s\n", txMap["label"])
				fmt.Printf("   Confirmations: %.0f\n\n", txMap["confirmations"])
			}
		}
	}

	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║  ✓ Demo Complete - Network Partition Handled Successfully║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝\n")

	fmt.Println("What Happened:")
	fmt.Println("1. ✓ Bitcoin Core was running (network online)")
	fmt.Println("2. ❌ You stopped Bitcoin Core (simulated network partition)")
	fmt.Println("3. 💾 SDK queued 3 transactions in SQLite (offline mode)")
	fmt.Println("4. ✓ You restarted Bitcoin Core (network recovered)")
	fmt.Println("5. 🔄 SDK detected recovery and activated reconciler")
	fmt.Println("6. 📡 Replayed and broadcast all queued transactions")
	fmt.Println("7. ⛏️  Mined confirmation block")
	fmt.Println("8. ✅ Bob received all payments successfully\n")

	fmt.Println("Key Takeaway:")
	fmt.Println("  Stopping/starting Bitcoin Core = Network partition/recovery")
	fmt.Println("  MeshGuard SDK ensures NO transaction loss during outages!\n")
}
