# EXAMPLES（日本語版）

## メイン workspace を開く

```sh
shelf
shelf review
shelf tree
```

## 読み取り専用の一覧取得

```sh
shelf ls --kind todo --status open
shelf ls --tag backend --json
shelf next
```

## link を直接操作する

```sh
shelf link --from 01AAA --to 01BBB --type depends_on
shelf unlink --from 01AAA --to 01BBB --type depends_on
shelf links 01AAA
```
