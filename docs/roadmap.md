# 路线图

## 总原则

- 当前只做 `v0.1`：项目初始化、规范、文档、最小 CLI、CI。
- 三平台正式支持属于 `v1.0`，不放进 `v0.1`。
- 每个版本都先补测试和文档，再扩大运行能力。
- 运行内核保持可选：`hybrid`、`browser`、`http`。
- 默认优先兼容性，只有 provider 明确声明安全时才走 HTTP 快速路径。
- Go 版长期主轴是高速、高效、轻量、高性能，并把并发能力和内存占用做到可验证的极限。
- `v1.0` 不做 GUI；后续轻量 GUI 必须作为薄外壳复用稳定 core/CLI。

## 版本路线

| 版本 | 目标 | 主要交付 |
| --- | --- | --- |
| `v0.1` | 项目初始化 | 仓库规范、中文文档、CI、最小 `surveyctl version`、架构占位 |
| `v0.2` | CLI 和配置基础 | 命令框架、配置 schema、配置迁移、结构化错误、日志事件 |
| `v0.3` | 领域模型和 provider 契约 | `SurveyDefinition`、`QuestionDefinition`、URL 识别、provider registry、fixture 测试框架 |
| `v0.4` | 答案策略纯函数 | 概率、严格比例、多选限制、强制题、随机文本、信效度基础算法 |
| `v0.5` | 运行编排基础 | `RunPlan`、`RunState`、worker pool、context 取消、事件流、mock engine 测试 |
| `v0.6` | HTTP 和缓存基础 | HTTP client、解析缓存、代理 transport、问卷星 HTML parser 原型 |
| `v0.7` | 浏览器内核基础 | Playwright Go 封装、browser session、preflight doctor、mockable page 接口 |
| `v0.8` | 腾讯与 Credamo parser 原型 | 腾讯 API parser、Credamo browser parser、三平台解析 fixture |
| `v0.9` | 运行时预览 | `browser/http/hybrid` 预览、基础提交判定、性能基准、稳定性回归 |
| `v1.0` | 三平台正式支持 | 问卷星、腾讯问卷、Credamo 的解析、配置生成、基础运行、性能回归、测试和文档闭环，不包含 GUI |
| `v1.1+` | 轻量 GUI 和扩展能力 | 轻量操作外壳、二维码辅助、反填增强、AI 主观题、打包与更新器 |

## V0.1 波次

### 波次 0：持久化探索结论

- 写入原 Python 项目深度分析。
- 写入 Go 目标架构说明。
- 写入版本路线图。
- 文档以中文为主，不写入本机绝对路径。

### 波次 1：初始化仓库

- 初始化 Go module。
- 创建最小 CLI。
- 支持 `surveyctl version`。

### 波次 2：开发规范

- 贡献指南。
- 开发指南。
- 行为准则。
- 安全政策。
- `.editorconfig` 和 `.gitignore`。

### 波次 3：GitHub 治理

- Issue 模板。
- PR 模板。
- CI 工作流。
- 发布流程骨架。

### 波次 4：架构占位

- `internal/provider`
- `internal/config`
- `internal/runner`
- `internal/engine`
- `hybrid`、`browser`、`http` 模式类型。

### 波次 5：验证

- `gofmt`
- `go test ./...`
- `go vet ./...`
- 通过 PR 合入。

## V0.2 细分

### Phase 1：CLI 框架

- 增加 `surveyctl version`、`surveyctl config validate`、`surveyctl doctor` 命令雏形。
- 统一命令错误输出。
- 明确退出码。

### Phase 2：配置 schema

- 定义 `RunConfig`。
- 定义 `schema_version`。
- 增加配置读写。
- 增加配置迁移入口。

### Phase 3：错误和日志

- 定义结构化错误码。
- 定义 `RunEvent`。
- 支持普通文本和 JSON Lines 输出。

## V0.3 细分

### Phase 1：领域模型

- 定义 provider、survey、question、option、page、condition。
- 定义题型枚举和 provider 原始字段保留策略。

### Phase 2：Provider 契约

- 定义 provider interface。
- 定义 capabilities。
- 定义 registry。
- 定义 URL matcher。

### Phase 3：Fixture 测试框架

- 建立 `testdata` 目录规范。
- 建立 parser fixture 断言工具。
- 为三平台准备最小样例。

## V0.4 细分

### Phase 1：基础概率

- 权重归一化。
- 单选/下拉/评分抽样。
- 多选概率和数量约束。

