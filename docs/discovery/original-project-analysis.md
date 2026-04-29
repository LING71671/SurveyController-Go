# 原项目分析

日期：2026-04-29

源项目：本地分析时提供的原 Python 版 SurveyController 工作副本。

## 概览

原 SurveyController 项目是 Python 3.11+ 桌面应用。它使用 PySide6 和 QFluentWidgets 构建界面，使用 Playwright 做浏览器自动化，使用 `httpx` 处理 HTTP 路径，使用 BeautifulSoup 解析 HTML，并使用 OpenPyXL 处理基于表格的反填。

原 README 将该工具描述为面向问卷星、腾讯问卷和 Credamo 见数平台的一站式问卷自动化程序，支持自定义答案分布、随机或指定地区 IP、作答时长控制、信度设置、二维码解析、配置导入导出，以及 AI 生成主观题答案。

## 仓库规模

已观察到的文件数量：

- 301 个 Python 文件。
- 7 个 JavaScript 文件。
- 5 个 Markdown 文件。
- 已有 GitHub 工作流和议题模板。

体量最大、迁移风险最高的 Python 文件包括：

- `tencent/provider/runtime_interactions.py`
- `credamo/provider/runtime.py`
- `software/network/browser/driver.py`
- `wjx/provider/questions/text.py`
- `software/integrations/ai/client.py`
- `wjx/provider/_submission_core.py`
- `credamo/provider/parser.py`
- `wjx/provider/runtime.py`
- `wjx/provider/questions/single.py`
- `tencent/provider/runtime_answerers.py`

这些文件是迁移热点，因为它们混合了平台 DOM 细节、运行态状态、重试逻辑和回退行为。

## 架构

主入口是 `SurveyController.py`，它会调用 `software.app.main.main`。

重要边界：

- `software/providers`：平台识别、平台适配器注册表、标准化问卷定义。
- `software/core/config`：运行配置架构和 JSON 编解码。
- `software/core/task`：执行配置、执行状态、线程进度、代理租约、反填状态。
- `software/core/engine`：工作线程执行循环、浏览器会话服务、提交服务、停止策略、驱动工厂。
- `wjx/provider`、`tencent/provider`、`credamo/provider`：平台专属解析器、运行时、导航、答题器和提交检测。
- `software/network/browser`：Playwright 封装和浏览器生命周期。
- `software/network/proxy`：代理来源、代理池、会话、配额和地区逻辑。
- `software/core/psychometrics`：信度和联合优化器行为。
- `software/core/reverse_fill`：表格反填解析和运行时协调。

当前 Python 代码已经有可用的平台适配器注册模型。Go 版应保留这个边界，而不是逐行平移。

## 平台适配观察

问卷星：

- 优先通过 HTTP 解析，失败后回退到 Playwright。
- 详细 HTML 解析被拆到选项、通用、矩阵和规则辅助模块。
- 已有无头 HTTP 提交路径，可捕获并复用浏览器生成的请求。

腾讯问卷：

- 在可行时使用 API 获取会话、元数据和题目。
- 必要时回退到 Playwright。
- 包含大量选择、文本、下拉、矩阵和星级题交互辅助逻辑。

Credamo 见数平台：

- 高度依赖 Playwright 和 DOM 执行。
- 解析时会通过预填答案处理动态显隐题。
- 包含强制选项和简单算术陷阱题检测。

## 测试资产

已有测试覆盖：

- 配置编解码和运行路径。
- 引擎循环、清理、运行时控制、提交服务。
- 平台通用行为、Credamo 解析器和运行时、问卷星反填、问卷缓存。
- 心理测量方向和联合优化器。
- 题目校验。
- 在线解析器回归。

Go 迁移应复用这些测试背后的行为意图，并逐步把重要用例沉淀为夹具。

## 迁移风险

- Python 使用动态字典表达题目元数据；Go 版需要显式结构体和带版本的配置架构。
- UI 与运行配置在多个路径上耦合；Go CLI 应拆开配置、应用编排和平台运行时。
- 平台 DOM 行为脆弱，必须隔离在平台适配器实现内部。
- Python 中的浏览器生命周期和跨线程清理问题，应在 Go 中转化为明确的 `context` 取消和资源所有权。
- HTTP 快速路径必须由平台适配器声明并经过测试，不应静默替代浏览器模式。

## 对 V0.1 的影响

`v0.1` 不应实现真实平台行为。它应持久化这些结论、初始化仓库，并为平台契约、运行内核模式、配置和运行器编排建立架构边界。
