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

# Update and move
./shelf set <task-id> --state done
./shelf mv <task-id> --parent root

# Link tasks
./shelf link --from <a> --to <b> --type depends_on
./shelf links <a>

# Check integrity
./shelf doctor
```

## Commands

- `shelf init [--root <dir>] [--force]`
- `shelf add [--root <dir>] [--title ... --kind ... --state ... --parent <id|root> --body ...]`
- `shelf ls [--root <dir>] [--kind ... --state ... --parent <id|root> --limit N --search ...]`
- `shelf tree [--root <dir>] [--from <id|root> --max-depth N --state ...]`
- `shelf show <id> [--root <dir>]`
- `shelf set <id> [--root <dir>] [--title ... --kind ... --state ... --parent ... --body ... --append-body ...]`
- `shelf mv <id> --parent <id|root> [--root <dir>]`
- `shelf done <id> [--root <dir>]`
- `shelf link [--root <dir>] [--from ... --to ... --type ...]`
- `shelf unlink [--root <dir>] [--from ... --to ... --type ...]`
- `shelf links <id> [--root <dir>]`
- `shelf doctor [--root <dir>]`

## Interactive by Default for Omitted Args

When required args/flags are omitted and stdin/stdout are TTY, gitshelf prompts interactively instead of failing.

- `add`: omitted `--title`
- `link` / `unlink`: omitted required flags
- `show` / `set` / `done` / `links`: omitted `<id>`
- `mv`: omitted `<id>` and/or `--parent`

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
```

Task file format (`.shelf/tasks/<id>.md`):

```md
+++
id = "01..."
title = "Example"
kind = "todo"
state = "open"
parent = "01..." # optional
created_at = "2026-03-05T12:34:56+09:00"
updated_at = "2026-03-05T12:34:56+09:00"
+++

Body text...
```

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

### Can I move `.shelf`?

Use `--root <dir>` on every command to target a specific project root.
