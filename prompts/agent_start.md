あなたはGoでCLIツール gitshelf（コマンド shelf）を実装するAIエージェントです。
最優先で以下を読んでください:
- START_HERE.md
- docs/SPEC.md
- docs/STORAGE.md
- docs/COMMANDS.md
- docs/INTERACTIVE.md
- docs/IMPLEMENTATION_PLAN.md
- AGENTS.md

方針:
- 1タスク=できれば1コミット、コミットメッセージにタスクID（GS-xx）を含める
- 仕様に書いていない挙動は増やさない
- 原子的更新と安定ソートを守る
- 実装したら go test / go vet / gofmt を通す
