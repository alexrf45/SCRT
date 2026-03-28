# Security Research Container Toolkit

![GitHub Release](https://img.shields.io/github/v/release/alexrf45/SCRT?display_name=tag&style=plastic)
![Docker Image Size](https://img.shields.io/docker/image-size/fonalex45/scrt)
![Docker Pulls](https://img.shields.io/docker/pulls/fonalex45/scrt)

> **SCRT** — a disposable, flexible, and repeatable container environment for security researchers, analysts, and enthusiasts.

---

`SCRT` is a containerized security research environment designed for offensive and defensive operations.

Whether it's recon, exploitation, log analysis, or tool testing, `scrt` gives you consistent environments every time. Avoid dependency hell and save countless hours configuring the host OS. `scrt` runs locally, on a remote lab host, or in a Kubernetes cluster.

---

## Features

- Lightweight Kali base image with opinionated tool selection
- Persistent containers, volumes, and workspaces — pick up where you left off
- Custom Starship prompt with pre-configured Tmux status line
- GUI app support (Mousepad, SQLiteViewer) via X11 forwarding
- Firefox pre-configured with bookmarks, FoxyProxy, and Cookie Modifier
- Built-in command history with `fzf` for fast filtering (`Ctrl+r`)
- **Remote lab mode** — `scrt serve` exposes a REST API and web status dashboard
- **Environment-agnostic** — auto-detects Docker, containerd, or runc at startup

---

## How to run `scrt`

### Dependencies

- Docker (or containerd / runc for daemon-less operation)
- Linux x86-64 or arm64 (prebuilt binary) — or Go 1.24+ to build from source

### Install the CLI binary

```bash
# Download latest release
curl -L https://github.com/alexrf45/SCRT/releases/latest/download/scrt-linux-amd64 \
  -o ~/.local/bin/scrt
chmod +x ~/.local/bin/scrt
```

Full release history and changelogs: https://github.com/alexrf45/SCRT/releases

### Build from source

```bash
git clone https://github.com/alexrf45/SCRT.git
cd SCRT/scrt
make build        # produces bin/scrt
make all          # vet + test + build
```

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
  import <file> --repo <repo>         Import a backup tar as an image
  list                                List all SCRT containers (TUI or table)
  serve [--addr :8080] [--token tok]  Start HTTP API and web UI
  config                              Show current configuration
  config edit                         Open config in $EDITOR
  version                             Show version information

EXAMPLES:
  scrt start myproject
  scrt start myproject --image fonalex45/scrt:dev
  scrt backup myproject --dir ./my-backups
  scrt destroy myproject --force
  scrt serve --addr :8080 --token $(openssl rand -hex 32)
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
| `SCRT_TOKEN` | Bearer token for `scrt serve` API | auto-generated |

---

## Remote Lab Deployment

`scrt serve` starts an HTTP API and web status dashboard. Pair it with a
reverse proxy for TLS and you have a persistent remote research environment.

### Docker Compose (single host)

```bash
cd deploy
cp .env.example .env
# edit .env: set SCRT_TOKEN and update Caddyfile with your domain
docker compose up -d
```

The compose stack runs `scrt` in serve mode behind Caddy (automatic HTTPS via
ACME). The `scrt` port is never exposed directly — only Caddy reaches it.

```
internet ──► Caddy :443 (TLS) ──► scrt :8080
                                      ↕
                               /var/run/docker.sock
```

Files: [`deploy/compose.yaml`](deploy/compose.yaml), [`deploy/Caddyfile`](deploy/Caddyfile)

### Kubernetes

Deploys `scrt` as a `StatefulSet` on a dedicated, tainted security research
node with containerd socket passthrough.

> [!WARNING]
> The k8s manifest uses `privileged: true`. This grants full node access.
> Deploy only to a **dedicated, tainted** node — never on shared infrastructure.

```bash
# Label and taint a dedicated node
kubectl label node <node> role=security-research
kubectl taint node <node> security-research=:NoSchedule

# Inject token (never commit a real value to secret.yaml)
kubectl create secret generic scrt-token -n scrt \
  --from-literal=token=$(openssl rand -hex 32)

# Apply manifests
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/secret.yaml
kubectl apply -f deploy/k8s/statefulset.yaml
kubectl apply -f deploy/k8s/service.yaml
kubectl apply -f deploy/k8s/ingress.yaml    # requires cert-manager + nginx ingress
```

Update `deploy/k8s/statefulset.yaml` with the correct containerd socket path
for your distribution, and replace `lab.yourdomain.com` in
`deploy/k8s/ingress.yaml` with your domain.

| Distribution | Socket path |
|---|---|
| Standard k8s / EKS / GKE / AKS | `/run/containerd/containerd.sock` |
| k3s | `/run/k3s/containerd/containerd.sock` |
| Docker-based node | `/var/run/docker.sock` |

Files: [`deploy/k8s/`](deploy/k8s/)

### Web UI

Navigate to your domain after deployment. Enter the bearer token to connect.
The dashboard auto-refreshes every 10 seconds and supports stop, backup, and
destroy actions on running containers.

---

### Contributing

This project uses [Conventional Commits](https://www.conventionalcommits.org/) to drive automated semantic versioning via [release-please](https://github.com/googleapis/release-please).

| Prefix | Version bump | When to use |
|---|---|---|
| `feat:` | minor | New feature |
| `fix:` | patch | Bug fix |
| `feat!:` / `BREAKING CHANGE:` | major | Breaking change |
| `chore:`, `docs:`, `ci:` | none | No release |

Merging a `feat:` or `fix:` commit to `main` causes release-please to open a release PR. Merging that PR creates the version tag, which triggers the release pipeline — tests, binary upload, and Docker image push. The pipeline can also be re-triggered manually from **Actions → Release → Run workflow**.

> [!NOTE]
> Avoid Go-style function calls (e.g. `foo(bar)`) in commit message bodies — the release-please parser treats `(` as a scope delimiter and will silently skip the commit.

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
