# 原 Python 项目深度分析报告

日期：2026-04-30

来源：本地分析时提供的原 Python 版 SurveyController 工作副本。

## 结论摘要

原项目不是一个简单脚本，而是一个以 PySide6 桌面界面为入口、以 Playwright 浏览器自动化为兼容性底座、以 `httpx` 为部分快速路径、并叠加问卷解析、答题策略、代理资源、反填、AI 填空、日志与更新能力的中型自动化应用。

Go 重写不应逐文件平移。更合适的方向是把原项目拆成更清晰的核心边界：

- 领域模型：问卷、题目、选项、运行配置、答案计划、提交结果。
- 平台适配器：问卷星、腾讯问卷、Credamo 见数各自实现解析和运行细节。
- 运行内核：`browser`、`http`、`hybrid` 三种可选模式，由平台能力声明决定是否可用。
- 编排层：CLI、配置加载、任务计划、并发 worker、取消和资源回收。
- 纯函数策略层：概率、严格比例、信效度、反填映射、题型默认配置。
- 基础设施层：浏览器、HTTP 客户端、代理、日志、缓存、文件 IO。

Go 版最能发挥优势的地方是：强类型配置、明确接口、`context.Context` 取消、goroutine worker 池、HTTP 连接池、低内存命令行运行、可测试的纯函数解析与策略，以及把“平台 DOM 脆弱逻辑”隔离在 provider 内部。

## 仓库规模

本次扫描到的主体规模：

| 类别 | 数量或规模 |
| --- | ---: |
| Python 文件 | 301 |
| Python 有效代码行 | 约 50,568 行 |
| JavaScript 文件 | 7 |
| 单元/现场测试文件 | 36 |
| 主要运行入口 | `SurveyController.py` |
| 主要桌面框架 | PySide6、PySide6-Fluent-Widgets |
| 主要自动化框架 | Playwright |
| 主要 HTTP 客户端 | httpx |
| 主要解析依赖 | BeautifulSoup |
| 主要表格依赖 | openpyxl |

按顶层目录粗略分布：

| 顶层目录 | Python 有效代码行 | 观察 |
| --- | ---: | --- |
| `software` | 约 34,683 | UI、核心引擎、配置、网络、代理、日志、AI、系统能力集中在此 |
| `wjx` | 约 7,939 | 问卷星 provider，解析、题型运行和无头 HTTP 提交逻辑最复杂 |
| `CI` | 约 3,776 | Python CI、单元测试、现场测试、Worker 辅助 |
| `tencent` | 约 2,535 | 腾讯问卷 provider，API 解析和 DOM 交互分离较明显 |
| `credamo` | 约 1,591 | Credamo provider，强依赖 Playwright 动态页面 |

体量和迁移风险最高的文件包括：

| 文件 | 主要风险 |
| --- | --- |
| `tencent/provider/runtime_interactions.py` | 腾讯问卷 DOM 交互细节集中，选择器、弹层、矩阵、下拉、星级题逻辑多 |
| `credamo/provider/runtime.py` | Credamo 动态题目、分页、等待、答题和提交耦合明显 |
| `software/network/browser/driver.py` | Playwright 生命周期、跨线程清理、代理、浏览器选择都集中在一个适配层 |
| `wjx/provider/questions/text.py` | 填空题、多项填空、位置题、AI 填空、随机文本逻辑复杂 |
| `software/integrations/ai/client.py` | 免费 AI 与第三方 AI 兼容、超时和日志处理较多 |
| `wjx/provider/_submission_core.py` | 问卷星提交按钮、抓包、cookie 迁移、HTTP 提交、代理重试全部集中 |
| `credamo/provider/parser.py` | 动态显隐题解析、预填翻页、陷阱题识别逻辑复杂 |
| `wjx/provider/runtime.py` | 问卷星题型分发、页内状态、反填索引和提交衔接 |
| `wjx/provider/questions/single.py` | 单选题附加填空、嵌入式下拉、DOM 兜底较多 |
| `tencent/provider/runtime_answerers.py` | 腾讯问卷各题型答题器，适合作为 Go provider 测试夹具来源 |

