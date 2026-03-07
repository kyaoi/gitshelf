# COMMANDS

Current public CLI surface for `shelf`.

## Common

- `--root <dir>` selects the project root that contains `.shelf/`
- if `--root` is omitted, `shelf` searches upward from the current directory
- if no local `.shelf/` is found, `shelf` falls back to the global `default_root`
- `init` and `completion` do not require an existing `.shelf/`
- `--show-id`, `-i` enables task IDs in text output and task selectors

## shelf

- on TTY: opens `Cockpit` in `calendar` mode
- on non-TTY: prints help

## shelf init

Initialize or refresh the current shelf.

Creates and keeps only:

- `.shelf/config.toml`
- `.shelf/tasks/`
- `.shelf/edges/`

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
- `--format compact|detail|kanban`
- `--json`

Unknown kind/status/tag values fail fast.

## shelf next

Read-only shortlist of actionable tasks.

Flags:

- `--limit <n>`
- `--json`

## Notes

The current public CLI intentionally does not expose standalone commands for add/edit/show/set/mv/snooze/link/archive/history/import/export/github/view/doctor.

Those operations are expected to happen inside Cockpit.
