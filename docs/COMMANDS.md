# COMMANDS

Current public CLI surface for `shelf`.

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

### shelf config copy-preset list

List saved advanced copy presets.

Flags:

- `--json`

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
- `--format compact|detail|kanban|tree`
- `--preset <now|review|board>`
- `--json`

Unknown kind/status/tag values fail fast.

`--preset` applies read-only defaults similar to the matching Cockpit view.
Explicit flags still win.

## shelf next

Read-only shortlist of actionable tasks.

Flags:

- `--limit <n>`
- `--json`

## shelf show

Show one task with inspector-style details.

Usage:

- `shelf show <task-id>`

Flags:

- `--json`

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

- `--json`

Text output uses tree/path labels so duplicate titles are distinguishable.
IDs stay hidden unless `--show-id` is enabled.

## Notes

Most daily editing happens inside Cockpit.
Link management is also available through standalone commands.