## 当前产品能力

从 README 和源码看，原项目提供以下能力：

- 支持问卷星、腾讯问卷、Credamo 见数。
- 支持图形界面配置问卷、题目、运行参数和日志。
- 支持自动解析问卷题目结构。
- 支持单选、多选、下拉、矩阵、量表、评分、滑块、排序、填空、多项填空、位置类题目。
- 支持自定义概率、严格比例、题间一致性、信度目标和联合心理测量优化。
- 支持随机 IP、指定地区 IP、代理池、代理额度和代理可用性检查。
- 支持随机 User-Agent。
- 支持作答时长、定时开放等待和提交间隔。
- 支持 Excel 反填。
- 支持 AI 主观题填空。
- 支持解析缓存、配置导入导出、日志保存、自动更新、状态页提示。

Go CLI 的初期版本不需要复刻桌面 UI，但必须保留这些能力在架构上的位置，否则后续会再次出现“大对象 + 动态字典 + provider 分支散落”的问题。

## 原项目运行链路

原项目主入口是：

1. `SurveyController.py` 调用 `software.app.main.main`。
2. `software.app.main` 初始化崩溃日志、Qt 应用、字体、HTTP 预热和主窗口。
3. `software.ui.shell.main_window` 创建主窗口和各页面。
4. 工作台页面通过 `RunController` 连接 UI 与运行引擎。
5. `RunController.parse_survey` 在后台线程调用 `software.providers.registry.parse_survey`。
6. provider 解析器返回标准化 `SurveyDefinition`。
7. `build_default_question_entries` 把题目元数据转为默认可配置项。
8. 用户调整配置后，`RunController.start_run` 校验题目配置、构造 `ExecutionConfig` 和 `ExecutionState`。
9. `RunController` 创建 worker 线程，每个线程调用 `software.core.engine.runner.run`。
10. `ExecutionLoop.run_thread` 创建浏览器会话、加载问卷、调用 provider 填写、调用提交服务判定结果。
11. 成功后提交计数、记录比例统计和反填样本；失败后按停止策略决定重试或停止。

Go 版可以把这条链路重写为：

1. CLI 解析命令和配置。
2. `app` 层构造不可变 `RunPlan`。
3. `runner` 层创建 `context.Context`、worker 池和资源池。
4. `provider` 根据 URL 和配置生成 `SurveyDefinition` 与 `AnswerPlan`。
5. `engine` 根据运行模式选择 `browser`、`http` 或 `hybrid`。
6. worker 执行单份任务，返回结构化 `SubmissionResult`。
7. `runner` 聚合进度、错误、统计和退出码。

## 当前模块边界

### 桌面入口与应用层

相关模块：

- `software/app/main.py`
- `software/app/config.py`
- `software/app/runtime_paths.py`
- `software/app/browser_probe.py`
- `software/ui/shell/*`
- `software/ui/controller/*`
- `software/ui/pages/*`
- `software/ui/widgets/*`

观察：

- UI 层非常重，占 `software` 代码的最大部分。
- `RunController` 是 UI 与 engine 的桥，承担了配置校验、初始化门禁、线程启动、随机 IP 开关、状态定时刷新、清理和提示弹窗。
- `EngineGuiAdapter` 把 UI 回调传入 engine，这让 engine 仍然知道 GUI 概念。
- 初始化门禁会在多线程无头模式前跑浏览器快检，避免直接拉起大量 worker 后失败。

Go 重写建议：

- CLI 版不要保留 GUI adapter。用 `app.Service` 或 `runner.Runner` 返回事件流。
- UI 状态和运行状态彻底分离。CLI 只订阅事件，未来 GUI/TUI 也通过同一事件接口订阅。
- 浏览器快检保留为 `surveyctl doctor browser` 和运行前可选 preflight。
- 所有可写路径通过 `internal/runtimepath` 或 `internal/storage` 管理，不散落在业务层。

### Provider 注册与平台识别

相关模块：

- `software/providers/common.py`
- `software/providers/contracts.py`
- `software/providers/registry.py`
- `software/providers/survey_cache.py`

