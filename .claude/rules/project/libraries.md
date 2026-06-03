---
paths:
  - "**/*.go"
  - "**/go.mod"
  - "**/go.sum"
---
# SCRT Required Libraries

Use ONLY these approved libraries for their respective purposes. Do not introduce alternatives without explicit approval.

## Docker Operations
- **`github.com/docker/docker/client`** — all container/image operations via Docker SDK
- **`os/exec`** — TTY-attached sessions only (`docker run -it`, `docker exec -it`)

## TUI Framework
- **`github.com/rivo/tview`** — the primary full-screen TUI: interactive menus, lists, dialogs
- **`github.com/gdamore/tcell/v2`** — pulled in by tview; may be used directly in the `tui` package

## Interactive Prompts
- **`charm.land/huh/v2`** — transient terminal prompts, confirmations, and forms (the
  `config` wizard, the `pull` tag picker, the `destroy` confirmation). Runs on bubbletea;
  this is the approved exception to the "no bubbletea as primary TUI" rule — huh is for
  short-lived prompts, never a full-screen app. Full-screen UIs stay in tview.

## CLI Framework
- **`github.com/spf13/cobra`** — command tree, flags, args
- **`charm.land/fang/v2`** — wraps the cobra root (`fang.Execute`) for styled help/errors,
  `--version`, manpage generation, and shell completion

## Styling
- **`charm.land/lipgloss/v2`** — terminal output styling (colors, borders, banners)

## Spinner
- **`charm.land/bubbletea/v2`** + **`charm.land/bubbles/v2`** — the loading spinner only
  (`internal/tui/spinner.go`) and as the engine huh runs on

> The charm libraries use the `charm.land/...` v2 module paths and require **Go 1.25+**.
> An indirect `github.com/charmbracelet/lipgloss` v1 is pulled transitively by
> `charmbracelet/log` (the logger has no v2 release); this is acceptable — project code
> imports only the v2 `charm.land` paths.

## Do Not Use
- Do not use `bubbletea` as the primary full-screen TUI framework — tview is the chosen
  library; bubbletea is allowed only for the spinner and as huh's engine (see above)
- Do not use the Docker CLI (`os/exec docker ...`) for non-TTY operations — use the SDK instead
