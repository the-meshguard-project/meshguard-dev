@echo off
setlocal enabledelayedexpansion

echo === Bitcoin Core Regtest Full Setup ===
echo.

REM -------------------------
REM 1. Check installation
REM -------------------------
echo [1/6] Checking Bitcoin Core installation...
where bitcoind >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: bitcoind not found. Install Bitcoin Core first.
    echo Download from: https://bitcoin.org/en/download
    pause
    exit /b 1
)
echo Bitcoin Core found [OK]
echo.

REM -------------------------
REM 2. Create bitcoin.conf
REM -------------------------
echo [2/6] Creating bitcoin.conf...
set BITCOIN_DIR=%APPDATA%\Bitcoin
set CONF_FILE=%BITCOIN_DIR%\bitcoin.conf

if not exist "%BITCOIN_DIR%" mkdir "%BITCOIN_DIR%"

(
echo regtest=1
echo server=1
echo daemon=1
echo rpcuser=alice
echo rpcpassword=strongpassword123
echo fallbackfee=0.0002
) > "%CONF_FILE%"

echo Config file created at %CONF_FILE% [OK]
echo.

REM -------------------------
REM 3. Start bitcoind
REM -------------------------
echo [3/6] Starting Bitcoin Core...
start /B bitcoind -regtest
timeout /t 5 /nobreak >nul
echo Bitcoin Core started [OK]
echo.

REM -------------------------
REM 4. Create wallets
REM -------------------------
echo [4/6] Creating wallets...
bitcoin-cli -regtest createwallet "Alice" 2>nul || echo Alice wallet exists
bitcoin-cli -regtest createwallet "Bob" 2>nul || echo Bob wallet exists
echo.

REM -------------------------
REM 5. Generate addresses
REM -------------------------
echo [5/6] Generating addresses...
for /f "delims=" %%i in ('bitcoin-cli -regtest -rpcwallet=Alice getnewaddress') do set ALICE_ADDR=%%i
for /f "delims=" %%i in ('bitcoin-cli -regtest -rpcwallet=Bob getnewaddress') do set BOB_ADDR=%%i

echo Alice Address: %ALICE_ADDR%
echo Bob Address:   %BOB_ADDR%
echo.

REM -------------------------
REM 6. Mine blocks
REM -------------------------
echo [6/6] Mining blocks to Alice...
bitcoin-cli -regtest -rpcwallet=Alice generatetoaddress 101 %ALICE_ADDR% >nul
echo.

echo === FINAL BALANCES ===
for /f "delims=" %%i in ('bitcoin-cli -regtest -rpcwallet=Alice getbalance') do set ALICE_BAL=%%i
for /f "delims=" %%i in ('bitcoin-cli -regtest -rpcwallet=Bob getbalance') do set BOB_BAL=%%i

echo Alice: %ALICE_BAL% BTC
echo Bob:   %BOB_BAL% BTC
echo.

echo Setup complete [OK]
echo.
echo Now run: go run sdk/examples/alice_bob_demo.go
echo.
pause
