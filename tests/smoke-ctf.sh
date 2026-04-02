#!/bin/sh
# smoke-ctf.sh — tool presence check for Dockerfile.ctf
# Runs inside the container via bind-mount: sh /tests/smoke-ctf.sh

PASS=0
FAIL=0

ok()   { printf '  ok   %s\n' "$*"; PASS=$((PASS + 1)); }
fail() { printf '  FAIL %s\n' "$*"; FAIL=$((FAIL + 1)); }

check_cmd() {
    if command -v "$1" >/dev/null 2>&1; then ok  "cmd:  $1"
    else                                    fail "cmd:  $1 not found"; fi
}

check_file() {
    if [ -f "$1" ]; then ok  "file: $1"
    else                 fail "file: $1 not found"; fi
}

check_dir() {
    if [ -d "$1" ]; then ok  "dir:  $1"
    else                 fail "dir:  $1 not found"; fi
}

check_py() {
    if python3 -c "$1" >/dev/null 2>&1; then ok  "py:   $1"
    else                                     fail "py:   $1 (import failed)"; fi
}

printf '\n=== smoke-ctf ===\n\n'

# --- static binaries (downloader stage) ---
echo '-- static binaries'
check_cmd pspy
check_cmd linpeas
check_cmd busybox
check_cmd miniserve

# --- debuggers / analysis (apt) ---
echo '-- debuggers and analysis'
check_cmd gdb
check_cmd radare2
check_cmd ltrace
check_cmd strace
check_cmd binwalk

# --- forensics (apt) ---
echo '-- forensics'
check_cmd foremost
check_cmd steghide
check_cmd strings
check_cmd xxd

# --- Python tooling (pip) ---
echo '-- Python (pip)'
check_py 'import pwn'
check_py 'from Crypto.Cipher import AES'
check_py 'import gmpy2'
check_py 'import volatility3'
check_cmd ROPgadget

# --- pwndbg ---
echo '-- pwndbg'
check_dir /opt/pwndbg

# --- environment ---
echo '-- environment'
UID_ACTUAL="$(id -u)"
if [ "$UID_ACTUAL" -eq 1000 ]; then ok  "uid=1000 (kali)"
else                                 fail "uid=$UID_ACTUAL (expected 1000)"; fi

check_file "$HOME/.oh-my-zsh/custom/themes/kali.zsh-theme"
check_file "$HOME/.history"
check_file "$HOME/.tools/shell-upgrade.sh"
check_cmd  fd

printf '\n==> %d passed, %d failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ] && printf '[PASS] ctf\n\n' || { printf '[FAIL] ctf\n\n'; exit 1; }
