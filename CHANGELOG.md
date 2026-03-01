# Changelog

## [3.0.3](https://github.com/alexrf45/SCRT/compare/v3.0.2...v3.0.3) (2026-03-01)


### Bug Fixes

* remove logger.Info calls from Stop, Destroy, and Enter ([71eb468](https://github.com/alexrf45/SCRT/commit/71eb468d55f02ab6ae20bab73a97869e5e164bf5))

## [3.0.2](https://github.com/alexrf45/SCRT/compare/v3.0.1...v3.0.2) (2026-03-01)


### Bug Fixes

* run tui list callbacks in goroutines to prevent event loop freeze ([fae7606](https://github.com/alexrf45/SCRT/commit/fae76061008f893de07294bdbe1228f5a50f5629))

## [3.0.1](https://github.com/alexrf45/SCRT/compare/v3.0.0...v3.0.1) (2026-03-01)


### Bug Fixes

* resolve double-tag bug in pull dialog and add bubbletea spinner ([bd52445](https://github.com/alexrf45/SCRT/commit/bd52445500be7a78e7bb899d4538ab7b67729fd0))

## [3.0.0](https://github.com/alexrf45/SCRT/compare/v2.0.3...v3.0.0) (2026-03-01)


### ⚠ BREAKING CHANGES

* scrt list now launches a tview TUI instead of a static bubbles table when running in a TTY.

### Features

* migrate to Docker SDK and tview interactive TUI (v3.0.0) ([26f069b](https://github.com/alexrf45/SCRT/commit/26f069b5b4150ab1d00d174239e05aee3baef4e9))

## [2.0.2](https://github.com/alexrf45/SCRT/compare/v2.0.1...v2.0.2) (2026-02-26)


### Bug Fixes

* restore status column in list output ([9db636c](https://github.com/alexrf45/SCRT/commit/9db636c0c0c46585a60302360f9d76b1c007e3b6))

## [2.0.1](https://github.com/alexrf45/SCRT/compare/v2.0.0...v2.0.1) (2026-02-26)


### Bug Fixes

* allow env vars to re-enable host networking, X11, and GPU ([9998d9a](https://github.com/alexrf45/SCRT/commit/9998d9a2ceaa614992f599784c29923792613d1f))