已有优点：

- 已经有 provider 常量、URL 识别、标准化 `SurveyDefinition`。
- `registry.py` 把问卷星、腾讯问卷、Credamo 适配器统一分发。
- `contracts.py` 对题目字段做了一次归一化。
- `survey_cache.py` 有解析缓存和远端指纹思路。

主要问题：

- provider contract 仍然依赖 `Any`、字典和动态字段。
- adapter 是内部类，不利于插件式扩展和能力声明。
- `fill_survey`、验证页识别、完成页识别、提交成功信号分散在多个函数上。
- 解析缓存直接绑定运行目录和 HTTP 实现，测试时需要更多 mock。

Go 重写建议：

- `provider.Provider` 必须是显式接口，包含 `ID`、`MatchURL`、`Capabilities`、`Parse`、`NewSession` 等方法。
- `SurveyDefinition`、`Question`、`Option`、`Page`、`Condition`、`ValidationRule` 使用结构体表达。
- provider 能力必须声明：是否支持 HTTP 解析、浏览器解析、HTTP 提交、浏览器提交、混合提交、完成页识别。
- 缓存不应由 provider 直接写磁盘；应由 `parser.Cache` 包装 provider parser。

### 核心配置模型

相关模块：

- `software/core/config/schema.py`
- `software/core/config/codec.py`
- `software/core/task/task_context.py`
- `software/core/questions/schema.py`

当前模型分两层：

- `RuntimeConfig`：UI 和磁盘配置使用。
- `ExecutionConfig`：运行前固定下来的配置。
- `ExecutionState`：运行中的动态状态、锁、计数、代理租约、反填状态、联合信效度样本。
- `QuestionEntry`：用户对单题的配置项。

已有优点：

- 已经有配置 schema version，当前为 v5。
- 已经有迁移逻辑和旧字段拒绝逻辑。
- `ExecutionConfig` 与 `ExecutionState` 开始分离静态和动态状态。

主要问题：

- `ExecutionConfig` 字段非常多，按题型拆成多个并行数组，例如 `single_prob`、`matrix_prob`、`texts`、`question_config_index_map`。
- 题目元数据仍然大量使用 `Dict[str, Any]`。
- `ExecutionState` 为兼容旧路径实现了 `__getattr__` 和 `__setattr__`，说明静态/动态边界还没完全清理。
- 多处通过 `getattr` 获取字段，缺少编译期约束。

Go 重写建议：

- 以 `RunConfig` 表达用户输入，以 `RunPlan` 表达运行前编译结果，以 `RunState` 表达运行状态。
- 避免按题型并行数组，改为每题一个强类型 `AnswerStrategy` 或 `QuestionPlan`。
- 配置文件必须有 `schema_version`，迁移函数只在 `internal/config/migrate` 中出现。
- `RunPlan` 启动后不可变，worker 只能读；运行统计集中在并发安全的 `RunState`。

### 运行引擎

相关模块：

- `software/core/engine/runner.py`
- `software/core/engine/execution_loop.py`
- `software/core/engine/browser_session_service.py`
- `software/core/engine/submission_service.py`
- `software/core/engine/run_stop_policy.py`
- `software/core/engine/provider_common.py`
- `software/core/engine/cleanup.py`

当前链路：

- `runner.run` 只是薄入口。
- `ExecutionLoop` 是单个 worker 的主循环。
- `BrowserSessionService` 负责浏览器会话、代理选择、User-Agent、信号量和资源释放。
- `SubmissionService` 负责提交后的完成页、验证页、失败归因和成功计数。
- `RunStopPolicy` 管理失败阈值、目标达成、暂停和终止。
- `provider_common` 负责每轮答题前的信效度计划、persona 和一致性上下文。

已有优点：

- worker 循环已经有比较完整的失败归因。
- 浏览器会话服务已从执行循环拆出。
- 提交判定已从执行循环拆出。
- 已经用信号量限制浏览器实例数量。

主要问题：

