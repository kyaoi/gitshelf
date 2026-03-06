# gitshelf

`gitshelf` is a lightweight Git-friendly CLI task manager.

- CLI command: `shelf`
- Data root: `.shelf/`
- Tasks: `.shelf/tasks/<id>.md` (flat files, one task per file)
- Links: `.shelf/edges/<src_id>.toml` (outbound edges only)
- Tree: represented by each task's `parent`
- Extra relations: represented by links (`depends_on`, `related`)

## Install

```bash
go build -o shelf ./cmd/shelf
```

## Quick Start

```bash
# Initialize in current directory
./shelf init

# Initialize global shelf (writes global config + creates global .shelf)
./shelf init --global

# Add tasks
./shelf add --title "Weekly Goal"
./shelf add --title "Monday Plan" --parent root

# List and inspect
./shelf ls
./shelf tree
./shelf show <task-id>
./shelf edit <task-id>

# Update and move
./shelf set <task-id> --status done
./shelf mv <task-id> --parent root
./shelf start <task-id>
./shelf block <task-id>
./shelf cancel <task-id>
./shelf next
./shelf agenda
./shelf today
./shelf snooze <task-id> --by 2d
./shelf archive <task-id>
./shelf unarchive <task-id>
./shelf reopen <task-id>
./shelf undo
./shelf redo
./shelf history
./shelf explain <task-id>

# Link tasks
./shelf link --from <a> --to <b> --type depends_on
./shelf links <a> --transitive

# Manage views / backup
./shelf view list
./shelf view set focus --ready --limit 20
./shelf preset set ls_focus --command ls --view active --format detail --limit 20
./shelf deps <task-id> --transitive
./shelf export --out backup.json
./shelf import --validate-only --in backup.json
./shelf import --merge --in backup.json
./shelf completion zsh

# Check integrity
./shelf doctor
./shelf doctor --fix
```

## Commands

- `shelf init [--root <dir>] [--force]`
- `shelf add [--root <dir>] [--title ... --kind ... --status ... --due YYYY-MM-DD|today|tomorrow|+Nd|-Nd|next-week|mon..sun --repeat-every <N>d|<N>w|<N>m|<N>y --parent <id|root> --body ...]`
- `shelf ls [--root <dir>] [--preset <name> --view <name> --kind ... --status ... --not-kind ... --not-status ... --ready --blocked-by-deps --due-before ... --due-after ... --overdue --no-due --parent <id|root> --limit N --search ... --json]`
- `shelf view list|show|set|copy|rename|merge|delete [--root <dir>] ...`
- `shelf preset list|show|set|delete [--root <dir>] ...`
- `shelf next [--root <dir>] [--view <name> --limit N --json]`
- `shelf agenda [--root <dir>] [--preset <name> --view <name> --days N --kind ... --status ... --not-kind ... --not-status ... --json]`
- `shelf today [--root <dir>] [--preset <name> --view <name> --carry-over --yes --kind ... --status ... --not-kind ... --not-status ... --json]`
- `shelf tree [--root <dir>] [--preset <name> --view <name> --from <id|root> --max-depth N --kind ... --status ... --not-kind ... --not-status ... --json]`
- `shelf show <id> [--root <dir>] [--no-body --only-body --json]`
- `shelf explain <id> [--root <dir>] [--view <name> --json]`
- `shelf edit [id] [--root <dir>]`
- `shelf set <id> [--root <dir>] [--title ... --kind ... --status ... --due YYYY-MM-DD|today|tomorrow --clear-due --repeat-every ... --clear-repeat --parent ... --body ... --append-body ...]`
- `shelf snooze <id> [--root <dir>] (--by <Nd> | --to YYYY-MM-DD|today|tomorrow)`
- `shelf archive <id> [--root <dir>]`
- `shelf unarchive <id> [--root <dir>]`
- `shelf mv <id> --parent <id|root> [--root <dir>]`
- `shelf done <id> [--root <dir>] [--recurring-action create|reopen]`
- `shelf start <id> [--root <dir>]`
- `shelf block <id> [--root <dir>]`
- `shelf cancel <id> [--root <dir>]`
- `shelf reopen <id> [--root <dir>]`
- `shelf link [--root <dir>] [--from ... --to ... --type ...]`
- `shelf unlink [--root <dir>] [--from ... --to ... --type ...]`
- `shelf links <id> [--root <dir>] [--transitive --json]`
- `shelf deps <id> [--root <dir>] [--transitive --reverse --json]`
- `shelf export [--root <dir>] [--out <path>|-]`
- `shelf import [--root <dir>] [--in <path>|- --validate-only --dry-run --merge --replace]`
- `shelf undo [--root <dir>]`
- `shelf redo [--root <dir>]`
- `shelf history [--root <dir>] [--limit N --json]`
- `shelf completion bash|zsh|fish|powershell`
- `shelf doctor [--root <dir>] [--fix --json]`

Global display flags:

- `--show-id`, `-i`: show IDs in `ls` / `tree` / interactive task selectors
- task selectors always show body preview by default (`(empty body)` when body is empty)

Color output:

- TTY output is colorized by default for readability.
- `NO_COLOR=1` disables color.
- `CLICOLOR_FORCE=1` forces color even when output is not a TTY.

