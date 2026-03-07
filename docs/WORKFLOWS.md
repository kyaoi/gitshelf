# WORKFLOWS

Practical workflows for the current `shelf` toolset.

Use this file when you want to understand how the tool is intended to be used today.

- CLI spec: [`COMMANDS.md`](COMMANDS.md)
- Command-by-command operator guide: [`COMMAND_GUIDE.md`](COMMAND_GUIDE.md)
- Interactive behavior: [`INTERACTIVE.md`](INTERACTIVE.md)

## Main Entry Points

The current tool has one main interactive workspace and several launcher commands around it.

- `shelf`
  - on TTY: opens Cockpit directly
  - on non-TTY: prints help
- `shelf cockpit`
  - opens the main interactive workspace explicitly
- `shelf calendar`
  - opens Cockpit in `calendar` mode
- `shelf tree`
  - opens Cockpit in `tree` mode on TTY
  - use `--plain` for text tree output
- `shelf board`
  - opens Cockpit in `board` mode
- `shelf review`
  - opens Cockpit in `review` mode on TTY
  - use `--plain` or `--json` when needed
- `shelf now`
  - opens Cockpit in `now` mode on TTY
  - `today` remains an alias

Recommended rule:

- start with `shelf` or `shelf cockpit`
- use other commands when you want a narrower entry point or a non-interactive/scriptable path

## Recommended Daily Workflow

### 1. Capture quickly

Use `capture` when you do not want to classify the task yet.

```bash
shelf capture "Call vendor"
shelf capture "Investigate parser regression" --tag backend --due tomorrow
```

Result:

- `kind=inbox`
- `status=open`

### 2. Triage inbox

Use `triage` to process captured items.

```bash
shelf triage
shelf triage --auto done
```

Use this when you want to:

- change kind/status
- archive throwaway items
- turn inbox notes into real tasks

### 3. Work from Cockpit

Open the main workspace:

```bash
shelf
# or
shelf cockpit
```

Then switch modes as needed:

- `C`: calendar
- `T`: tree
- `B`: board
- `R`: review
- `N`: now

### 4. Inspect and edit deeply

When one task needs closer attention:

```bash
shelf show <id>
shelf edit <id>
shelf set <id> --status blocked --append-body "Waiting on API answer."
```

### 5. Clean up structure

When parent-child structure matters:

```bash
shelf tree
shelf mv <id> --parent root
```

Or in Cockpit `tree` mode:

- `h` / `l`: collapse / expand
- `m`: move current task or marked tasks
- `v` / `V`: mark one or mark a range

## Choosing Between `capture` and `add`

Use `capture` when:

- you want speed
- you do not know the final kind/parent yet
- you are collecting inbox material

Use `add` when:

- you know this is a real task now
- you want to set kind/status/parent immediately
- you want a reviewed interactive flow

Examples:

```bash
shelf capture "Check flaky CI failure"
shelf add --title "Refactor parser" --kind todo --status in_progress --parent root
```

## Current `add` Interactive Flow

When you run `shelf add` without `--title` on TTY, the current flow is:

1. Select `Kind`
2. Select `Status`
3. Review and edit:
   - `Title`
   - `Kind`
   - `Status`
   - `Tags`
   - `Due`
   - `Repeat`
   - `Parent`
4. Create or cancel

Important details:

- `Title` is edited in the review screen
- `Title` is still required
- `Ctrl+S` creates immediately from the review screen
- `Ctrl+Enter` also creates on terminals that send a matching sequence
- parent selection uses hierarchical tree labels
- long selection lists scroll automatically

## Cockpit Modes

### `calendar`

Best when:

- you think in dates first
- you want to plan around due dates
- you want to add directly on a focused day

Use:

- `t`: jump to today
- `n` / `p`: switch tasks on the focused day
- `a`: add a task on the focused day
- `o/i/b/d/c`: change status

### `tree`

Best when:

- parent-child decomposition matters
- you want to reorganize a subtree

Use:

- `h` / `l`: collapse / expand
- `m`: move current/marked tasks
- `v`: mark one
- `V`: mark a range
- `u`: clear all marks

### `board`

Best when:

- you want to think in statuses
- you want to batch-change statuses visually

Use:

- `v` / `V`: mark tasks
- `u`: clear marks
- `o/i/b/d/c`: update selected or marked tasks

### `review`

Best when:

- you want a compact decision dashboard

Sections:

- `Inbox`
- `Overdue`
- `Today`
- `Blocked`
- `Ready`

### `now`

Best when:

- you want to focus on today’s execution

Main layout:

- `Focused Day`
- `Overdue`
- `Today`

## When to Stay Outside Cockpit

Use plain commands when:

- you are scripting
- you want JSON
- you want one direct answer, not a workspace

Examples:

```bash
shelf ls --status open --json
shelf review --plain
shelf tree --plain
shelf calendar --json --months 3
```

## Common Task Maintenance Commands

### Status

```bash
shelf start <id>
shelf block <id>
shelf done <id>
shelf cancel <id>
shelf reopen <id>
```

### Metadata and notes

```bash
shelf set <id> --tag backend
shelf set <id> --due next-week
shelf set <id> --append-body "Need benchmark results."
```

### Schedule adjustment

```bash
shelf snooze <id> --by 2d
shelf snooze <id> --to tomorrow
```

### Relationships

```bash
shelf link --from <a> --to <b> --type depends_on
shelf link --from <a> --to <b> --type related
shelf deps <id> --transitive
shelf links <id> --transitive
```

## Notes on Current Terminology

- `status` is the unified term everywhere
- supported default statuses:
  - `open`
  - `in_progress`
  - `blocked`
  - `done`
  - `cancelled`
- supported link types:
  - `depends_on`
  - `related`

`depends_on` direction:

- `A depends_on B` means `B` must happen before `A`

## Suggested Starting Point

If you are not sure where to begin, use:

```bash
shelf
```

Then:

1. `R` for review
2. `N` for now
3. `T` for structure
4. `B` for status cleanup
5. `C` for date planning
