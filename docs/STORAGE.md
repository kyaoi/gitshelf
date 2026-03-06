# STORAGE (Formats and Invariants)

## Directory Layout

```text
.shelf/
  config.toml
  tasks/
    <id>.md
  edges/
    <src_id>.toml
  history/
    index.json
    actions.log
    snapshots/
```

## Config File (`.shelf/config.toml`)

Core keys:

- `kinds`
- `statuses`
- `link_types`
- `default_kind`
- `default_status`

Optional saved views:

```toml
[views."active"]
not_statuses = ["done", "cancelled"]

[views."only_done"]
statuses = ["done"]
```

Supported view keys:

- `kinds`, `statuses`, `not_kinds`, `not_statuses`
- `ready`, `blocked_by_deps`
- `due_before`, `due_after`, `overdue`, `no_due`
- `parent`, `search`, `limit`

Optional output presets:

```toml
[output_presets."ls_focus"]
command = "ls"
view = "active"
format = "detail"
limit = 20
```

## IDs

- ULID is used for new tasks.
- Short display ID uses the first 8 characters.

## Task File (`.shelf/tasks/<id>.md`)

- Markdown body with TOML front matter (`+++ ... +++`).
- Front matter is structured metadata.
- Required front matter keys:
  - `id`
  - `title`
  - `kind`
  - `status`
  - `created_at`
  - `updated_at`
- Optional:
  - `due_on` (`YYYY-MM-DD`)
  - `repeat_every` (`<N>d|<N>w|<N>m|<N>y`)
  - `archived_at` (RFC3339)
  - `parent`
  - body text (freeform notes)

CLI accepts `today` / `tomorrow` for due input, but stores normalized `YYYY-MM-DD`.

Body is intentionally freeform. Typical usage:

- detailed description
- supplementary context
- execution memo
- progress log
- idea draft
- references

Key order is stable:

1. `id`
2. `title`
3. `kind`
4. `status`
5. `due_on` (if present)
6. `repeat_every` (if present)
7. `archived_at` (if present)
8. `parent` (if present)
9. `created_at`
10. `updated_at`

Timestamps use RFC3339.

## Edge File (`.shelf/edges/<src_id>.toml`)

`[[edge]]` array with:

- `to`
- `type`

Edge output is stable sorted by:

1. `to` ascending
2. `type` ascending

Duplicate `(to, type)` is removed on write.

## Invariants

### Tasks

- filename ID and front matter `id` must match
- `title` must be non-empty
- `kind` must exist in config `kinds`
- `status` must exist in config `statuses`
- `due_on` must match `YYYY-MM-DD` when present
- `repeat_every` must match `<N>d|<N>w|<N>m|<N>y` when present
- `archived_at` must be RFC3339 when present
- when `parent` exists:
  - parent task must exist
  - cannot point to self
  - must not create parent cycle

### Edges

- source task should exist
- `type` must exist in config `link_types` (`depends_on`, `related`)
- destination task must exist
- duplicate `(type, to)` is invalid

### Link Direction

`A depends_on B` always means:

`A --depends_on--> B`

## Atomic Writes

All writes use temp file -> rename in same filesystem to avoid partial corruption.

## Undo Snapshot

Mutating commands push snapshots under `.shelf/history/snapshots`.
`index.json` tracks undo/redo stacks.
`actions.log` stores JSONL action records (`apply`/`undo`/`redo`).

## Error Messages

Error output should include file path and cause where possible, for example:

`.shelf/tasks/<id>.md: unknown kind: ...`
