# STORAGE

Current storage layout for the Cockpit-first toolset.

## Directory Layout

Default layout:

```text
.shelf/
  config.toml
  tasks/
    <id>.md
  edges/
    <src_id>.toml
```

Config always lives at `.shelf/config.toml`.
Task data lives under `storage_root`, which defaults to `.shelf`.

Example root-level layout:

```toml
storage_root = "."
```

```text
.shelf/
  config.toml
tasks/
  <id>.md
edges/
  <src_id>.toml
```

## Config

Current config stores:

- `kinds`
- `statuses`
- `tags`
- `storage_root`
- `link_types`
- `default_kind`
- `default_status`
- `[commands.calendar]`
- `[commands.cockpit]`

Link type config:

```toml
[link_types]
names = ["depends_on", "related"]
blocking = "depends_on"
```

- `names` lists all allowed link type names
- `blocking` names the relation used for readiness and cycle checks
- the default blocking relation also rejects links from a child task to one of its ancestors

Current calendar config:

```toml
[commands.calendar]
default_range_unit = "days"
default_days = 7
default_months = 6
default_years = 2
```

Current cockpit config:

```toml
[commands.cockpit]
copy_separator = "\n"
post_exit_git_action = "none"
commit_message = "chore: update shelf data"

[[commands.cockpit.copy_presets]]
name = "subtree_path"
scope = "subtree"
template = "{{path}}\n{{subtree}}"
join_with = "\n\n"
```

- `copy_separator`
- `copy_presets[].name`
- `copy_presets[].scope`: `task` or `subtree`
- `copy_presets[].template`: supports `{{title}}`, `{{path}}`, `{{body}}`, `{{subtree}}`
- `copy_presets[].join_with`: optional; falls back to `copy_separator`

Storage location config:

```toml
storage_root = ".shelf"
```

- `storage_root` is the common parent directory for `tasks/` and `edges/`
- relative paths are resolved from the project root
- `.` places `tasks/` and `edges/` directly under the project root

## Task File

Each task is stored as `<storage_root>/tasks/<id>.md`.

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

## Edge File

Each task can have one outbound edge file:

- `<storage_root>/edges/<src_id>.toml`

Format:

```toml
[[edge]]
to = "01..."
type = "depends_on"
```

Supported link types come from `config.toml` `link_types.names`.
Default names are `depends_on` and `related`.

## Invariants

- one task per markdown file
- task files are keyed by full task ID
- edge files store outbound links only
- parent-child hierarchy is represented by `parent`
- link graph is represented by edge files
