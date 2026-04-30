# 架构说明

## 目标

SurveyController-go 是原 Python 版 SurveyController 的 Go 语言重写。当前阶段优先做 CLI 工具，目标是更轻、更快、更容易测试和部署。

`v0.1` 只做项目初始化、规范、文档和最小 CLI，不实现真实问卷运行。三平台正式支持属于 `v1.0` 目标。

## 设计原则

- 先强类型建模，再写平台细节。
- 先 provider 契约，再写具体平台。
- 先解析和配置生成，再接真实运行。
- 先纯函数测试，再做浏览器集成。
- 运行内核可选：`hybrid`、`browser`、`http`。
- 平台验证、登录要求、反滥用页面必须停止并报告，不做绕过。

## 总体分层

```text
cmd/surveyctl
  -> internal/app
      -> internal/config
      -> internal/provider
      -> internal/runner
          -> internal/engine
              -> internal/browser
              -> internal/httpclient
          -> internal/answer
          -> internal/proxy
          -> internal/sample
      -> internal/logging
```

## 包职责

| 包 | 职责 |
| --- | --- |
| `cmd/surveyctl` | CLI 入口，只负责参数、退出码和人类可读输出 |
| `internal/app` | 用例编排：解析、生成配置、运行任务、doctor 检查 |
| `internal/config` | 配置 schema、迁移、校验、读写 |
| `internal/domain` | 后续可放核心领域模型：问卷、题目、答案、错误码 |
| `internal/provider` | provider 接口、注册表、能力声明、标准模型 |
| `internal/provider/wjx` | 问卷星实现 |
| `internal/provider/tencent` | 腾讯问卷实现 |
| `internal/provider/credamo` | Credamo 见数实现 |
| `internal/runner` | worker 池、进度、停止策略、运行状态聚合 |
| `internal/engine` | 单份问卷执行流程和运行模式选择 |
| `internal/browser` | Playwright Go 封装、浏览器池、页面会话 |
| `internal/httpclient` | HTTP 连接池、代理 transport、重试和超时 |
| `internal/answer` | 题型配置、答案计划、概率、严格比例、信效度 |
| `internal/proxy` | 代理租约、代理池、地区、TTL、健康检查 |
| `internal/sample` | 反填数据源、样本租约、提交和回滚 |
| `internal/logging` | 结构化日志、事件输出、脱敏 |
| `internal/testkit` | fixture、mock provider、mock browser、集成测试辅助 |

## 核心数据模型

后续应逐步建立这些强类型模型：

- `ProviderID`：平台标识。
- `QuestionKind`：题型枚举。
- `SurveyDefinition`：解析后的标准问卷。
- `QuestionDefinition`：单题结构、选项、矩阵行、页码、provider 原始 id。
- `QuestionConfig`：用户对单题的配置。
- `RunConfig`：用户输入或配置文件。
- `RunPlan`：启动前编译好的不可变计划。
- `QuestionPlan`：某题的答案策略和运行元数据。
- `Answer`：运行时产生的单题答案。
- `SubmissionResult`：单份提交结果。
- `RunState`：并发安全的运行状态。
- `RunEvent`：CLI 和未来 UI 订阅的进度事件。

## Provider 契约

Provider 不只是 parser。它应声明自己能做什么。

```go
type Provider interface {
    ID() ProviderID
    MatchURL(rawURL string) bool
    Capabilities() Capabilities
    Parse(ctx context.Context, req ParseRequest) (*SurveyDefinition, error)
    NewRunner(ctx context.Context, plan *RunPlan) (ProviderRunner, error)
}
```

建议能力声明：

```go
type Capabilities struct {
    ParseHTTP       bool
    ParseBrowser    bool
    RunBrowser      bool
    SubmitHTTP      bool
    SubmitBrowser   bool
    SupportsHybrid  bool
    RequiresLoginOK bool
}
```

Provider runner 可以继续细分：

```go
type ProviderRunner interface {
    Prepare(ctx context.Context, session EngineSession) error
    Fill(ctx context.Context, session EngineSession, answers []Answer) error
    Submit(ctx context.Context, session EngineSession) (*SubmissionResult, error)
    Detect(ctx context.Context, session EngineSession) (*PageState, error)
    Close(ctx context.Context) error
}
```

## 运行内核

运行模式是用户可选项。

| 模式 | 语义 |
| --- | --- |
| `browser` | 全程浏览器执行，优先兼容性 |
| `http` | 全程 HTTP 快速路径，provider 不支持时直接失败 |
| `hybrid` | 默认模式，浏览器保证兼容性，安全可复用请求时才走 HTTP |

关键规则：

- `http` 模式不能静默降级为浏览器。
- `browser` 模式不能偷偷走 HTTP 提交，除非用户允许 hybrid。
- `hybrid` 也必须由 provider 显式声明支持。
- provider 必须给出失败原因，例如“不支持 HTTP 提交”“需要登录”“命中验证”。

