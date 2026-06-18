# MeshGuard SDK Examples

Examples demonstrating the SDK's transaction coordination capabilities.

## Examples

### 1. Basic Demo (`regtest_demo.go`)
Demonstrates core SDK functionality without external dependencies.

```bash
go run sdk/examples/regtest_demo.go
```

**What it does:**
- Simulates network partition
- Queues 5 Bitcoin transactions in SQLite
- Reconciles events after "network recovery"
- Shows state transitions and final results

### 2. Bitcoin Core Integration (`bitcoin_regtest.go`)
Full integration with Bitcoin Core regtest network.

```bash
go run sdk/examples/bitcoin_regtest.go
```

## Running with Bitcoin Core Regtest

### Prerequisites
1. Bitcoin Core installed
2. Access to `bitcoin-cli` and `bitcoind`

### Setup

**Option 1: Bitcoin Core Data Directory**

If Bitcoin Core is installed at default location, update `bitcoin_regtest.go`:

```go
btc := &BitcoinRPC{
    URL:      "http://127.0.0.1:18443",
    User:     "your-rpc-user",
    Password: "your-rpc-password",
}
```

Get RPC credentials from your bitcoin.conf file.

**Option 2: Quick Start with bitcoin-cli**

1. Start Bitcoin Core in regtest:
```bash
bitcoind -regtest -daemon -rpcuser=user -rpcpassword=password
```

2. Generate some blocks:
```bash
bitcoin-cli -regtest -rpcuser=user -rpcpassword=password createwallet "test"
bitcoin-cli -regtest -rpcuser=user -rpcpassword=password generatetoaddress 101 $(bitcoin-cli -regtest -rpcuser=user -rpcpassword=password getnewaddress)
```

3. Run the example:
```bash
go run sdk/examples/bitcoin_regtest.go
```

### Expected Output

```
=== MeshGuard SDK - Bitcoin Core Regtest Integration ===

Testing Bitcoin Core connection...
✓ Connected to Bitcoin Core (chain: regtest, blocks: 101)

Getting wallet address...
Address: bcrt1q...

--- Simulating Network Partition ---
Queuing Bitcoin transactions...

  Queued: btc-payment-1 (seq: 1)
  Queued: btc-payment-2 (seq: 2)

✓ Transactions queued during partition

--- Network Recovered ---
Running reconciliation...

--- Reconciliation Summary ---
Duration:        523ms
Events Checked:  2
Events Settled:  2
Events Failed:   0

✓ Successfully processed 2 transactions

=== Demo Complete ===
```

## What the SDK Demonstrates

1. **Event Queuing**: Transactions persist to SQLite during network issues
2. **State Machine**: Events transition: pending → processing → reconciling → completed
3. **Atomic Ordering**: Sequence numbers ensure correct replay order
4. **Reconciliation**: Automatic replay and settlement after recovery
5. **WAL Mode**: Write-Ahead Logging for durability

## Architecture

```
Network Partition Detected
         ↓
   Queue Events Locally (SQLite WAL)
         ↓
   Network Recovers
         ↓
   Reconciler Wakes Up
         ↓
   Replay Events in Order
         ↓
   Update State → Completed
```

## Cleanup

Stop Bitcoin Core:
```bash
bitcoin-cli -regtest stop
```

Database files are auto-cleaned by the examples.
