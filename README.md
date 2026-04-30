# SurveyController-go

SurveyController-go 是 SurveyController 的 Go 语言重写版本，目标是做成轻量、高性能的命令行工具，用于获得授权的问卷自动化学习与测试。

> 本项目仅供获得授权的学习与测试使用。请勿用于污染第三方问卷数据、绕过平台保护机制，或生成虚假答卷。

## 当前状态

`v0.1` 是项目初始化版本：

- Go 模块和最小 `surveyctl` 命令行工具。
- 原 Python 项目的探索结论已持久化到文档。
- 架构与路线图文档。
- GitHub 议题和拉取请求模板。
- 覆盖格式检查、测试、`go vet`、竞态检查和跨平台构建的 CI。

三平台支持是 `v1.0` 目标。`v0.1` 只预留架构边界，不实现真实问卷运行能力。

## 快速开始

```powershell
go test ./...
go vet ./...
go run ./cmd/surveyctl version
go run ./cmd/surveyctl config validate examples/run.yaml
go run ./cmd/surveyctl doctor
```

预期版本输出：

```text
surveyctl v0.1.0
```

## 运行内核方向

后续版本会支持三种可选运行内核：

- `hybrid`：默认模式。优先保证浏览器兼容性；当平台适配器明确支持安全复用请求时，启用 HTTP 快速路径。
- `browser`：纯浏览器模式，优先追求兼容性。
- `http`：纯 HTTP 模式；当所选平台不支持时，必须清晰报错，不自动降级。

## 开发文档

建议先阅读：

- [开发指南](docs/development.md)
- [架构说明](docs/architecture.md)
- [路线图](docs/roadmap.md)
- [原项目分析](docs/discovery/original-project-analysis.md)
