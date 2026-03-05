# TRACEABILITY（仕様→実装タスクの対応）

- SPEC / STORAGE / COMMANDS / INTERACTIVE が仕様の単一ソース
- 実装タスクは IMPLEMENTATION_PLAN の GS-xx を参照

| 仕様 | 主担当タスク |
|---|---|
| 保存形式（tasks/edges/config） | GS-02, GS-03 |
| 不変条件（kind/state/link_types, parent循環, edge重複） | GS-03, GS-08, GS-10 |
| add（対話/非対話） | GS-05, GS-06 |
| tree | GS-07, GS-08 |
| link/unlink/links | GS-09 |
| doctor | GS-10 |
