# 原 Python 项目运行闭环对齐复盘

日期：2026-04-30

来源：本地 `B:\SurveyController\SurveyController-main` 工作副本。

## 本次对齐目标

Go 版已经进入 `v0.9` 运行时预览：已有提交判定契约、`engine.SubmissionResult`、runner 提交预览状态和 worker pool mock 提交流程。本次复盘用于防止后续开发只沿着 Go 当前抽象惯性推进，而遗漏原 Python 项目的真实运行闭环。

本次重点抽样：

- 根入口、依赖和打包清单。
- `software.core.engine` 的提交判定、停止策略和执行循环。
- `software.providers` 的标准契约和归一化字段。
- `software.core.config` 的运行配置字段。
- 代理、反填、二维码等用户可见能力。

## 原项目能力轮廓

原项目不是单纯 parser 或 runner，而是完整桌面应用：

- GUI：PySide6 + Fluent Widgets。
- 浏览器：Playwright，运行时偏向真实浏览器链路。
- 平台：问卷星、腾讯问卷、Credamo。
- 输入：URL、二维码图片、配置文件、问卷星 Excel 反填。
- 策略：概率、严格比例、信效度、联合样本、AI 主观题。
- 网络：随机 IP、指定地区 IP、自定义代理 API、代理配额与坏代理记录。
- 运行控制：暂停、停止、失败阈值、目标达成停止、浏览器启动失败归因。
- 打包：PyInstaller，Windows 桌面应用优先。

Go 版当前选择 CLI-first 是合理的，而且符合新项目定位：高速、高效、轻量、高性能，并把并发和内存占用做到可验证的极限。不能误以为 CLI-first 等于只做 parser；`v1.0` 至少要跑通核心运行闭环，不做 GUI。后续可以在核心稳定后增加轻量化 GUI，但它应作为独立外壳复用 CLI/core 能力，而不是复刻原 PySide6 桌面应用。

## 已对齐的部分

| 原项目能力 | Go 当前状态 | 结论 |
| --- | --- | --- |
| Provider registry 与 URL 分发 | 已有 provider registry、URL matcher、capability | 方向正确 |
| 标准问卷模型 | 已有 `domain.SurveyDefinition`、题型、`ProviderRaw` | 比原项目 dict 更稳，应继续扩字段 |
| HTTP client 与缓存 | 已有 `httpclient`、`parsecache` | 与原项目 `httpx` + cache 思路一致 |
| 三平台 parser 原型 | 已有 WJX/Tencent/Credamo parser fixture | 仍是 parser 原型，不是完整平台支持 |
| 浏览器抽象 | 已有 `browser.Page`、fake、HTML fetch | mockability 方向正确 |
| 提交判定 | 已有 `provider.SubmissionDetector`、`engine.SubmissionResult` | 已对齐原项目 `SubmissionOutcome` 的核心字段 |
| Runner 状态 | 已有成功、失败、停止请求、事件 | 已接近 `RunStopPolicy` 的最小形态 |
| Worker pool | 已支持 `SubmissionTask` 预览路径 | 已能跑 mock runtime preview |

## 新发现的关键差距

### 1. 失败原因需要独立模型

原项目有 `FailureReason`：

- `browser_start_failed`
- `proxy_unavailable`
- `page_load_failed`
- `fill_failed`
- `submission_verification_required`
- `device_quota_limit`
- `user_stopped`

Go 当前已有 `apperr.Code`，但 runner 的 `SubmissionResult` 事件还没有稳定记录失败原因字段。下一步应该把失败归因写入事件和状态快照，避免后续只靠 message 做判断。

建议：

- 在 runner 事件字段中稳定输出 `error_code` 或 `failure_reason`。
- `engine.SubmissionResult` 保持结构化错误，不把错误降级为字符串。
- 设备次数上限、验证码、登录、用户取消必须是终止性原因。

### 2. 停止策略不只是阈值

原项目 `RunStopPolicy` 处理：

- 暂停等待。
- 成功后清零连续失败。
- 失败阈值。
- 目标达成停止。
- 反填样本不足停止。
- 设备上限和验证页停止。
- 成功后触发随机 IP 换新。

Go 当前 runner 已有目标、失败阈值、显式停止，但还没有：

- 连续失败清零语义。
- 失败阈值是否启用的配置项。
- 停止分类。
- 成功后代理轮换动作。
- 样本租约提交/回滚。

这些应作为 `v0.9` 后续，而不是等到最后再补。

### 3. ExecutionLoop 的资源租约顺序很重要

原项目单轮执行大致顺序：

1. 等待暂停/停止。
2. 准备浏览器 session。
3. 加载问卷。
4. 识别设备次数上限。
5. 重置比例/分布临时状态。
6. 预留联合信效度样本。
7. 获取反填样本。
8. provider 填写。
9. 提交后判定。
10. 成功提交样本和分布；失败回滚样本。
11. 必要时释放浏览器或代理。

Go 当前 worker pool 已能跑提交预览，但还没有“单份执行事务”概念。后续应新增 `ExecutionAttempt` 或类似模型，先用 fake resource lease 测试提交/回滚顺序，再接真实浏览器。

### 4. 配置 schema 明显偏薄

原项目 `RuntimeConfig` 包含：

