<div align="center">

### < wlocks >

*A terminal tool that shows which processes have which files open - an interactive `lsof` / `fuser` alternative.*

</div>

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev)
[![MIT](https://img.shields.io/badge/License-MIT-238636?style=for-the-badge)](LICENSE)
[![Release](https://img.shields.io/github/v/release/programmersd21/wlocks?style=for-the-badge&label=release&color=8957E5)](https://github.com/programmersd21/wlocks/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/programmersd21/wlocks/ci.yml?style=for-the-badge&label=CI)](https://github.com/programmersd21/wlocks/actions)
[![AUR](https://img.shields.io/aur/version/wlocks-bin?style=for-the-badge&logo=archlinux)](https://aur.archlinux.org/packages/wlocks-bin)

<img src="demo/demo.gif" width="720" alt="wlocks demo">

## Install

```bash
# arch linux
yay -S wlocks-bin

# linux (amd64 / arm64)
curl -L https://github.com/programmersd21/wlocks/releases/latest/download/wlocks_linux_$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/').tar.gz | tar xz
sudo mv wlocks /usr/local/bin/

# from source (go 1.25+)
git clone https://github.com/programmersd21/wlocks.git && cd wlocks
go build -o wlocks ./cmd/wlocks && sudo mv wlocks /usr/local/bin/
```

## Usage

```bash
wlocks                    # current directory
wlocks /path/to/file      # specific file or folder
wlocks --theme linear     # theme override
wlocks --debug /path      # debug diagnostics
```

## Keys

`j`/`k` navigate · `enter` detail view · `/` fuzzy search · `esc` back
`K` kill · `F` force kill · `P` pause/resume · `r` refresh
`s` sort · `S` reverse sort · `T` cycle theme · `ctrl+p` palette · `?` help · `q` quit

## Features

- Auto-refresh polls `/proc` every second for live process and file changes
- Detail view: pid, cmdline, cwd, open file descriptors, lock duration
- Fuzzy search across process name, command, and file path
- Sort by name, duration, pid, or access mode
- Nine themes: `default`, `tokyo`, `catppuccin`, `everforest`, `nord`, `gruvbox`, `apple`, `linear`, `neon`
- Kill / pause / force-kill with confirmation safeguards
- Config persisted to `~/.config/wlocks/config.toml`

## Config

```toml
theme = "linear"
default_sort = "duration"
live_refresh_rate = 1
animation_speed = "normal"
```

## How it works

Reads `/proc/[pid]/fd/*` symlinks, resolving paths via `filepath.EvalSymlinks` for bind mounts and namespaces. Read/write mode comes from decoding `O_ACCMODE` bits in `/proc/[pid]/fdinfo/[fd]` - no guessing. Inaccessible processes are skipped gracefully (visible in `--debug`). Pure Go, no cgo, no runtime dependencies.

## License

[MIT](LICENSE) © programmersd21
