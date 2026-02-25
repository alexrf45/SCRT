#!/bin/bash
set -euo pipefail

base_system() {
  sudo apt-get install \
    -y --no-install-recommends \
    cmake \
    djvulibre-bin \
    file \
    iputils-ping \
    libpcap-dev \
    man-db \
    nfs-common \
    openssh-server \
    rpcbind \
    supervisor
}

base_desktop() {
  sudo apt-get install \
    -y --no-install-recommends \
    autocutsel \
    dbus-x11 \
    feh \
    firefox-esr \
    kali-themes \
    mousepad \
    x11-utils \
    xpdf
}

base_languages() {
  sudo apt-get install \
    -y --no-install-recommends \
    pipx \
    python3-pip \
    python3-virtualenv \
    ruby \
    ruby-dev \
    upx-ucl
}

base_tools() {
  sudo apt-get install \
    -y --no-install-recommends \
    hexcurse \
    ltrace \
    p7zip-full \
    rlwrap \
    traceroute
}

main_tools() {
  sudo apt-get install \
    -y --no-install-recommends \
    aria2 \
    bat \
    fzf \
    jq \
    mkcert \
    fastfetch \
    starship
}

network() {
  sudo apt-get install -y \
    --no-install-recommends \
    braa \
    dnsutils \
    ftp \
    iproute2 \
    masscan \
    mitmproxy \
    netcat-traditional \
    netdiscover \
    netexec \
    nmap \
    onesixtyone \
    proxychains \
    raven \
    snmp-mibs-downloader \
    snmpcheck \
    socat \
    swaks \
    tcpdump \
    telnet
}

active_directory() {
  sudo apt-get install -y \
    --no-install-recommends \
    bloodhound.py \
    enum4linux-ng \
    evil-winrm \
    ldap-utils \
    powershell \
    responder \
    smbclient
}

osint_tools() {
  sudo apt-get install -y \
    --no-install-recommends \
    cewl \
    csvtool \
    exiflooter \
    h8mail \
    sn0int \
    sqlitebrowser \
    vinetto
}

base_system
base_desktop
base_languages
base_tools
main_tools
network
active_directory
osint_tools

sudo apt-get clean && sudo rm -rf /var/lib/apt/lists/*

mkdir -p "$HOME/.local/bin" "$HOME/.logs" "$HOME/.tools" "$HOME/.proxychains"
