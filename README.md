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

### Recommended: install directly with `go install`

```bash
go install github.com/kyaoi/gitshelf/cmd/shelf@latest
```

### Local development: clone and build

```bash
git clone https://github.com/kyaoi/gitshelf.git
cd gitshelf
go install ./cmd/shelf
```


## Shell Completion

Generate completion for your shell:

```bash
shelf completion zsh
shelf completion bash
shelf completion fish
shelf completion powershell
```

Examples:

### zsh

```bash
mkdir -p "${HOME}/.zsh/completions"
shelf completion zsh > "${HOME}/.zsh/completions/_shelf"
echo 'fpath=("${HOME}/.zsh/completions" $fpath)' >> ~/.zshrc
echo 'autoload -Uz compinit && compinit' >> ~/.zshrc
```

### bash

```bash
mkdir -p "${HOME}/.local/share/bash-completion/completions"
shelf completion bash > "${HOME}/.local/share/bash-completion/completions/shelf"
```

### fish

```bash
mkdir -p "${HOME}/.config/fish/completions"
shelf completion fish > "${HOME}/.config/fish/completions/shelf.fish"
```

### PowerShell

```powershell
shelf completion powershell | Out-String | Invoke-Expression
```

## Quick Start

```bash
# initialize
shelf init

# main workspace
shelf
shelf cockpit

# cockpit launchers
shelf calendar
shelf tree
shelf board
shelf review
shelf now

# script-friendly queries
shelf ls --status open --json
shelf link --from 01AAA --to 01BBB --type depends_on
shelf links 01AAA
shelf next
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
- `shelf link`
- `shelf unlink`
- `shelf links`
- `shelf ls`
- `shelf next`

Most daily editing still happens inside Cockpit, but link management is also available through standalone commands.

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

- names from `config.toml` `link_types.names`
- one blocking relation from `config.toml` `link_types.blocking`
- default names are `depends_on` and `related`

## Quality Checks

```bash
gofmt -w .
go test ./...
go test -race ./...
go vet ./...
```
