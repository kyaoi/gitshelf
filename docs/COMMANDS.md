# COMMANDS (CLI Specification)

## Common

- All commands support `--root <dir>` to explicitly select the project root (directory that contains `.shelf/`).
- If `--root` is omitted, commands search upward from the current directory for `.shelf/`.
- If local `.shelf/` is not found, commands fall back to global config `default_root`.
- If `.shelf/` cannot be found, commands fail with a non-zero exit code.
- `init` is the only command that does not require existing `.shelf/`.
- `completion` is also available without existing `.shelf/`.
- `--show-id`, `-i`: show IDs in list/tree/link-like text outputs and interactive task selectors.
- Task selectors always show a body preview (or `(empty body)`).
- Enum selectors and non-selection commands do not show body preview.
- Colorized output is enabled by default on TTY.
- Set `NO_COLOR=1` to disable color output.
- Set `CLICOLOR_FORCE=1` to force color output.

## shelf init

Initialize `.shelf/` layout.

- Creates:
  - `.shelf/config.toml`
  - `.shelf/tasks/`
  - `.shelf/edges/`
- Existing directories are preserved.
- Existing `config.toml` is preserved unless `--force` is passed.
- `--global` writes global config (`GlobalConfigPath`) and initializes `.shelf/` under `default_root`.

Flags:

- `--force`: overwrite `config.toml` with defaults.
- `--global`: initialize global default root + global config.

## shelf add

Create a task.

- Non-interactive mode: `--title` is required.
- Interactive mode (TTY only): guided steps (Title -> Kind -> Status) then review/edit (`Tags`/`Due`/`Repeat`/`Parent`) before create.

Flags:

- `--title <str>`
- `--kind <kind>` (defaults to config `default_kind`)
- `--status <status>` (defaults to config `default_status`)
- `--tag <tag>` (repeatable; free input, new tags are added to config catalog)
- `--due <YYYY-MM-DD|today|tomorrow|+Nd|-Nd|next-week|this-week|mon..sun|next-mon..next-sun|in N days>` (optional)
- `--repeat-every <N>d|<N>w|<N>m|<N>y>` (optional)
- `--parent <id|root>`
- `--body <str>`

Output includes full ID for copy-paste.

## shelf capture

Quick capture command for inbox queue.

- Always creates `kind=inbox` and `status=open`
- Accepts title by positional args (`capture <title...>`) or `--title`
- If title is missing:
  - TTY: prompt for title
  - non-TTY: fail with `<title> を指定してください`

Flags:

- `--title <str>`
- `--tag <tag>` (repeatable)
- `--due <YYYY-MM-DD|today|tomorrow|+Nd|-Nd|next-week|this-week|mon..sun|next-mon..next-sun|in N days>`
- `--body <str>`

## shelf ls

Flat task list.
ID is omitted from default display. Parent is shown as `root` or parent title.

Flags:

- `--view <name>` (built-in: `active|ready|blocked|overdue`, or config view)
- `--preset <name>` (output preset for `ls`)
- `--kind <kind>` (repeatable include filter)
- `--status <status>` (repeatable include filter)
- `--tag <tag>` (repeatable include filter)
- `--not-kind <kind>` (repeatable exclude filter)
- `--not-status <status>` (repeatable exclude filter)
- `--not-tag <tag>` (repeatable exclude filter)
- `--ready` (actionable tasks only)
- `--blocked-by-deps` (tasks blocked by unresolved `depends_on`)
- `--due-before <YYYY-MM-DD>`
- `--due-after <YYYY-MM-DD>`
- `--overdue`
- `--no-due`
- `--parent <id|root>`
- `--limit <n>` (default: 50)
- `--search <query>` (title/body partial match)
- `--json`

Default ordering is ULID ascending (creation order).
Unknown `kind` / `status` / `tag` values return an error.

Examples:

- `shelf ls --kind todo --status open`
- `shelf ls --tag backend`
- `shelf ls --not-status done --not-status cancelled`
- `shelf ls --not-tag wip`
- `shelf ls --status open --status in_progress --status blocked`
- `shelf ls --kind todo --not-status done --not-status cancelled`
- `shelf ls --ready --overdue`
- `shelf ls --json`

## shelf next

List actionable tasks (`open`/`in_progress` and unblocked by dependencies).

Flags:

- `--view <name>` (built-in or config view)
- `--preset <name>` (output preset for `next`)
- `--limit <n>` (default: 50)
- `--json`

