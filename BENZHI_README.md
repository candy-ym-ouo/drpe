# drpe 打包说明

数据保留策略执行器（DRPE）按策略对历史数据执行归档、匿名化或清理，并记录审计日志。

## 本地验证

```bash
go mod download
go test ./...
go vet ./...
go build -o drpe ./cmd/drpe
./drpe -demo
```

启动服务：

```bash
./drpe
```

健康检查：`curl http://127.0.0.1:8080/healthz`

## Docker 打包

```bash
./build_benzhi_docker.sh drpe linux/amd64
./build_benzhi_docker.sh drpe linux/arm64
```

镜像保留完整 Go 工具链，便于在容器内继续测试和修改。
