# COMMANDS

Current public CLI surface for `shelf`.

For machine-readable schemas, see [`OUTPUTS.md`](OUTPUTS.md).

## Common

- `--root <dir>` selects the project root that contains `.shelf/`
- `--root` and global `default_root` accept `~` shorthand such as `~/DailyTodo`
- if the supplied path points to `<root>/.shelf`, `shelf` normalizes it back to `<root>`
- paths inside `.shelf/` such as `.shelf/tasks` are rejected as invalid roots
- if `--root` is omitted, `shelf` searches upward from the current directory
- if no local `.shelf/` is found, `shelf` falls back to the global `default_root`
- `init` and `completion` do not require an existing `.shelf/`
- `--show-id`, `-i` enables task IDs in text output and task selectors
- `--git-on-exit <none|commit|commit_push>` overrides the post-Cockpit git action
- `--git-message <text>` overrides the commit message used by `--git-on-exit`
- task data location is controlled by `storage_root` in `.shelf/config.toml`

## shelf

- on TTY: opens `Cockpit` in `calendar` mode
- on non-TTY: prints help

## shelf init

Initialize or refresh the current shelf.

Creates and keeps only:

- `.shelf/config.toml`
- `<storage_root>/tasks/`
- `<storage_root>/edges/`

Flags:

- `--force`: rewrite `config.toml` with defaults
- `--global`: initialize the global default root and global config

## shelf completion

Generate shell completion.

Subcommands:

- `completion bash`
- `completion zsh`
- `completion fish`
- `completion powershell`

## shelf cockpit

Main TUI workspace.

Aliases:

- `cp`

TTY only.

Flags:

- `--mode <calendar|tree|board|review|now>`
- `--start <YYYY-MM-DD|today|tomorrow>`
- `--days <n>`
- `--months <n>`
- `--years <n>`
- `--limit <n>`
- `--kind <kind>` (repeatable)
- `--status <status>` (repeatable)
- `--tag <tag>` (repeatable)
- `--not-kind <kind>` (repeatable)
- `--not-status <status>` (repeatable)
- `--not-tag <tag>` (repeatable)

## shelf config

Persist user-facing config values.

### shelf config show

Show the effective config for the current root.

Flags:

- `--json`
- `--format compact|tsv|csv|jsonl`
- `--fields <name,...>` for `--format tsv|csv`
- `--header`
- `--no-header`

### shelf config copy-preset list

List saved advanced copy presets.

Flags:

- `--json`
- `--format compact|tsv|csv|jsonl`
- `--fields <name,...>` for `--format tsv|csv`
- `--header`
- `--no-header`

### shelf config copy-preset get

Show one saved advanced copy preset.

Usage:

- `shelf config copy-preset get <name>`

Flags:

- `--json`

### shelf config copy-preset rm

Remove one saved advanced copy preset.

Usage:

- `shelf config copy-preset rm <name>`

### shelf config copy-preset set

Create or update one saved advanced copy preset used by Cockpit `M`.

Flags:

- `--name <preset-name>`
- `--scope <task|subtree>`
- `--subtree-style <indented|tree>` optional, controls `{{subtree}}` rendering
- `--template <text>`
- `--join-with <text>` optional, defaults to `commands.cockpit.copy_separator`

Supported template placeholders:

- `{{title}}`
- `{{path}}`
- `{{body}}`
- `{{subtree}}`

## Launcher Commands

These commands are thin wrappers around `shelf cockpit --mode ...`.

### shelf calendar

Aliases:

- `cal`

Starts Cockpit in `calendar` mode.

### shelf tree

Aliases:

- `tr`

Starts Cockpit in `tree` mode.

### shelf board

Aliases:

- `kb`

Starts Cockpit in `board` mode.

### shelf review

Aliases:

- `rv`

Starts Cockpit in `review` mode.

### shelf now

Aliases:

- `nw`

Starts Cockpit in `now` mode.

All launcher commands accept the same Cockpit flags:

- `--start`
- `--days`
- `--months`
- `--years`
- `--limit`
- `--kind`
- `--status`
- `--tag`
- `--not-kind`
- `--not-status`
- `--not-tag`

All launcher commands are TTY-only.

## shelf ls

Read-only task listing for scripts and quick inspection.

Flags:

- `--kind <kind>` (repeatable)
- `--status <status>` (repeatable)
- `--tag <tag>` (repeatable)
- `--not-kind <kind>` (repeatable)
- `--not-status <status>` (repeatable)
- `--not-tag <tag>` (repeatable)
- `--ready`
- `--blocked-by-deps`
- `--due-before <date>`
- `--due-after <date>`
- `--overdue`
- `--no-due`
- `--parent <id|root>`
- `--search <text>`
- `--limit <n>`
- `--include-archived`
- `--only-archived`
- `--format compact|detail|kanban|tree|tsv|csv|jsonl`
- `--preset <now|review|board>`
- `--fields <name,...>` for `--format tsv|csv`
- `--header`
- `--no-header`
- `--sort <id|title|path|kind|status|due_on|created_at|updated_at>`
- `--reverse`
- `--group-by <status|kind|parent>`
- `--count`
- `--json`
- `--schema <v1|v2>` for `--json`, `--format jsonl`, `--format tsv`, and `--format csv`