## Runner 设计

Runner 只处理任务生命周期：

1. 加载 `RunConfig`。
2. 编译 `RunPlan`。
3. 创建 `context.Context`。
4. 初始化 provider、代理池、样本源、浏览器池、HTTP 客户端。
5. 启动 worker pool。
6. 聚合 `RunEvent`。
7. 根据目标份数、失败阈值、用户取消或终止错误退出。

Runner 不应知道平台 DOM 选择器，也不应决定某题怎么点。

## Engine 设计

Engine 处理单份任务：

1. 申请 worker 资源。
2. 申请代理和样本。
3. 创建浏览器或 HTTP session。
4. 调用 provider 准备页面。
5. 调用答案计划生成答案。
6. 调用 provider 填写。
7. 调用 provider 提交。
8. 调用提交结果判定。
9. 成功则提交状态，失败则按错误类型回滚资源。

Engine 不应直接依赖 CLI，也不应输出 UI 文案。

## 浏览器层

浏览器层提供项目自己的最小接口，不复刻 Selenium 风格 API。

```go
type BrowserPool interface {
    NewSession(ctx context.Context, opts BrowserSessionOptions) (BrowserSession, error)
    Close(ctx context.Context) error
}

type BrowserSession interface {
    Page() Page
    Close(ctx context.Context) error
}
```

设计要求：

- session 所有权归创建它的 worker。
- 所有导航、等待、点击、输入都必须支持 context 超时。
- 关闭动作使用 `defer`，必要时有兜底 kill，但不能成为常规路径。
- 浏览器错误映射为稳定错误码。

## HTTP 层

HTTP 层围绕 Go `net/http` 设计：

- 按代理、TLS、超时策略缓存 `Transport`。
- 统一设置 User-Agent、headers、cookie jar。
- 支持请求记录和脱敏日志。
- 支持 provider 注入平台专属 headers。
- 支持 fixture 测试中的 fake client。

HTTP 快速路径必须通过 provider 能力声明开启。

## 答案策略层

答案策略层应尽量是纯函数：

- 概率归一化。
- 权重抽样。
- 严格比例纠偏。
- 多选最小/最大限制。
- 强制选项和陷阱题。
- 随机文本。
- 反填映射。
- 维度一致性。
- 联合信效度计划。

Provider 只负责“把答案填进去”，不负责“决定答案是什么”。

## 配置

配置文件必须包含 schema version。

建议结构：

```yaml
schema_version: 1
survey:
  url: ""
  provider: ""
run:
  target: 1
  concurrency: 1
  mode: hybrid
questions: []
proxy: {}
answer: {}
```

迁移要求：

- 所有迁移集中在 `internal/config/migrate`。
- 旧字段不在业务层兼容。
- 不支持的旧配置给出清晰错误。

## 错误模型

错误应结构化，而不是只返回字符串。

建议错误码：

- `config_invalid`
- `provider_unsupported`
- `parse_failed`
- `browser_start_failed`
- `page_load_failed`
- `fill_failed`
- `submit_failed`
- `verification_required`
- `login_required`
- `device_quota_limited`
- `proxy_unavailable`
- `sample_exhausted`
- `user_cancelled`

CLI 根据错误码决定退出码和输出。

## 事件模型

运行时向外发送事件：

- `run_started`
- `worker_started`
- `worker_progress`
- `submission_success`
- `submission_failure`
- `provider_warning`
- `verification_required`
- `run_paused`
- `run_stopped`
- `run_finished`

CLI 默认渲染人类可读文本，`--json` 输出 JSON Lines。

## 测试策略

测试分层：

| 层级 | 内容 |
| --- | --- |
| 单元测试 | 配置、URL 识别、概率、信效度、强制题识别 |
| fixture 测试 | 问卷星 HTML、腾讯 API JSON、Credamo DOM 抽取结果 |
| mock engine 测试 | worker、停止、失败阈值、资源回收 |
| 浏览器集成测试 | Playwright Go，按标签或 nightly 运行 |
| benchmark | parser、answer plan、HTTP client、runner |

## 性能方向

Go 版重点优化：

- CLI 启动无需加载桌面 UI。
- HTTP 解析路径优先使用连接池。
- Parser 纯函数化，避免启动浏览器。
- Worker 池复用浏览器底座。
- 配置编译为不可变 `RunPlan`，worker 只读。
- 日志事件走 channel，减少锁竞争。
- 关键路径建立 benchmark，避免盲目优化。

## 当前 v0.1 边界

`v0.1` 只保留最小占位：

- `cmd/surveyctl version`
- `internal/provider` 契约雏形
- `internal/config` 配置占位
- `internal/runner` 运行器占位
- `internal/engine` 运行模式
- 文档、规范、CI

真实平台解析和运行从后续版本开始。
