# 脚本参考

本页记录 `scripts/` 下的本地验证和压测入口。所有脚本都只调用本地 mock 或本地 dry-run 路径，不执行真实问卷平台提交。

## PowerShell 兼容性

脚本可以在 Windows PowerShell 和 PowerShell Core 上运行。矩阵脚本和 `verify-local.ps1` 会通过 `scripts/lib/powershell.ps1` 自动选择可用的 `pwsh` 或 `powershell` 来启动子脚本。

Windows 本地常用：

```powershell
powershell -ExecutionPolicy Bypass -File scripts\verify-local.ps1
```

PowerShell Core 常用：

```powershell
pwsh -File scripts/verify-local.ps1
```

## verify-local.ps1

默认验证入口：

```powershell
.\scripts\verify-local.ps1
```

默认执行：

- `go test ./...`
- `go vet ./...`
- `staticcheck ./...`
- 轻量 mock stress matrix

常用参数：

```powershell
.\scripts\verify-local.ps1 -SkipStress
.\scripts\verify-local.ps1 -SkipStaticcheck
.\scripts\verify-local.ps1 -IncludeFullStress
.\scripts\verify-local.ps1 -IncludeWJXHTTPDryRunStress
.\scripts\verify-local.ps1 -IncludeWJXHTTPDryRunStress -IncludeFullStress
```

CI smoke 使用：

```powershell
.\scripts\verify-local.ps1 -SkipGoChecks -SkipStaticcheck -SkipStress -IncludeWJXHTTPDryRunStress
```

`-SkipGoChecks` 只用于上层流程已经单独跑过 Go 检查的场景，例如 CI quality job。普通本地提交前不建议跳过 Go 检查。

## mock-stress.ps1

本地 mock runner 压测入口，不访问网络：

```powershell
.\scripts\mock-stress.ps1
.\scripts\mock-stress.ps1 -Target 1000 -Concurrency 1000 -Json
.\scripts\mock-stress.ps1 -Target 5 -Concurrency 1 -FailEvery 2
```

可用预算参数：

```powershell
.\scripts\mock-stress.ps1 -MinThroughput 1 -MaxGoroutines 1
.\scripts\mock-stress.ps1 -ExpectFailureThreshold true -FailEvery 2
```

## mock-stress-matrix.ps1

一次运行多组 mock profile：

```powershell
.\scripts\mock-stress-matrix.ps1 -SkipFull
.\scripts\mock-stress-matrix.ps1
```

`-SkipFull` 只跑轻量 smoke 和失败阈值 profile。默认包含 1000x1000 profile，适合发布前或性能专项验证。

## wjx-http-dryrun-stress.ps1

问卷星 HTTP dry-run 压测入口。它会读取本地 fixture，经过 runner、worker pool、答案计划和 HTTP pipeline，最后由本地 dry-run executor 记录 draft，不访问网络：

```powershell
.\scripts\wjx-http-dryrun-stress.ps1
.\scripts\wjx-http-dryrun-stress.ps1 -Target 1000 -Concurrency 1000
.\scripts\wjx-http-dryrun-stress.ps1 -Target 1000 -Concurrency 1000 -Json
```

默认输出压缩 summary，包含成功数、失败数、吞吐、资源指标、draft 数量和首个 draft 摘要；不会默认打印所有 drafts。

## wjx-http-dryrun-stress-matrix.ps1

一次运行多组 WJX HTTP dry-run profile：

```powershell
.\scripts\wjx-http-dryrun-stress-matrix.ps1 -SkipFull
.\scripts\wjx-http-dryrun-stress-matrix.ps1
```

`-SkipFull` 只跑轻量 smoke 和预算 profile。默认包含 1000x1000 profile。CI 只运行 `-SkipFull`，完整 profile 由本地显式触发。

## 选择建议

- 日常提交前：`.\scripts\verify-local.ps1`
- 只看 Go 代码：`.\scripts\verify-local.ps1 -SkipStress`
- WJX dry-run 快速回归：`.\scripts\verify-local.ps1 -IncludeWJXHTTPDryRunStress`
- 发布前性能检查：`.\scripts\verify-local.ps1 -IncludeWJXHTTPDryRunStress -IncludeFullStress`
- 调试单个 profile：直接运行对应的 `*-stress.ps1`

## 安全边界

- mock stress 只使用本地 mock submitter。
- WJX HTTP dry-run 只读取本地 fixture，并使用本地 dry-run executor。
- 脚本输出中的 `network: disabled (...)` 是安全边界提示。
- 后续真实运行入口进入 CLI 前，必须继续保持 provider 能力门控，并在登录、验证、风控、设备次数上限或频控信号出现时停止并报告。
