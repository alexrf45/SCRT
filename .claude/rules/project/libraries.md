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
- **`github.com/rivo/tview`** — interactive menus, lists, dialogs
- **`github.com/gdamore/tcell/v2`** — pulled in by tview; may be used directly in the `tui` package

## Styling
- **`github.com/charmbracelet/lipgloss`** — terminal output styling (colors, borders, banners)

## Do Not Use
- Do not use `bubbletea` as the primary TUI framework — tview is the chosen library
- Do not use the Docker CLI (`os/exec docker ...`) for non-TTY operations — use the SDK instead
