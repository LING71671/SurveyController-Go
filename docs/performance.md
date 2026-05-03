# 性能与压测

本项目的性能目标是高速、高效、轻量、高性能。`v1.0` 前的压测以本地 mock run 为主：它会经过配置编译、答案计划生成、worker pool、运行状态和报告输出，但不会访问任何问卷平台。

## 本地 mock 压测

推荐从 1000 并发开始验证：

```powershell
.\scripts\mock-stress.ps1
```

等价于：

```powershell
go run ./cmd/surveyctl run --mock examples/mock-run.yaml --target 1000 --concurrency 1000 --seed 7
```

输出会包含：

- `successes` / `failures` / `completed`
- `duration_ms`
- `throughput_per_second`
- `goroutines`
- `heap_alloc_bytes`
- `heap_alloc_delta_bytes`
- `total_alloc_delta_bytes`
- `failure_threshold_reached`

这些数字用于观察趋势，不作为跨机器绝对承诺。比较不同实现时，优先在同一机器、同一配置、同一 seed 下重复运行。

## 压测矩阵

如果希望一次验证成功路径、失败阈值和 1000 并发路径，可以运行：

```powershell
.\scripts\mock-stress-matrix.ps1
```

快速 smoke 可以跳过完整 1000 并发 profile：

```powershell
.\scripts\mock-stress-matrix.ps1 -SkipFull
```

矩阵脚本会输出每个 profile 的 target、concurrency、成功/失败数、吞吐、heap delta、goroutine 和失败阈值状态。它内部仍然只调用本地 mock run，不访问网络。

## 本地验证入口

提交 PR 前推荐跑：

```powershell
.\scripts\verify-local.ps1
```

默认会跑 Go 测试、`go vet`、`staticcheck` 和轻量 mock stress matrix。要包含完整 1000 并发 profile：

```powershell
.\scripts\verify-local.ps1 -IncludeFullStress
```

要同时包含 WJX HTTP dry-run matrix：

```powershell
.\scripts\verify-local.ps1 -IncludeWJXHTTPDryRunStress
.\scripts\verify-local.ps1 -IncludeWJXHTTPDryRunStress -IncludeFullStress
```

CI 使用 `-SkipGoChecks` 跑轻量 smoke，避免重复执行已经单独跑过的 Go 检查，同时覆盖本地验证脚本的编排逻辑。

完整脚本说明见 [脚本参考](scripts.md)。

## JSON 汇总

脚本可输出单个 JSON 汇总，方便后续 CI 或轻量 GUI 读取：

```powershell
.\scripts\mock-stress.ps1 -Json
```

JSON 汇总不要和 `--events jsonl` 混用。前者是最终报告，后者是事件流。

## WJX HTTP 预览

问卷星 HTTP 路径目前先提供本地预览，不执行网络请求：

```powershell
go run ./cmd/surveyctl run --wjx-http-preview examples/wjx-http-preview.yaml --fixture internal/provider/wjx/testdata/survey.html
go run ./cmd/surveyctl run --wjx-http-preview examples/wjx-http-preview.yaml --fixture internal/provider/wjx/testdata/survey.html --json
```

预览会经过配置编译、答案计划生成、问卷星 HTTP form 映射和 plan/fixture 兼容性校验，但不会调用 HTTP executor。输出中的 `network: disabled (preview)` 是这个阶段的安全边界。

当前示例覆盖单选、多选、文本、评分和矩阵题，用于观察 answer plan 到 `processjq.ashx` form 的映射。矩阵题在本地 draft 中使用稳定的行级表示，例如 `q5_r1:5;q5_r2:1`，便于脚本检查每一行是否都被计划和映射。后续真实网络运行必须继续复用 provider 能力门控，并在登录、验证、风控、设备次数上限或频控信号出现时停止并报告。

## WJX HTTP Dry-Run

需要验证完整 runner 和 HTTP pipeline 时，可以使用本地 dry-run executor。它会记录生成的 HTTP draft，但不会执行网络请求：