- `ExecutionLoop` 仍然同时处理浏览器、代理、页面加载、信效度、反填、provider、提交、失败、等待间隔。
- 停止信号使用 `threading.Event`，状态散落在多个对象里。
- 浏览器清理受到 Python Playwright sync API 和线程所有权限制，需要跨线程 `taskkill` 兜底。
- provider 和 engine 通过大量 `Any` 连接。

Go 重写建议：

- `runner` 只负责任务生命周期和 worker 池。
- `engine` 只负责单份执行流程。
- `provider` 负责平台页面操作和提交动作。
- `browser` 负责浏览器生命周期，资源所有者必须明确。
- 全链路统一 `context.Context`，停止、超时、取消和 preflight 都走 context。
- 错误使用结构化类型，例如 `ErrBrowserStart`、`ErrPageLoad`、`ErrProviderUnsupported`、`ErrVerificationRequired`、`ErrDeviceQuota`。

### 浏览器封装

相关模块：

- `software/network/browser/driver.py`

当前能力：

- 延迟加载 Playwright。
- 支持 Edge 和 Chrome fallback。
- 支持 headless viewport。
- 支持 context 级代理、User-Agent。
- 支持 persistent browser manager 和 transient driver。
- 模拟 Selenium 风格的 `find_element`、`find_elements`、`execute_script`。
- 能识别 Playwright 启动环境错误、代理隧道错误、浏览器断开错误。
- 跨线程清理时避免直接调用 sync API，必要时强制结束浏览器进程树。

Go 重写建议：

- 不复刻 Selenium 风格 API；直接定义项目自己的最小 `browser.Page` 接口。
- `BrowserPool` 管理浏览器进程，`Session` 管理 context/page。
- `Session` 必须由创建它的 worker 拥有和关闭。
- 通过 context 控制导航、等待、点击、输入、脚本执行超时。
- 浏览器错误要能映射到稳定错误码，便于 CLI 退出和日志归因。

### HTTP 客户端

相关模块：

- `software/network/http/client.py`
- `wjx/provider/_submission_core.py`
- `tencent/provider/parser.py`
- `software/providers/survey_cache.py`

当前能力：

- `httpx.Client` 连接池按 proxy、verify、redirect、trust_env 缓存。
- 支持 requests 风格 `proxies` 参数兼容。
- 支持流式响应封装。
- 启动时预热 httpx/httpcore/ssl，避免后台线程首次初始化崩溃。
- 问卷星解析优先 HTTP，失败回退浏览器。
- 腾讯问卷解析优先平台 API，失败回退浏览器。
- 问卷星无头提交可抓取浏览器生成的 `processjq` 请求，再用 httpx 发送。

Go 重写建议：

- Go 的 `net/http` 连接池和 `Transport` 是性能优势，应抽象出 `HTTPClient`，按代理和 TLS 配置复用 transport。
- HTTP 快速路径必须显式声明，不应在用户选择 `http` 模式时静默回退浏览器。
- 混合模式下，浏览器可负责兼容性和请求生成，HTTP 负责安全可复用的快速提交。
- 解析缓存应以规范化 URL + 指纹为 key，缓存层包裹 parser，而非散落在 provider 内部。

### 问卷星 Provider

相关模块：

- `wjx/provider/parser.py`
- `wjx/provider/html_parser*.py`
- `wjx/provider/runtime.py`
- `wjx/provider/navigation.py`
- `wjx/provider/questions/*.py`
- `wjx/provider/_submission_core.py`
- `wjx/provider/submission.py`

解析策略：

- 优先 HTTP 获取 HTML。
- 检测暂停问卷、未开放问卷。
- 使用 BeautifulSoup 解析 `divQuestion`、`fieldset`、`topic`、题型、题号、页码、选项、矩阵行、跳转规则、显示条件、多选限制、滑块范围等。
- HTTP 失败后使用 headless Playwright 打开页面并读取 HTML。

运行策略：

- `runtime.py` 按题型分发到 `single`、`multiple`、`dropdown`、`matrix`、`scale`、`score`、`slider`、`text`、`reorder` 等模块。
- 答题前会处理开始作答、恢复弹窗、分页导航。
- 多处使用 JS 触发 input/change/click 事件，解决自定义控件不响应普通点击的问题。

