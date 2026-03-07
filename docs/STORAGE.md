# STORAGE

Current storage layout for the Cockpit-first toolset.

## Directory Layout

```text
.shelf/
  config.toml
  tasks/
    <id>.md
  edges/
    <src_id>.toml
```

Legacy directories such as `.shelf/templates/` and `.shelf/history/` are no longer part of the active layout and are removed by `shelf init`.

## Config

Current config stores:

- `kinds`
- `statuses`
- `tags`
- `link_types`
- `default_kind`
- `default_status`
- `[commands.calendar]`

Current calendar config:

```toml
[commands.calendar]
default_range_unit = "days"
default_days = 7
default_months = 6
default_years = 2
```

Legacy sections like `[views.*]` and `[output_presets.*]` may still be readable for migration, but they are no longer written back.

## Task File

Each task is stored as `.shelf/tasks/<id>.md`.

Current front matter fields:

- `id`
- `title`
- `kind`
- `status`
- `tags`
- `due_on`
- `repeat_every`
- `archived_at`
- `parent`
- `created_at`
- `updated_at`

The task body is freeform notes.

Current note:

- legacy fields such as `github_urls`, `estimate_minutes`, `spent_minutes`, and `timer_started_at` may still be readable, but current writes omit them

## Edge File

Each task can have one outbound edge file:

- `.shelf/edges/<src_id>.toml`

Format:

```toml
[[edge]]
to = "01..."
type = "depends_on"
```

Supported link types:

- `depends_on`
- `related`

## Invariants

- one task per markdown file
- task files are keyed by full task ID
- edge files store outbound links only
- parent-child hierarchy is represented by `parent`
- link graph is represented by edge files
- current writes normalize to the active schema and omit legacy sections/fields
