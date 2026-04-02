#!/bin/sh
# smoke-ad.sh — tool presence check for Dockerfile.ad
# Runs inside the container via bind-mount: sh /tests/smoke-ad.sh

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

printf '\n=== smoke-ad ===\n\n'

# --- static binaries (downloader stage) ---
echo '-- static binaries'
check_cmd kerbrute
check_cmd chisel
check_cmd miniserve
check_cmd pspy
check_cmd pspys
check_cmd linpeas

# --- system tools (apt) ---
echo '-- system tools'
check_cmd nmap
check_cmd smbclient
check_cmd ldapsearch
check_cmd netcat

# --- AD tooling (kali apt + pip) ---
echo '-- AD tools'
check_cmd nxc
check_cmd evil-winrm
check_cmd enum4linux-ng
check_py 'import impacket'
check_py 'import impacket.smbconnection'

# --- Windows payloads ---
echo '-- Windows payloads'
check_file /opt/windows/rubeus.exe
check_file /opt/windows/certify.exe
check_file /opt/smbserver.py

# --- PowerShell frameworks ---
echo '-- PowerShell frameworks'
check_dir /opt/ps/nishang
check_dir /opt/ps/powerview

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
[ "$FAIL" -eq 0 ] && printf '[PASS] ad\n\n' || { printf '[FAIL] ad\n\n'; exit 1; }
