#!/bin/sh
# smoke-web.sh — tool presence check for Dockerfile.web
# Runs inside the container via bind-mount: sh /tests/smoke-web.sh
# Accumulates all failures before exiting so you get a full picture.

PASS=0
FAIL=0

ok()   { printf '  ok   %s\n' "$*"; PASS=$((PASS + 1)); }
fail() { printf '  FAIL %s\n' "$*"; FAIL=$((FAIL + 1)); }

check_cmd() {
    if command -v "$1" >/dev/null 2>&1; then ok  "cmd: $1"
    else                                    fail "cmd: $1 not found"; fi
}

check_py() {
    if python3 -c "$1" >/dev/null 2>&1; then ok  "py:  $1"
    else                                     fail "py:  $1 (import failed)"; fi
}

check_dir() {
    if [ -d "$1" ]; then ok  "dir: $1"
    else                 fail "dir: $1 not found"; fi
}

check_file() {
    if [ -f "$1" ]; then ok  "file: $1"
    else                 fail "file: $1 not found"; fi
}

printf '\n=== smoke-web ===\n\n'

# --- ProjectDiscovery suite (static binaries from downloader stage) ---
echo '-- ProjectDiscovery'
check_cmd httpx
check_cmd subfinder
check_cmd nuclei
check_cmd katana
check_cmd dnsx
check_cmd naabu
check_cmd notify

# --- tomnomnom tools ---
echo '-- tomnomnom'
check_cmd httprobe
check_cmd waybackurls
check_cmd unfurl
check_cmd gron

# --- misc static tools ---
echo '-- misc static'
check_cmd ffuf
check_cmd rush
check_cmd miniserve
check_cmd go-dork

# --- Python tools (pip) ---
echo '-- Python (pip)'
check_cmd sqlmap
check_py  'import mitmproxy'

# --- system tools (apt) ---
echo '-- system'
check_cmd exiftool
check_cmd curl
check_cmd wget
check_cmd jq
check_cmd fzf
check_cmd git
check_cmd tmux

# --- environment ---
echo '-- environment'
UID_ACTUAL="$(id -u)"
if [ "$UID_ACTUAL" -eq 1000 ]; then ok  "uid=1000 (kali)"
else                                 fail "uid=$UID_ACTUAL (expected 1000)"; fi

check_dir  "$HOME/.config/nuclei"
check_dir  "$HOME/.config/subfinder"
check_file "$HOME/.oh-my-zsh/custom/themes/kali.zsh-theme"
check_file "$HOME/.history"
check_file "$HOME/.tools/shell-upgrade.sh"
check_cmd  fd

printf '\n==> %d passed, %d failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ] && printf '[PASS] web\n\n' || { printf '[FAIL] web\n\n'; exit 1; }