```powershell
go run ./cmd/surveyctl run --wjx-http-dry-run examples/wjx-http-preview.yaml --fixture internal/provider/wjx/testdata/survey.html
go run ./cmd/surveyctl run --wjx-http-dry-run examples/wjx-http-preview.yaml --fixture internal/provider/wjx/testdata/survey.html --target 1000 --concurrency 1000 --json
go run ./cmd/surveyctl run --wjx-http-dry-run examples/wjx-http-preview.yaml --fixture internal/provider/wjx/testdata/survey.html --target 1000 --concurrency 1000 --min-throughput 1 --max-goroutines 1
go run ./cmd/surveyctl run --wjx-http-dry-run examples/wjx-http-preview.yaml --fixture internal/provider/wjx/testdata/survey.html --target 1000 --concurrency 1000 --events jsonl
.\scripts\wjx-http-dryrun-stress.ps1
.\scripts\wjx-http-dryrun-stress.ps1 -Target 1000 -Concurrency 1000 -MinThroughput 1 -MaxGoroutines 1
.\scripts\wjx-http-dryrun-stress.ps1 -Target 1000 -Concurrency 1000 -Json
.\scripts\wjx-http-dryrun-stress-matrix.ps1 -SkipFull
.\scripts\wjx-http-dryrun-stress-matrix.ps1
```

text 输出只展示汇总和首个 draft，避免高并发 dry-run 时刷屏；JSON 输出包含完整 `drafts`，适合脚本检查 answer plan 到 form 的稳定性，包括矩阵题的行级答案。输出中的 `network: disabled (dry-run)` 是安全边界。

`--events text|jsonl` 可用于观察 dry-run 期间的 runner 事件。高并发 profile 中建议优先使用 `--json` 汇总或脚本预算；事件流更适合小规模诊断和后续轻量 UI 订阅。

WJX HTTP dry-run 支持和 mock run 相同的预算参数。预算失败时 CLI 会先输出 dry-run 报告，再以非零退出码返回失败原因，便于 CI 和脚本保留诊断信息。

脚本默认输出压缩后的 summary，包括成功数、失败数、吞吐、资源指标、draft 数量和首个 draft 的端点/答案数量；`-Json` 输出同样是压缩 summary，不会默认打印完整 drafts。

矩阵脚本会复用单 profile 脚本，默认包含 smoke、预算和 1000x1000 profile；`-SkipFull` 只跑轻量 profile，适合本地快速回归。

CI 会执行 `.\scripts\wjx-http-dryrun-stress-matrix.ps1 -SkipFull` 作为脚本 smoke。完整 1000x1000 profile 不放进默认 CI，避免把每个 PR 的必跑路径变重；发布前或性能专项验证时再显式运行完整 profile。

## 预算断言

脚本支持轻量预算断言，适合本地提交前或后续 CI 使用。预算参数会透传给 `surveyctl run --mock`，由 CLI 基于同一份 `RunPlanReport` 统一判定；脚本只负责输出 JSON 或人类可读摘要：

```powershell
.\scripts\mock-stress.ps1 -Target 1000 -Concurrency 1000 -MinThroughput 1000 -MaxGoroutines 1
```

可用预算包括：

- `-MinThroughput`：要求 `throughput_per_second` 不低于指定值。
- `-MaxHeapDelta`：要求 `heap_alloc_delta_bytes` 不高于指定值。
- `-MaxGoroutines`：要求运行结束后的 goroutine 数不高于指定值。
- `-ExpectFailureThreshold`：要求 `failure_threshold_reached` 等于指定布尔值。

预算失败时 CLI 会先输出 mock run 报告，再以非零退出码返回失败原因。这样本地脚本、矩阵脚本和后续 CI 可以共享同一套判定语义。预算应先设得保守，主要用于抓明显退化；严苛性能门槛必须先在 CI 机器上积累基线。

## 失败注入

可以用失败注入验证失败阈值和停止行为：

```powershell
.\scripts\mock-stress.ps1 -Target 5 -Concurrency 1 -FailEvery 2
.\scripts\mock-stress.ps1 -Target 5 -Concurrency 1 -FailEvery 2 -ExpectFailureThreshold true
```

预期会在第二次 mock submission 失败后触发 `failure_threshold_reached: true`。这仍然是本地 mock，不访问网络。

## 解释原则

- `duration_ms` 和 `throughput_per_second` 主要用于对比同一台机器上的相对变化。
- `heap_alloc_delta_bytes` 和 `total_alloc_delta_bytes` 用于观察运行期间新增分配。
- `goroutines` 用于确认 worker 已回收，长期目标是运行后不留下额外 goroutine。
- 真实平台运行加入后，仍要保持 provider 能力门控；登录、验证、风控、设备次数上限必须停止并报告。
