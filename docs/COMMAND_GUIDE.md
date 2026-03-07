# COMMAND_GUIDE

Practical guide for deciding which `shelf` command to use, why it exists, and how to use it well.

`docs/COMMANDS.md` is the CLI specification.
This file is the operator guide.

## Quick Routing

| Goal | Command |
|---|---|
| Initialize `.shelf/` | `shelf init` |
| Capture something quickly | `shelf capture` |
| Create a structured task | `shelf add` |
| Triage inbox items | `shelf triage` |
| See everything flat | `shelf ls` |
| See hierarchy | `shelf tree` |
| Inspect one task deeply | `shelf show` |
| Open the raw task file | `shelf edit` |
| Update metadata/body | `shelf set` |
| Move in the tree | `shelf mv` |
| Change status quickly | `shelf done`, `start`, `block`, `cancel`, `reopen` |
| See what to work on now | `shelf next`, `review`, `today`, `agenda` |
| See due dates on a calendar | `shelf calendar` |
| Manage status visually | `shelf board` |
| Add links between tasks | `shelf link`, `unlink`, `links`, `deps` |
| Link GitHub issues/PRs | `shelf github`, `shelf sync github` |
| Track estimate/time | `shelf estimate`, `shelf track` |
| Trigger local reminders | `shelf notify` |
| Save reusable trees | `shelf template` |
| Save reusable filters/output | `shelf view`, `shelf preset` |
| Explain why a task is shown/blocked | `shelf explain` |
| Repair or verify data | `shelf doctor` |
| Export/import all data | `shelf export`, `shelf import` |
| Undo/redo changes | `shelf undo`, `shelf redo`, `history` |
| Generate shell completion | `shelf completion` |

## Setup

### `shelf init`

Meaning:
Create the `.shelf/` directory layout and default config.

Use when:
Start a new project-local shelf or initialize the global fallback shelf.

Common patterns:

```bash
shelf init
shelf init --root /path/to/project
shelf init --global
```

Notes:
- Safe to rerun.
- `--force` replaces the config with defaults.

### `shelf completion`

Meaning:
Generate shell completion scripts.

Use when:
You want tab completion for commands and flags in `bash`, `zsh`, `fish`, or PowerShell.

## Capture and Creation

### `shelf capture`

Meaning:
Fast inbox entry point.

Use when:
You want to save something before deciding where it belongs.

What it does:
- always creates `kind=inbox`
- always creates `status=open`
- accepts optional tags, due date, and body

Common patterns:

```bash
shelf capture "Call vendor"
shelf capture --title "Research API options" --tag backend
```

### `shelf add`

Meaning:
Create a regular task with explicit structure.

Use when:
You already know the title, kind, status, parent, or other metadata.

Common patterns:

```bash
shelf add --title "Weekly Goal"
shelf add --title "Monday Plan" --parent root
shelf add --title "Refactor parser" --kind todo --status in_progress
```

Notes:
- Interactive mode is used when `--title` is omitted in TTY.
- Parent selection is hierarchical by default.

### `shelf template`

Meaning:
Save and reuse a task subtree as a template.

Use when:
You repeat the same structure every week, sprint, or project.

Subcommands:
- `template save`: snapshot a subtree
- `template list|ls`: show saved templates
- `template show`: inspect a template
- `template apply`: recreate tasks from a template
- `template delete|rm`: remove a template

## Inbox Processing and Review

### `shelf triage`

Meaning:
Process inbox items in batch.

Use when:
You want to classify or finish items created by `capture`.

Modes:
- interactive: edit one task at a time
- auto: apply one status action to the current inbox slice

### `shelf review`

Meaning:
Daily dashboard for decision making.

Use when:
You want one screen that answers:
- what is still in inbox
- what is overdue
- what is due today
- what is blocked
- what is ready

### `shelf next`

Meaning:
Show actionable tasks only.

Use when:
You want a short list of tasks that are ready right now.

### `shelf today`

Meaning:
Focus on overdue and today-due work.

Use when:
You are planning the current day.

Special mode:
- `--carry-over` moves overdue active tasks to today

### `shelf agenda`

Meaning:
Bucket tasks by due timing.

Use when:
You want a due-oriented review with `Overdue`, `Today`, `Tomorrow`, `Upcoming`, `Later`, and `No due`.

### `shelf calendar`

Meaning:
Due-date calendar view.

Use when:
You want a date-first perspective instead of a task-first list.

Notes:
- `calendar` opens a TUI by default.
- Use `--json` in non-TTY contexts.
- If `--days` is omitted, config `calendar_default_days` is used.
- `--days` controls the range that the TUI can navigate inside.
- `--months` lets you open whole-month ranges such as one month or three months.
- The focused day has a task list and a body preview panel.
- You can add a task directly on the focused day with `a`; kind/status come from config defaults.
- You can edit or snooze the selected task directly from the TUI.
- You can also change the selected task status directly with `o/i/b/d/c`.
- If you change a task to a status outside the current filter, calendar keeps it visible until the next reload so context is not lost.

### `shelf board`

Meaning:
TUI board organized by status columns.

Use when:
You want to move through active work visually and change statuses in place.

Notes:
- TTY only.
- Columns follow config `statuses`.

