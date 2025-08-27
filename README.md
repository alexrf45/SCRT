# toolkit

![GitHub Release](https://img.shields.io/github/v/release/alexrf45/toolkit?display_name=tag&style=plastic)
![Logo](https://img.shields.io/docker/image-size/fonalex45/toolkit) ![Logo](https://img.shields.io/docker/pulls/fonalex45/gr3ysh3ll)

> **toolkit** — a disposable, flexible, and repeatable container environment for security researchers, analysts, and enthusiasts.  
> **Launch anywhere. Burn after use. Repeat.**

---

## What is toolkit?

**toolkit** is a containerized security research environment designed for offensive and defensive operations. Whether you're doing recon, exploitation, analysis, or tool testing, `toolkit` gives you:

- **Containerized environments** — consistent environments every time
- **Burnable instances** — throwaway containers that keep the host OS clean
- ☁️**Portable deployments** — run it locally, in the cloud or kubernetes
  \*\*> [!NOTE]
  > container image has not been tested on K8s.

---

## Features

- Lightweight base image with common pentesting utilities
- Persistant containers, volumes and workspaces
- ZSH-powered shell with rich prompt
- Pre-configured Tmux configuration
- GUI apps such as burpsuite and firefox
- Bash script executable
- Cloud-ready: works on any cloud providers

---

## 🛠️ Getting Started

### 🔧 Requirements

- Docker
- (Optional) Docker Compose
- Bash or ZSH

### 📥 Pull the Image

```bash
docker pull fonalex45/toolkit:latest

```

### Custom aliases included

```bash
alias cme='nxc'
alias port-scan='sudo nmap -sC -sV -p- $IP > scan.txt'
alias udp-scan='sudo nmap -sU --top-ports 10 $IP -v > udp.scan.txt'
alias stealth-scan='sudo nmap --data-length 6 -T3 -A -ttl 64 -p- $IP > stealth-scan.txt'
alias public='curl wtfismyip.com/text'
alias t='tmux new -f ~/.tmux.conf -s $1'
alias :q='exit'
alias home='cd ~'
alias :r='. ~/.bashrc'
alias update='sudo apt update'
alias upgrade='sudo apt upgrade -y'
alias i='sudo apt install -y'
alias ls='ls --color=auto'
alias command='cat $HOME/.commands'
alias proxy='proxychains'
alias serve='sudo python3 -m http.server 80'
```

## Command history

- Useful commands are already built into the container history. Simple type `CTRL+r' to pull up the fzf window and filter for commands. fzf makes navigating commands and files a breeze.

### 🤘 Contributing

Have an idea, bug, or tool request? Open an issue or submit a PR.
