# SCRT — Security Research Container Toolkit

![GitHub Release](https://img.shields.io/github/v/release/alexrf45/SCRT?display_name=tag&style=plastic)
![Docker Image Size](https://img.shields.io/docker/image-size/fonalex45/scrt)
![Docker Pulls](https://img.shields.io/docker/pulls/fonalex45/scrt)

> A disposable, flexible, and repeatable container environment for security researchers, analysts, and enthusiasts.

SCRT is a CLI wrapper around Docker that provides project-isolated security research environments. Whether it's recon, exploitation, log analysis, or tool testing, `scrt` gives you consistent environments with persistent workspaces, X11 forwarding, GPU passthrough, and host networking — all from a single command.

## Features

- Lightweight Kali base image with opinionated tool selection
- Persistent containers, volumes, and workspaces — pick up where you left off
- Custom Starship prompt with pre-configured Tmux status line
- GUI app support (Burpsuite, Wireshark, Firefox) via X11 forwarding
- Firefox pre-configured with bookmarks, FoxyProxy, and Cookie Modifier
- Built-in command history with `fzf` for fast filtering (`Ctrl+r`)
- Project directory scaffolding (recon, exploit, pivot, privesc, report)
- Compressed project backups with timestamps

## Dependencies

- Docker (with BuildKit for `docker-build` targets)
- Bash or ZSH

## Installation

### Option 1: Docker Build (recommended)

No local Go toolchain required. Uses a clean, reproducible Docker build environment:

```bash
git clone https://github.com/alexrf45/SCRT.git
cd SCRT/scrt

# Build the binary — outputs to bin/scrt
make docker-build

# Copy to PATH
cp bin/scrt ~/.local/bin/
```

### Option 2: Local Go Build

Requires Go 1.24+:

```bash
git clone https://github.com/alexrf45/SCRT.git
cd SCRT/scrt

go mod tidy
make all        # vet → test → build
make install    # copies bin/scrt to ~/.local/bin/
```

## Usage

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
  config                              Create default configuration
  version                             Show version information

EXAMPLES:
  scrt start myproject
  scrt start myproject --image fonalex45/scrt:dev
  scrt backup myproject --dir ./my-backups
  scrt destroy myproject --force
```

## Configuration

SCRT loads configuration from `~/.scrt.conf.json`. Generate a default config:

```bash
scrt config
```

Settings can be overridden via environment variables:

| Variable | Description | Default |
|---|---|---|
| `SCRT_IMAGE` | Docker image to use | `fonalex45/scrt:latest` |
| `SCRT_SHELL` | Shell inside container | `/bin/zsh` |
| `SCRT_HOST_NET` | Enable host networking | `true` |
| `SCRT_X11` | Enable X11 forwarding | `true` |
| `SCRT_GPU` | Enable GPU passthrough | `true` |
| `SCRT_CAPS` | Linux capabilities (comma-separated) | `NET_ADMIN,CAP_SYS_TIME` |
| `SCRT_MOUNTS` | Extra mounts (comma-separated) | — |
| `SCRT_WORKDIR` | Base working directory | current directory |

## Build Targets

| Target | Description |
|---|---|
| `make all` | vet → test → build (local toolchain) |
| `make build` | Compile binary with version injection |
| `make test` | Run tests with race detector |
| `make vet` | Static analysis |
| `make lint` | Run golangci-lint |
| `make docker-build` | Build binary inside Docker (no local Go needed) |
| `make docker-test` | Run vet + tests inside Docker |
| `make docker-all` | Full CI pipeline in Docker |
| `make install` | Build and copy to `~/.local/bin/` |
| `make clean` | Remove build artifacts |

## Project Structure

```
scrt/
├── cmd/scrt/           # CLI entrypoint (main + cobra commands)
│   └── main.go
├── internal/
│   ├── config/         # Configuration loading, validation, env overrides
│   ├── container/      # Docker CLI operations (lifecycle, builder, labels)
│   └── project/        # Project directory scaffolding
├── Dockerfile.build    # Reproducible build environment
├── Makefile
├── go.mod
└── go.sum
```

## Project Workspace

Each `scrt start <project>` creates:

```
<project>/
├── recon/
├── www/
├── exploit/
├── pivot/
├── privesc/
├── report/
└── .scrt-logs/
```

## Architecture

SCRT is written in Go using `os/exec` to shell out to the Docker CLI for container operations and Cobra for CLI parsing. The codebase follows these principles:

- **Config as code**: JSON config loaded once at startup, validated, passed explicitly — no globals (CFG-1, CFG-2)
- **Structured logging**: `slog` with consistent fields (OBS-1)
- **Sentinel errors**: Package-level error types with `errors.Is`/`errors.As` for control flow (ERR-2, ERR-3)
- **Table-driven tests**: All packages tested with race detector enabled (T-1, G-3)
- **Reproducible builds**: `-trimpath` with version injection via `-ldflags` (CI-2)
- **Minimal dependencies**: stdlib preferred; Cobra is the only external dep — Docker operations use `os/exec` (MD-1)

## Container Image

The SCRT container image (`fonalex45/scrt`) is a Kali Linux base with curated security tooling. See `sources/` for the full tool list. Key categories:

- **Network**: nmap, masscan, netexec, tcpdump, socat, mitmproxy
- **Web**: Burpsuite, ffuf, sqlmap, whatweb
- **Active Directory**: bloodhound.py, evil-winrm, responder, smbclient
- **OSINT**: cewl, sn0int, exiflooter, h8mail
- **Privesc**: linpeas, winpeas, pspy, chisel
- **Desktop**: Firefox, Mousepad, Wireshark (via X11)

## License

See [LICENSE](LICENSE) for details.
