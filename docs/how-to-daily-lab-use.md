# How to use SCRT for daily security research in a lab environment

**Audience:** Security researchers comfortable with Docker and the Linux command line.

**Scope:** Lab environments only. This guide assumes your host machine is on an isolated research network or a dedicated lab host. Use on live targets, production networks, or shared infrastructure is out of scope.

---

## Prerequisites

- Docker daemon running on the host (`docker info` returns without error)
- `scrt` binary on your `$PATH` — see [build from source](#build-from-source) below if needed
- A dedicated lab host or isolated network segment

### Build from source

```bash
git clone https://github.com/alexrf45/SCRT.git
cd SCRT/scrt
make build          # produces bin/scrt
cp bin/scrt ~/.local/bin/scrt
```

Verify:

```bash
scrt version
```

---

## 1. First-time setup

Before running any containers, review your configuration. SCRT writes defaults to `~/.scrt.conf.json` on first edit.

```bash
scrt config edit
```

Your editor opens with the full config. Key fields to review for a lab environment:

| Field | Default | What to check |
|---|---|---|
| `docker_image` | `fonalex45/scrt:latest` | Set to the image you want to use day-to-day |
| `container_shell` | `/bin/zsh` | Change to `/bin/bash` if preferred |
| `host_networking` | `true` | See [Lab safety practices](#10-lab-safety-practices) before changing |
| `enable_x11` | `true` | Disable if your lab host has no display — see safety notes |
| `work_dir_base` | current directory at first run | Set to a stable base directory, e.g. `~/lab` |

Save and exit. SCRT validates the file immediately — it will report any errors before returning to the prompt.

To confirm what loaded:

```bash
scrt config
```

---

## 2. Starting a new research project

Each project gets its own container and workspace directory under `work_dir_base`.

```bash
scrt start <project>
```

**Example:**

```bash
scrt start htb-machine
```

This creates `~/lab/htb-machine/` and drops you into an interactive shell inside the container. The project name becomes the container name and is injected as `$PROJECT` and `$TARGET` inside the shell.

To use a different image for a specific project without changing your config:

```bash
scrt start htb-machine --image fonalex45/scrt:dev
```

**Project names** must be lowercase alphanumeric with hyphens only (no spaces, no underscores).

---

## 3. Re-entering a project

When you exit the shell, the container keeps running. To get back in:

```bash
scrt enter <project>
```

If the container was stopped (e.g. after a host reboot), `enter` starts it automatically before opening the shell.

```bash
scrt enter htb-machine
```

Your workspace files under `~/lab/htb-machine/` persist between sessions via a bind mount.

---

## 4. Browsing and managing containers

The `list` command opens an interactive TUI when run in a terminal:

```bash
scrt list
```

| Key | Action |
|---|---|
| `e` | Enter the selected container |
| `s` | Stop the selected container |
| `d` | Destroy the selected container |
| `b` | Backup the selected container |
| `r` | Refresh the container list |
| `q` | Quit |

Arrow keys or `j`/`k` move selection. When piped or run without a TTY, `list` outputs a plain table instead of the TUI.

---

## 5. Updating your image

Pull the latest version of your configured image:

```bash
scrt pull
```

Without `--image`, an interactive dialog lets you select a tag (`latest`, `dev`, or a custom value). To pull non-interactively:

```bash
scrt pull --image fonalex45/scrt:dev
```

Existing running containers are not affected — they continue using the image they were started with. New containers started after the pull will use the updated image.

---

## 6. Backing up project data

Create a tar archive of a project's workspace directory:

```bash
scrt backup <project>
```

By default, archives land in `./backups/`. To specify a different location:

```bash
scrt backup htb-machine --dir ~/archives
```

The output filename includes the project name and a timestamp:

```
~/archives/htb-machine-20260402-143012.tar.gz
```

**When to back up:**

- Before destroying a project you may want to revisit
- After completing a significant phase of research
- Before pulling a new image version

---

## 7. Importing a backup for analysis

To load a backup archive back as a Docker image — useful for forensic analysis of a prior research session:

```bash
scrt import <file> --repo <repo> [--tag <tag>]
```

**Example:**

```bash
scrt import ~/archives/htb-machine-20260402-143012.tar.gz \
  --repo fonalex45/scrt-archive \
  --tag htb-machine-apr2026
```

The archive is imported as a new local image. You can then start a fresh container from it:

```bash
scrt start review --image fonalex45/scrt-archive:htb-machine-apr2026
```

The `--tag` flag defaults to `imported` if omitted.

---

## 8. Stopping and destroying a project

### Stop (preserves data)

Stops the container but leaves the workspace directory and container intact:

```bash
scrt stop <project>
```

The container can be re-entered later with `scrt enter`.

### Destroy (removes container and data)

Removes the container and deletes the project directory:

```bash
scrt destroy <project>
```

You will be prompted to confirm. To skip the prompt in scripts:

```bash
scrt destroy <project> --force
```

**This is not reversible.** Back up first if you need the data.

---

## 9. Lab safety practices

SCRT's defaults are tuned for an isolated lab environment. Before adjusting them, understand the trade-offs.

### Host networking

`host_networking: true` gives the container direct access to every interface on the host, including any network your lab machine is connected to. This is intentional for lab use — tools like `nmap` and `tcpdump` need it.

**If your lab machine has any interface connected to a network you do not own or control, set `host_networking` to `false`** or use per-project network namespaces. You can override per session via environment variable:

```bash
SCRT_HOST_NET=false scrt start isolated-project
```

### X11 forwarding

`enable_x11: true` mounts the host's X11 socket into the container, which allows GUI tools (Firefox, Mousepad) to run. Any process inside the container can interact with your host display session.

Disable it on headless hosts or when you do not need GUI tools:

```bash
scrt config edit   # set enable_x11 to false
```

Or per session:

```bash
SCRT_X11=false scrt start no-gui-project
```

### Linux capabilities

Every SCRT container runs with `NET_ADMIN` and `SYS_TIME` by default:

| Capability | Why it's included | Risk if abused |
|---|---|---|
| `NET_ADMIN` | Required for interface manipulation, traffic shaping, VPN tools | Can reconfigure host network interfaces if combined with host networking |
| `SYS_TIME` | Required for time-based tooling and clock sync in offline labs | Can skew the host clock |

To reduce the capability set for a less privileged session, edit `custom_caps` in config before starting the container:

```bash
scrt config edit   # remove capabilities not needed for your current task
scrt start minimal-project
```

### Destroy vs stop

Do not use `destroy --force` as a routine cleanup shortcut. It silently deletes the project directory. Use `stop` unless you are certain you no longer need the data.

---

## Quick reference

```
scrt start <project> [--image img:tag]   Create workspace, start container
scrt enter <project>                     Enter running container (starts if stopped)
scrt stop <project>                      Stop container, keep workspace
scrt destroy <project> [--force]         Remove container and workspace
scrt list                                Interactive container browser (TUI)
scrt pull [--image img:tag]              Pull/update image
scrt backup <project> [--dir path]       Archive workspace to tar
scrt import <file> --repo <repo>         Import backup as Docker image
scrt config                              Show current configuration
scrt config edit                         Edit configuration in $EDITOR
```
