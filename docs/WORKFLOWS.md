# WORKFLOWS

Practical usage for the current Cockpit-first toolset.

## Default Rule

Use this as the default mental model:

- start with `shelf`
- stay inside `Cockpit`
- use `ls` or `next` only when you need a direct textual answer or JSON

## Main Entry Points

- `shelf`
  - on TTY, opens Cockpit directly
- `shelf cockpit`
  - same workspace, explicit form
- `shelf calendar`
  - open Cockpit in calendar mode
- `shelf tree`
  - open Cockpit in tree mode
- `shelf board`
  - open Cockpit in board mode
- `shelf review`
  - open Cockpit in review mode
- `shelf now`
  - open Cockpit in now mode

## Recommended Daily Flow

### 1. Open the workspace

```bash
shelf
```

### 2. Move between modes

Inside Cockpit:

- `C`: calendar
- `T`: tree
- `B`: board
- `R`: review
- `N`: now
- `Ctrl+H` / `Ctrl+L`: previous / next mode

### 3. Use calendar as the default planning view

Useful keys:

- `t`: jump to today
- `h/l`: previous/next day
- `j/k`: previous/next week
- `[` / `]`: previous/next month
- `n/p`: cycle tasks on the focused day
- `a`: create from the current context
- `A`: quick capture

### 4. Change task state in place

Useful keys:

- `o`: open
- `i`: in progress
- `b`: blocked
- `d`: done
- `c`: cancelled
- `x`: archive toggle
- `z`: snooze

### 5. Reorganize structure in tree mode

Useful keys:

- `h/l`: collapse / expand
- `m`: move selected task or marked tasks
- `v`: toggle mark for the current task
- `V`: start or stop range marking
- `u`: clear all marks

### 6. Use ls / next only for direct answers

Examples:

```bash
shelf ls --status open --json
shelf ls --kind todo --not-status done --not-status cancelled
shelf next
shelf next --json
```

## When to Use Which Mode

- `calendar`: date-first planning
- `tree`: parent-child structure and moves
- `board`: status-driven work management
- `review`: inbox / overdue / blocked / ready scan
- `now`: focused execution for today

## What Is No Longer a Top-Level Command

The current tool intentionally does not rely on separate top-level commands for:

- add/edit/show/set
- move/snooze/archive
- history/undo/redo
- import/export
- GitHub sync
- saved views or presets

Those workflows are expected to stay inside Cockpit.
