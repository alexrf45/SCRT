# Security Research Container Toolkit

![GitHub Release](https://img.shields.io/github/v/release/alexrf45/SCRT?display_name=tag&style=plastic)
![Docker Image Size](https://img.shields.io/docker/image-size/fonalex45/scrt)
![Docker Pulls](https://img.shields.io/docker/pulls/fonalex45/scrt)

> **SCRT** — a disposable, flexible, and repeatable container environment for security researchers, analysts, and enthusiasts.

---

`SCRT` is a containerized security research environment designed for offensive and defensive operations.

Whether it's recon, exploitation, log analysis, or tool testing, `scrt` gives you consistent environments every time. Avoid dependency hell and save countless hours configuring the host OS. `scrt` runs locally or in the cloud.

> [!NOTE]
> The container image has not been tested on Kubernetes.

---

## Features

- Lightweight Kali base image with opinionated tool selection
- Persistent containers, volumes, and workspaces — pick up where you left off
- Custom Starship prompt with pre-configured Tmux status line
- GUI app support (Mousepad, SQLiteViewer) via X11 forwarding
- Firefox pre-configured with bookmarks, FoxyProxy, and Cookie Modifier
- Built-in command history with `fzf` for fast filtering (`Ctrl+r`)

---

## How to run `scrt`

### Dependencies

- Docker
- Bash or ZSH

---

```
USAGE:
  scrt <command> [arguments]

COMMANDS:
  start <project> [--image <image>]   Start a new container
  enter <project>                     Enter a running container
  stop <project>                      Stop a container
  destroy <project> [--force]         Destroy container and data
  backup <project> [--dir <path>]     Backup project data
  pull [--image <image>]              Pull/update container image
  list                                List all SCRT containers
  config                              Show current configuration
  config edit                         Open config in $EDITOR
  version                             Show version information

EXAMPLES:
  scrt start myproject
  scrt start myproject --image fonalex45/scrt:dev
  scrt backup myproject --dir ./my-backups
  scrt destroy myproject --force
  scrt config edit
```

### Configuration

SCRT loads configuration from `~/.scrt.conf.json`. View or edit:

```bash
scrt config        # display current settings
scrt config edit   # open in $VISUAL / $EDITOR / vi
```

| Variable | Description | Default |
|---|---|---|
| `SCRT_IMAGE` | Docker image to use | `fonalex45/scrt:latest` |
| `SCRT_SHELL` | Shell inside container | `/bin/zsh` |
| `SCRT_HOST_NET` | Set to `false` to disable host networking | `true` |
| `SCRT_X11` | Set to `false` to disable X11 forwarding | `true` |
| `SCRT_GPU` | Set to `false` to disable GPU passthrough | `true` |
| `SCRT_WORKDIR` | Base working directory | current directory |

---

### Contributing

This project uses [Conventional Commits](https://www.conventionalcommits.org/) to drive automated semantic versioning.

| Prefix | Version bump | When to use |
|---|---|---|
| `feat:` | minor | New feature |
| `fix:` | patch | Bug fix |
| `feat!:` / `BREAKING CHANGE:` | major | Breaking change |
| `chore:`, `docs:`, `ci:` | none | No release |

---

### Custom aliases included

```bash
# daily use
alias c='clear'
alias t='tmux new -f ~/.tmux.conf -s $1'
alias i='sudo apt install -y'
alias q='exit'
alias r='. ~/.zshrc'
alias update='sudo apt update'
alias upgrade='sudo apt upgrade'
alias get="curl -O -L"
alias cat='batcat'
alias weather='curl https://wttr.in'
alias public='curl wtfismyip.com/text'
alias download='aria2c'
alias home='cd ~'

# pentesting aliases
alias cme='nxc'
alias port-scan='sudo nmap -sC -sV -p- $IP > scan.txt'
alias udp-scan='sudo nmap -sU --top-ports 10 $IP -v > udp.scan.txt'
alias stealth-scan='sudo nmap --data-length 6 -T3 -A -ttl 64 -p- $IP > stealth-scan.txt'
alias proxy='proxychains'
alias serve='sudo python3 -m http.server 8888'
alias notepad='mousepad notes.md > /dev/null 2>&1 &'

# python3
alias py-virt='python3 -m venv .venv && source .venv/bin/activate'
alias freeze='pip freeze > requirements.txt'
alias py-install='pip install -r requirements.txt'
alias py-list='pipx list | grep package'
```