## Kind and Status

- `kind`: task category (`todo`, `idea`, `memo`, ...)
- `status`: task progress (`open`, `in_progress`, `blocked`, `done`, `cancelled`)
- `due_on` (`YYYY-MM-DD`): optional for all kinds (`todo`/`memo`/`idea` etc.)
- `repeat_every` (`<N>d|<N>w|<N>m|<N>y`): optional recurring interval
- `archived_at` (RFC3339): set by `archive`, cleared by `unarchive`
- CLI input for due also accepts `today`, `tomorrow`, `+Nd`, `-Nd`, `next-week`, `mon..sun` and stores normalized date

## Link Types

Supported `link_types` are only:

- `depends_on`
- `related`

`derived_from` is not used.

## ls Filter Examples

```bash
./shelf ls --kind todo --status open
./shelf ls --not-status done --not-status cancelled
./shelf ls --status open --status in_progress --status blocked
./shelf ls --kind todo --not-status done --not-status cancelled
./shelf ls --ready --overdue
./shelf ls --blocked-by-deps
./shelf ls --view active
./shelf ls --include-archived
./shelf ls --json
```

`ls` / `tree` output omits IDs by default for readability.
Use `show` to inspect full metadata and hierarchy details.

## Interactive by Default for Omitted Args

When required args/flags are omitted and stdin/stdout are TTY, gitshelf prompts interactively instead of failing.

- `add`: omitted `--title`
- `link` / `unlink`: omitted required flags
- `show` / `set` / `done` / `links`: omitted `<id>`
- `mv`: omitted `<id>` and/or `--parent`

`add` and `set` interactive flows use an editable field session:

- choose field (`Title`/`Kind`/`Status`/`Due`/`Parent`/`Body`)
- update repeatedly
- confirm and save

Task selectors always show body preview.

## Saved Views

`--view <name>` can use:

- built-in views: `active`, `ready`, `blocked`, `overdue`
- custom views from `.shelf/config.toml`:

```toml
[views."only_done"]
statuses = ["done"]
```

You can also manage this from CLI:

```bash
./shelf view list
./shelf view show only_done
./shelf view set only_done --status done
./shelf view copy only_done only_done_copy
./shelf view rename only_done_copy done_only
./shelf view merge done_union --from only_done --from active --strategy union
./shelf view delete only_done
```

Output presets:

```bash
./shelf preset set ls_focus --command ls --view active --format detail --limit 20
./shelf ls --preset ls_focus
```

In non-TTY mode, interactive prompts are disabled and missing values produce clear errors.

## Global Shelf and Fallback

gitshelf supports a global default root configured at:

- `~/.config/gitshelf/config.toml` (Linux default location via `os.UserConfigDir()`)

Global config format:

```toml
default_root = "/abs/path/to/store"
```

Resolution order for commands:

1. Use `--root` when provided.
2. Otherwise search upward from cwd for `.shelf/config.toml`.
3. If not found, use global config `default_root`.
4. If global config is missing, command fails with guidance to run:
   - `shelf init --global`

## Storage Format

```text
.shelf/
  config.toml
  tasks/
    <id>.md
  edges/
    <src_id>.toml
  history/
    index.json
    actions.log
    snapshots/
```

Task file format (`.shelf/tasks/<id>.md`):

```md
+++
id = "01..."
title = "Example"
kind = "todo"
status = "open"
due_on = "2026-03-31" # optional
repeat_every = "1w" # optional
archived_at = "2026-03-31T11:22:33+09:00" # optional
parent = "01..." # optional
created_at = "2026-03-05T12:34:56+09:00"
updated_at = "2026-03-05T12:34:56+09:00"
+++

Body text...
```

Task files are split into:

- front matter: structured metadata (`title`, `kind`, `status`, `due_on`, `parent`, timestamps)
- body: freeform notes (`details`, `supplements`, `progress logs`, `ideas`, `references`)

`shelf show <id>` displays both metadata and body so the task context stays in one place.

Edge file format (`.shelf/edges/<src_id>.toml`):

```toml
[[edge]]
to = "01..."
type = "depends_on"
```

## FAQ

### What does `depends_on` mean?

`A depends_on B` means `B` must be completed before `A`.
The CLI always displays it as:

`A --depends_on--> B`

### Is interactive mode always available?

No. Interactive mode is enabled only when stdin/stdout are TTY.
In non-TTY mode, required flags must be provided.

## Local Quality Gate (mise + lefthook)

```bash
mise install
mise run hooks-install
mise run hooks-pre-commit
mise run hooks-pre-push
```

Hooks:

- `pre-commit`: staged gofmt check + `go test ./...`
- `pre-push`: `go test ./...`, `go test -race ./...`, `go vet ./...`

## Automated Backup Script

```bash
SHELF_ROOT=/path/to/repo ./scripts/backup_shelf.sh
```

Environment variables:

- `SHELF_BIN` (default: `shelf`)
- `SHELF_BACKUP_DIR` (default: `${SHELF_ROOT}/.shelf/backups`)
- `SHELF_BACKUP_KEEP` (default: `30`)

### Can I move `.shelf`?

Use `--root <dir>` on every command to target a specific project root.
