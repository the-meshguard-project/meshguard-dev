#!/bin/bash
set -e

echo "=== Bitcoin Core Regtest Full Setup ==="

# -------------------------
# 1. Check installation
# -------------------------
echo "[1/6] Checking Bitcoin Core installation..."
if ! command -v bitcoind &> /dev/null; then
    echo "ERROR: bitcoind not found. Install Bitcoin Core first."
    exit 1
fi
echo "Bitcoin Core found ✔"

# -------------------------
# 2. Create bitcoin.conf
# -------------------------
echo "[2/6] Creating bitcoin.conf..."
BITCOIN_DIR="$HOME/.bitcoin"
CONF_FILE="$BITCOIN_DIR/bitcoin.conf"

mkdir -p "$BITCOIN_DIR"

cat > "$CONF_FILE" <<EOF
regtest=1
server=1
daemon=1
rpcuser=alice
rpcpassword=strongpassword123
fallbackfee=0.0002
EOF

echo "Config file created at $CONF_FILE ✔"

# -------------------------
# 3. Start bitcoind
# -------------------------
echo "[3/6] Starting Bitcoin Core..."
bitcoind -regtest -daemon
sleep 5
echo "Bitcoin Core started ✔"

# -------------------------
# 4. Create wallets
# -------------------------
echo "[4/6] Creating wallets..."
bitcoin-cli -regtest createwallet "Alice" 2>/dev/null || echo "Alice wallet exists"
bitcoin-cli -regtest createwallet "Bob" 2>/dev/null || echo "Bob wallet exists"

# -------------------------
# 5. Generate addresses
# -------------------------
echo "[5/6] Generating addresses..."
ALICE_ADDR=$(bitcoin-cli -regtest -rpcwallet=Alice getnewaddress)
BOB_ADDR=$(bitcoin-cli -regtest -rpcwallet=Bob getnewaddress)

echo "Alice Address: $ALICE_ADDR"
echo "Bob Address:   $BOB_ADDR"

# -------------------------
# 6. Mine blocks (give Alice coins)
# -------------------------
echo "[6/6] Mining blocks to Alice..."
bitcoin-cli -regtest -rpcwallet=Alice generatetoaddress 101 "$ALICE_ADDR"

echo ""
echo "=== FINAL BALANCES ==="
echo "Alice: $(bitcoin-cli -regtest -rpcwallet=Alice getbalance) BTC"
echo "Bob:   $(bitcoin-cli -regtest -rpcwallet=Bob getbalance) BTC"
echo ""
echo "Setup complete ✔"
echo ""
echo "Now run: go run sdk/examples/alice_bob_demo.go"
