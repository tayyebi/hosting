#!/usr/bin/env bash
set -euo pipefail

echo "[Platform] Stopping any running instance..."
# On Windows/Git Bash use taskkill via cmd; fall back to pkill on real Linux
if command -v cmd &>/dev/null; then
    cmd //c "taskkill /F /IM platform.exe" 2>/dev/null || true
else
    pkill -f "./platform" 2>/dev/null || true
fi

echo "[Platform] Cleaning old binary..."
rm -f ./platform.exe ./platform

echo "[Platform] Building..."
go build -o platform.exe .

echo "[Platform] Starting on :3000"
./platform.exe
