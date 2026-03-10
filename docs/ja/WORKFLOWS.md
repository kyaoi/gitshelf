# WORKFLOWS（日本語版）

現在の Cockpit-first な使い方です。

## 基本ルール

- まずは `shelf`
- 普段の作業は Cockpit に留まる
- text / JSON の単発確認だけ `ls` / `next` を使う
- relation を script から直接触るときだけ `link` / `unlink` / `links` を使う

詳細な keybind は [`INTERACTIVE.md`](INTERACTIVE.md) のみにまとめます。

## 主な入口

- `shelf`
  - TTY なら Cockpit を直接開く
- `shelf cockpit`
  - 同じ workspace を明示的に開く
- `shelf calendar`
  - calendar mode で開く
- `shelf tree`
  - tree mode で開く
- `shelf board`
  - board mode で開く
- `shelf review`
  - review mode で開く
- `shelf now`
  - now mode で開く

## おすすめの日次フロー

### 1. workspace を開く

```bash
shelf
```

### 2. Cockpit 内で確認と整理を進める

- 日付基準の計画は `calendar`
- 親子構造や移動は `tree`
- status 基準の整理は `board`
- inbox / overdue / blocked / ready の見直しは `review`
- 今日の実行に寄せるなら `now`

non-calendar mode では sidebar の `Calendar` と `Selected Day` が main selection に同期します。

### 3. 編集は TUI 内で行う

- 現在の文脈から task を追加する
- status / kind / tag / due / link をその場で更新する
- add / link / tag / filter などは中央 popup で扱う

### 4. 直接 command を使うのは script または単発確認だけ

```bash
shelf ls --status open --json
shelf next --json
shelf link --from 01AAA --to 01BBB --type depends_on
shelf links 01AAA
```

## mode の使い分け

- `calendar`: 日付基準で計画する
- `tree`: 親子構造や移動を扱う
- `board`: status 基準で整理する
- `review`: inbox / overdue / blocked / ready を俯瞰する
- `now`: 今日の実行に集中する
