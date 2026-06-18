# MeshGuard SDK

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Tests](https://img.shields.io/badge/tests-passing-success)](./sdk/examples/TEST_RESULTS.md)
[![Coverage](https://img.shields.io/badge/coverage-88%25-brightgreen)](./sdk/examples/TEST_RESULTS.md)
[![License](https://img.shields.io/badge/license-MIT-blue)](./LICENSE)

**Resilient transaction coordination for Bitcoin.** When networks fail, transactions shouldn't.

MeshGuard is a transaction coordination system that ensures Bitcoin payments survive network partitions. When connectivity drops, transactions queue locally in SQLite. When the network returns, the reconciliation engine replays and settles queued events atomically.

---

##  Quick Start

### Option 1: Run Demo (No Bitcoin Core Required)

```bash
git clone https://github.com/the-meshguard-project/meshguard-dev.git
cd meshguard-dev
go mod tidy
go run sdk/examples/regtest_demo.go
```

**Output:**
```
=== MeshGuard SDK Regtest Demo ===

✓ Initialized SQLite store and reconciler

--- Scenario: Network Partition Detected ---
Simulating 5 Bitcoin transactions queued during partition...

  Queued: btc-tx-001 (seq: 1, status: pending)
  Queued: btc-tx-002 (seq: 2, status: pending)
  Queued: btc-tx-003 (seq: 3, status: pending)
  Queued: btc-tx-004 (seq: 4, status: failed)
  Queued: btc-tx-005 (seq: 5, status: pending)

✓ All transactions queued in local SQLite WAL

--- Network Recovered ---
Starting reconciliation engine...

--- Reconciliation Summary ---
Duration: 2.32s
Events Checked:  5
Events Settled:  5 ✓
Events Failed:   0 ✗

✓ SDK Demo Complete
```

### Option 2: Full Bitcoin Core Integration

**Prerequisites:**
- [Bitcoin Core](https://bitcoin.org/en/download) installed
- Go 1.21+

**Windows:**
```cmd
scripts\setup_regtest.bat
go run sdk/examples/alice_bob_demo.go
```

**Linux/macOS:**
```bash
chmod +x scripts/setup_regtest.sh
./scripts/setup_regtest.sh
go run sdk/examples/alice_bob_demo.go
```

---

##  Table of Contents

- [Features](#-features)
- [Architecture](#-architecture)
- [Installation](#-installation)
- [Usage Examples](#-usage-examples)
- [Running Tests](#-running-tests)
- [Test Results](#-test-results)
- [Documentation](#-documentation)
- [API Reference](#-api-reference)
- [Troubleshooting](#-troubleshooting)
- [Contributing](#-contributing)

---

##  Features

- **Network Partition Resilience** - Transactions survive connectivity failures
- **SQLite WAL Persistence** - Durable local queue with Write-Ahead Logging
- **Event State Machine** - Validated transitions: pending → processing → completed
- **Atomic Sequencing** - Guaranteed correct replay order with concurrent safety
- **Automatic Reconciliation** - Replays queued events after network recovery
- **Pure Go** - No CGO required, cross-platform (Windows, Linux, macOS)
- **Bitcoin Core Integration** - Real RPC calls, actual blockchain transactions
- **Pause/Resume** - Fine-grained control over reconciliation process
- **Comprehensive Testing** - 88% average test coverage

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                       │
│              (Your Payment/Lightning System)                │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                    MeshGuard SDK                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │    Types     │  │    Queue     │  │   Engine     │     │
│  │              │  │              │  │              │     │
│  │ • Events     │→ │ • SQLite WAL │→ │ • Reconciler │     │
│  │ • Clock      │  │ • EventStore │  │ • Pause/Resume│    │
│  │ • State      │  │ • CRUD Ops   │  │ • Replay     │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                   Bitcoin Network                           │
│              (Bitcoin Core RPC / Lightning)                 │
└─────────────────────────────────────────────────────────────┘
```

### Event Flow

```
1. Payment Request
       ↓
2. Network Check → ✗ PARTITION DETECTED
       ↓
3. Queue Event (SQLite WAL)
   • ID: unique identifier
   • Sequence: atomic counter
   • Status: pending
   • Payload: transaction data
       ↓
4. Persist to Disk
       ↓
5. Wait for Network Recovery...
       ↓
6. Network Restored ✓
       ↓
7. Reconciler Activates
   • Fetch pending events
   • Sort by sequence
   • Transition: pending → processing
       ↓
8. Broadcast Transactions
   • Send to Bitcoin Network
   • Transition: processing → reconciling
       ↓
9. Confirm Settlement
   • Verify confirmations
   • Transition: reconciling → completed
       ↓
10. Update Queue
    • Mark completed
    • Generate summary
```

---

## 📦 Installation

### Requirements

- **Go**: 1.21 or higher
- **Bitcoin Core** (optional): For full integration
- **Git**: For cloning the repository

### Install

```bash
# Clone repository
git clone https://github.com/the-meshguard-project/meshguard-dev.git
cd meshguard-dev

# Install dependencies
go mod tidy

# Run tests to verify installation
go test ./sdk/...
```

---

## 💡 Usage Examples

### Example 1: Basic Event Queuing

```go
package main

import (
    "github.com/meshguard/sdk/sdk/queue"
    "github.com/meshguard/sdk/sdk/types"
    "time"
)

func main() {
    // Initialize store
    store, _ := queue.NewSQLiteStore("events.db")
    defer store.Close()

    // Create clock
    clock := types.NewClock(0)

    // Queue event
    event := &types.MeshGuardEvent{
        ID:        "payment-001",
        Sequence:  clock.Next(),
        Type:      "bitcoin_payment",
        Status:    types.EventStatusPending,
        Payload:   []byte(`{"amount": 0.001, "to": "address"}`),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    store.Save(event)
    // Event is now persisted and survives crashes
}
```

### Example 2: Reconciliation

```go
package main

import (
    "github.com/meshguard/sdk/sdk/engine"
    "github.com/meshguard/sdk/sdk/queue"
)

func main() {
    store, _ := queue.NewSQLiteStore("events.db")
    defer store.Close()

    reconciler := engine.NewReconciler(store)

    // Run reconciliation
    summary, _ := reconciler.Reconcile()

    fmt.Printf("Events Settled: %d\n", summary.EventsSettled)
    fmt.Printf("Events Failed: %d\n", summary.EventsFailed)
}
```

### Example 3: State Transitions

```go
event := &types.MeshGuardEvent{
    Status: types.EventStatusPending,
}

// Valid transitions
event.Transition(types.EventStatusProcessing)  // ✓ true
event.Transition(types.EventStatusCompleted)   // ✓ true

// Invalid transition
event.Transition(types.EventStatusPending)     // ✗ false
```

---

## 🧪 Running Tests

### Run All Tests

```bash
go test ./sdk/... -v
```

### Run with Coverage

```bash
go test ./sdk/... -cover
```

### Run Specific Package

```bash
# Types package
go test ./sdk/types/... -v

# Queue package  
go test ./sdk/queue/... -v

# Engine package
go test ./sdk/engine/... -v
```

### Run Integration Tests

```bash
# SQLite integration
go test ./sdk/queue/... -run TestSQLite -v

# Reconciler integration
go test ./sdk/engine/... -run TestReconcile -v
```

---

## 📊 Test Results

### Summary

| Package | Coverage | Tests | Status |
|---------|----------|-------|--------|
| `sdk/types` | 100.0% | 4 |  PASS |
| `sdk/queue` | 87.2% | 4 | PASS |
| `sdk/engine` | 77.5% | 4 | PASS |
| **Average** | **88.2%** | **12** | ** ALL PASS** |

### Detailed Results

```
=== RUN   TestEventTransition
=== RUN   TestEventTransition/pending_to_processing
=== RUN   TestEventTransition/pending_to_failed
=== RUN   TestEventTransition/processing_to_completed
=== RUN   TestEventTransition/processing_to_reconciling
=== RUN   TestEventTransition/reconciling_to_completed
--- PASS: TestEventTransition (0.02s)
    --- PASS: TestEventTransition/pending_to_processing (0.00s)
    --- PASS: TestEventTransition/pending_to_failed (0.00s)
    --- PASS: TestEventTransition/processing_to_completed (0.00s)
    --- PASS: TestEventTransition/processing_to_reconciling (0.00s)
    --- PASS: TestEventTransition/reconciling_to_completed (0.00s)
PASS

=== RUN   TestClockConcurrency
--- PASS: TestClockConcurrency (0.01s)
PASS

=== RUN   TestSQLiteStoreIntegration
--- PASS: TestSQLiteStoreIntegration (1.86s)
PASS

=== RUN   TestReconcileWithPendingEvents
--- PASS: TestReconcileWithPendingEvents (0.00s)
PASS

 All tests passing
```

See [detailed test results](./sdk/examples/TEST_RESULTS.md) for more information.

---

##  Documentation

### Core Documentation

- **[Quick Start Guide](./QUICK_START.md)** - Get started in 5 minutes
- **[Alice & Bob Demo](./sdk/examples/ALICE_BOB_DEMO.md)** - End-to-end transaction demo
- **[Bitcoin Regtest Guide](./BITCOIN_REGTEST_GUIDE.md)** - Bitcoin Core integration
- **[Examples README](./sdk/examples/README.md)** - All available examples
- **[Scripts Documentation](./scripts/README.md)** - Automation scripts
- **[Test Results](./sdk/examples/TEST_RESULTS.md)** - Comprehensive test output

### Package Documentation

Generate Go documentation:
```bash
go doc github.com/meshguard/sdk/sdk/types
go doc github.com/meshguard/sdk/sdk/queue
go doc github.com/meshguard/sdk/sdk/engine
```

---

##  API Reference

### Types Package

#### `MeshGuardEvent`
```go
type MeshGuardEvent struct {
    ID          string
    Sequence    uint64
    Type        string
    Status      EventStatus
    Payload     []byte
    CreatedAt   time.Time
    UpdatedAt   time.Time
    Retries     int
    LastError   string
}
```

#### `EventStatus`
```go
const (
    EventStatusPending     EventStatus = "pending"
    EventStatusProcessing  EventStatus = "processing"
    EventStatusCompleted   EventStatus = "completed"aa
    EventStatusFailed      EventStatus = "failed"
    EventStatusReconciling EventStatus = "reconciling"
)
```

#### `Clock`
```go
func NewClock(start uint64) *Clock
func (c *Clock) Next() uint64
func (c *Clock) Current() uint64
```

### Queue Package

#### `EventStore` Interface
```go
type EventStore interface {
    Save(event *MeshGuardEvent) error
    Get(id string) (*MeshGuardEvent, error)
    GetByStatus(status EventStatus) ([]*MeshGuardEvent, error)
    Update(event *MeshGuardEvent) error
    Delete(id string) error
    Close() error
}
```

#### `SQLiteStore`
```go
func NewSQLiteStore(path string) (*SQLiteStore, error)
```

### Engine Package

#### `Reconciler`
```go
func NewReconciler(store queue.EventStore) *Reconciler
func (r *Reconciler) Pause()
func (r *Reconciler) Resume()
func (r *Reconciler) IsPaused() bool
func (r *Reconciler) Reconcile() (*types.ReconciliationSummary, error)
```

---

##  Troubleshooting

### Bitcoin Core Not Found

**Problem:**
```
ERROR: bitcoind not found. Install Bitcoin Core first.
```

**Solution:**
Download and install Bitcoin Core from https://bitcoin.org/en/download

### Connection Refused

**Problem:**
```
Error: dial tcp 127.0.0.1:18443: connectex: No connection could be made
```

**Solution:**
```bash
# Start Bitcoin Core
bitcoind -regtest -daemon -rpcuser=alice -rpcpassword=strongpassword123

# Wait a few seconds, then verify
bitcoin-cli -regtest getblockchaininfo
```

### Port Already in Use

**Problem:**
```
Error: bind: address already in use
```

**Solution:**
```bash
# Windows
netstat -ano | findstr :18443
taskkill /F /PID <pid>

# Linux/macOS
lsof -i :18443
kill <pid>
```

### Database Locked

**Problem:**
```
Error: database is locked
```

**Solution:**
```bash
# Close other processes using the database
# Or remove lock files
rm -f events.db-shm events.db-wal
```

### Test Failures

**Problem:**
Tests fail with "no such table: events"

**Solution:**
```bash
# Clean and rebuild
rm -f *.db*
go clean -testcache
go test ./sdk/... -v
```

---

##  Contributing

Contributions welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/meshguard-dev.git
cd meshguard-dev

# Install dependencies
go mod tidy

# Run tests
go test ./sdk/... -v

# Run linter (if available)
golangci-lint run
```

### Code Style

- Follow Go conventions
- Write tests for new features
- Update documentation
- Keep commits atomic and descriptive

---

##  License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- Bitcoin Core team for the incredible infrastructure
- Go community for excellent tooling
- SQLite team for the reliable database engine

---

##  Contact & Support

- **Issues**: [GitHub Issues](https://github.com/the-meshguard-project/meshguard-dev/issues)
- **Documentation**: [Full Docs](./sdk/examples/)
- **Examples**: [Examples Directory](./sdk/examples/)

---

##  Roadmap

- [ ] Lightning Network integration (LND driver)
- [ ] REST API server
- [ ] WebSocket real-time notifications
- [ ] Dashboard UI (React + Vite)
- [ ] Multi-node coordination
- [ ] Event replay with time-travel debugging
- [ ] Prometheus metrics export
- [ ] Docker compose setup
- [ ] Kubernetes deployment manifests

---

**MeshGuard SDK** - When networks fail, transactions shouldn't. 

