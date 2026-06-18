# How to Simulate Network Partition with Bitcoin Core

## Understanding Network Partition Simulation

**Yes!** Stopping Bitcoin Core server = Network partition (offline mode)  
**Starting Bitcoin Core** = Network recovery (online mode)

This simulates real-world scenarios where:
- Internet connection drops
- Bitcoin node becomes unreachable  
- Network routing fails
- Firewall blocks connections

## Quick Demo

### Method 1: Manual Start/Stop

```bash
# 1. Start Bitcoin Core
bitcoind -regtest -daemon -rpcuser=alice -rpcpassword=strongpassword123

# 2. Run your application (it queues transactions)
go run sdk/examples/alice_bob_demo.go

# 3. STOP Bitcoin Core (simulate network partition)
bitcoin-cli -regtest stop

# 4. Application continues working, queues transactions locally

# 5. START Bitcoin Core again (simulate recovery)
bitcoind -regtest -daemon -rpcuser=alice -rpcpassword=strongpassword123

# 6. Application detects recovery and processes queued transactions
```

### Method 2: Interactive Demo

This demo walks you through the process step-by-step:

```bash
# 1. Setup Bitcoin Core first
scripts\setup_regtest.bat  # Windows
./scripts/setup_regtest.sh  # Linux/macOS

# 2. Run interactive demo
go run sdk/examples/interactive_partition_demo.go
```

The interactive demo will:
1. ✓ Check Bitcoin Core is running
2. ⚠️ Prompt you to STOP Bitcoin Core
3. 💾 Queue transactions while offline
4. 🔌 Prompt you to START Bitcoin Core
5. 🔄 Automatically reconcile and broadcast
6. ✅ Show final results

## What Happens During Each Phase

### Phase 1: Normal Operation (Bitcoin Core Running)

```
Application → Bitcoin Core RPC
              ↓
         [Connected ✓]
              ↓
    Transactions Broadcast
         Immediately
```

**Status**: All transactions go through directly

### Phase 2: Network Partition (Bitcoin Core Stopped)

```
Application → Bitcoin Core RPC
              ↓
         [Connection Refused ❌]
              ↓
         Detect Failure
              ↓
    Queue to SQLite WAL 💾
              ↓
    Transactions Pending
```

**Status**: Transactions queued locally, no data loss

**What you'll see:**
```
Error: dial tcp 127.0.0.1:18443: connectex: No connection could be made
```

### Phase 3: Network Recovery (Bitcoin Core Restarted)

```
   Reconciler Activates
              ↓
    Fetch Pending Events
              ↓
    Sort by Sequence
              ↓
    Broadcast Each TX
              ↓
    Update Status → Complete ✓
```

**Status**: All queued transactions processed in order

## Complete Example Flow

### Step-by-Step Commands

```bash
# Terminal 1: Setup and start Bitcoin Core
scripts\setup_regtest.bat
# Bitcoin Core is now running on port 18443

# Terminal 2: Start your application
go run sdk/examples/alice_bob_demo.go
# Application is now connected and working

# Terminal 1: Simulate network failure
bitcoin-cli -regtest stop
# Bitcoin Core stopped - network is "down"

# Terminal 2: Application detects failure
# ❌ Connection refused
# 💾 Queuing transactions to SQLite
# ✓ No data loss

# Wait a few moments...

# Terminal 1: Simulate network recovery
bitcoind -regtest -daemon -rpcuser=alice -rpcpassword=strongpassword123
# Bitcoin Core restarted - network is "up"

# Terminal 2: Application detects recovery
# ✓ Network online
# 🔄 Reconciling...
# 📡 Broadcasting queued transactions
# ✅ All transactions settled
```

## Testing Different Scenarios

### Scenario 1: Brief Outage (< 1 minute)

```bash
# Stop
bitcoin-cli -regtest stop

# Wait 30 seconds
sleep 30

# Restart
bitcoind -regtest -daemon -rpcuser=alice -rpcpassword=strongpassword123
```

**Result**: Small queue, quick recovery

### Scenario 2: Extended Outage (several minutes)