提交策略：

- 普通模式点击提交按钮和确认按钮。
- 无头模式下，先通过 Playwright route 捕获问卷星提交请求，再迁移 headers/cookies/body 用 HTTP 提交。
- 支持提交代理切换和短路成功信号。
- 可识别验证码、智能验证、设备填写次数上限等。

Go 重写重点：

- 问卷星最适合作为第一个完整 provider，因为它已有 HTTP 解析和混合提交雏形。
- HTML parser 应先做成纯函数和 fixture 测试。
- 题型答题器应按 question kind 拆成小接口，不要形成一个大 runtime。
- 无头 HTTP 提交要做成 `Submitter` 能力，只有当浏览器捕获到完整请求且 provider 声明可复用时启用。

### 腾讯问卷 Provider

相关模块：

- `tencent/provider/parser.py`
- `tencent/provider/runtime.py`
- `tencent/provider/runtime_answerers.py`
- `tencent/provider/runtime_interactions.py`
- `tencent/provider/runtime_flow.py`
- `tencent/provider/navigation.py`
- `tencent/provider/submission.py`

解析策略：

- 从 URL 提取 survey id 和 hash。
- 优先调用腾讯问卷 API：`session`、`meta`、`questions`。
- 多 locale 尝试。
- 检测登录要求。
- API 失败后回退 Playwright，在页面内 fetch API。
- 把腾讯 provider type 映射到内部 type code，例如 radio、checkbox、select、text、nps、star、matrix_radio、matrix_star。

运行策略：

- `runtime_answerers.py` 负责单选、下拉、文本、量表/评分、矩阵、多选、矩阵星级。
- `runtime_interactions.py` 包含大量 DOM 操作底层函数：等待题目可见、点击输入、打开下拉、选择弹层项、矩阵单元格、星级单元格、多选约束。
- `runtime_flow.py` 处理完成页、验证页、分页问题。

Go 重写重点：

- 腾讯问卷解析优先做 HTTP/API provider，性能收益会比浏览器解析更明显。
- Runtime 由于 DOM 细节很多，应晚于解析器迁移。
- API payload 到标准题目模型的映射要做大量 fixture 测试。
- 登录要求必须变成明确错误，不进入运行。

### Credamo 见数 Provider

相关模块：

- `credamo/provider/parser.py`
- `credamo/provider/runtime.py`
- `credamo/provider/submission.py`

解析策略：

- 强依赖 Playwright。
- 页面题目通过 JS 从 `.answer-page .question` 动态提取。
- 通过预填当前页面题目触发动态显隐题，再收集新出现的题。
- 支持多页解析和去重。
- 能识别强制作答提示、简单算术陷阱题、强制填空文本。

运行策略：

- 动态等待题目根节点。
- 按 runtime question key 防重复。
- 支持单选、多选、文本、下拉、量表、排序。
- 分页与提交按钮通过可见文本识别。

Go 重写重点：

- Credamo 是三平台里最依赖浏览器的一类，不适合早期追求 HTTP 快速路径。
- 解析和运行都需要稳定的 Playwright Go 封装后再迁移。
- 预填触发动态显隐的行为需要严格隔离在 provider 内，避免污染通用 parser 模型。
- 陷阱题识别适合做成纯函数包并先迁移测试。

## 题型与答案策略

核心模块：

- `software/core/questions/schema.py`
- `software/core/questions/default_builder.py`
- `software/core/questions/normalization.py`
- `software/core/questions/validation.py`
- `software/core/questions/tendency.py`
- `software/core/questions/distribution.py`
- `software/core/questions/consistency.py`
- `software/core/questions/strict_ratio.py`
- `software/core/psychometrics/*`

当前题型模型：

- `QuestionEntry` 表示用户配置，包括题型、概率、文本、行数、选项数、provider 字段、AI、随机文本、附加填空、维度、心理测量偏向。
- `configure_probabilities` 把 `QuestionEntry` 编译进 `ExecutionConfig` 的多个题型数组。
- `default_builder` 从解析出来的题目元数据生成默认题目配置。
- `validation` 负责启动前检查。

