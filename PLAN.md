# Product Roadmap

## Vision

`gitshelf` is a Git-friendly task manager centered around one TUI workspace: `Cockpit`.

The core product shape is:

- open `shelf`
- stay inside `Cockpit`
- manage hierarchy, links, dates, and workflow state without leaving the terminal
- keep storage text-based, reviewable, and stable under Git

This project should continue to optimize for:

- Cockpit-first daily operation
- predictable, inspectable file formats
- scriptable read-side commands
- explicit user control over writes and Git integration

This project should not drift toward:

- a daemon or background service
- a database-backed architecture
- a web UI / GUI-first product
- multi-parent DAG trees
- hidden or speculative automation

## Design Principles

### 1. One Workspace, Many Views

`calendar`, `tree`, `board`, `review`, and `now` are not separate products.
They are different operational views into the same workspace.

Improvements should prefer:

- deeper mode consistency
- better cross-mode handoff
- shared action patterns

over:

- mode-specific one-off behaviors
- duplicated command surfaces

### 2. Git-Native By Default

Storage should stay human-readable and diff-friendly.
Writes should remain explicit, reviewable, and safe to undo.

Improvements should prefer:

- stable formatting
- atomic writes
- useful diff boundaries
- opt-in Git automation

### 3. Cockpit Over Command Sprawl

Most editing should happen in the TUI.
Standalone CLI commands should exist primarily for:

- read-only inspection
- scripting
- config and preset management
- link operations where direct shell usage matters

### 4. Power Without Ambiguity

The tool can be dense and fast, but behavior must stay explainable.
Keybindings, popups, copy/export behavior, and config should feel consistent and composable.

## Roadmap

### Phase 1: Cockpit Flow Polish

Priority: highest

This phase deepens the main workspace instead of widening the product.

Goals:

- reduce context switches out of `Cockpit`
- make high-frequency editing faster than opening an external editor
- make bulk operations feel first-class

Planned improvements:

- expand inline editing for common metadata and note-taking
  - due date
  - repeat
  - parent / placement
  - body editing for short-to-medium notes
- strengthen multi-select / marked-task workflows
  - consistent bulk status / kind / tag / snooze / copy actions
  - clearer feedback for actions affecting multiple tasks
- unify popup ergonomics
  - one mental model for focus, confirm, cancel, save, and preview
  - reduce special-case controls between add / tag / link / copy / filter flows
- finish the copy/export surface as a share-ready workflow
  - saved presets
  - stable preview
  - better command discoverability
  - consistent subtree rendering options
- improve discoverability inside Cockpit
  - clearer help overlays
  - more explicit action grouping
  - leave room for a future action picker / command palette

### Phase 2: CLI Symmetry And Scriptability

Priority: medium

This phase makes the CLI surface a better companion to the Cockpit-first model.

Goals:

- expose the right information to scripts without turning the tool into command sprawl
- make config and preset management feel intentional, not incidental

Planned improvements:

- align public docs with the actual command surface
  - include `config` and preset flows everywhere they are public
- expand read-side symmetry with Cockpit views
  - improve query outputs so scripting can reuse what users see in Cockpit
  - prefer text + JSON surfaces that map cleanly to existing concepts
- strengthen config management
  - make preset/config operations consistent
  - ensure command-based config management is documented and stable
- improve export / report-oriented CLI outputs
  - better reusable task labels
  - safer machine-readable outputs
  - clearer view-specific summaries

### Phase 3: Git-Native Safety And Review

Priority: medium

This phase improves confidence in write-heavy workflows.

Goals:

- make Git integration useful without making it scary
- improve recovery, review, and sharing workflows

Planned improvements:

- improve undo / snapshot / recovery discoverability
- make post-exit Git actions easier to understand and safer to trust
- add better review-oriented output around changed task data
- keep commit boundaries aligned with task-manager semantics rather than internal implementation details

## Near-Term Execution Order

1. Cockpit inline edit and bulk action polish
2. popup consistency and discoverability cleanup
3. public CLI/config surface cleanup and documentation parity
4. Git safety, undo, and review workflow refinement

## Current Baseline To Preserve

The roadmap assumes these are already part of the product baseline:

- configurable `storage_root`
- cursor-aware text input
- advanced copy presets
- `config copy-preset set`
- q-only Cockpit quit

Future planning should not reopen these as missing features unless a regression is found.
