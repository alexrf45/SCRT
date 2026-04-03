#!/bin/bash
set -euo pipefail

miniserve_latest_url() {
  curl -s https://api.github.com/repos/svenstaro/miniserve/releases/latest |
    grep "browser_download_url" | grep "x86_64-unknown-linux-gnu" | cut -d '"' -f 4
}

web() {
  (
    cd "$HOME/.tools"

    # ffuf — web fuzzer
    wget -q -O ffuf.tgz \
      "https://github.com/ffuf/ffuf/releases/download/v2.1.0/ffuf_2.1.0_linux_amd64.tar.gz"
    tar -xzf ffuf.tgz ffuf
    mv ffuf "$HOME/.local/bin/ffuf"
    chmod +x "$HOME/.local/bin/ffuf"
    rm ffuf.tgz

    # hurl — HTTP runner
    wget -q -O hurl.tgz \
      "https://github.com/Orange-OpenSource/hurl/releases/download/7.1.0/hurl-7.1.0-x86_64-unknown-linux-gnu.tar.gz"
    tar -xzf hurl.tgz
    mv hurl-*/bin/hurl "$HOME/.local/bin/hurl"
    chmod +x "$HOME/.local/bin/hurl"
    rm -rf hurl-* hurl.tgz

    # sqlmap — SQL injection testing
    git clone --depth 1 https://github.com/sqlmapproject/sqlmap.git sqlmap
    chmod +x sqlmap/sqlmap.py
    ln -sf "$HOME/.tools/sqlmap/sqlmap.py" "$HOME/.local/bin/sqlmap"

    # whatweb — web fingerprinting
    git clone --depth 1 https://github.com/urbanadventurer/WhatWeb.git whatweb
    chmod +x whatweb/whatweb
    ln -sf "$HOME/.tools/whatweb/whatweb" "$HOME/.local/bin/whatweb"

    # exiftool — metadata extraction (Perl; FindBin resolves lib/ from symlink target)
    wget -q -O exiftool.tgz \
      "https://github.com/exiftool/exiftool/archive/refs/tags/13.54.tar.gz"
    tar -xzf exiftool.tgz
    mv exiftool-* exiftool
    chmod +x exiftool/exiftool
    ln -sf "$HOME/.tools/exiftool/exiftool" "$HOME/.local/bin/exiftool"
    rm exiftool.tgz
  )

  # MySQL and PostgreSQL CLI clients — drop-in replacements for mysql/psql
  pip3 install --no-cache-dir --break-system-packages mycli pgcli
}

web_server() {
  curl -Lo "$HOME/.local/bin/miniserve" "$(miniserve_latest_url)"
  chmod +x "$HOME/.local/bin/miniserve"
}

extras() {
  # mkcert — local TLS cert generator (used by cert-gen() in functions.sh)
  wget -q -O "$HOME/.local/bin/mkcert" \
    "https://github.com/FiloSottile/mkcert/releases/download/v1.4.4/mkcert-v1.4.4-linux-amd64"
  chmod +x "$HOME/.local/bin/mkcert"
}

password() {
  sudo apt-get install -y \
    --no-install-recommends \
    crunch
}

payload() {
  (
    cd "$HOME/.tools"
    wget -q -O nc.exe \
      "https://github.com/ShutdownRepo/Exegol-resources/raw/main/windows/nc.exe"
    wget -q -O nc \
      "https://github.com/andrew-d/static-binaries/raw/master/binaries/linux/x86_64/ncat"
    chmod +x nc
  )
}

active_directory() {
  (
    cd "$HOME/.tools"
    wget -q -O rubeus.exe \
      "https://github.com/r3motecontrol/Ghostpack-CompiledBinaries/raw/master/Rubeus.exe"
    wget -q -O certify.exe \
      "https://github.com/r3motecontrol/Ghostpack-CompiledBinaries/raw/master/Certify.exe"
  )
}

pivot() {
  (
    cd "$HOME/.tools"
    wget -q -O chisel.gz \
      "https://github.com/jpillora/chisel/releases/download/v1.10.1/chisel_1.10.1_linux_amd64.gz"
    gunzip chisel.gz
    chmod +x chisel
    wget -q -O win-chisel.gz \
      "https://github.com/jpillora/chisel/releases/download/v1.10.1/chisel_1.10.1_windows_amd64.gz"
    gunzip win-chisel.gz
  )
}

privesc() {
  (
    cd "$HOME/.tools"
    wget -q -O linpeas \
      "https://github.com/peass-ng/PEASS-ng/releases/latest/download/linpeas.sh"
    chmod +x linpeas
    wget -q -O winpeas.exe \
      "https://github.com/peass-ng/PEASS-ng/releases/latest/download/winPEASx64_ofs.exe"
    wget -q -O pspys \
      "https://github.com/DominicBreuker/pspy/releases/download/v1.2.1/pspy64s"
    chmod +x pspys
    wget -q -O pspy \
      "https://github.com/DominicBreuker/pspy/releases/download/v1.2.1/pspy64"
    chmod +x pspy
  )
}

extra() {
  (
    cd "$HOME/.tools"
    git clone --depth 1 https://github.com/samratashok/nishang.git nishang
    git clone --depth 1 https://github.com/gustanini/PowershellTools.git powershell-tools
    git clone --depth 1 https://github.com/aniqfakhrul/powerview.py powerview
  )
}

sudo apt-get update

web
web_server
extras
password
payload
active_directory
pivot
privesc
extra

wget -q -O "$HOME/.local/bin/busybox" \
  "https://busybox.net/downloads/binaries/1.35.0-x86_64-linux-musl/busybox"
chmod +x "$HOME/.local/bin/busybox"

wget -q -O nvim.tar.gz \
  "https://github.com/neovim/neovim/releases/latest/download/nvim-linux-x86_64.tar.gz"
tar -xzf nvim.tar.gz
mv nvim-linux-x86_64/bin/nvim "$HOME/.local/bin/nvim"
rm -rf nvim.tar.gz nvim-linux-x86_64