当前答案策略：

- 普通概率：按权重抽样。
- 严格比例：根据运行统计纠偏。
- 维度一致性：同一份问卷内按维度生成基准倾向。
- 联合信效度：按目标 Cronbach alpha 生成整批答案计划。
- 强制选项：题目提示要求选择特定答案时锁定。
- 反填：从表格样本为指定题目取答案。
- AI 填空：按题目标题和上下文生成文本。
- 随机文本：姓名、手机号、身份证号、整数范围等。

Go 重写建议：

- 把 `QuestionEntry` 拆成 `QuestionDefinition`、`QuestionConfig`、`QuestionPlan`。
- 答案生成统一为 `AnswerPlanner`，输出 `Answer`，provider 只负责把答案落实到页面或请求。
- 概率、严格比例、信效度、强制题、随机文本都应是纯函数，可用 Go table tests 覆盖。
- 不要在 provider DOM 代码里直接决定概率；provider 只接收已经算好的答案。

## 并发与状态

当前并发模型：

- UI 解析用后台线程。
- 每个运行 worker 是一个 Python `threading.Thread`。
- worker 共用一个 `ExecutionState`，内部用 `threading.Lock`、`Semaphore`、`Event`。
- 浏览器实例数量由 semaphore 限制。
- 随机 IP、代理池、反填样本、联合信效度样本都在共享状态里加锁管理。
- 日志缓冲使用后台队列线程。

问题：

- 共享状态对象过大，容易出现隐式耦合。
- 部分状态通过 thread name 作为 key。
- Python sync Playwright 和线程所有权导致清理复杂。
- UI 线程和 worker 线程之间需要大量 adapter 和队列。

Go 重写建议：

- 用 `context.Context` 统一取消。
- 用 `errgroup` 或自定义 worker pool 管理 worker 生命周期。
- `RunState` 提供小而明确的方法：`ReserveSample`、`CommitSuccess`、`RecordFailure`、`SnapshotProgress`。
- worker id 使用显式整数，不使用线程名字符串。
- 事件流通过 channel 输出，CLI 可直接渲染进度；未来 GUI/TUI 也订阅同一事件。

## 代理与网络资源

相关模块：

- `software/network/proxy/api/provider.py`
- `software/network/proxy/pool/pool.py`
- `software/network/proxy/pool/prefetch.py`
- `software/network/proxy/policy/source.py`
- `software/network/proxy/policy/settings.py`
- `software/network/proxy/session/auth.py`
- `software/network/session_policy.py`

当前能力：

- 支持默认、福利、自定义代理来源。
- 支持代理地区代码。
- 支持代理额度和会话。
- 支持批量提取、可用性检查、TTL 判断、坏代理剔除。
- 无头提交时可单独使用提交代理。

Go 重写建议：

- 代理作为 `Lease` 资源，由 `ProxyPool` 管理。
- `Lease` 应包含地址、过期时间、来源、是否可复用、失败次数。
- 浏览器代理和 HTTP 提交代理要分开建模。
- 网络请求统一走可注入的 `HTTPClient`，便于测试。
- 代理可用性检查、TTL 判断、地区解析都应独立成纯函数或小服务。

## 反填与外部数据

相关模块：

- `software/core/reverse_fill/*`
- `software/io/spreadsheets/wjx_excel.py`
- `CI/unit_tests/engine/test_reverse_fill_runtime.py`
- `CI/unit_tests/providers/test_wjx_reverse_fill.py`

当前能力：

- 从表格解析样本。
- 建立题目到表格列的映射。
- worker 运行前获取反填样本。
- 提交成功后提交样本，失败时可重排或丢弃。
- 当样本不足以达到目标份数时终止。

Go 重写建议：

- 反填要从一开始作为 `SampleSource` 接口设计。
- Excel、CSV、JSONL 可以是不同实现。
- worker 通过 `SampleLease` 获取样本，提交成功 `Commit`，失败 `Release` 或 `Discard`。
- 反填答案应编译进 `AnswerPlan`，provider 不直接读表格。

