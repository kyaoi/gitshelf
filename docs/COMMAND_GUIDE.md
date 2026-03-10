# COMMAND GUIDE

A concise guide to which command to use.

## Use `shelf` most of the time

If you are on a TTY and just want to work, use:

```bash
shelf
```

That opens `Cockpit`, the main workspace.

## Use `shelf cockpit` when you want to be explicit

```bash
shelf cockpit
shelf cockpit --mode tree
shelf cockpit --mode board --months 3
```

## Use launcher commands when you know the starting view

- `shelf calendar`
- `shelf tree`
- `shelf board`
- `shelf review`
- `shelf now`

These are only starting presets for the same TUI.

## Inside Cockpit

Most daily editing happens inside `Cockpit`.

- `a`: add as a child of the selected task
- `A`: add at root
- `K`: edit kind
- `#`: edit tags
- `o/i/b/d/c`: change status
- `L` / `U`: add or remove links
- `y`: copy the selected title

See [`INTERACTIVE.md`](INTERACTIVE.md) for the full key map.

## Use `shelf ls` when you want text or JSON

Examples:

```bash
shelf ls --status open
shelf ls --kind todo --json
shelf ls --ready --json
```

## Use `shelf next` when you want the short answer

Examples:

```bash
shelf next
shelf next --json
```

## Use link commands when you need direct scripting

```bash
shelf link --from 01AAA --to 01BBB --type depends_on
shelf unlink --from 01AAA --to 01BBB --type depends_on
shelf links 01AAA
```

## Use `shelf init` only for setup or cleanup

Examples:

```bash
shelf init
shelf init --force
shelf init --global
```

## Use `shelf completion` only for shell setup

Examples:

```bash
shelf completion zsh
shelf completion bash
```
