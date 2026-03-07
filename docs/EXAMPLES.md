# EXAMPLES（運用例）

## 週→曜日→具体タスク
```sh
shelf add --title "1週間の目標" --kind todo
# => 01JWEEK...

shelf add --title "月曜日にやること" --kind todo --parent 01JWEEK...
# => 01JMON...

shelf add --title "英単語100個" --kind todo --parent 01JMON...
shelf add --title "量子力学 教科書P50まで" --kind todo --parent 01JMON...
```

### 子同士の関係（例: 先に英単語の単語帳を準備）
```sh
shelf add --title "単語帳を用意する" --kind todo --parent 01JMON...
# => 01JBOOK...

shelf link --from 01JWORD... --to 01JBOOK... --type depends_on
# 英単語100個 depends_on 単語帳を用意する
```

## tree 表示
```sh
shelf tree --from 01JWEEK...
```

## ls のフィルタ
```sh
shelf ls --kind todo --status open
shelf ls --not-status done --not-status cancelled
```

## done にする
```sh
shelf set 01JWORD... --status done
```

## links 確認
```sh
shelf links 01JWORD...
```
