#!/bin/bash
set -euo pipefail

httprobe_install() {
  wget -q -O httprobe.tgz \
    "https://github.com/tomnomnom/httprobe/releases/download/v0.2/httprobe-linux-amd64-0.2.tgz"
  tar -xzf httprobe.tgz
  chmod +x httprobe
  mv httprobe "$HOME/.local/bin/httprobe"
  rm httprobe.tgz
}

httpx_install() {
  wget -q "https://github.com/projectdiscovery/httpx/releases/download/v1.3.6/httpx_1.3.6_linux_amd64.zip"
  unzip -q httpx_1.3.6_linux_amd64.zip -d httpx
  chmod +x httpx/httpx
  mv httpx/httpx "$HOME/.local/bin/http-x"
  rm -rf httpx httpx_1.3.6_linux_amd64.zip
}

go_dorks_install() {
  wget -q -O go-dork \
    "https://github.com/dwisiswant0/go-dork/releases/download/v1.0.2/go-dork_1.0.2_linux_amd64"
  chmod +x go-dork
  mv go-dork "$HOME/.local/bin/go-dork"
}

gron_install() {
  wget -q "https://github.com/tomnomnom/gron/releases/download/v0.7.1/gron-linux-amd64-0.7.1.tgz"
  tar -xzf gron-linux-amd64-0.7.1.tgz
  chmod +x gron
  mv gron "$HOME/.local/bin/gron"
  rm gron-linux-amd64-0.7.1.tgz
}

rush_install() {
  wget -q -O rush.tar.gz \
    "https://github.com/shenwei356/rush/releases/download/v0.5.2/rush_linux_amd64.tar.gz"
  tar -xzf rush.tar.gz
  chmod +x rush
  mv rush "$HOME/.local/bin/rush"
  rm rush.tar.gz
}

katana_install() {
  wget -q -O katana.zip \
    "https://github.com/projectdiscovery/katana/releases/download/v1.0.4/katana_1.0.4_linux_amd64.zip"
  unzip -q katana.zip katana
  chmod +x katana
  mv katana "$HOME/.local/bin/katana"
  rm katana.zip
}

chaos_install() {
  wget -q -O chaos.zip \
    "https://github.com/projectdiscovery/chaos-client/releases/download/v0.5.1/chaos-client_0.5.1_linux_amd64.zip"
  unzip -q chaos.zip chaos-client
  chmod +x chaos-client
  mv chaos-client "$HOME/.local/bin/chaos"
  rm chaos.zip
}

dnsx_install() {
  wget -q -O dnsx.zip \
    "https://github.com/projectdiscovery/dnsx/releases/download/v1.1.4/dnsx_1.1.4_linux_amd64.zip"
  unzip -q dnsx.zip dnsx
  chmod +x dnsx
  mv dnsx "$HOME/.local/bin/dnsx"
  rm dnsx.zip
}

waybackurls_install() {
  wget -q -O waybackurls.tgz \
    "https://github.com/tomnomnom/waybackurls/releases/download/v0.1.0/waybackurls-linux-amd64-0.1.0.tgz"
  tar -xzf waybackurls.tgz
  chmod +x waybackurls
  mv waybackurls "$HOME/.local/bin/waybackurls"
  rm waybackurls.tgz
}

unfurl_install() {
  wget -q -O unfurl.tgz \
    "https://github.com/tomnomnom/unfurl/releases/download/v0.4.3/unfurl-linux-amd64-0.4.3.tgz"
  tar -xzf unfurl.tgz
  chmod +x unfurl
  mv unfurl "$HOME/.local/bin/unfurl"
  rm unfurl.tgz
}

subfinder_install() {
  wget -q -O subfinder.zip \
    "https://github.com/projectdiscovery/subfinder/releases/download/v2.6.3/subfinder_2.6.3_linux_amd64.zip"
  unzip -q subfinder.zip subfinder
  chmod +x subfinder
  mv subfinder "$HOME/.local/bin/subfinder"
  rm subfinder.zip
}

notify_install() {
  wget -q -O notify.zip \
    "https://github.com/projectdiscovery/notify/releases/download/v1.0.5/notify_1.0.5_linux_amd64.zip"
  unzip -q notify.zip notify
  chmod +x notify
  mv notify "$HOME/.local/bin/notify"
  rm notify.zip
}

sudo apt-get update
sudo apt-get install -y --no-install-recommends amass
sudo apt-get clean && sudo rm -rf /var/lib/apt/lists/*

httprobe_install
httpx_install
go_dorks_install
gron_install
rush_install
katana_install
chaos_install
dnsx_install
waybackurls_install
unfurl_install
subfinder_install
notify_install
