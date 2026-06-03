# Changelog

## [4.3.0](https://github.com/alexrf45/SCRT/compare/v4.2.0...v4.3.0) (2026-06-03)


### Features

* **cli:** migrate to charm v2 stack and add fang CLI polish ([83e6424](https://github.com/alexrf45/SCRT/commit/83e64241245b5c68a4f7593fe2905155aee878d3))
* **tui:** add container search/filter and live auto-refresh ([b3f7aff](https://github.com/alexrf45/SCRT/commit/b3f7aff94831edc28b6deac5a2c9bfbaa6d19038))
* **tui:** add huh prompts, pull picker, and config wizard ([de06aa3](https://github.com/alexrf45/SCRT/commit/de06aa36978a1d77ddaa69c0ae0a9698a81af18f))
* **tui:** add in-browser log viewer and file copy ([2e9abdd](https://github.com/alexrf45/SCRT/commit/2e9abdde55f0cfaacde7967d97670cf8bed3fd76))

## [4.2.0](https://github.com/alexrf45/SCRT/compare/v4.1.1...v4.2.0) (2026-04-03)


### Features

* **base:** reduce image size ~210-280 MB; add smoke-base test ([c6dc724](https://github.com/alexrf45/SCRT/commit/c6dc72412e730702c00c9ff5e77ae49e8dcee0f6))
* **base:** reduce image size ~210-280 MB; add smoke-base test ([75e68c8](https://github.com/alexrf45/SCRT/commit/75e68c8f54681e760843d413d778f2806a4c38d4))

## [4.1.1](https://github.com/alexrf45/SCRT/compare/v4.1.0...v4.1.1) (2026-04-03)


### Bug Fixes

* **docker:** add apt-get update to 0-base.sh; replace web apt deps wi… ([#51](https://github.com/alexrf45/SCRT/issues/51)) ([b3113c9](https://github.com/alexrf45/SCRT/commit/b3113c9f10b97cd6913d9743a20609527fa75b05))
* **docker:** add apt-get update to 0-base.sh; replace web apt deps with direct downloads ([1fdc6e7](https://github.com/alexrf45/SCRT/commit/1fdc6e7f0adef56411b660bffd7ffb04b7198173))

## [4.1.0](https://github.com/alexrf45/SCRT/compare/v4.0.0...v4.1.0) (2026-04-02)


### Features

* add scenario-specific Docker images with smoke tests ([a8865f4](https://github.com/alexrf45/SCRT/commit/a8865f46e5c97a3de11fecd2eb165d6110de147e))
* **web:** container IP, background script runner, jobs panel ([5e63922](https://github.com/alexrf45/SCRT/commit/5e6392240d97ca2fafbbeda82786c4880aa78ae6))
* **web:** start containers, TLS, file transfer, and themes ([52d20c1](https://github.com/alexrf45/SCRT/commit/52d20c1325a31d808d914b51768e6a3b7fc03d2d))

## [4.0.0](https://github.com/alexrf45/SCRT/compare/v3.2.0...v4.0.0) (2026-03-28)


### ⚠ BREAKING CHANGES

* scrt list now launches a tview TUI instead of a static bubbles table when running in a TTY.

### Features

* add deployment manifests and update README for remote lab ([eeb40a8](https://github.com/alexrf45/SCRT/commit/eeb40a8ea4027cc51d78a56ba3418aaca42371ce))
* add remote lab design doc, control plane Dockerfile, and dev CI image build ([ed233c2](https://github.com/alexrf45/SCRT/commit/ed233c2927c4c551e0dd0426c5daac3b993549fb))
* **backend:** introduce Backend interface and tier detection ([62e7a88](https://github.com/alexrf45/SCRT/commit/62e7a88c618dc49f974d4190e81a50ba7f63590d))
* migrate to Docker SDK and tview interactive TUI (v3.0.0) ([26f069b](https://github.com/alexrf45/SCRT/commit/26f069b5b4150ab1d00d174239e05aee3baef4e9))
* remote lab foundation — backend abstraction, serve mode, deploy manifests ([#42](https://github.com/alexrf45/SCRT/issues/42)) ([b45ae49](https://github.com/alexrf45/SCRT/commit/b45ae497416d5f34beb90da7681c3bc98a83fdeb))
* **serve:** add HTTP API, embedded web UI, and scrt serve command ([0bbec1d](https://github.com/alexrf45/SCRT/commit/0bbec1d286ce698037d2202745f416845c49041a))
* **web-tty:** add WebSocket terminal for browser-based container shell ([75118bf](https://github.com/alexrf45/SCRT/commit/75118bf19cb74021bf93eb4dba082554d753e138))


### Bug Fixes

* allow env vars to re-enable host networking, X11, and GPU ([9998d9a](https://github.com/alexrf45/SCRT/commit/9998d9a2ceaa614992f599784c29923792613d1f))
* **api:** correct Info JSON tag Names→Name so web UI displays container names ([12824db](https://github.com/alexrf45/SCRT/commit/12824dbe0da27ce95a7727da168dd94ba86c6b89))
* **main:** renaming repo ([1a0d3e5](https://github.com/alexrf45/SCRT/commit/1a0d3e557d188b835f5ba4f5dc7bbedb2f9f4acf))
* remove logger.Info calls from Stop, Destroy, and Enter ([71eb468](https://github.com/alexrf45/SCRT/commit/71eb468d55f02ab6ae20bab73a97869e5e164bf5))
* resolve double-tag bug in pull dialog and add bubbletea spinner ([bd52445](https://github.com/alexrf45/SCRT/commit/bd52445500be7a78e7bb899d4538ab7b67729fd0))
* restore status column in list output ([9db636c](https://github.com/alexrf45/SCRT/commit/9db636c0c0c46585a60302360f9d76b1c007e3b6))
* run tui list callbacks in goroutines to prevent event loop freeze ([fae7606](https://github.com/alexrf45/SCRT/commit/fae76061008f893de07294bdbe1228f5a50f5629))
* **ui:** handle non-JSON error responses in apiFetch; prevent redirect-to-HTML bug ([9a8d03f](https://github.com/alexrf45/SCRT/commit/9a8d03ffe30ac5a7651a779343b4503e467dadca))

## [3.2.0](https://github.com/alexrf45/SCRT/compare/v3.1.1...v3.2.0) (2026-03-28)


### Features

* **web-tty:** add WebSocket terminal for browser-based container shell ([75118bf](https://github.com/alexrf45/SCRT/commit/75118bf19cb74021bf93eb4dba082554d753e138))

## [3.1.1](https://github.com/alexrf45/SCRT/compare/v3.1.0...v3.1.1) (2026-03-28)


### Bug Fixes

* **api:** correct Info JSON tag Names→Name so web UI displays container names ([12824db](https://github.com/alexrf45/SCRT/commit/12824dbe0da27ce95a7727da168dd94ba86c6b89))
* **ui:** handle non-JSON error responses in apiFetch; prevent redirect-to-HTML bug ([9a8d03f](https://github.com/alexrf45/SCRT/commit/9a8d03ffe30ac5a7651a779343b4503e467dadca))

## [3.1.0](https://github.com/alexrf45/SCRT/compare/v3.0.4...v3.1.0) (2026-03-28)


### Features

* add deployment manifests and update README for remote lab ([eeb40a8](https://github.com/alexrf45/SCRT/commit/eeb40a8ea4027cc51d78a56ba3418aaca42371ce))
* add remote lab design doc, control plane Dockerfile, and dev CI image build ([ed233c2](https://github.com/alexrf45/SCRT/commit/ed233c2927c4c551e0dd0426c5daac3b993549fb))
* **backend:** introduce Backend interface and tier detection ([62e7a88](https://github.com/alexrf45/SCRT/commit/62e7a88c618dc49f974d4190e81a50ba7f63590d))
* remote lab foundation — backend abstraction, serve mode, deploy manifests ([#42](https://github.com/alexrf45/SCRT/issues/42)) ([b45ae49](https://github.com/alexrf45/SCRT/commit/b45ae497416d5f34beb90da7681c3bc98a83fdeb))
* **serve:** add HTTP API, embedded web UI, and scrt serve command ([0bbec1d](https://github.com/alexrf45/SCRT/commit/0bbec1d286ce698037d2202745f416845c49041a))

## [3.0.4](https://github.com/alexrf45/SCRT/compare/v3.0.3...v3.0.4) (2026-03-28)


### Bug Fixes

* **main:** renaming repo ([1a0d3e5](https://github.com/alexrf45/SCRT/commit/1a0d3e557d188b835f5ba4f5dc7bbedb2f9f4acf))

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