Unknown kind/status/tag values fail fast.

`--preset` applies read-only defaults similar to the matching Cockpit view.
Explicit flags still win.

`--schema v1` keeps the current compatibility shape.
`--schema v2` uses canonical machine-readable field names.

`--format tsv` and `--format csv` use the same task fields.
TSV defaults to no header. CSV defaults to a header row.
`--header` and `--no-header` override those defaults.

Task fields:

- `v1`: `id`, `title`, `path`, `kind`, `status`, `due_on`, `repeat_every`, `archived_at`, `parent`, `parent_path`, `tags`, `file`
- `v2`: `id`, `title`, `path`, `kind`, `status`, `due_on`, `repeat_every`, `archived_at`, `parent_id`, `parent_path`, `tags`, `file`

`--fields` can reorder or reduce those columns.
`--format jsonl` prints one task object per line using the same record shape as `--json`.
`--group-by` adds a `group` field to tabular/JSON output and prints grouped sections in text output.
`--group-by` cannot be combined with `--format kanban`, `--format tree`, or `--count`.
`--count` prints only the total number of matching tasks. With `--json`, it returns `{ "count": N }`.

## shelf next

Read-only shortlist of actionable tasks.

Flags:

- `--limit <n>`
- `--format compact|tsv|csv|jsonl`
- `--fields <name,...>` for `--format tsv|csv`
- `--header`
- `--no-header`
- `--sort <id|title|path|kind|status|due_on|created_at|updated_at>`
- `--reverse`
- `--count`
- `--json`
- `--schema <v1|v2>` for `--json`, `--format jsonl`, `--format tsv`, and `--format csv`

`--count` prints only the total number of ready tasks. With `--json`, it returns `{ "count": N }`.

`--schema v1` keeps the current compatibility shape.
`--schema v2` uses canonical machine-readable field names.

## shelf show

Show one task with inspector-style details.

Usage:

- `shelf show <task-id>`

Flags:

- `--format compact|tsv|csv|jsonl`
- `--fields <name,...>` for `--format tsv|csv`
- `--header`
- `--no-header`
- `--json`
- `--schema <v1|v2>` for `--json`, `--format jsonl`, `--format tsv`, and `--format csv`

`--schema v1` keeps the current compatibility shape.
`--schema v2` keeps only canonical task field names in machine-readable output.

`--format tsv` and `--format csv` print one row for the selected task.
Available fields:

- `v1`: `id`, `title`, `path`, `kind`, `status`, `tags`, `due_on`, `repeat_every`, `archived_at`, `parent_id`, `parent`, `parent_path`, `file`, `created_at`, `updated_at`, `body`
- `v2`: `id`, `title`, `path`, `kind`, `status`, `tags`, `due_on`, `repeat_every`, `archived_at`, `parent_id`, `parent_path`, `file`, `created_at`, `updated_at`, `body`
- `outbound_count`, `inbound_count`

`--format jsonl` prints one task object per line.

## shelf link

Create an outbound link.

Flags:

- `--from <id>`
- `--to <id>`
- `--type <link-type>`

If `--type` is omitted, the configured blocking link type is used.

## shelf unlink

Remove an outbound link.

Flags:

- `--from <id>`
- `--to <id>`
- `--type <link-type>`

If `--type` is omitted, the configured blocking link type is used.

## shelf links

Show outbound and inbound links for one task.

Usage:

- `shelf links <task-id>`

Flags:

- `--format compact|tsv|csv|jsonl`
- `--fields <name,...>` for `--format tsv|csv`
- `--header`
- `--no-header`
- `--summary`
- `--json`
- `--schema <v1|v2>` for `--json`, `--format jsonl`, `--format tsv`, and `--format csv`

`--schema v1` keeps the current compatibility shape.
`--schema v2` keeps only canonical edge field names in machine-readable output.

`--format tsv` and `--format csv` print one row per edge.
Available fields:

- `v1`: `direction`, `type`, `source_id`, `source_title`, `source_path`, `source_file`, `target_id`, `target_title`, `target_path`, `target_file`, `task_id`, `task_title`, `task_path`, `task_file`, `other_id`, `other_title`, `other_path`, `other_file`
- `v2`: `direction`, `type`, `source_id`, `source_title`, `source_path`, `source_file`, `target_id`, `target_title`, `target_path`, `target_file`

`--format jsonl` prints one edge object per line.
`--summary` switches the output to aggregated `direction/type/count` rows.

Text output uses tree/path labels so duplicate titles are distinguishable.
IDs stay hidden unless `--show-id` is enabled.

## Notes

Most daily editing happens inside Cockpit.
Link management is also available through standalone commands.
