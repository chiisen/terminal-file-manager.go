# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**gofm** - A high-performance Terminal File Manager written in Go (Golang). It's a keyboard-first TUI application for file browsing, operation, and management.

Target: Faster than GUI file managers, more intuitive than shell ls/cd, supports large directories (10k+ files).

## Key Commands

```bash
# Initialize Go module (first time setup)
go mod init gofm

# Build the application
go build -o gofm ./cmd/gofm

# Run in development
go run ./cmd/gofm

# Run with a specific directory
gofm /var/www

# Format code (required before commit)
go fmt ./...

# Run static analysis
go vet ./...

# Run tests
go test ./...
```

## Architecture

The project follows the architecture defined in `docs/PRD.md`:

```
Terminal UI (TUI Renderer)
        │
        v
Event System (keyboard input)
        │
        v
State Manager (current path)
        │
        v
FileSystem Layer
```

### Package Structure

- `cmd/gofm/` - Application entry point (main.go)
- `internal/app/` - Application state and logic
- `internal/ui/` - TUI rendering (layout, renderer, components)
- `internal/fs/` - File system operations
- `internal/input/` - Keymap and event handling
- `internal/state/` - State management
- `internal/preview/` - File preview functionality

### Framework

- **Bubble Tea** - Go TUI framework (event-driven, reactive UI)
- **Lip Gloss** - Terminal styling

### State Machine

The app operates in these states:
- `Normal` - Default navigation mode
- `Search` - Fuzzy search mode (triggered with `/`)
- `Rename` - File renaming mode
- `ConfirmDelete` - Delete confirmation

## Keybindings (Default)

| Key | Action |
|-----|--------|
| ↑ / k | Move up |
| ↓ / j | Move down |
| Enter / l | Open |
| h | Parent directory |
| y | Copy |
| p | Paste |
| d | Delete |
| r | Rename |
| a | New file |
| A | New directory |
| / | Search |

## Performance Targets

- Directory render: < 50ms
- Navigation latency: < 16ms
- Memory usage: < 50MB
- Startup time: < 200ms

## Important Notes

- User is learning Go - provide explanatory comments for Go-specific concepts (goroutines, channels, defer, interfaces, pointers)
- All new Go code should include explanatory comments in Traditional Chinese
- Config file: `~/.config/gofm/config.toml`
- Log file: `~/.config/gofm/log.txt`

## Data Models

```go
type FileEntry struct {
    Name    string
    Path    string
    Size    int64
    Mode    os.FileMode
    ModTime time.Time
    IsDir   bool
}

type AppState struct {
    CurrentPath string
    Entries     []FileEntry
    Cursor      int
    Selected    map[string]bool
}
```
