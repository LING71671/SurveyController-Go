# SurveyController-Go

SurveyController-Go 是 SurveyController 的 Go 语言重写版本，目标是做成高速、高效、轻量、高性能的命令行和核心运行工具，用于获得授权的问卷自动化学习与测试。

> 本项目仅供获得授权的学习与测试使用。请勿用于污染第三方问卷数据、绕过平台保护机制，或生成虚假答卷。

## 当前状态

当前主线已经进入 `v0.9` 运行时预览阶段，正在围绕三平台解析、提交判定、runner 状态、worker pool 和执行事务骨架推进。

已完成的基础能力包括：

- CLI 基础命令、配置校验、doctor 检查。
- 强类型问卷模型、provider registry、URL matcher。
- 问卷星、腾讯问卷、Credamo 的 parser 原型和 fixture 回归。
- 答案策略纯函数、运行计划、运行状态、worker pool。
- HTTP client、解析缓存、浏览器抽象和 fake page。
- provider 提交判定契约、engine submission result、runner 提交预览任务。
- 跨平台 CI、`go vet`、`staticcheck`、race test。

`v1.0` 目标是三平台解析、配置生成、基础运行、性能回归、测试和文档闭环，不包含 GUI。后续可以做轻量化 GUI，但必须作为 core/CLI 的薄外壳，不把业务逻辑塞进 UI。

## 设计目标

- **高速**：热路径优先预编译计划、复用资源和少分配。
- **高效**：worker、browser session、HTTP transport、proxy/sample lease 都要有明确生命周期。
- **轻量**：core 优先使用标准库和小接口，避免 GUI 或重量级桌面依赖进入核心。
- **高性能**：并发和内存占用要有 benchmark、race 回归和资源上限。
- **安全边界清晰**：登录、验证、风控、设备次数上限必须停止并报告，不做绕过。

## 快速开始

```powershell
go test ./...
go vet ./...
go run ./cmd/surveyctl version
go run ./cmd/surveyctl config validate examples/run.yaml
go run ./cmd/surveyctl config generate --provider wjx --fixture internal/provider/wjx/testdata/survey.html --url https://www.wjx.cn/vm/example.aspx
go run ./cmd/surveyctl run --dry-run examples/run.yaml
go run ./cmd/surveyctl run --mock examples/mock-run.yaml --seed 7
go run ./cmd/surveyctl run --mock examples/mock-run.yaml --events jsonl
go run ./cmd/surveyctl run --wjx-http-preview examples/wjx-http-preview.yaml --fixture internal/provider/wjx/testdata/survey.html
go run ./cmd/surveyctl run --wjx-http-dry-run examples/wjx-http-preview.yaml --fixture internal/provider/wjx/testdata/survey.html
go run ./cmd/surveyctl run --wjx-http-dry-run examples/wjx-http-preview.yaml --fixture internal/provider/wjx/testdata/survey.html --events jsonl
.\scripts\mock-stress.ps1
.\scripts\wjx-http-dryrun-stress.ps1
go run ./cmd/surveyctl doctor
go run ./cmd/surveyctl doctor browser
```

预期版本输出：

```text
surveyctl v0.1.0
```

版本命令当前仍使用基础 CLI 版本号；功能里程碑通过 Git tag 标记。

## 运行内核方向

后续版本会支持三种可选运行内核：

- `hybrid`：默认模式。优先保证浏览器兼容性；当平台适配器明确支持安全复用请求时，启用 HTTP 快速路径。
- `browser`：纯浏览器模式，优先追求兼容性。
- `http`：纯 HTTP 模式；当所选平台不支持时，必须清晰报错，不自动降级。

## 输出方向

运行事件会同时支持人类可读文本和 JSON Lines。默认文本用于终端查看，JSON Lines 用于后续脚本、CI 和 UI 订阅。

结构化事件会优先携带机器可读字段，例如提交状态、错误码、失败归因、是否停止、是否需要轮换代理。高并发运行时不能依赖解析人类文案做决策。

当前 `surveyctl run` 已经支持两个本地预览入口：

- `--dry-run`：只编译配置和运行计划，不生成提交任务，不访问网络。
- `--mock`：使用本地 mock submitter 执行 runner/worker pool/答案计划链路，不访问网络。