## Listing and Inspection

### `shelf ls`

Meaning:
Flat list of tasks with powerful filtering.

Use when:
You want search, include/exclude filters, views, or JSON output.

Good examples:

```bash
shelf ls --kind todo --status open
shelf ls --not-status done --not-status cancelled
shelf ls --tag backend --ready
```

### `shelf tree`

Meaning:
Hierarchical view based on `parent`.

Use when:
The tree structure matters more than a flat filtered list.

### `shelf show`

Meaning:
Deep inspection for one task.

What it shows:
- front matter
- body
- hierarchy path
- context tree
- readiness
- inbound/outbound links

Use when:
You need the full picture for a single task.

### `shelf explain`

Meaning:
Explain readiness and view/filter matching for one task.

Use when:
You need to know why a task appears, why it does not appear, or why it is blocked.

### `shelf edit`

Meaning:
Open the raw task markdown file in `$VISUAL`, `$EDITOR`, or `vi`.

Use when:
You want full-file editing including front matter and body.

## Editing and Movement

### `shelf set`

Meaning:
Structured task update command.

Use when:
You want to change title, kind, status, tags, due date, repeat, parent, body, GitHub links, or worklog fields.

### `shelf mv`

Meaning:
Change the `parent` of a task.

Use when:
You want to reorganize the tree without touching other metadata.

### Status shortcuts

Commands:
- `shelf done`
- `shelf start`
- `shelf block`
- `shelf cancel`
- `shelf reopen`

Meaning:
Small convenience wrappers around `shelf set --status ...`.

Use when:
You want fast status transitions with less typing.

### `shelf snooze`

Meaning:
Move `due_on` by relative or absolute date input.

Use when:
The task stays valid, but the date should move.

Notes:
- `--by` shifts from the current due date (or from today if due is empty)
- `--to` sets the due date directly
- In TTY, if neither is provided, the command first offers common presets like `today`, `tomorrow`, `+3 days`, and `next-week`
- Custom `by` / `to` input is still available from the same selector

### `shelf archive` / `shelf unarchive`

Meaning:
Hide or restore tasks without deleting them.

Use when:
The task should leave normal working views temporarily or permanently.

## Relations and Dependencies

### `shelf link`

Meaning:
Create an outbound relationship from one task to another.

Supported link types:
- `depends_on`
- `related`

Important invariant:
`A depends_on B` means `B` must be done before `A`.

### `shelf unlink`

Meaning:
Remove an existing relationship.

### `shelf links`

Meaning:
Show inbound and outbound links for one task.

Use when:
You want relationship context, not only dependency context.

Special mode:
- `--suggest` proposes candidate `related` links with reasons

### `shelf deps`

Meaning:
Show dependency prerequisites and dependents for `depends_on`.

Use when:
You care specifically about dependency order.

Special modes:
- `--graph`: render dependency graph
- `--transitive`: recurse through the dependency chain
- `--suggest`: propose likely prerequisites with reasons

## GitHub, Time, and Notifications

### `shelf github`

Meaning:
Attach or inspect GitHub issue / pull request URLs on tasks.

Subcommands:
- `github link`
- `github unlink`
- `github show`

### `shelf sync github`

Meaning:
Pull title/state from GitHub into task metadata.

What it updates:
- `title`
- `status` (`open` -> `open`, `closed` -> `done`)

### `shelf estimate`

Meaning:
Set or inspect estimated and spent work.

Use when:
You want lightweight planning without a separate time system.

### `shelf track`

Meaning:
Start/stop a simple timer on a task.

Use when:
You want to accumulate `spent_minutes` from actual work sessions.

### `shelf notify`

Meaning:
Run a local shell command for due/overdue active tasks.

Use when:
You want desktop notifications or external automation.

## Views and Output Presets

### `shelf view`

Meaning:
Manage saved task filters.

Use when:
You want named reusable query logic such as `active`, `blocked`, or team-specific slices.

### `shelf preset`

Meaning:
Manage output presets for commands like `ls`, `tree`, `agenda`, or `today`.

Use when:
You want stable output formatting and limits for repeated workflows.

## History, Safety, and Exchange

### `shelf undo` / `shelf redo`

Meaning:
Restore or reapply snapshots created by mutating commands.

### `shelf history`

Meaning:
Inspect recorded write actions and snapshots.

Use when:
You want to know what changed and when.

### `shelf doctor`

Meaning:
Validate `.shelf/` integrity.

Use when:
You suspect inconsistent data or want a pre-commit health check.

What it checks:
- task metadata validity
- parent existence and cycles
- edge validity and duplicates
- unknown kind/status/tag
- invalid GitHub URLs

### `shelf export`

Meaning:
Serialize config, tasks, and edges as JSON.

Use when:
You want backup, migration, or reviewable snapshots.

### `shelf import`

Meaning:
Load data from JSON export.

Modes:
- validate-only
- dry-run
- merge
- replace

## Recommended Daily Flow

```bash
shelf capture "idea or interruption"
shelf triage
shelf review
shelf next
shelf today
```

## Recommended Weekly Flow

```bash
shelf agenda --days 14
shelf calendar
shelf tree
shelf template apply weekly-plan
shelf doctor --strict
```
