@echo off
echo === Cleaning up Bitcoin Core Regtest ===
echo.

echo Stopping Bitcoin Core...
bitcoin-cli -regtest stop 2>nul

timeout /t 3 /nobreak >nul

echo Removing regtest data directory...
set BITCOIN_DIR=%APPDATA%\Bitcoin\regtest
if exist "%BITCOIN_DIR%" (
    rmdir /s /q "%BITCOIN_DIR%"
    echo Regtest data removed [OK]
) else (
    echo No regtest data found
)

echo.
echo Cleanup complete!
pause