## 日志、错误与可观测性

相关模块：

- `software/logging/log_utils.py`
- `software/core/engine/failure_reason.py`
- `software/core/engine/submission_service.py`
- `software/ui/pages/workbench/log_panel/*`

当前能力：

- 异步日志缓冲。
- 日志分类为 INFO、OK、WARNING、ERROR。
- 过滤敏感 token 和第三方噪声。
- 未处理异常接入日志。
- 运行失败有粗粒度 `FailureReason`。

Go 重写建议：

- 使用结构化日志，字段至少包含 run id、worker id、provider、phase、question id、attempt、error code。
- CLI 默认人类可读，`--json` 输出机器可读事件。
- 错误模型分层：配置错误、平台不支持、浏览器错误、网络错误、验证命中、提交失败、用户取消。
- 敏感字段必须在日志层统一脱敏。

## 测试资产

已有测试覆盖方向：

- 配置编解码和运行路径。
- 浏览器快检、执行循环、停止策略、提交服务、浏览器会话服务。
- provider 通用行为、问卷缓存。
- Credamo 解析器、运行等待、运行时。
- 问卷星反填。
- 随机 IP 会话和额度归一化。
- 心理测量方向、Cronbach alpha、联合优化器。
- 题目配置校验。
- 现场解析器测试。

Go 重写的测试迁移建议：

- 第一阶段迁移纯函数测试：URL 识别、配置迁移、概率归一化、题型默认配置、强制题识别、信效度算法。
- 第二阶段迁移 parser fixture：问卷星 HTML、腾讯 API JSON、Credamo 页面抽取结果。
- 第三阶段迁移 engine mock 测试：worker、停止、失败阈值、提交结果判定。
- 第四阶段迁移 Playwright 集成测试：只在可用环境或 nightly CI 运行。
- 第五阶段建立 benchmark：解析耗时、配置编译耗时、答案计划生成耗时、HTTP 提交路径耗时。

## Go 版可发挥优势的点

### 强类型领域模型

Python 版大量依赖动态字典和 `Any`。Go 版应把核心数据变为强类型：

- `SurveyDefinition`
- `QuestionDefinition`
- `QuestionKind`
- `ProviderID`
- `RunConfig`
- `RunPlan`
- `QuestionPlan`
- `Answer`
- `SubmissionResult`
- `FailureReason`

这样能在编译期消灭大量字段名拼写、类型断言和 `getattr` 分支。

### 可组合接口

Provider 能力不应靠主流程 `if provider == ...` 判断，而应由接口表达：

- `Parser`
- `Runner`
- `Submitter`
- `CompletionDetector`
- `VerificationDetector`
- `CapabilityReporter`

这样后续加平台时不改主循环。

### 并发和取消

Go 的 goroutine、channel、context 更适合 CLI 并发运行：

- `context.Context` 贯穿解析、运行、提交、等待、清理。
- worker pool 控制并发。
- `RunEvent` channel 输出进度。
- `sync.Mutex` 只出现在小状态对象里。
- 不需要通过 UI adapter 回调跨线程调度。

### HTTP 性能

Go 的 `net/http` 可直接复用连接池和 transport：

- 腾讯问卷 API 解析可优先吃到收益。
- 问卷星 HTML 解析和缓存可更轻量。
- 混合提交路径可减少浏览器参与时间。
- 可以按代理维度缓存 `http.Transport`。

### 低内存 CLI

去掉 PySide6 和桌面 UI 后，CLI 版天然更轻：

- 启动更快。
- 后台运行更容易。
- CI 和服务器环境更友好。
- 日志和进度可结构化输出。

## 迁移风险

