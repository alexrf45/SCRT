#!/bin/bash

web() {
  sudo apt-get install -y \
    --no-install-recommends \
    burpsuite \
    default-mysql-client \
    exiftool \
    ffuf \
    hurl \
    postgresql \
    sqlmap \
    sqsh \
    whatweb
}

web-server() {
  latest-release() {
    curl -s https://api.github.com/repos/svenstaro/miniserve/releases/latest |
      grep "browser_download_url" | grep "x86_64-unknown-linux-gnu" | cut -d '"' -f 4
  }

  release_url=$(latest-release)

  curl -Lo "$HOME/.local/bin/miniserve" "$release_url"

  chmod +x "$HOME/.local/bin/miniserve"
}

password() {
  sudo apt-get install -y \
    --no-install-recommends \
    crunch
}

payload() {
  cd "$HOME/.tools/" &&
    wget -q -O nc.exe \
      "https://github.com/ShutdownRepo/Exegol-resources/raw/main/windows/nc.exe" &&
    wget -q -O nc \
      "https://github.com/andrew-d/static-binaries/raw/master/binaries/linux/x86_64/ncat"
}

active_directory() {
  cd "$HOME/.tools/" &&
    wget -q -O rubeus.exe \
      "https://github.com/r3motecontrol/Ghostpack-CompiledBinaries/raw/master/Rubeus.exe" &&
    wget -q -O certify.exe \
      "https://github.com/r3motecontrol/Ghostpack-CompiledBinaries/raw/master/Certify.exe" &&
    wget -q -O sharp.ps1 \
      "https://github.com/BloodHoundAD/BloodHound/raw/master/Collectors/SharpHound.ps1" &&
    wget -q -O SharpHound.exe \
      "https://raw.githubusercontent.com/BloodHoundAD/BloodHound/master/Collectors/SharpHound.exe"
}

pivot() {
  cd "$HOME/.tools/" &&
    wget -q -O chisel.gz \
      "https://github.com/jpillora/chisel/releases/download/v1.10.1/chisel_1.10.0_linux_amd64.gz" &&
    gunzip chisel.gz &&
    wget -q -O win-chisel.gz \
      "https://github.com/jpillora/chisel/releases/download/v1.10.1/chisel_1.10.0_windows_amd64.gz" &&
    gunzip win-chisel.gz
}

privesc() {
  cd "$HOME/.tools/" &&
    wget -q -O linpeas \
      "https://github.com/peass-ng/PEASS-ng/releases/latest/download/linpeas.sh" &&
    wget -q -O winpeas.exe \
      "https://github.com/peass-ng/PEASS-ng/releases/download/20250701-bdcab634/winPEASx64_ofs.exe" &&
    wget -q -O pspys \
      "https://github.com/DominicBreuker/pspy/releases/download/v1.2.1/pspy64s" &&
    wget -q -O pspy \
      "https://github.com/DominicBreuker/pspy/releases/download/v1.2.1/pspy64"
}

extra() {
  cd "$HOME/.tools/" &&
    git clone https://github.com/samratashok/nishang.git nishang &&
    git clone https://github.com/gustanini/PowershellTools.git powershell-tools &&
    git clone https://github.com/aniqfakhrul/powerview.py powerview
}

web
web-server
password
payload
active_directory
pivot
privesc
extra

wget -q -O "$HOME/.local/bin/busybox" \
  "https://busybox.net/downloads/binaries/1.35.0-x86_64-linux-musl/busybox"

wget -q -O "$HOME/.local/bin/nvim" \
  "https://github.com/neovim/neovim/releases/download/stable/nvim-linux-x86_64.appimage" &&
  chmod +x "$HOME/.local/bin/nvim"