## shelf view

Manage saved views in `.shelf/config.toml`.

Subcommands:

- `shelf view list|ls [--json]`
- `shelf view show <name> [--json]`
- `shelf view set <name> [filter flags...]`
- `shelf view copy <src> <dst>`
- `shelf view rename <src> <dst>`
- `shelf view merge <dst> --from <name> --from <name> [--strategy overlay|union]`
- `shelf view delete|rm <name>`

`view set` supports:

- `--kind`, `--status`, `--tag`, `--not-kind`, `--not-status`, `--not-tag` (repeatable)
- `--ready`, `--blocked-by-deps`
- `--due-before`, `--due-after`, `--overdue`, `--no-due`
- `--parent`, `--search`, `--limit`

Rules:

- built-in views (`active|ready|blocked|overdue`) cannot be overwritten/deleted
- at least one filter flag is required for `view set`

## shelf preset

Manage output presets in `.shelf/config.toml`.

Subcommands:

- `shelf preset list|ls [--json]`
- `shelf preset show <name> [--json]`
- `shelf preset set <name> --command <ls|tree|next|agenda|today> [--format ... --view ... --limit ...]`
- `shelf preset delete|rm <name>`

## shelf agenda

Due-oriented daily list.

Default target statuses are `open`, `in_progress`, `blocked`.

Flags:

- `--view <name>` (built-in or config view)
- `--preset <name>` (output preset for `agenda`)
- `--days <n>` (upcoming range, default 7)
- `--kind <kind>` (repeatable include filter)
- `--status <status>` (repeatable include filter)
- `--tag <tag>` (repeatable include filter)
- `--not-kind <kind>` (repeatable exclude filter)
- `--not-status <status>` (repeatable exclude filter)
- `--not-tag <tag>` (repeatable exclude filter)
- `--json`

## shelf today

Show only overdue + today tasks.

Default target statuses are `open`, `in_progress`, `blocked`.

Flags:

- `--view <name>` (built-in or config view)
- `--preset <name>` (output preset for `today`)
- `--kind`, `--status`, `--not-kind`, `--not-status` (repeatable)
- `--format <compact|detail>`
- `--json`
- `--include-archived`
- `--only-archived`
- `--carry-over` (move overdue active tasks to today)
- `--yes` (required on non-TTY with `--carry-over`)

## shelf tree

Render tree based on `parent`.
ID is omitted from tree output by default.

Flags:

- `--view <name>` (built-in or config view; due/readiness views are rejected)
- `--preset <name>` (output preset for `tree`)
- `--from <id|root>` (default: `root`)
- `--max-depth <n>` (`0` means unlimited)
- `--kind <kind>` (repeatable include filter)
- `--status <status>` (repeatable include filter)
- `--not-kind <kind>` (repeatable exclude filter)
- `--not-status <status>` (repeatable exclude filter)
- `--json`

## shelf show <id>

Show task details:

- front matter fields
- body (freeform notes)
- hierarchy path + subtree
- outbound and inbound link summary

Flags:

- `--no-body` (hide body section)
- `--only-body` (print body only)
- `--json`

## shelf explain <id>

Explain why a task matches or does not match built-in views and default command filters.

Also prints current readiness (`ready`, unresolved `depends_on`).

Flags:

- `--view <name>` (add custom/built-in view explanation)
- `--json`

## shelf edit [id]

Open task file (`.shelf/tasks/<id>.md`) in editor.

- Editor resolution order: `$VISUAL` -> `$EDITOR` -> `vi`
- Opens the whole task file (front matter + body)
- If `<id>` is omitted:
  - TTY: task selector is shown
  - non-TTY: command fails with `<id> を指定してください`

## shelf set <id>

Update task fields.

Flags:

- `--title <str>`
- `--kind <kind>`
- `--status <status>`
- `--tag <tag>` (repeatable add)
- `--untag <tag>` (repeatable remove)
- `--clear-tags`
- `--due <YYYY-MM-DD|today|tomorrow|+Nd|-Nd|next-week|this-week|mon..sun|next-mon..next-sun|in N days>`
- `--clear-due`
- `--repeat-every <N>d|<N>w|<N>m|<N>y>`
- `--clear-repeat`
- `--parent <id|root>`
- `--body <str>` (replace body)
- `--append-body <str>` (append text)