### Phase 2：强制题与随机文本

- 强制选项识别结果的运行表达。
- 随机姓名、手机号、身份证号、整数范围。

### Phase 3：信效度

- 维度一致性。
- Cronbach alpha 工具函数。
- 联合答案计划最小实现。

## V0.5 细分

### Phase 1：RunPlan

- 将配置编译为不可变运行计划。
- 将题目配置编译为 `QuestionPlan`。

### Phase 2：RunState

- 成功/失败计数。
- worker 进度。
- 失败阈值。
- 目标达成停止。

### Phase 3：Worker Pool

- context 取消。
- 并发限制。
- 事件流。
- mock provider 和 mock engine 测试。

## V0.6 细分

### Phase 1：HTTP Client

- 连接池。
- 代理 transport。
- 超时。
- header/cookie 处理。

### Phase 2：解析缓存

- URL 归一化。
- 指纹。
- TTL。
- 可注入存储。

### Phase 3：问卷星 HTML Parser

- HTML fixture。
- 标准题目输出。
- 暂停和未开放检测。

## V0.7 细分

### Phase 1：Playwright Go 封装

- browser pool。
- browser session。
- page 接口。
- 生命周期和错误映射。

### Phase 2：Doctor

- `surveyctl doctor browser`。
- 浏览器环境检查。
- 代理连通性检查占位。

### Phase 3：浏览器 Parser 支撑

- headless 打开页面。
- 读取 HTML。
- context 超时和取消。

## V0.8 细分

### Phase 1：腾讯 API Parser

- URL 识别。
- survey id/hash 提取。
- session/meta/questions API client skeleton。
- locale fallback。
- 可注入 HTTP client。
- 登录要求识别。
- provider type 映射。
- 附加填空选项、多选上下限、矩阵行列映射。

### Phase 2：Credamo Browser Parser

- DOM snapshot parser。
- 页面题目抽取 JS 输出结构。
- 强制选项、算术陷阱题、强制填空纯函数识别。
- 动态显隐题采集 skeleton。
- 翻页和去重策略。
- 强制选项和强制填空。

### Phase 3：三平台 Fixture

- 问卷星 HTML。
- 腾讯 API JSON。
- Credamo DOM 抽取 JSON。
- 原 Python 项目关键 parser 单测迁移。

## V0.9 细分

### Phase 1：运行内核预览

- `browser` 模式跑通 mock 和最小真实链路。
- `http` 模式能力检查。
- `hybrid` 模式能力选择。
- provider runner/detector 契约。

### Phase 1.5：HTTP-first 高并发主路径

- WJX HTTP submission draft。
- WJX HTTP mock executor，不访问网络。
- WJX HTTP response detector。
- AnswerPlan 最小闭环，先覆盖单选、多选、评分/量表。
- HTTP path benchmark，覆盖 1000、5000 轻量任务和分配基线。
- Provider capability gate：只有显式 `SubmitHTTP` 才允许 `http` 模式真实执行。
- Browser 保持小池兼容兜底，不作为高并发主线。

### Phase 2：提交判定

- 完成页识别。
- 验证页识别。
- 登录要求识别。
- 设备次数上限识别。
- 提交成功短路信号。
- provider 校验文案提取。

### Phase 3：性能和稳定性

- Parser benchmark。
- Answer planner benchmark。
- Worker pool benchmark。
- Worker pool 高并发压测。
- Runner 内存分配基准。
- Parser 大 fixture 内存分配基准。
- 浏览器资源回收回归。
- 真实 browser doctor probe，按环境跳过集成测试。
- 运行事件和状态不能依赖 UI。
- core 包避免引入 GUI 或重量级桌面依赖。

## V1.0 边界

`v1.0` 必须满足：

- 三平台可解析。
- 三平台可生成默认配置。
- 三平台基础运行链路可用。
- 命中登录、验证、反滥用页面时停止并报告。
- 配置、运行、provider、答案策略有测试。
- CLI 文档完整。
- 仍保持授权学习与测试使用边界。
- 不包含 GUI；后续轻量 GUI 只作为 core/CLI 的薄外壳。
- 保留基础 benchmark 或性能回归用例，确保轻量高效目标可验证。
- 高并发运行必须受资源池和配置上限约束，不能靠无界 goroutine 堆并发。
- 核心运行链路需要内存分配基线，防止 parser、runner、event buffer 长会话膨胀。
