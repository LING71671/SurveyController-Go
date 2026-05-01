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

## JSON 汇总

脚本可输出单个 JSON 汇总，方便后续 CI 或轻量 GUI 读取：

```powershell
.\scripts\mock-stress.ps1 -Json
```

JSON 汇总不要和 `--events jsonl` 混用。前者是最终报告，后者是事件流。

## 失败注入

可以用失败注入验证失败阈值和停止行为：

```powershell
.\scripts\mock-stress.ps1 -Target 5 -Concurrency 1 -FailEvery 2
```

预期会在第二次 mock submission 失败后触发 `failure_threshold_reached: true`。这仍然是本地 mock，不访问网络。

## 解释原则

- `duration_ms` 和 `throughput_per_second` 主要用于对比同一台机器上的相对变化。
- `heap_alloc_delta_bytes` 和 `total_alloc_delta_bytes` 用于观察运行期间新增分配。
- `goroutines` 用于确认 worker 已回收，长期目标是运行后不留下额外 goroutine。
- 真实平台运行加入后，仍要保持 provider 能力门控；登录、验证、风控、设备次数上限必须停止并报告。
