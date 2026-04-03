#!/bin/bash
set -euo pipefail

base_system() {
  sudo apt-get install \
    -y --no-install-recommends \
    file \
    iputils-ping \
    libpcap0.8
}

base_languages() {
  sudo apt-get install \
    -y --no-install-recommends \
    python3-pip \
    python3-venv \
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
    fastfetch \
    fzf \
    jq
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

sudo apt-get update

base_system
base_languages
base_tools
main_tools
network
active_directory
osint_tools

mkdir -p "$HOME/.local/bin" "$HOME/.logs" "$HOME/.tools" "$HOME/.proxychains"
