# MeshGuard Bitcoin Regtest Scripts

Automated scripts for setting up and testing MeshGuard SDK with Bitcoin Core regtest.

## Quick Start

### Windows

```cmd
REM 1. Setup Bitcoin Core regtest
scripts\setup_regtest.bat

REM 2. Run Alice & Bob demo
go run sdk/examples/alice_bob_demo.go

REM 3. Cleanup
scripts\cleanup_regtest.bat
```

### Linux/macOS

```bash
# 1. Setup Bitcoin Core regtest
chmod +x scripts/setup_regtest.sh
./scripts/setup_regtest.sh

# 2. Run Alice & Bob demo
go run sdk/examples/alice_bob_demo.go

# 3. Cleanup
chmod +x scripts/cleanup_regtest.sh
./scripts/cleanup_regtest.sh
```

## What the Setup Script Does

1. **Checks Installation** - Verifies bitcoind is available
2. **Creates bitcoin.conf** - Sets up regtest configuration
   ```
   regtest=1
   server=1  
   rpcuser=alice
   rpcpassword=strongpassword123
   ```
3. **Starts Bitcoin Core** - Launches bitcoind in regtest mode
4. **Creates Wallets** - Sets up Alice and Bob wallets
5. **Generates Addresses** - Creates receiving addresses
6. **Mines Blocks** - Mines 101 blocks to fund Alice (50 BTC)

## RPC Credentials

All examples use these credentials:
- **RPC User**: `alice`
- **RPC Password**: `strongpassword123`
- **RPC Port**: `18443` (regtest default)

## Demo Flow

The alice_bob_demo.go demonstrates:

```
Alice (50 BTC) → Network Partition Detected
       ↓
   Queue 3 Transactions:
   - 1.5 BTC to Bob
   - 2.0 BTC to Bob  
   - 0.5 BTC to Bob
       ↓
   Network Recovered
       ↓
   MeshGuard SDK Reconciles
       ↓
   Broadcast Transactions
       ↓
   Mine Confirmation Block
       ↓
   Bob receives 4.0 BTC ✓
```

## Manual Bitcoin Core Commands

If you prefer manual setup:

```bash
# Start
bitcoind -regtest -daemon -rpcuser=alice -rpcpassword=strongpassword123

# Create wallets
bitcoin-cli -regtest createwallet "Alice"
bitcoin-cli -regtest createwallet "Bob"

# Get addresses
bitcoin-cli -regtest -rpcwallet=Alice getnewaddress
bitcoin-cli -regtest -rpcwallet=Bob getnewaddress

# Mine blocks
bitcoin-cli -regtest -rpcwallet=Alice generatetoaddress 101 <alice-address>

# Check balances
bitcoin-cli -regtest -rpcwallet=Alice getbalance
bitcoin-cli -regtest -rpcwallet=Bob getbalance

# Send transaction
bitcoin-cli -regtest -rpcwallet=Alice sendtoaddress <bob-address> 1.5

# Stop
bitcoin-cli -regtest stop
```

## Troubleshooting

### Port Already in Use
```bash
# Find process using port 18443
netstat -ano | findstr :18443  # Windows
lsof -i :18443                 # Linux/macOS

# Kill existing bitcoind
taskkill /F /IM bitcoind.exe   # Windows
pkill bitcoind                 # Linux/macOS
```

### Connection Refused
- Verify bitcoind is running: `bitcoin-cli -regtest getblockchaininfo`
- Check credentials match in bitcoin.conf
- Wait a few seconds after starting bitcoind

### Wallet Exists Error
```bash
# List existing wallets
bitcoin-cli -regtest listwallets

# Load existing wallet
bitcoin-cli -regtest loadwallet "Alice"

# Or delete and recreate
rm -rf ~/.bitcoin/regtest/wallets/Alice  # Linux/macOS
```

## Directory Locations

### Windows
- Config: `%APPDATA%\Bitcoin\bitcoin.conf`
- Data: `%APPDATA%\Bitcoin\regtest\`

### Linux/macOS
- Config: `~/.bitcoin/bitcoin.conf`
- Data: `~/.bitcoin/regtest/`

## Script Files

- **setup_regtest.bat** / **.sh** - Initial setup and blockchain creation
- **cleanup_regtest.bat** / **.sh** - Stop daemon and remove data
- **README.md** - This file

## Integration with MeshGuard SDK

The SDK connects to Bitcoin Core and:
1. Queues transactions during network partition
2. Persists them in SQLite with WAL mode
3. Replays transactions on network recovery
4. Tracks state transitions
5. Provides reconciliation summary

See `sdk/examples/alice_bob_demo.go` for full implementation.
