# TESTING（日本語版）

`gitshelf` では heavy suite を標準の test suite として扱います。

## Local hooks

ローカル hook の install:

```bash
lefthook install
```

hook の役割:

- `pre-commit`: 軽い local guardrail だけを実行する
- `pre-push`: merge 前に必要な heavy suite を実行する
- `pre-commit` では public docs が削除済み CLI schema 用語へ戻っていないことも確認する

## Required checks

merge 前には次を実行します。

- `gofmt -w .`
- `go test ./...`
- `bash scripts/check_coverage_ratchet.sh`
- `go test -race ./...`
- `go vet ./...`
- `bash scripts/check_public_contract_docs.sh`

`pre-push` と同じ heavy suite をまとめて回すなら:

```bash
bash scripts/run_heavy_checks.sh
```

## Regression policy

- bug fix は `repro test first` を原則にする
- machine-readable CLI output は public contract として扱う
- golden fixture の更新は意図的な変更時だけ行う
- package coverage を [`../../scripts/coverage_baseline.txt`](../../scripts/coverage_baseline.txt) 未満に下げない

## Golden fixtures

CLI output の snapshot は次にあります。

- `internal/cli/testdata/outputs/`

public output contract を意図的に変えたときだけ更新します。

```bash
UPDATE_GOLDEN=1 go test ./internal/cli -run TestCLIMachineReadableOutputGoldens
```

## Important checks

- `init` の冪等性
- `link` / `unlink` の edge 更新と重複抑止
- `links` の outbound / inbound 解決
- Cockpit の mode 切り替え時に selection が可能な限り維持されること
- sidebar の `Calendar` と `Selected Day` が main pane と同期すること
- 継承 due が calendar / tree / review / now で一貫表示されること

## Stable diffs

- `edges` の順序が deterministic であること
- `tree` の順序が deterministic であること

## Conflict resistance

- task 本文編集と link 操作が別ファイルに分離されること

## Release smoke

`v1.3` のような tag release では:

1. `go install github.com/kyaoi/gitshelf/cmd/shelf@v1.3`
2. `shelf --version` が `v1.3` を表示する
3. `mise use -g go:github.com/kyaoi/gitshelf/cmd/shelf@latest`
4. `shelf --version` が `dev` ではなく最新 tag を表示する

GitHub Actions でも [`../../.github/workflows/release-smoke.yml`](../../.github/workflows/release-smoke.yml)
を使って、tag push 時と manual 実行時の release smoke を回します。
