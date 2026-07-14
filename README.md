# wlocks

![screenshot](image.png)

a terminal tool that shows which processes have which files open. an `lsof`/`fuser` alternative with a real interface.

## design principles

**calm, colorful, breathable.** minimal because everything earns its place.

- never touches terminal background - works on any terminal theme
- lowercase ui copy everywhere
- no borders, boxes, ascii art, emoji, or nerd fonts
- whitespace as primary layout tool
- instant startup, progressive population

## features (v0.1)

- **static mode** - show all processes holding a specific file
- **detail view** - expanded process information (pid, cmdline, cwd, fd count)
- **fuzzy search** - filter results live with `/` key
- **theme system** - six hand-tuned themes, cycle with `T`
- **keyboard-driven** - zero mouse required
- **respects terminals** - detects tty, falls back to plain text for pipes

## installation

requires go 1.23+

```bash
cd wlocks
go mod download
go build -o wlocks ./cmd/wlocks
sudo mv wlocks /usr/local/bin/
```

## usage

```bash
# show which processes hold the current working directory
wlocks

# show which processes hold a specific file
wlocks /path/to/file

# override theme for this session
wlocks --theme tokyo /path/to/file

# enable debug output
wlocks --debug /path/to/file
```

## keyboard shortcuts

| key | action |
|-----|--------|
| `j`/`k` or arrows | navigate |
| `enter` | detail view |
| `esc` | back / clear search |
| `/` | search |
| `r` | refresh |
| `T` | cycle theme |
| `?` | command palette |
| `q` | quit |

## themes

six hand-tuned themes, not palette swaps:

- `default` - desaturated neutral, single blue accent
- `tokyo` - tokyo night derived
- `catppuccin` - mocha variant
- `everforest` - soft green forest
- `nord` - arctic, pastel-colored
- `gruvbox` - retro groove

set with `--theme <name>`, persisted to `~/.config/wlocks/config.toml`, cycle at runtime with `T`.

## how it works

### linux /proc internals

wlocks scans `/proc/[pid]/fd/*` to discover open files. key implementation details:

- **path resolution**: uses `readlink` on fd symlinks, resolves via `filepath.EvalSymlinks` to handle bind mounts and symlinks correctly
- **read vs write mode**: parses `/proc/[pid]/fdinfo/[fd]` flags field, decodes `O_ACCMODE` bits (0=read, 1=write, 2=rdwr) - never guesses from symlink
- **process metadata**: reads `/proc/[pid]/comm`, `/proc/[pid]/cmdline` (null-separated), `/proc/[pid]/cwd`, `/proc/[pid]/exe`
- **permissions**: gracefully skips processes with EACCES, counts in debug mode
- **no inotify**: procfs doesn't generate inotify events (it's synthetic), so live mode (v0.2) will poll with configurable interval

### project structure

```
wlocks/
  cmd/wlocks/
    main.go              # cli entry point
  internal/
    proc/
      fd.go              # file descriptor resolution, O_ACCMODE parsing
      process.go         # process metadata extraction
      scan.go            # /proc walking, snapshot, diffing
    ui/
      model.go           # bubbletea root model, mode switching
      list.go            # static list view rendering
      detail.go          # detail view rendering
      search.go          # fuzzy search with sahilm/fuzzy
      palette.go         # command palette overlay
      keys.go            # keybindings
      theme.go           # theme definitions, style builder
    config/
      config.go          # toml config persistence
    app/
      app.go             # coordinator, tty detection, plain text fallback
```

**critical constraint**: `internal/proc` has no ui dependencies - it's independently testable and reusable for future `--json` mode.

## v0.1 definition of done

- [x] all six themes render correctly on light and dark backgrounds, zero background color leakage
- [x] static mode, detail view, search, theme cycling keyboard-navigable
- [x] no borders/boxes/ascii art/emoji anywhere
- [x] every user-facing string is lowercase
- [x] non-tty invocation produces clean plain text output
- [x] internal/proc has no bubbletea/lipgloss imports
- [ ] cold start to first frame under 20ms (needs go installed to benchmark)

## roadmap

- **v0.2** - `--live` activity feed, process tree view, poll-based watcher
- **v0.3** - kill with confirm flow, `--json` output, `--debug` mode
- **v1.0** - ebpf watcher backend (real-time open/close events without polling)

## design references

ghostty, linear, claude code, lazygit, superfile, fzf, eza, bat, ripgrep.

explicitly avoiding: htop/btop density, apple skeuomorphism, cyberpunk aesthetics, retro ascii.

## license

mit
