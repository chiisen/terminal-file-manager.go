# PRD.md

Project: Go Terminal File Manager Codename: gofm

## 1. Product Overview

建立一個高效能 Terminal 檔案總管 (Terminal File Manager)，使用 Go
(Golang) 開發，提供鍵盤導向的檔案瀏覽、操作與管理能力。

目標： - 比 GUI 檔案總管更快 - 比 shell ls/cd 更直覺 - 支援大型目錄
(10k+ files)

## 2. Target Users

  User               Needs
  ------------------ -----------------------------
  Backend engineer   快速瀏覽與操作 server files
  DevOps             SSH 操作
  Power user         keyboard-first workflow

## 3. Core Principles

### Keyboard First

所有操作必須可以純鍵盤完成

### Minimal Latency

-   directory render \< 50ms
-   navigation latency \< 16ms

### Low Memory

-   memory \< 50MB

## 4. Core Features

### File Navigation

-   瀏覽目錄
-   進入資料夾
-   返回上一層

  Key         Action
  ----------- ------------
  ↑ / k       move up
  ↓ / j       move down
  enter / l   open
  h           parent dir

### File Operations

  Action     Shortcut
  ---------- ----------
  copy       y
  paste      p
  delete     d
  rename     r
  new file   a
  new dir    A

### Preview Pane

右側顯示： - text preview - image info - binary info

限制： - max preview size = 1MB

### Search

快捷鍵：`/` - fuzzy search - highlight results

### Sorting

-   name
-   size
-   modified
-   type

## 5. UI Layout

    +----------------------------------------------------+
    | PATH: /home/sam/projects                           |
    +----------------------+-----------------------------+
    |                      |                             |
    | file list            | preview                     |
    |                      |                             |
    | > main.go            | package main                |
    |   go.mod             |                             |
    |   README.md          | func main() {               |
    |   cmd/               |     fmt.Println("hello")    |
    |   pkg/               | }                           |
    |                      |                             |
    +----------------------+-----------------------------+
    | STATUS BAR                                        |
    +----------------------------------------------------+

## 6. Keybinding System

config: \~/.config/gofm/config.toml

範例：

``` toml
[keymap]
up = "k"
down = "j"
open = "l"
back = "h"
delete = "d"
copy = "y"
paste = "p"
```

## 7. System Architecture

    Terminal UI (TUI Renderer)
            |
            v
    Event System (keyboard input)
            |
            v
    State Manager (current path)
            |
            v
    FileSystem Layer

## 8. Package Structure

    gofm/

    cmd/
      gofm/
        main.go

    internal/
      app/
        app.go

      ui/
        layout.go
        renderer.go
        components.go

      fs/
        filesystem.go
        operations.go

      input/
        keymap.go
        event.go

      state/
        state.go

      preview/
        preview.go

## 9. Data Model

``` go
type FileEntry struct {
    Name string
    Path string
    Size int64
    Mode os.FileMode
    ModTime time.Time
    IsDir bool
}
```

``` go
type AppState struct {
    CurrentPath string
    Entries []FileEntry
    Cursor int
    Selected map[string]bool
}
```

## 10. Rendering Engine

推薦： - Bubble Tea (Go TUI framework) - Lip Gloss (terminal styling)

原因： - event-driven - reactive UI

## 11. Performance Strategy

Lazy load metadata

流程： 1. read dir names 2. render 3. async stat files

## 12. Error Handling

  Scenario            Strategy
  ------------------- --------------
  permission denied   show message
  file deleted        auto refresh
  broken symlink      highlight

## 13. Logging

log file: \~/.config/gofm/log.txt

## 14. MVP Scope

第一版： - file navigation - open directory - delete - rename - copy /
paste

不包含： - remote filesystem - plugins - git integration

## 15. Future Features

### Git integration

顯示： - M modified - A added - D deleted

### Plugin System

\~/.config/gofm/plugins

### Remote Filesystem

-   SSH
-   SFTP

## 16. CLI Usage

    gofm

指定目錄：

    gofm /var/www

## 17. Build

    go build -o gofm ./cmd/gofm

## 18. Success Metrics

  metric       target
  ------------ ----------
  startup      \< 200ms
  navigation   \< 16ms
  memory       \< 50MB

## 19. State Machine

    State
     ├── Normal
     ├── Search
     ├── Rename
     └── ConfirmDelete
