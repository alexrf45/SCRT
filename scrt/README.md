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
- GUI app support (Mousepad, SQLiteViewer, Firefox) via X11 forwarding
- Firefox pre-configured with bookmarks, FoxyProxy, and Cookie Modifier
- Built-in command history with `fzf` for fast filtering (`Ctrl+r`)
- Project directory scaffolding (recon, exploit, pivot, privesc, report)
- Compressed project backups with timestamps
- Color terminal output via [Charm](https://charm.land) (lipgloss + bubbles)

## Dependencies

- Docker
- Linux x86-64 (for the prebuilt binary) or Go 1.24+ (to build from source)

## Installation

### Option 1: Download prebuilt binary (recommended)

Grab the latest release directly from GitHub:

```bash
curl -L https://github.com/alexrf45/SCRT/releases/latest/download/scrt-linux-amd64 \
  -o ~/.local/bin/scrt
chmod +x ~/.local/bin/scrt
```

Or browse all releases: https://github.com/alexrf45/SCRT/releases

### Option 2: Docker Build

No local Go toolchain required. Uses a clean, reproducible Docker build environment:

```bash
git clone https://github.com/alexrf45/SCRT.git
cd SCRT/scrt

# Build the binary — outputs to bin/scrt
make docker-build

# Copy to PATH
cp bin/scrt ~/.local/bin/
```

### Option 3: Local Go Build

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

## Configuration

SCRT loads configuration from `~/.scrt.conf.json`. View or edit:

```bash
scrt config          # display current settings
scrt config edit     # open in $VISUAL / $EDITOR / vi
```

If no config file exists, `config edit` seeds it with defaults before opening.

Settings can also be overridden at runtime via environment variables:

| Variable | Description | Default |
|---|---|---|
| `SCRT_IMAGE` | Docker image to use | `fonalex45/scrt:latest` |
| `SCRT_SHELL` | Shell inside container | `/bin/zsh` |
| `SCRT_HOST_NET` | Set to `false` to disable host networking | `true` |
| `SCRT_X11` | Set to `false` to disable X11 forwarding | `true` |
| `SCRT_GPU` | Set to `false` to disable GPU passthrough | `true` |
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
├── cmd/scrt/           # CLI entrypoint (Cobra commands)
│   ├── main.go         # Command definitions and wiring
│   └── style.go        # Shared lipgloss styles and render helpers
├── internal/
│   ├── config/         # Configuration loading, validation, env overrides
│   ├── container/      # Docker CLI operations (lifecycle, builder, labels)
│   └── project/        # Project directory scaffolding and backup
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
- **Structured logging**: [`charmbracelet/log`](https://github.com/charmbracelet/log) with colored level badges; no timestamps in CLI output (OBS-1)
- **Sentinel errors**: Package-level error types with `errors.Is`/`errors.As` for control flow (ERR-2, ERR-3)
- **Table-driven tests**: All packages tested with race detector enabled (T-1, G-3)
- **Reproducible builds**: `-trimpath` with version injection via `-ldflags` (CI-2)
- **Terminal UI**: [lipgloss](https://github.com/charmbracelet/lipgloss) for styling and [bubbles](https://github.com/charmbracelet/bubbles) for the container list table

## Roadmap

### v3.0.0 (next major release)

v3.0.0 is the first release with full software development fundamentals in place. The groundwork laid in v2.x:

| Area | Status |
|---|---|
| Go binary rewrite (from Python) | ✅ Complete |
| Charm TUI — color output, styled list, config display | ✅ Complete |
| `config edit` — open config in `$EDITOR` | ✅ Complete |
| Conventional Commits + automated semantic versioning | ✅ Complete |
| CI pipeline — `go vet` + `go test -race` on every PR | ✅ Complete |
| Automated GitHub Releases with prebuilt linux/amd64 binary | ✅ Complete |
| Docker image semver tagging (`x.y.z`, `x.y`, `x`, `latest`) | ✅ Complete |
| CHANGELOG.md auto-generated by release-please | ✅ Complete |

v3.0.0 will be cut with a `feat!:` commit once any remaining pre-release work is complete. The major version bump reflects the breaking transition from the Python-based wrapper to the Go binary.

## Contributing

This project uses [Conventional Commits](https://www.conventionalcommits.org/) to drive automated semantic versioning via [release-please](https://github.com/googleapis/release-please).

| Prefix | Version bump | When to use |
|---|---|---|
| `feat:` | minor (x.**Y**.0) | New user-facing feature |
| `fix:` | patch (x.y.**Z**) | Bug fix |
| `feat!:` or `BREAKING CHANGE:` in body | major (**X**.0.0) | Breaking change |
| `chore:`, `docs:`, `ci:`, `refactor:`, `test:` | none | No release needed |

```bash
# Examples
git commit -m "feat: add JSON output flag to list command"
git commit -m "fix: handle missing DISPLAY gracefully when X11 disabled"
git commit -m "feat!: rename config fields for consistency

BREAKING CHANGE: docker_image renamed to image in config file"
```

Merging a `feat:` or `fix:` commit to `main` causes release-please to open a release PR. Merging that PR creates the version tag, which triggers the full release pipeline (binary build + Docker image push).

## Container Image

The SCRT container image (`fonalex45/scrt`) is a Kali Linux base with curated security tooling. See `sources/` for the full tool list. Key categories:

- **Network**: nmap, masscan, netexec, tcpdump, socat, mitmproxy
- **Web**: ffuf, sqlmap, whatweb
- **Active Directory**: bloodhound.py, evil-winrm, responder, smbclient
- **OSINT**: cewl, sn0int, exiflooter, h8mail
- **Privesc**: linpeas, winpeas, pspy, chisel
- **Desktop**: Firefox, Mousepad, Wireshark (via X11)

## License

See [LICENSE](LICENSE) for details.
