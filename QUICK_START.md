# MeshGuard SDK - Quick Start Guide

Get up and running with the MeshGuard SDK and Bitcoin Core regtest in 5 minutes.

## What is MeshGuard?

MeshGuard is a resilient transaction coordination system that ensures Bitcoin payments survive network partitions. When connectivity drops, transactions queue locally. When the network returns, the reconciliation engine replays and settles queued events atomically.

**Core Philosophy**: When networks fail, transactions shouldn't.

## Installation

```bash
# Clone the repository
git clone https://github.com/the-meshguard-project/meshguard-dev.git
cd meshguard-dev

# Install dependencies
go mod tidy
```

## Run the Demo (No Bitcoin Core Required)

```bash
go run sdk/examples/regtest_demo.go
```

This demonstrates the SDK queuing and reconciling 5 transactions without any external dependencies.

## Full Bitcoin Core Integration

### Step 1: Install Bitcoin Core

Download from: https://bitcoin.org/en/download

Verify installation:
```bash
bitcoind --version
```

### Step 2: Setup Regtest Environment

**Windows:**
```cmd
scripts\setup_regtest.bat
```

**Linux/macOS:**
```bash
chmod +x scripts/setup_regtest.sh
./scripts/setup_regtest.sh
```

This creates Alice and Bob wallets, mines blocks, and funds Alice with 50 BTC.

### Step 3: Run Alice → Bob Demo

```bash
go run sdk/examples/alice_bob_demo.go
```

Expected output:
```
=== MeshGuard SDK - Alice & Bob Bitcoin Regtest Demo ===

Connected: chain=regtest, blocks=101
Alice: 50.00000000 BTC
Bob:   0.00000000 BTC

Network Partition - Queuing Transactions
  Queued: 1.50000000 BTC (seq: 1)
  Queued: 2.00000000 BTC (seq: 2)
  Queued: 0.50000000 BTC (seq: 3)

Network Recovered - Processing Transactions
  TxID: abc123...
  TxID: def456...
  TxID: ghi789...

Reconciliation: 3 settled, 0 failed

Final Balances
Alice: 45.99998560 BTC
Bob:   4.00000000 BTC

Demo Complete!
```

### Step 4: Cleanup

**Windows:**
```cmd
scripts\cleanup_regtest.bat
```

**Linux/macOS:**
```bash
./scripts/cleanup_regtest.sh
```

## Project Structure

```
meshguard-dev/
├── sdk/
│   ├── types/              # Event definitions, state machine
│   ├── queue/              # SQLite persistence layer
│   ├── engine/             # Reconciliation logic
│   └── examples/           # Demos and integration tests
├── scripts/                # Regtest automation
├── go.mod                  # Go dependencies
└── README files            # Documentation
```

## SDK Components

### 1. Event Queue (`sdk/queue`)
- Persists transactions in SQLite with WAL mode
- Survives process crashes and network failures
- Query events by status

### 2. Reconciler (`sdk/engine`)
- Replays queued events after network recovery
- Pause/Resume controls
- Atomic state transitions

### 3. Types (`sdk/types`)
- Event state machine: pending → processing → completed
- Atomic sequence clock for ordering
- Reconciliation summary tracking

## Test Coverage

```
✓ sdk/types   100.0% coverage
✓ sdk/queue    87.2% coverage (includes SQLite integration)
✓ sdk/engine   77.5% coverage
```

Run tests:
```bash
go test ./sdk/... -v
```

## Available Demos

1. **Basic Demo** - No dependencies
   ```bash
   go run sdk/examples/regtest_demo.go
   ```

2. **Bitcoin Core Integration** - Requires bitcoind
   ```bash
   go run sdk/examples/bitcoin_regtest.go
   ```

3. **Alice & Bob Full Demo** - Complete end-to-end
   ```bash
   go run sdk/examples/alice_bob_demo.go
   ```

## Documentation

- **Full SDK Examples**: `sdk/examples/README.md`
- **Alice & Bob Demo**: `sdk/examples/ALICE_BOB_DEMO.md`
- **Bitcoin Regtest Guide**: `BITCOIN_REGTEST_GUIDE.md`
- **Test Results**: `sdk/examples/TEST_RESULTS.md`
- **Script Documentation**: `scripts/README.md`

## Common Issues

### Bitcoin Core Not Found
```
ERROR: bitcoind not found
```
Install Bitcoin Core from https://bitcoin.org/en/download

### Connection Refused
```
Error: dial tcp 127.0.0.1:18443: connectex: No connection could be made
```
Run the setup script: `scripts/setup_regtest.bat` or `.sh`

### Port Already in Use
```bash
# Windows
netstat -ano | findstr :18443
taskkill /F /PID <pid>

# Linux/macOS
lsof -i :18443
kill <pid>
```

## What Gets Demonstrated

✅ **Network Partition Handling** - Transactions queue locally  
✅ **SQLite WAL Persistence** - Survives crashes  
✅ **Event State Machine** - Valid transitions enforced  
✅ **Atomic Sequencing** - Correct replay order  
✅ **Reconciliation** - Automatic settlement after recovery  
✅ **Bitcoin Integration** - Real RPC calls and transactions  
✅ **Pure Go** - No CGO required, works on Windows  

## Next Steps

1. **Explore the Code**
   - Read `sdk/types/event.go` for the state machine
   - Check `sdk/engine/reconciler.go` for reconciliation logic
   - Review `sdk/queue/sqlite_store.go` for persistence

2. **Modify the Demos**
   - Change transaction amounts
   - Add more participants
   - Simulate different failure scenarios

3. **Build Your Application**
   - Use the SDK as a library
   - Integrate with your payment system
   - Add REST API endpoints
   - Connect to Lightning Network

4. **Read the Architecture**
   - See how events flow through the system
   - Understand state transitions
   - Learn about atomic sequencing

## Get Help

- 📖 Read the documentation in `sdk/examples/`
- 🐛 Check test files for usage examples
- 💬 Open an issue on GitHub

## Contributing

The SDK is in active development. Contributions welcome!

```bash
git checkout -b feature/my-feature
# Make changes
git commit -m "feat: add my feature"
git push origin feature/my-feature
```

---

**MeshGuard SDK** - Resilient transaction coordination for Bitcoin

When networks fail, transactions shouldn't. ⚡
