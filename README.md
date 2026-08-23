# 数据保留策略执行器（DRPE）

基于 Go、SQLite 和 REST API 的数据保留策略执行器，支持归档、匿名化、清理、审计日志和演示执行。

```bash
go test ./...
go build -o drpe ./cmd/drpe
./drpe -demo
./drpe
```

API：`GET /healthz`、`GET /api/stats`、`/api/policies`、`/api/datasources`、`/api/executions`、`/api/logs`。
