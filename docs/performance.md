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

## JSON 汇总

脚本可输出单个 JSON 汇总，方便后续 CI 或轻量 GUI 读取：

```powershell
.\scripts\mock-stress.ps1 -Json
```

JSON 汇总不要和 `--events jsonl` 混用。前者是最终报告，后者是事件流。

## 预算断言

脚本支持轻量预算断言，适合本地提交前或后续 CI 使用：

```powershell
.\scripts\mock-stress.ps1 -Target 1000 -Concurrency 1000 -MinThroughput 1000 -MaxGoroutines 1
```

可用预算包括：

- `-MinThroughput`：要求 `throughput_per_second` 不低于指定值。
- `-MaxHeapDelta`：要求 `heap_alloc_delta_bytes` 不高于指定值。
- `-MaxGoroutines`：要求运行结束后的 goroutine 数不高于指定值。
- `-ExpectFailureThreshold`：要求 `failure_threshold_reached` 等于指定布尔值。

这些预算应先设得保守，主要用于抓明显退化。严苛性能门槛必须先在 CI 机器上积累基线。

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