mock run 默认输出汇总信息，包括目标数、并发数、成功数、失败数、完成率、成功率、耗时、吞吐和 worker 数。需要临时压测不同规模时，可以不用修改 YAML，直接覆盖目标数和并发数：

```powershell
go run ./cmd/surveyctl run --mock examples/mock-run.yaml --target 1000 --concurrency 1000 --seed 7
```

也可以使用脚本复现默认 1000 并发 mock 压测：

```powershell
.\scripts\mock-stress.ps1
.\scripts\mock-stress.ps1 -Json
.\scripts\mock-stress.ps1 -MinThroughput 1 -MaxGoroutines 1
.\scripts\mock-stress.ps1 -Target 5 -Concurrency 1 -FailEvery 2
.\scripts\wjx-http-dryrun-stress.ps1
.\scripts\wjx-http-dryrun-stress-matrix.ps1 -SkipFull
.\scripts\wjx-http-dryrun-stress.ps1 -Target 1000 -Concurrency 1000 -Json
.\scripts\verify-local.ps1
.\scripts\verify-local.ps1 -IncludeWJXHTTPDryRunStress
```

mock 压测预算也可以直接走 CLI，例如：

```powershell
go run ./cmd/surveyctl run --mock examples/mock-run.yaml --target 1000 --concurrency 1000 --min-throughput 1 --max-goroutines 1
```

需要观察运行事件时可加：

```powershell
go run ./cmd/surveyctl run --mock examples/mock-run.yaml --events text
go run ./cmd/surveyctl run --mock examples/mock-run.yaml --events jsonl
```

`--events jsonl` 面向后续脚本、CI 和轻量 GUI 外壳；`v1.0` 前不会把 GUI 放进核心，但事件流会保持足够稳定，让 UI 只做薄订阅层。当前 mock run 和 WJX HTTP dry-run 共用这套事件协议。

问卷星 HTTP 路径目前提供本地预览入口，用于检查 answer plan 到 HTTP form 的映射，不执行网络请求。预览会校验配置计划和本地 fixture 的 URL、题目 ID、题型是否一致：

```powershell
go run ./cmd/surveyctl run --wjx-http-preview examples/wjx-http-preview.yaml --fixture internal/provider/wjx/testdata/survey.html
go run ./cmd/surveyctl run --wjx-http-preview examples/wjx-http-preview.yaml --fixture internal/provider/wjx/testdata/survey.html --json
```

需要经过完整 runner/worker pool 但仍然禁用网络时，可以使用问卷星 HTTP dry-run。它会从本地 fixture 构建 survey schema，运行 answer plan 和 HTTP pipeline，并用本地 dry-run executor 记录 draft：

```powershell
go run ./cmd/surveyctl run --wjx-http-dry-run examples/wjx-http-preview.yaml --fixture internal/provider/wjx/testdata/survey.html
go run ./cmd/surveyctl run --wjx-http-dry-run examples/wjx-http-preview.yaml --fixture internal/provider/wjx/testdata/survey.html --target 1000 --concurrency 1000 --json
go run ./cmd/surveyctl run --wjx-http-dry-run examples/wjx-http-preview.yaml --fixture internal/provider/wjx/testdata/survey.html --target 1000 --concurrency 1000 --min-throughput 1 --max-goroutines 1
go run ./cmd/surveyctl run --wjx-http-dry-run examples/wjx-http-preview.yaml --fixture internal/provider/wjx/testdata/survey.html --events text
.\scripts\wjx-http-dryrun-stress.ps1 -Target 1000 -Concurrency 1000
.\scripts\wjx-http-dryrun-stress-matrix.ps1
```

## 开发节奏

当前采用小步推进：

1. 先开 issue 明确范围。
2. 分支实现并补测试。
3. 本地跑 `go test ./...`、`go vet ./...`、`staticcheck`、race test。
4. 开 PR，等待 GitHub Actions 和 CodeRabbit。
5. squash merge 后按里程碑打 tag。

新功能进入 core 前要考虑并发、内存占用、资源上限和可测试性。

## 开发文档

建议先阅读：

- [开发指南](docs/development.md)
- [架构说明](docs/architecture.md)
- [路线图](docs/roadmap.md)
- [性能与压测](docs/performance.md)
- [原项目分析](docs/discovery/original-project-analysis.md)
- [运行闭环对齐复盘](docs/discovery/original-runtime-alignment-2026-04-30.md)
