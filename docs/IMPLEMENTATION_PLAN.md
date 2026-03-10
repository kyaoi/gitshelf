# IMPLEMENTATION_PLAN

Historical release plan for the current Cockpit-first tool.

## Phase 1: bootstrap

- module and entrypoint
- `init`
- stable `.shelf/` layout

## Phase 2: storage and invariants

- task / edge / config parsing
- ULID generation
- stable formatting and atomic writes

## Phase 3: Cockpit workspace

- calendar / tree / board / review / now
- task editing inside TUI
- sidebar synchronization and popup editors

## Phase 4: read-only and scripting commands

- `ls`
- `next`
- `link` / `unlink` / `links`

## Phase 5: docs and verification

- align docs with the implemented CLI and TUI
- keep formatting, tests, race tests, and vet green