```bash
# Stop
bitcoin-cli -regtest stop

# Queue many transactions while offline
# (run your application, it will queue)

# Wait 5 minutes
sleep 300

# Restart
bitcoind -regtest -daemon -rpcuser=alice -rpcpassword=strongpassword123
```

**Result**: Larger queue, batch reconciliation

### Scenario 3: Multiple Outages

```bash
# Stop
bitcoin-cli -regtest stop
# Queue some transactions

# Restart
bitcoind -regtest -daemon -rpcuser=alice -rpcpassword=strongpassword123
# Process queue

# Stop again
bitcoin-cli -regtest stop
# Queue more transactions

# Restart again
bitcoind -regtest -daemon -rpcuser=alice -rpcpassword=strongpassword123
# Process second queue
```

**Result**: Multiple reconciliation cycles

## How the SDK Detects Network Status

### Connection Check

```go
func checkConnection(btc *BitcoinRPC) bool {
    _, err := btc.Call("getblockchaininfo")
    return err == nil  // true = online, false = offline
}
```

### Offline Detection

```go
if err != nil {
    // Connection failed = Network is down
    // Switch to offline mode
    queueTransaction(event)
    return
}
```

### Recovery Detection

```go
// Periodic check (every N seconds)
for {
    if checkConnection(btc) {
        // Network is back!
        reconciler.Reconcile()
        break
    }
    time.Sleep(5 * time.Second)
}
```

## Verifying Queue Persistence

### Check SQLite Database

```bash
# While Bitcoin Core is stopped, check the queue
sqlite3 meshguard.db "SELECT id, status, sequence FROM events;"

# You'll see pending transactions:
# tx-1|pending|1
# tx-2|pending|2
# tx-3|pending|3
```

### Verify WAL Mode

```bash
# Check WAL files exist
ls -la meshguard.db*

# Output:
# meshguard.db       - Main database
# meshguard.db-shm   - Shared memory
# meshguard.db-wal   - Write-Ahead Log
```

The `-wal` file ensures transactions are durable even if the process crashes.

## Real-World Equivalents

| Simulation | Real-World Scenario |
|------------|-------------------|
| `bitcoin-cli stop` | Internet connection drops |
| Keep stopped | Extended network outage |
| `bitcoind -daemon` | Internet restored |
| Restart multiple times | Intermittent connectivity |

## Monitoring During Partition

### Check Queue Status

```bash
# In your Go code
pending, _ := store.GetByStatus(types.EventStatusPending)
fmt.Printf("Queued: %d transactions\n", len(pending))
```

### Check Reconciliation Progress

```bash
summary, _ := reconciler.Reconcile()
fmt.Printf("Settled: %d\n", summary.EventsSettled)
fmt.Printf("Failed: %d\n", summary.EventsFailed)
```

## Troubleshooting

### Bitcoin Core Won't Stop

```bash
# Force kill
# Windows
taskkill /F /IM bitcoind.exe

# Linux/macOS
pkill -9 bitcoind
```

### Bitcoin Core Won't Start

```bash
# Check if already running
bitcoin-cli -regtest getblockchaininfo

# If port is in use
netstat -ano | findstr :18443  # Windows
lsof -i :18443                 # Linux/macOS
```

### Queue Not Processing

```bash
# Check network is actually back
bitcoin-cli -regtest getblockchaininfo

# Manually trigger reconciliation
go run sdk/examples/alice_bob_demo.go
```

## Key Takeaways

✅ **Stopping Bitcoin Core** = Network partition (offline)  
✅ **Starting Bitcoin Core** = Network recovery (online)  
✅ **SQLite WAL** = Transactions survive crashes  
✅ **Atomic Sequencing** = Correct replay order  
✅ **Reconciliation** = Automatic after recovery  
✅ **No Data Loss** = All transactions eventually settle  

## Next Steps

1. Run `go run sdk/examples/interactive_partition_demo.go`
2. Follow the prompts to stop/start Bitcoin Core
3. Watch transactions queue and reconcile
4. Experiment with different timing scenarios
5. Build your own resilient payment system!

---

**Remember**: In production, you'd detect real network failures, not manually stop servers. But the principle is the same - MeshGuard SDK ensures transactions survive any network partition.
