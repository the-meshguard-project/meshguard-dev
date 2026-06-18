# Alice & Bob Bitcoin Transaction Demo

Complete end-to-end demonstration of MeshGuard SDK with Bitcoin Core regtest.

## Overview

This demo simulates a real-world scenario where Alice sends Bitcoin to Bob, but network issues cause transactions to be queued. The MeshGuard SDK ensures all transactions are eventually settled once the network recovers.

## Prerequisites

- Bitcoin Core installed (bitcoind, bitcoin-cli)
- Go 1.21+
- MeshGuard SDK dependencies (`go mod tidy`)

## Quick Start

### Step 1: Setup Bitcoin Core Regtest

**Windows:**
```cmd
scripts\setup_regtest.bat
```

**Linux/macOS:**
```bash
chmod +x scripts/setup_regtest.sh
./scripts/setup_regtest.sh
```

This will:
- Start Bitcoin Core in regtest mode
- Create Alice and Bob wallets
- Mine 101 blocks to fund Alice with 50 BTC
- Display addresses and balances

### Step 2: Run the Demo

```bash
go run sdk/examples/alice_bob_demo.go
```

### Step 3: Cleanup (Optional)

**Windows:**
```cmd
scripts\cleanup_regtest.bat
```

**Linux/macOS:**
```bash
chmod +x scripts/cleanup_regtest.sh
./scripts/cleanup_regtest.sh
```

## Demo Walkthrough

### Initial State
```
Alice: 50.00000000 BTC (from mining)
Bob:   0.00000000 BTC
```

### Scenario: Network Partition

Alice wants to send Bitcoin to Bob, but the network is unreliable. The MeshGuard SDK queues transactions locally:

1. **Transaction 1**: Alice → Bob: 1.5 BTC
2. **Transaction 2**: Alice → Bob: 2.0 BTC
3. **Transaction 3**: Alice → Bob: 0.5 BTC

All transactions are stored in SQLite with Write-Ahead Logging (WAL) for durability.

### Network Recovery

When the network comes back online:

1. **Reconciler wakes up**
2. **Fetches pending events** from the queue
3. **Replays transactions** in sequence order
4. **Broadcasts to Bitcoin network**
5. **Updates state**: pending → processing → completed
6. **Mines confirmation block**

### Final State
```
Alice: ~45.99999XXX BTC (sent 4.0 BTC + fees)
Bob:   4.00000000 BTC (received payments)
```

## Expected Output

```
=== MeshGuard SDK - Alice & Bob Bitcoin Regtest Demo ===

Testing Bitcoin Core connection...
Connected: chain=regtest, blocks=101

Step 1: Creating Wallets
Created: Alice, Bob

Step 2: Generating Addresses
Alice: bcrt1q...
Bob:   bcrt1q...

Step 3: Mining 101 blocks to Alice...
Mined 101 blocks

Alice: 50.00000000 BTC
Bob:   0.00000000 BTC

Step 4: Initialize MeshGuard SDK
SDK initialized

Step 5: Network Partition - Queuing Transactions
  Queued: 1.50000000 BTC (seq: 1)
  Queued: 2.00000000 BTC (seq: 2)
  Queued: 0.50000000 BTC (seq: 3)

Step 6: Network Recovered - Processing Transactions
Processing tx-alice-bob-1: 1.50000000 BTC
  TxID: abc123...
Processing tx-alice-bob-2: 2.00000000 BTC
  TxID: def456...
Processing tx-alice-bob-3: 0.50000000 BTC
  TxID: ghi789...

Reconciliation: 3 settled, 0 failed

Step 7: Mining Confirmation Block
Block mined

Step 8: Final Balances
Alice: 45.99998560 BTC
Bob:   4.00000000 BTC

Bob's Transactions:
1. 1.50000000 BTC (confirmations: 1)
2. 2.00000000 BTC (confirmations: 1)
3. 0.50000000 BTC (confirmations: 1)

Demo Complete!
```

## Architecture

```
┌─────────────┐
│   Alice     │
│  50 BTC     │
└──────┬──────┘
       │
       │ Network Partition!
       ↓
┌─────────────────────┐
│  MeshGuard SDK      │
│  ┌───────────────┐  │
│  │ Event Queue   │  │
│  │ (SQLite WAL)  │  │
│  ├───────────────┤  │
│  │ tx-1: 1.5 BTC │  │
│  │ tx-2: 2.0 BTC │  │
│  │ tx-3: 0.5 BTC │  │
│  └───────────────┘  │
└──────┬──────────────┘
       │
       │ Network Recovered!
       ↓
┌─────────────────────┐
│  Reconciler         │
│  ┌───────────────┐  │
│  │ Replay Events │  │
│  │ In Order      │  │
│  └───────────────┘  │
└──────┬──────────────┘
       │
       │ Broadcast TXs
       ↓
┌─────────────────────┐
│  Bitcoin Network    │
│  (Regtest)          │
└──────┬──────────────┘
       │
       │ Mine Block
       ↓
┌─────────────┐
│     Bob     │
│   4 BTC     │
└─────────────┘
```

## What This Demonstrates

### 1. Resilient Transaction Coordination
- Transactions survive network failures
- Local persistence with SQLite WAL
- Guaranteed eventual settlement

### 2. Event State Machine
```
pending → processing → reconciling → completed
              ↓
            failed → (retry)
```

### 3. Atomic Sequencing
- Clock ensures correct replay order
- Concurrent transactions handled safely
- No lost or duplicate events

### 4. Real Bitcoin Integration
- Actual RPC calls to Bitcoin Core
- Real transaction broadcasting
- Real blockchain confirmations

## Key SDK Components

### Event Queue (`sdk/queue`)
- SQLite persistence
- WAL mode for durability
- Query by status

### Reconciler (`sdk/engine`)
- Pause/Resume capability
- Atomic event processing
- Summary reporting

### Types (`sdk/types`)
- Event state machine
- Atomic clock
- Reconciliation tracking

## Testing Different Scenarios

### Scenario 1: All Transactions Succeed
```bash
go run sdk/examples/alice_bob_demo.go
```

### Scenario 2: Insufficient Funds
Modify the amounts to exceed Alice's balance and watch the SDK handle failures gracefully.

### Scenario 3: Network Interruption
Stop bitcoind mid-demo and restart to see SDK resilience.

## Troubleshooting

### Bitcoin Core Not Running
```
Error: connection failed: dial tcp 127.0.0.1:18443: connectex: No connection could be made
```
**Solution**: Run `scripts/setup_regtest.bat` or `.sh`

### Wallet Already Exists
```
Error: Wallet file verification failed. Failed to load database path...
```
**Solution**: Run cleanup script and setup again

### RPC Authentication Failed
```
Error: RPC error: Authorization failed
```
**Solution**: Check bitcoin.conf has correct credentials:
```
rpcuser=alice
rpcpassword=strongpassword123
```

## Learn More

- Full SDK Documentation: `sdk/examples/README.md`
- Regtest Guide: `BITCOIN_REGTEST_GUIDE.md`
- Test Results: `sdk/examples/TEST_RESULTS.md`
- Script Documentation: `scripts/README.md`

## Next Steps

1. Modify transaction amounts
2. Add more participants (Carol, Dave)
3. Simulate different failure scenarios
4. Integrate with Lightning Network (LND)
5. Add webhook notifications
6. Build a REST API around the SDK

The MeshGuard SDK provides the foundation for building resilient payment systems on Bitcoin!
