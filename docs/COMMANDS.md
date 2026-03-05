# COMMANDS (CLI Specification)

## Common

- All commands support `--root <dir>` to explicitly select the project root (directory that contains `.shelf/`).
- If `--root` is omitted, commands search upward from the current directory for `.shelf/`.
- If local `.shelf/` is not found, commands fall back to global config `default_root`.
- If `.shelf/` cannot be found, commands fail with a non-zero exit code.
- `init` is the only command that does not require existing `.shelf/`.

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
- Interactive mode (TTY only): Title -> Kind -> Parent.

Flags:

- `--title <str>`
- `--kind <kind>` (defaults to config `default_kind`)
- `--status <status>` (defaults to config `default_status`)
- `--parent <id|root>`
- `--body <str>`

Output includes full ID for copy-paste.

## shelf ls

Flat task list.

Flags:

- `--kind <kind>`
- `--status <status>`
- `--parent <id|root>`
- `--limit <n>` (default: 50)
- `--search <query>` (title/body partial match)

Default ordering is ULID ascending (creation order).

## shelf tree

Render tree based on `parent`.

Flags:

- `--from <id|root>` (default: `root`)
- `--max-depth <n>` (`0` means unlimited)
- `--status <status>` (display filter)

## shelf show <id>

Show task details:

- front matter fields
- body
- outbound and inbound link summary

## shelf set <id>

Update task fields.

Flags:

- `--title <str>`
- `--kind <kind>`
- `--status <status>`
- `--parent <id|root>`
- `--body <str>` (replace body)
- `--append-body <str>` (append text)

Parent updates validate existence and reject cycles.

## shelf mv <id>

Thin wrapper of `set` for parent updates.

Flags:

- `--parent <id|root>` (required)

## shelf done <id>

Shortcut for `set --status done`.

## shelf link

Create outbound link.

- Non-interactive mode requires `--from --to --type`.
- Interactive mode (TTY only): source -> destination -> type.

Flags:

- `--from <id>`
- `--to <id>`
- `--type <link_type>`

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
