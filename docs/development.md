# 开发指南

## 默认约定

- GitHub 相关工作优先使用 GitHub CLI（`gh`）。
- 引导提交之后，开发流程为：议题 -> 分支 -> 拉取请求。
- 保持 `main` 随时可发布。
- 优先提交小而清晰、便于审查的改动。

## 本地命令

```powershell
go run ./cmd/surveyctl version
go test ./...
go test -race ./...
go vet ./...
gofmt -w (git ls-files '*.go')
```

在 Windows 上，`go test -race` 需要 CGO 和 `gcc` 之类的 C 编译器。如果本地竞态检查失败并提示 `C compiler "gcc" not found`，可以先运行普通测试，并依赖 Ubuntu CI 中的竞态检查，直到本地安装好 C 工具链。

常用本地验证可以直接跑：

```powershell
.\scripts\verify-local.ps1
```

默认会执行 `go test ./...`、`go vet ./...`、`staticcheck` 和轻量 mock stress matrix。需要完整 1000 并发 profile 时：

```powershell
.\scripts\verify-local.ps1 -IncludeFullStress
```

如果只是快速检查 Go 代码、不跑压测：

```powershell
.\scripts\verify-local.ps1 -SkipStress
```

## 运行预览

`surveyctl run` 当前只开放不会访问网络的预览能力：

```powershell
go run ./cmd/surveyctl run --dry-run examples/run.yaml
go run ./cmd/surveyctl run --mock examples/mock-run.yaml --seed 7
go run ./cmd/surveyctl run --mock examples/mock-run.yaml --target 1000 --concurrency 1000 --seed 7
.\scripts\mock-stress.ps1
go run ./cmd/surveyctl run --mock examples/mock-run.yaml --events text
go run ./cmd/surveyctl run --mock examples/mock-run.yaml --events jsonl
```

`--dry-run` 用于验证配置能否编译成运行计划；`--mock` 会实际经过答案计划生成、worker pool、运行状态和事件输出，但 submitter 是本地 mock，不访问任何平台。

`--target` 和 `--concurrency` 可以覆盖配置中的运行规模，用于本地压测和验证资源上限。覆盖后的计划仍会重新走 runner 校验，因此 `browser` 模式不会被临时参数放大到超过小池限制。

常用压测入口见 [性能与压测](performance.md)。默认脚本会运行 1000 target / 1000 concurrency 的本地 mock；`-Json` 输出最终 JSON 汇总，`-FailEvery` 可验证失败阈值和停止行为。

事件流和汇总输出的边界要保持清晰：

- `--json` 只输出最终汇总 JSON。
- `--events text` 输出人类可读事件，再输出最终汇总。
- `--events jsonl` 输出 JSON Lines 事件，再输出最终汇总。
- `--events` 不和 `--json` 同时使用，避免把事件流和单个 JSON 汇总混成不稳定协议。

新增真实运行能力时，优先复用 runner 层的 `RunPlanReport` 和 logging 事件类型。CLI、CI、脚本和后续轻量 GUI 都应订阅同一套 core 事件，不要为 UI 单独分叉业务状态。

## 性能习惯

热路径优先做编译型结构，例如 `WeightedPicker`、`SelectionPicker`、`AnswerPlanBuilder`。配置解析、规则校验、权重归一化这类稳定工作应尽量在编译阶段完成，运行阶段只做必要的随机抽样和提交调度。

性能相关 PR 至少说明以下内容：

- 是否减少每次提交或每题的分配。
- 是否改变并发上限或 worker 生命周期。
- 是否增加 benchmark 或保留可复测的 benchmark 命令。
- 是否影响 browser 小池兜底路径。

本地 benchmark 不作为绝对性能承诺，但要写清机器上的相对变化，例如 `PickMany` 与编译后 picker 的 ns/op、B/op、allocs/op 对比。

## 先有议题

修改行为前先创建或认领议题：

```powershell
gh issue create
gh issue view 123
```

议题应写清用户目标、范围、验收标准和安全约束。

## 分支

使用简短、关联议题的分支名：

```text
codex/issue-123-config-schema
codex/issue-124-provider-contract
```

## 提交

提交信息保持简洁：

```text
chore: bootstrap go project
feat: add runtime engine mode parser
test: cover provider capability checks
```

## 拉取请求

使用 GitHub CLI 打开草稿拉取请求：

```powershell
gh pr create --draft --fill
```

每个拉取请求必须包含：

- 关联议题。
- 改了什么。
- 为什么改。
- 运行过哪些测试。
- 涉及时说明性能影响。
- 风险与回滚说明。

## 风格规则

- 参考 Google Go 风格指南和 Go 代码审查建议。
- 函数保持单一职责。
- 接口尽量靠近消费侧。
- 包名要清晰。
- 可取消的工作使用 `context.Context`。
- 遇到平台验证、登录或反滥用页面时停止并报告。

## 引导提交例外

由于远程仓库最初为空，`v0.1` 允许一次直接提交到 `main` 的引导提交。这个例外不适用于后续正常功能开发。
