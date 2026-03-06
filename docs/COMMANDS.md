# COMMANDS (CLI Specification)

## Common

- All commands support `--root <dir>` to explicitly select the project root (directory that contains `.shelf/`).
- If `--root` is omitted, commands search upward from the current directory for `.shelf/`.
- If local `.shelf/` is not found, commands fall back to global config `default_root`.
- If `.shelf/` cannot be found, commands fail with a non-zero exit code.
- `init` is the only command that does not require existing `.shelf/`.
- `--show-id`, `-i`: show IDs in `ls` / `tree` / interactive task selectors.
- Task selectors always show a body preview (or `(empty body)`).
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
- Interactive mode (TTY only): Title -> Kind -> Status -> Parent.

Flags:

- `--title <str>`
- `--kind <kind>` (defaults to config `default_kind`)
- `--status <status>` (defaults to config `default_status`)
- `--due <YYYY-MM-DD>` (optional)
- `--parent <id|root>`
- `--body <str>`

Output includes full ID for copy-paste.

## shelf ls

Flat task list.
ID is omitted from default display. Parent is shown as `root` or parent title.

Flags:

- `--kind <kind>` (repeatable include filter)
- `--status <status>` (repeatable include filter)
- `--not-kind <kind>` (repeatable exclude filter)
- `--not-status <status>` (repeatable exclude filter)
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
Unknown `kind` / `status` values return an error.

Examples:

- `shelf ls --kind todo --status open`
- `shelf ls --not-status done --not-status cancelled`
- `shelf ls --status open --status in_progress --status blocked`
- `shelf ls --kind todo --not-status done --not-status cancelled`
- `shelf ls --ready --overdue`
- `shelf ls --json`

## shelf next

List actionable tasks (`open`/`in_progress` and unblocked by dependencies).

Flags:

- `--limit <n>` (default: 50)
- `--json`

## shelf tree

Render tree based on `parent`.
ID is omitted from tree output by default.

Flags:

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
- `--due <YYYY-MM-DD>`
- `--clear-due`
- `--parent <id|root>`
- `--body <str>` (replace body)
- `--append-body <str>` (append text)

Parent updates validate existence and reject cycles.
When no update flags are passed on TTY, `set` opens an interactive multi-field editor.

## shelf mv <id>

Thin wrapper of `set` for parent updates.

Flags:

- `--parent <id|root>` (required)

## shelf done <id>

Shortcut for `set --status done`.

## shelf start <id>

Shortcut for `set --status in_progress`.

## shelf block <id>

Shortcut for `set --status blocked`.

## shelf cancel <id>

Shortcut for `set --status cancelled`.

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

`Linked: [A] --depends_on--> [B]`

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
- `--json`
