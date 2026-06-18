#!/bin/bash

echo "=== Cleaning up Bitcoin Core Regtest ==="
echo ""

echo "Stopping Bitcoin Core..."
bitcoin-cli -regtest stop 2>/dev/null

sleep 3

echo "Removing regtest data directory..."
BITCOIN_DIR="$HOME/.bitcoin/regtest"
if [ -d "$BITCOIN_DIR" ]; then
    rm -rf "$BITCOIN_DIR"
    echo "Regtest data removed ✔"
else
    echo "No regtest data found"
fi

echo ""
echo "Cleanup complete!"