- `submit_interval`
- `answer_duration`
- `timed_mode_enabled`
- `timed_mode_interval`
- `random_ip_enabled`
- `proxy_source`
- `custom_proxy_api`
- `proxy_area_code`
- `random_ua_enabled`
- `random_ua_keys`
- `random_ua_ratios`
- `fail_stop_enabled`
- `pause_on_aliyun_captcha`
- `reliability_mode_enabled`
- `psycho_target_alpha`
- `headless_mode`
- AI 配置
- reverse fill 配置
- answer rules、dimension groups、question entries

Go 当前 `RunConfig` 只有 `target/concurrency/mode` 和 map 占位。`v1.0` 前至少需要补：

- `failure_threshold` 和 `fail_stop_enabled`。
- `headless`。
- `submit_interval` 与 `answer_duration`。
- `proxy` 结构化配置。
- `reverse_fill` 结构化配置占位。
- `random_ua` 配置占位。

AI、GUI 设置、打包配置可以后置。

### 5. Provider 标准字段还需继续扩展

原项目 `normalize_survey_questions` 中的字段远多于 Go 当前模型：

- page、provider page id、provider question id、provider type。
- row texts、text inputs、text input labels。
- forced option、forced text。
- fillable options、attached option selects。
- location、rating、multi text、text like、slider matrix。
- jump rules、display conditions。
- slider min/max/step。
- multi min/max limit。
- unsupported 与 unsupported reason。

Go 当前已有 `ProviderRaw`，但核心字段还不足。下一轮 parser 对齐应优先补“运行会直接依赖”的字段：

- 多选上下限。
- 附加填空选项。
- 文本输入数量和标签。
- unsupported reason。
- provider page/question/type。

### 6. 原项目的非 parser 用户能力需要排期

以下能力当前 Go 版尚未纳入近期路线：

- 二维码解析：`zxing-cpp` + QImage。
- 问卷星 Excel 反填：`openpyxl` 读取导出表。
- 代理策略：默认源、自定义 API、地区、占用分钟、坏代理记录。
- IP 使用记录报表。
- 更新器和桌面打包。

建议 `v1.0` 只承诺运行核心：

- 二维码解析可作为 CLI 辅助命令，非运行闭环必需。
- Excel 反填需要先做 schema 和样本租约，不急于支持全部格式。
- 代理策略至少要有接口和 fake pool，真实 API 可以后置。
- GUI、打包、更新器放到 `v1.1+`。

## 对当前 Go 方向的校准

最近两轮 `engine.SubmissionResult`、runner 状态、`RunSubmissions` 没有跑偏，它们对应原项目：

- `SubmissionOutcome`
- `RunStopPolicy.record_success/record_failure`
- `ExecutionLoop` 中提交后的成功/失败/停止分支

但下一步不宜继续只加 worker pool 功能。更应该补三个基础边界：

1. 结构化失败归因进入事件和状态。
2. 单份执行事务和资源租约提交/回滚。
3. 配置 schema 补齐运行闭环所需字段。

同时，后续实现要遵守性能取舍：

- 优先 CLI/core，避免把 UI 状态渗透到运行核心。
- 优先标准库和小型接口，避免引入重量级依赖。
- 浏览器链路只在兼容性需要时使用，能用 HTTP/fixture/mock 验证的地方不启动浏览器。
- `v0.9` 起保留 benchmark 和 race 回归，防止性能目标在迭代中变成口号。
- 轻量 GUI 只做操作入口和状态展示，不重新承载业务逻辑。
- 高并发必须由 worker pool、browser pool、HTTP transport pool 和代理池共同限流，不能靠无界 goroutine 堆吞吐。
- 内存占用要纳入基准：parser 大 fixture、runner 长会话、事件缓冲、缓存 TTL 都需要可测上限。
- 热路径优先预编译计划和复用资源，减少每次提交的临时对象分配。

## 建议下一批 issue

| 顺序 | 建议 issue | 对齐来源 | 目标 |
| --- | --- | --- | --- |
| 1 | runner 事件增加失败归因字段 | `FailureReason`、`RunStopPolicy` | 让运行结果可机器判断 |
| 2 | execution attempt 事务骨架 | `ExecutionLoop` | 测试代理/样本/浏览器资源提交回滚顺序 |
| 3 | config runtime schema 扩展 | `RuntimeConfig` | 增加失败阈值、headless、间隔、代理、反填占位 |
| 4 | provider model 补运行关键字段 | `contracts.normalize_survey_questions` | 多选上下限、附加填空、unsupported reason |
| 5 | proxy lease fake pool | proxy policy/pool | 为运行预览准备代理租约接口 |
| 6 | reverse fill sample schema | reverse fill schema | 为 Excel 反填准备样本租约 |
| 7 | runtime benchmark baseline | 高性能目标 | 建立 worker/parser/runner 并发和内存基线 |
| 8 | QR decode CLI 调研 | QR utils | 后续作为辅助输入能力 |

## 当前边界判断

- `v0.9` 应继续聚焦 runtime preview，不要转向 GUI。
- `v1.0` 应定义为“三平台解析 + 基础运行闭环 + 安全停止报告”，明确不做 GUI。
- 轻量化 GUI 可以放到 `v1.1+` 或单独里程碑，作为薄客户端调用稳定 core/CLI。
- 高速、高效、轻量、高性能是 Go 版长期主轴，功能迁移必须服务这个目标。
- 并发和内存占用是硬指标，后续功能进入 core 前要考虑 benchmark、race 和资源上限。
- 原项目中的 AI 主观题、更新器、PyInstaller 打包、Fluent UI 等桌面应用能力是后续版本能力。
