#!/bin/sh
# smoke-base.sh — tool presence check for Dockerfile (base Kali image)
# Runs inside the container via bind-mount: sh /tests/smoke-base.sh
# Accumulates all failures before exiting so you get a full picture.

PASS=0
FAIL=0

ok()   { printf '  ok   %s\n' "$*"; PASS=$((PASS + 1)); }
fail() { printf '  FAIL %s\n' "$*"; FAIL=$((FAIL + 1)); }

check_cmd() {
    if command -v "$1" >/dev/null 2>&1; then ok  "cmd:  $1"
    else                                    fail "cmd:  $1 not found"; fi
}

check_py() {
    if python3 -c "$1" >/dev/null 2>&1; then ok  "py:   $1"
    else                                     fail "py:   $1 (import failed)"; fi
}

check_dir() {
    if [ -d "$1" ]; then ok  "dir:  $1"
    else                 fail "dir:  $1 not found"; fi
}

check_file() {
    if [ -f "$1" ]; then ok  "file: $1"
    else                 fail "file: $1 not found"; fi
}

check_absent() {
    if ! command -v "$1" >/dev/null 2>&1; then ok  "absent: $1"
    else                                      fail "absent: $1 should NOT be present"; fi
}

# User-level binaries are installed to ~/.local/bin which is not in the
# default PATH for non-login sh. Add it so command -v finds them.
export PATH="/home/kali/.local/bin:$PATH"

printf '\n=== smoke-base ===\n\n'

# --- environment ---
echo '-- environment'
UID_ACTUAL="$(id -u)"
if [ "$UID_ACTUAL" -eq 1000 ]; then ok  "uid=1000 (kali)"
else                                 fail "uid=$UID_ACTUAL (expected 1000)"; fi

check_file "$HOME/.oh-my-zsh/custom/themes/kali.zsh-theme"
check_file "$HOME/.history"
check_file "$HOME/.tmux.conf"
check_file "$HOME/.proxychains/proxychains.conf"
check_file "$HOME/.zshrc"

# starship must be functional (not just present)
if starship init zsh >/dev/null 2>&1; then ok  "starship init zsh"
else                                      fail "starship init zsh failed"; fi

# --- core apt tools ---
echo '-- core tools'
check_cmd curl
check_cmd wget
check_cmd git
check_cmd tmux
check_cmd vim
check_cmd nano
check_cmd python3
check_cmd pip3
check_cmd ruby
check_cmd jq
check_cmd fzf
check_cmd batcat
check_cmd 7z
check_cmd rlwrap
check_cmd traceroute
check_cmd aria2c
check_cmd fastfetch
check_cmd upx

# --- network / pentest apt tools ---
echo '-- network tools'
check_cmd nmap
check_cmd masscan
check_cmd tcpdump
check_cmd socat
check_cmd nc
check_cmd proxychains
check_cmd dig
check_cmd netdiscover
check_cmd braa
check_cmd onesixtyone
check_cmd swaks
check_cmd ftp
check_cmd telnet

# --- AD / pentest apt tools ---
echo '-- AD tools'
check_cmd nxc
check_cmd evil-winrm
check_cmd enum4linux-ng
check_cmd ldapsearch
check_cmd smbclient
check_cmd responder
check_py  'import bloodhound'

# --- OSINT tools ---
echo '-- OSINT tools'
check_cmd cewl
check_cmd csvtool
check_cmd sn0int

# --- binary tools (~/.local/bin) ---
echo '-- binary tools'
check_cmd ffuf
check_cmd hurl
check_cmd miniserve
check_file "$HOME/.tools/chisel"
check_cmd busybox
check_cmd nvim
check_cmd sqlmap
check_cmd exiftool
check_cmd mkcert
check_cmd mycli

# --- tool files ---
echo '-- tool files'
check_file "$HOME/.tools/linpeas"
check_file "$HOME/.tools/pspy"
check_file "$HOME/.tools/rubeus.exe"
check_dir  "$HOME/.tools/nishang"
check_dir  "$HOME/.tools/sqlmap"

# --- regression guards: these must NOT be present ---
echo '-- regression guards'
check_absent firefox
check_absent mousepad
check_absent supervisord
check_absent cmake

printf '\n==> %d passed, %d failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ] && printf '[PASS] base\n\n' || { printf '[FAIL] base\n\n'; exit 1; }