Parent updates validate existence and reject cycles.
When no update flags are passed on TTY, `set` opens an interactive multi-field editor.
Interactive `set` includes `Body (replace)` and `Body (append)`, then shows a review step before apply.

## shelf snooze <id>

Adjust task due date.

Flags:

- `--by <Nd>` (relative day shift, e.g. `2d`, `-1d`)
- `--to <YYYY-MM-DD|today|tomorrow|+Nd|-Nd|next-week|this-week|mon..sun|next-mon..next-sun|in N days>` (absolute set)

Rules:

- exactly one of `--by` or `--to` is required
- if `<id>` is omitted and TTY is available, task selector is shown

## shelf archive <id>

Set `archived_at` to current local RFC3339 timestamp.

## shelf unarchive <id>

Clear `archived_at`.

## shelf mv <id>

Thin wrapper of `set` for parent updates.

Flags:

- `--parent <id|root>` (required)

## shelf done <id>

Shortcut for `set --status done`.

Recurring tasks (`repeat_every` present):

- `--recurring-action create`: mark current task done and create next task
- `--recurring-action reopen`: keep same task and advance due/status to open

On non-TTY, recurring `done` requires `--recurring-action`.

## shelf start <id>

Shortcut for `set --status in_progress`.

## shelf block <id>

Shortcut for `set --status blocked`.

## shelf cancel <id>

Shortcut for `set --status cancelled`.

## shelf reopen <id>

Shortcut for `set --status open`.

## shelf link

Create outbound link.

- Non-interactive mode requires `--from --to --type`.
- Interactive mode (TTY only): source -> destination -> type.

Flags:

- `--from <id>`
- `--to <id>`
- `--type <link_type>`

Supported link types:

- `depends_on`
- `related`

Output keeps direction explicit:

`Linked: A --depends_on--> B`

With `--show-id`, short IDs are included.

## shelf unlink

Remove outbound link.

- Non-interactive mode requires `--from --to --type`.
- Interactive mode (TTY only) lets users select existing outbound edge from a source task.

Flags:

- `--from <id>`
- `--to <id>`
- `--type <link_type>`

## shelf links <id>

Show links of a task:

- outbound: from `.shelf/edges/<id>.toml`
- inbound: reverse lookup by scanning all edge files

Flags:

- `--transitive` (show recursive `depends_on` closure)
- `--json`

## shelf deps <id>

Show `depends_on` prerequisites and dependents of a task.

Flags:

- `--transitive` (recursive closure)
- `--reverse` (print dependents first)
- `--graph` (render ASCII graph)
- `--json`

## shelf export

Export full `.shelf` data as JSON (`config`, `tasks`, `edges`).

Flags:

- `--out <path>` (`-` for stdout)

## shelf import

Import full `.shelf` data from JSON export format.

This operation replaces current `tasks`/`edges` and writes `config`.
Undo snapshot is taken before import.

Flags:

- `--in <path>` (`-` for stdin)
- `--validate-only` (validate payload only)
- `--dry-run` (show summary without write)
- `--merge` (merge into current shelf; incoming wins on conflicts)
- `--replace` (replace current shelf; default mode)

## shelf undo

Restore last snapshot taken by mutating commands (`add/set/mv/done/start/block/cancel/reopen/snooze/archive/unarchive/link/unlink/import`).

Flags:

- `--steps <n>` (default: 1)

## shelf redo

Re-apply undone mutating actions.

Flags:

- `--steps <n>` (default: 1)

## shelf history

Show mutating action history (`apply`/`undo`/`redo`) from `.shelf/history/actions.log`.

Flags:

- `--limit <n>` (default: 50)
- `--json`

### shelf history show <entry|snapshot_id>

Show details for one history entry (index from `history` list output or snapshot ID).

Flags:

- `--json`

## shelf doctor

Integrity checker for `.shelf/`:

- task parent existence
- parent cycle detection
- unknown `kind` / `status`
- edge destination existence
- unknown `link_type`
- duplicate edge detection
- edge source existence

Outputs file path + task ID + issue message for manual fixes.

Flags:

- `--fix` (apply safe normalization before checks)
- `--strict` (emit additional warnings, e.g. `todo` without `due_on`)
- `--json`

Doctor output includes per-issue `hint` text for common recovery paths.
Mutating commands acquire `.shelf/.write.lock` to prevent concurrent writes.

## shelf completion

Generate shell completions:

- `shelf completion bash`
- `shelf completion zsh`
- `shelf completion fish`
- `shelf completion powershell`
