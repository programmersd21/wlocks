# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.1.0] - 2026-07-14

### Added

- scan `/proc/[pid]/fd/*` to discover which processes hold files open
- real time auto refresh, polls `/proc` every 2 seconds
- detail view with pid, command, cwd, open fd count, and lock duration
- fuzzy search across process name, command, and file path
- sort by name, duration, pid, or mode with reverse support
- six hand tuned themes: default, tokyo, catppuccin, everforest, nord, gruvbox
- smooth 60fps animations with cubic easing for scrolling and view transitions
- command palette with quick access to all actions
- interactive help view and statistics view
- status message feedback for all user actions
- tty detection with plain text fallback for pipes and non interactive use
- persistent theme preference in `~/.config/wlocks/config.toml`
- keyboard driven navigation with vim style bindings
- linux only, reads `/proc` directly, no external dependencies at runtime

### Fixed

- goreleaser `format` field replaced with `formats` (v2 deprecation)
- golangci lint compatibility pinned to Go 1.24

### Changed

- auto refresh is always on, removed `--live` flag and toggle
- command palette moved to `ctrl+p`

### CI/CD

- github actions workflows for ci and release
- goreleaser v2 configuration for automated multi arch linux builds
- shields.io badges for go version, license, release, ci status, and goreleaser
