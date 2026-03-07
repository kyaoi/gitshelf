# gitshelf

`gitshelf` is a Git-friendly task manager centered around one TUI workspace: `Cockpit`.

- CLI command: `shelf`
- Main entry: `shelf` or `shelf cockpit`
- Storage root: `.shelf/`
- Tasks: `.shelf/tasks/<id>.md`
- Links: `.shelf/edges/<src_id>.toml`

## Documentation

- CLI spec: [`docs/COMMANDS.md`](docs/COMMANDS.md)
- Command guide: [`docs/COMMAND_GUIDE.md`](docs/COMMAND_GUIDE.md)
- Workflow guide: [`docs/WORKFLOWS.md`](docs/WORKFLOWS.md)
- Interactive behavior: [`docs/INTERACTIVE.md`](docs/INTERACTIVE.md)
- Storage: [`docs/STORAGE.md`](docs/STORAGE.md)
- Japanese user docs: [`docs/ja/README.md`](docs/ja/README.md)

## Install

```bash
go build -o shelf ./cmd/shelf
```

## Quick Start

```bash
# initialize
./shelf init

# main workspace
./shelf
./shelf cockpit

# cockpit launchers
./shelf calendar
./shelf tree
./shelf board
./shelf review
./shelf now

# script-friendly queries
./shelf ls --status open --json
./shelf next

# shell completion
./shelf completion zsh
```

## Command Surface

Only these top-level commands are part of the current public CLI surface:

- `shelf init`
- `shelf completion`
- `shelf cockpit`
- `shelf calendar`
- `shelf tree`
- `shelf board`
- `shelf review`
- `shelf now`
- `shelf ls`
- `shelf next`

Everything else is expected to happen inside Cockpit.

## Cockpit-First Usage

`Cockpit` is the main workspace.

- `shelf` on TTY opens `Cockpit`
- `shelf cockpit` opens it explicitly
- `calendar/tree/board/review/now` are just launcher presets for the same workspace
- creating, editing, moving, snoozing, linking, archiving, and status changes are handled inside the TUI

Recommended starting point:

```bash
shelf
```

## Current Data Model

Task metadata uses:

- `title`
- `kind`
- `status`
- `tags`
- `due_on`
- `repeat_every`
- `archived_at`
- `parent`
- timestamps

Links use only:

- `depends_on`
- `related`

## Quality Checks

```bash
gofmt -w .
go test ./...
go test -race ./...
go vet ./...
```