| 风险 | 描述 | Go 版对策 |
| --- | --- | --- |
| 平台 DOM 易变 | 三个平台 DOM 和 API 都可能变化 | provider 内部隔离选择器，建立 fixture 和现场测试 |
| 动态配置复杂 | 原配置字段多且历史兼容复杂 | v0.2 先设计 schema 和 migration，不急着实现平台 |
| 浏览器资源泄漏 | Python 版已有跨线程清理兜底 | Go 版明确 session 所有权，所有关闭走 context 和 defer |
| HTTP 快速路径误用 | 不是所有平台和状态都能安全 HTTP 提交 | provider 显式声明能力，用户选择 `http` 时不可静默降级 |
| 反填样本一致性 | 并发下样本领取、提交、失败回滚复杂 | `SampleLease` 显式 commit/release/discard |
| 信效度与严格比例 | 算法逻辑细，迁移易偏差 | 先迁移纯函数和测试，再接运行时 |
| 代理质量 | 代理 TTL、地区、额度、失败归因复杂 | 代理池作为独立服务，提交代理和浏览器代理分离 |
| 合规边界 | 命中验证、登录、反滥用页面时不能绕过 | 错误模型明确停止，文档保持授权测试边界 |

## 建议的 Go 架构映射

| Python 现状 | Go 目标 |
| --- | --- |
| `software.providers.registry` | `internal/provider/registry` |
| `software.providers.contracts` | `internal/domain/survey` 或 `internal/provider` 标准模型 |
| `software.core.config` | `internal/config` |
| `software.core.task` | `internal/runner/state` 与 `internal/runner/plan` |
| `software.core.engine` | `internal/engine` |
| `software.network.browser` | `internal/browser` |
| `software.network.http` | `internal/httpclient` |
| `software.network.proxy` | `internal/proxy` |
| `software.core.questions` | `internal/answer` |
| `software.core.psychometrics` | `internal/answer/psychometric` |
| `software.core.reverse_fill` | `internal/sample` |
| `wjx/provider` | `internal/provider/wjx` |
| `tencent/provider` | `internal/provider/tencent` |
| `credamo/provider` | `internal/provider/credamo` |
| `software.logging` | `internal/logging` |
| `CI/unit_tests` | `internal/...` 的 table tests 和 fixture tests |

## 推荐迁移顺序

1. 配置和领域模型：先建立强类型结构体、schema version、错误码。
2. URL 识别和 provider registry：不依赖浏览器，适合早期落地。
3. 题型默认配置和答案策略：迁移纯函数，快速建立测试信心。
4. 解析器 fixture：先问卷星 HTML，再腾讯 API JSON，再 Credamo 页面抽取数据。
5. 浏览器抽象：封装 Playwright Go，建立 mockable 接口。
6. 运行器：worker pool、context、事件流、停止策略。
7. 混合内核：先声明模式，再逐步开启安全的 HTTP 快速路径。
8. 平台 runtime：逐平台接入真实运行，优先最容易测试的一条链路。

## 对当前版本规划的影响

`v0.1` 仍然只做初始化和文档，不实现真实运行。

但后续版本应更细：

- `v0.2`：CLI 框架、配置 schema、错误模型、日志事件。
- `v0.3`：领域模型、provider registry、URL 识别、fixture 测试框架。
- `v0.4`：答案策略纯函数迁移，包括概率、严格比例、强制题和随机文本。
- `v0.5`：浏览器抽象和 Playwright Go 封装。
- `v0.6`：HTTP 客户端、缓存、问卷星 HTML parser。
- `v0.7`：腾讯 API parser。
- `v0.8`：Credamo browser parser。
- `v0.9`：runner、browser/http/hybrid 内核预览和性能基准。
- `v1.0`：三平台基础解析、配置生成、运行、提交判定、测试和文档闭环。

## 最终判断

原 Python 项目的最大价值不是目录结构，而是沉淀下来的平台经验：哪些题型容易错、哪些页面要等待、哪些提交可以 HTTP 化、哪些场景必须停下来报告验证或登录要求。

Go 重写应保留这些经验，但用更硬的工程边界重新组织：

- 平台细节关在 provider。
- 运行生命周期关在 runner。
- 浏览器和 HTTP 关在 engine 基础设施。
- 答案策略变成纯函数。
- 配置和错误都强类型。
- CLI 通过事件流观察运行，而不是成为业务核心的一部分。

这样才能既保持兼容性，又真正得到 Go 的轻量、高性能、可测试和可维护优势。
