# Bitcoin Core Regtest with MeshGuard SDK

## Quick Start

### 1. Start Bitcoin Core in Regtest Mode

```bash
bitcoind -regtest -daemon -rpcuser=user -rpcpassword=password
```

### 2. Create Wallet and Generate Blocks

```bash
# Create a test wallet
bitcoin-cli -regtest -rpcuser=user -rpcpassword=password createwallet "meshguard_test"

# Generate 101 blocks (need 100 confirmations for coinbase maturity)
bitcoin-cli -regtest -rpcuser=user -rpcpassword=password generatetoaddress 101 $(bitcoin-cli -regtest -rpcuser=user -rpcpassword=password getnewaddress)
```

### 3. Verify Setup

```bash
# Check blockchain info
bitcoin-cli -regtest -rpcuser=user -rpcpassword=password getblockchaininfo

# Check wallet balance
bitcoin-cli -regtest -rpcuser=user -rpcpassword=password getbalance
```

Expected output: 50 BTC (from block rewards)

### 4. Run SDK Demo

```bash
# Basic demo (no Bitcoin Core required)
go run sdk/examples/regtest_demo.go

# Full Bitcoin Core integration
go run sdk/examples/bitcoin_regtest.go
```

## Update RPC Credentials

Edit `sdk/examples/bitcoin_regtest.go` if using different credentials:

```go
btc := &BitcoinRPC{
    URL:      "http://127.0.0.1:18443", // regtest default
    User:     "your-rpc-user",
    Password: "your-rpc-password",
}
```

## What the Demo Shows

1. **Network Partition Simulation**
   - Transactions queue locally in SQLite
   - WAL mode ensures durability
   
2. **Event State Machine**
   ```
   pending → processing → reconciling → completed
                ↓
              failed → pending (retry)
   ```

3. **Reconciliation After Recovery**
   - Events replay in order (atomic sequences)
   - State transitions tracked
   - Summary shows success/failure counts

## Testing Scenarios

### Scenario 1: Normal Flow
```bash
go run sdk/examples/regtest_demo.go
```
All events transition: pending → completed

### Scenario 2: With Failures
Modify the demo to add more failed events, then watch reconciliation handle retries.

### Scenario 3: Concurrent Events
The atomic clock ensures proper ordering even with concurrent transactions.

## Clean Up

```bash
# Stop Bitcoin Core
bitcoin-cli -regtest stop

# Or kill process
pkill bitcoind
```

## Architecture Flow

```
┌─────────────────┐
│ Bitcoin Payment │
│   Request       │
└────────┬────────┘
         │
         ↓
┌─────────────────┐      Network
│  Queue Event    │◄──── Partition
│  (SQLite WAL)   │      Detected
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│ Wait for        │
│ Network         │
└────────┬────────┘
         │
         ↓
┌─────────────────┐      Network
│  Reconciler     │◄──── Recovered
│  Wakes Up       │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│ Replay Events   │
│ In Order        │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│ Mark Completed  │
│ or Failed       │
└─────────────────┘
```

## SDK Test Results

```
✓ types:  100.0% coverage - Event state machine
✓ queue:   87.2% coverage - SQLite persistence  
✓ engine:  77.5% coverage - Reconciliation logic
```

## Troubleshooting

**Bitcoin Core not starting:**
```bash
# Check if already running
ps aux | grep bitcoind

# Check regtest data directory
ls ~/.bitcoin/regtest/
```

**Connection refused:**
- Verify Bitcoin Core is running: `bitcoin-cli -regtest getblockchaininfo`
- Check RPC credentials match in bitcoin.conf or command line
- Verify port 18443 is not blocked

**Insufficient balance:**
```bash
# Generate more blocks
bitcoin-cli -regtest generatetoaddress 10 $(bitcoin-cli -regtest getnewaddress)
```
