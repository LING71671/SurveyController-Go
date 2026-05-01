# SurveyController-main 界面设置与功能分析报告

日期：2026-04-30  
对象项目：`B:\SurveyController\SurveyController-main`  
输出位置：`B:\SurveyController\SurveyController-go\docs\discovery`  
分析范围：桌面端 PySide6/QFluentWidgets 应用、问卷解析与执行引擎、运行配置、提供商适配、随机 IP/UA、AI 填空、反填、更新、日志、支持与辅助页面。

## 1. 总览结论

`SurveyController-main` 是一个面向问卷自动作答的 Windows 桌面应用。入口 `SurveyController.py` 启动 PySide6 GUI，核心界面由 `software/ui/shell/main_window.py` 组装，业务执行由 `software/ui/controller/run_controller.py` 及其 parts 模块驱动，最终下沉到 `software/core/engine` 与 `software/providers/*`。

应用的主功能链路是：

1. 用户输入问卷链接或上传/粘贴二维码。
2. 应用识别平台并解析问卷结构。
3. 用户在题目配置向导中设置每题策略、比例、随机文本、AI 填空、矩阵倾向等。
4. 用户配置运行参数，例如目标份数、并发、随机 IP、随机 UA、作答时间、AI 服务、信效度提升。
5. 引擎用 Playwright 启动 Edge/Chrome 浏览器，按线程循环作答、提交、统计进度。
6. 可选启用规则约束、维度分组、反填 Excel 样本、代理 IP、随机 UA、AI 主观题生成。

当前支持的问卷平台：

| 平台 | 解析 | 自动作答 | 主要支持题型 | 备注 |
| --- | --- | --- | --- | --- |
| 问卷星/WJX | 支持 | 支持 | 单选、多选、填空、多空、下拉、矩阵、量表/评分、滑块、排序等 | 功能最完整，支持 headless HTTP 提交、反填、阿里云验证检测 |
| 腾讯问卷/QQ | 支持 | 支持 | 单选、多选、下拉、文本、NPS/星级、矩阵单选/星级 | 支持登录限制检测，部分题型会标记 unsupported |
| Credamo | 支持 | 支持 | 单选、多选、下拉、量表、排序、文本、多空 | 运行侧较轻量，高级能力覆盖不如 WJX/QQ 完整 |

## 2. 项目结构与入口

关键文件与模块：

| 路径 | 作用 |
| --- | --- |
| `SurveyController.py` | 程序入口，导入并执行 `software.app.main.main()` |
| `software/app/main.py` | 初始化日志、fault handler、Qt 应用、主题、主窗口、命令行探针 |
| `software/ui/shell/main_window.py` | 主窗口、导航、页面实例、控制器信号连接 |
| `software/ui/pages/workbench/*` | 概览、运行参数、题目策略、反填、日志等工作台页面 |
| `software/ui/pages/settings/settings.py` | 应用全局设置页面 |
| `software/ui/controller/*` | GUI 与运行引擎之间的控制器层 |
| `software/core/config/*` | RuntimeConfig、JSON 编解码、配置迁移 |
| `software/core/engine/*` | 多线程作答执行循环、浏览器会话、提交策略 |
| `software/providers/*` | WJX、Tencent、Credamo 平台解析与作答适配 |
| `software/network/*` | HTTP、浏览器、代理、随机 IP、状态监测 |
| `software/integrations/ai/*` | AI 填空客户端和运行时 |
| `software/core/reverse_fill/*` | Excel 反填校验、映射和样本队列 |
| `software/update/updater.py` | GitHub Release 更新检测与下载 |

启动阶段的可见行为：

- 设置组织/应用 QSettings 名称为 `SurveyController/Settings`。
- 应用主题色 `#2563EB`，加载 `software/ui/theme.json`。
- Windows 上启用 Mica 背景，设置窗口图标和默认尺寸。
- 恢复窗口置顶状态。
- 预热 `httpx` 连接，开启运行日志与崩溃日志。
- 支持 `--sc-browser-probe` 子命令，用于运行前浏览器可用性探测。

## 3. 主窗口与导航地图

主窗口使用 `MSFluentWindow`，左侧导航分为工作台入口、底部入口和“更多”菜单。

顶部/主导航：

| 导航项 | 页面类 | 是否懒加载 | 功能 |
| --- | --- | --- | --- |
| 概览 | `DashboardPage` | 否 | 链接/二维码解析、配置导入导出、快捷目标/并发/IP、开始/停止/恢复、进度 |
| 运行参数 | `RuntimePage` | 否 | 目标、并发、时间控制、随机 IP、随机 UA、AI、信效度、无头模式 |
| 题目策略 | `QuestionStrategyPage` | 否 | 条件规则、维度分组 |
| 反填 | `ReverseFillPage` | 否 | Excel 样本反填预检和映射，当前为预览功能 |
| 日志 | `LogPage` | 是 | 运行日志查看、导出、报错反馈 |

底部导航：

| 导航项 | 页面类/行为 | 功能 |
| --- | --- | --- |
| 社区 | `CommunityPage` | QQ 群、联系开发者、贡献入口、开源许可 |
| 设置 | `SettingsPage` | 全局偏好、更新源、缓存清理、重启 |
| 更多 | 菜单 | 更新日志、IP 使用记录、捐助、关于、退出 |

隐藏但常驻的数据页：

- `QuestionPage` 不直接展示，主要作为题目配置数据存储。
- 解析问卷后会打开 `QuestionWizardDialog`，实际题目配置在该向导内完成。

## 4. 可设置界面详表

### 4.1 概览页

源码位置：`software/ui/pages/workbench/dashboard/page.py` 与 `dashboard/parts/*`

主要设置与操作：

| 区域 | 控件/操作 | 设置项或行为 |
| --- | --- | --- |
| 问卷入口 | 链接输入框 | 输入问卷链接；也支持拖入/粘贴二维码图片 |
| 问卷入口 | 二维码按钮 | 上传二维码图片并解析链接 |
| 问卷入口 | 自动配置问卷 | 解析问卷结构并进入题目配置向导 |
| 配置操作 | 配置列表 | 打开运行时 `configs` 目录下的配置抽屉 |
| 配置操作 | 载入配置 | 读取 JSON 配置并应用到运行页、题目页、策略页、反填页 |
| 配置操作 | 保存配置 | 保存当前 RuntimeConfig 到 JSON |
| 快捷设置 | 目标份数 | 概览页范围为 1 到 99999 |
| 快捷设置 | 并发数 | 无头模式下最多 16，否则最多 8 |
| 快捷设置 | 随机 IP 开关 | 快速启用/关闭随机 IP，实际授权与额度由控制器处理 |
| 快捷设置 | 运行参数提示卡 | 跳转到运行参数页 |
| 快捷设置 | 服务状态卡 | 打开状态页 `https://status.hungrym0.top/status/surveycontroller` |
| 快捷设置 | 随机 IP 额度卡 | 展示已用/总额度、服务心跳、申请额度按钮 |
| 题目清单 | 新增/编辑/删除/清空 | 管理手动或解析生成的题目配置 |
| 线程进度 | 线程行 | 展示每个 worker 的状态、成功/失败、步骤、累计进度 |
| 运行控制 | 开始执行 | 构造 RuntimeConfig 并调用 `controller.start_run()` |
| 运行控制 | 继续 | 暂停后恢复运行 |
| 运行控制 | 停止 | 请求停止并清理浏览器 |

二维码与链接能力：

- 支持 `png/jpg/jpeg/bmp/gif` 文件。
- 支持剪贴板图片，包括微信截图类自定义 MIME。
- 二维码解码依赖 `zxing-cpp` 与 `Pillow`。
- 支持平台链接校验：WJX、Tencent、Credamo。

概览页运行逻辑：

- 启动前必须存在题目配置。
- 启动时合并运行页配置、题目配置、条件规则、维度分组、反填配置。
- 运行中主进度按 `已完成/目标份数` 计算。
- 初始化浏览器或等待执行时进度条可切换为不确定状态。
- 完成后按钮变为“重新开始”。
- 阿里云验证、随机 IP 额度不足、免费 AI 不稳定等事件会触发提示或快捷反馈。

### 4.2 运行参数页

源码位置：`software/ui/pages/workbench/runtime_panel/main.py`、`cards.py`、`ai.py`

运行参数页是应用最集中、最关键的设置页面。

#### 特性开关

| 设置项 | 控件 | 默认/范围 | 说明 |
| --- | --- | --- | --- |
| 随机 IP | Switch | 默认关闭 | 启用代理 IP，支持默认、限时福利、自定义来源 |
| IP 来源 | ComboBox | 默认来源 | 选项为默认、限时福利、自定义 |
| 自定义代理 API | LineEdit + 检测 | 空 | 支持 URL 中 `{num}` 占位或自动追加 `num` 参数 |
| API 免费试用 | 链接按钮 | 无 | 跳转/打开试用入口 |
| 指定地区 | 可搜索省市 ComboBox | 不限制 | 默认/福利源支持地区范围不同 |
| 随机 UA | Switch | 默认关闭 | 启用随机 User-Agent |
| UA 占比 | RatioSlider | 微信/手机/链接合计 100 | `wechat`、`mobile`、`pc` 三类权重 |

随机 IP 来源差异：

| 来源 | 是否需要授权 | 额度 | 地区支持 | 备注 |
| --- | --- | --- | --- | --- |
| 默认 | 是 | 服务端额度 | 支持省/市筛选 | 答题时长会折算代理分钟和额度成本 |
| 限时福利 | 是 | 福利池额度 | 部分城市 | 仅适配不超过 1 分钟的作答时长 |
| 自定义 | 否 | 本地视为自定义 | 取决于 API | API 返回需能解析出 `ip:port` 或代理字符串 |

#### 作答设置

| 设置项 | 默认/范围 | 说明 |
| --- | --- | --- |
| 目标份数 | 默认 10，范围 1 到 9999 | 与概览页快捷目标同步，但最大值小于概览页 |
| 并发浏览器 | 默认 2，范围 1 到 8/16 | 无头模式开启时可到 16，关闭时限制到 8 |
| 提升问卷信效度 | 默认开启 | 开启心理测量联合优化与维度内一致性控制 |
| 目标 Cronbach's α | 默认 0.9，范围 0.60 到 0.95 | 影响量表/矩阵类维度作答的一致性 |
| 无头模式 | 默认开启 | 开启后浏览器不显示，且可提高并发上限 |

#### 时间控制

| 设置项 | 默认/范围 | 说明 |
| --- | --- | --- |
| 提交间隔 | 0 到 300 秒 | 每份提交成功后的等待区间 |
| 作答时长 | 0 到 600 秒 | 每份最终提交前模拟作答耗时；等值区间会扩展约 ±20% 抖动 |
| 定时模式 | 默认关闭 | 用于抢名额/定时开放问卷，开启后会禁用普通提交间隔和作答时长输入 |

时间与代理联动：

- 作答时长会被换算为代理最小有效分钟数。
- 代理分钟数影响随机 IP 额度消耗。
- 福利 IP 源要求代理分钟数不超过 1。
- 定时模式配置中存在 `timed_mode_interval`，默认 3 秒，但当前运行参数页没有明显暴露独立输入项。

#### AI 填空助手

| 设置项 | 控件/默认 | 说明 |
| --- | --- | --- |
| AI 模式 | 免费/自定义服务商 | 免费模式使用项目服务端，自定义服务商使用兼容 API |
| 服务商 | DeepSeek、通义千问、SiliconFlow、火山方舟、自定义 | 非自定义时自动带出 Base URL 和模型列表 |
| API Key | PasswordLineEdit | 自定义服务商模式必填 |
| Base URL | 自定义时显示 | 支持 OpenAI 兼容 `/chat/completions` 或 `/responses` |
| 模型 ID | ComboBox/LineEdit | 可选预置模型或手输 |
| 系统提示词 | 可展开文本框 | 影响主观题答案风格 |
| 测试按钮 | PushButton | 测试 AI 配置可用性 |

内置服务商默认值：

| 服务商 | Base URL | 默认模型 |
| --- | --- | --- |
| DeepSeek | `https://api.deepseek.com/v1` | `deepseek-chat` |
| 通义千问 | DashScope OpenAI 兼容地址 | `qwen-turbo` |
| SiliconFlow | `https://api.siliconflow.cn/v1` | `deepseek-ai/DeepSeek-V3.2` |
| 火山方舟 | `https://ark.cn-beijing.volces.com/api/v3` | `doubao-seed-1-8-251228` |
| 自定义 | 用户填写 | 用户填写 |

注意事项：

- 免费 AI 与随机 IP 身份/设备 ID 有绑定关系。
- 免费 AI 超时在执行循环中有特殊容错，连续失败达到阈值后会提示“不稳定”。
- 自定义服务商 API Key 会进入运行配置对象，保存配置时需要注意敏感信息落盘风险。

### 4.3 题目配置向导

源码位置：`software/ui/dialogs/question_wizard/*`、`software/ui/pages/workbench/question_editor/*`

题目配置向导在解析问卷后打开，是所有题目级策略的主要入口。

支持题型：

| 内部类型 | 中文 | 主要配置能力 |
| --- | --- | --- |
| `single` | 单选 | 选项权重、倾向预设、其他项填写、嵌入下拉 |
| `multiple` | 多选 | 每个选项独立命中概率、最小/最大选择数、其他项填写 |
| `text` | 填空 | 答案列表、随机姓名/手机号/身份证/整数、AI |
| `multi_text` | 多空填空 | 每个空独立配置答案来源、随机类型、整数范围、AI |
| `dropdown` | 下拉 | 选项权重、其他项填写、嵌入下拉 |
| `matrix` | 矩阵 | 每行选项权重、行级倾向预设、矩阵量表/滑块处理 |
| `scale` | 量表 | 选项权重、倾向预设、信效度优化 |
| `score` | 评分 | 选项权重、倾向预设、信效度优化 |
| `slider` | 滑块 | 随机或目标值，按解析范围处理 |
| `order` | 排序 | 随机排序，自动识别 Top-N 限制 |

通用能力：

- 题目搜索：按题干、选项、矩阵行、附加下拉搜索。
- 题目导航：左侧/侧边快速跳转。
- 批量倾向预设：偏左、居中、偏右、自定义。
- 跳题/显隐逻辑标识：展示跳转和显示条件警告。
- 自动校验：避免所有权重为 0、整数范围缺失、嵌入下拉无有效权重。
- 取消时恢复进入向导前的题目快照。

题型细节：

| 题型 | 设置细节 |
| --- | --- |
| 单选/下拉/量表/评分 | 每个选项 0 到 100 权重，界面显示预计占比 |
| 多选 | 每个选项 0 到 100 独立命中概率，不要求总和为 100 |
| 矩阵 | 每行独立权重组，可按行设置倾向 |
| 滑块 | 可设置目标值，范围来自平台解析的 min/max，默认 0 到 100 |
| 排序 | 当前不提供比例配置，按随机排序执行 |
| 填空 | 可使用固定答案列表，也可切换随机姓名、随机手机号、随机身份证号、随机整数或 AI |
| 多空填空 | 每个空可以有独立的随机/AI/列表配置 |
| 其他项填写 | 单选、多选、下拉的可填写选项支持填写文本、随机值或 AI |
| 嵌入下拉 | 对挂载在选项上的二级下拉单独设置权重 |

手动新增题目对话框：

- 支持选择题型、策略、选项数、矩阵行数、答案数量。
- 支持完全随机或自定义配比。
- 文本题可直接启用 AI。
- 矩阵题可按行配比。
- 多选题的权重语义是独立命中率。

### 4.4 题目策略页

源码位置：`software/ui/pages/workbench/question_strategy/page.py` 及 parts 模块

题目策略页分为“条件规则”和“维度分组”两个标签页。

#### 条件规则

用途：按前题答案约束后题选项，规则越靠后优先级越高。

可设置项：

| 设置 | 说明 |
| --- | --- |
| 条件题目 | 作为触发条件的前置题 |
| 条件类型 | 选择了以下选项、未选择以下选项 |
| 条件选项 | 前置题命中的选项 |
| 目标题目 | 被约束的后置题 |
| 动作类型 | 一定选择以下选项、一定不选择以下选项 |
| 目标选项 | 后置题被强制或排除的选项 |

规则限制：

- 条件题必须早于目标题。
- 条件题与目标题不能相同。
- 必须选择条件选项和目标选项。
- 规则支持单选、多选、量表/评分、矩阵等可离散选择题型。
- 矩阵题需要选择具体行。
- 加载旧配置或解析变化后会清理不再适用的规则。

#### 维度分组

用途：为量表/评分/矩阵题建立心理测量维度，配合信效度优化。

可设置项：

| 设置 | 说明 |
| --- | --- |
| 新增维度 | 手动输入名称或选择预设 |
| 预设维度 | 满意度、信任感、使用意愿、感知价值、服务质量、产品质量 |
| 添加题目 | 从支持题型中选择题目加入维度 |
| 拖拽分组 | 题目可在未分组和自定义分组之间拖拽 |
| 重命名/删除 | 自定义维度可改名或删除，删除后题目回到未分组 |
| 倾向预设 | 每个维度内题目保留倾向标识 |

支持题型主要是 `scale`、`score`、`matrix`。

### 4.5 反填配置页

源码位置：`software/ui/pages/workbench/reverse_fill/page.py`、`software/core/reverse_fill/*`

反填是预览功能，用 Excel 样本数据按行级答卷回填问卷。当前运行侧主要支持 WJX。

界面设置：

| 区域 | 控件/操作 | 说明 |
| --- | --- | --- |
| 问卷结构 | 链接输入/二维码/解析按钮 | 解析问卷结构，用于建立 Excel 列与题目的映射 |
| 数据源 | 启用自动反填 | 开启后受支持题型按 Excel 样本覆盖常规配置 |
| 数据源 | Excel 路径 | 选择 `.xlsx` 文件 |
| 状态 | 格式/预检摘要 | 展示识别格式、样本数量、可反填/回退/阻塞情况 |
| 映射预览 | 表格 | 展示题号、解析题型、支持判定、关联列、执行策略 |
| 异常与回退 | 表格 | 展示阻塞原因、严重级别、推荐处理 |
| 一键定位 | 按钮 | 有问题时可打开题目配置向导定位缺失依赖 |

反填格式：

| 格式键 | 含义 |
| --- | --- |
| `auto` | 自动识别 |
| `wjx_sequence` | 问卷星按序号 |
| `wjx_score` | 问卷星按分数 |
| `wjx_text` | 问卷星按文本 |

支持与阻塞：

- 支持单选、下拉、量表/评分、文本、多空、矩阵等。
- 不支持或会阻塞的情况包括定位题、其他项填写、嵌入下拉、多选复合值、排序箭头值、滑块等。
- Excel 读取使用 `openpyxl`，首行表头按 `题号 + 标题` 识别。
- 样本不足、列缺失、列歧义、值无法匹配都会出现在预检问题表。

隐藏/未完全暴露点：

- `RuntimeConfig` 中存在 `reverse_fill_format` 和 `reverse_fill_start_row`。
- 页面代码中也有 `_FORMAT_CHOICES` 与 `_start_row_value`。
- 当前可见界面没有明显暴露格式选择器和起始行输入，主要通过配置加载或内部默认值生效。

### 4.6 日志页

源码位置：`software/ui/pages/workbench/log_panel/page.py`

功能：

- 实时显示运行日志。
- 自动刷新间隔约 500 ms。
- 最大显示块数约 2000。
- 支持日志高亮。
- 用户滚动时暂停自动跟随，回到底部后恢复。
- 可导出到文件。
- 可直接打开报错反馈表单。
- 启动时加载上一会话 `last_session.log`。

### 4.7 设置页

源码位置：`software/ui/pages/settings/settings.py`

全局设置项：

| 分组 | 设置项 | QSettings key | 默认值 | 行为 |
| --- | --- | --- | --- | --- |
| 界面外观 | 显示选中导航名称 | `navigation_selected_text_visible` | true | 控制导航选中项是否显示文字标签 |
| 界面外观 | 窗口置顶 | `window_topmost` | false | 设置 `WindowStaysOnTopHint` 并重新显示窗口 |
| 行为设置 | 关闭前询问是否保存 | `ask_save_on_close` | true | 关闭窗口时弹出保存当前配置提示 |
| 行为设置 | 执行期间阻止自动休眠 | `prevent_sleep_during_run` | true | 运行时调用 Windows 电源 API 阻止自动休眠 |
| 软件更新 | 启动时检查更新 | `auto_check_update` | true | 主窗口启动后后台检查 GitHub Release |
| 软件更新 | 下载源 | `download_source` | official | 控制更新安装包下载源 |

系统工具：

| 工具 | 行为 |
| --- | --- |
| 重启程序 | 确认后用当前 Python/EXE 参数重新启动，并跳过关闭保存提示 |
| 恢复默认设置 | 清除若干 QSettings key，并将开关恢复默认 |
| 删除问卷解析缓存 | 调用 `clear_survey_parse_cache()` 删除本地解析缓存 |

下载源选项来自 `software/app/config.py`：

| key | 标签 | 行为 |
| --- | --- | --- |
| `official` | 官方服务器（推荐） | 可配置直连下载地址 |
| `github` | GitHub 原始地址 | 使用 Release 原始下载地址 |
| `ghfast` | ghfast.top 镜像 | 给 GitHub URL 加镜像前缀 |
| `ghproxy` | ghproxy.net 镜像 | 给 GitHub URL 加镜像前缀 |

观察点：

- “恢复默认设置”清除了导航、置顶、关闭保存、阻止休眠、自动更新，但没有移除 `download_source`。因此下载源下拉可能保留用户选择。

### 4.8 社区、关于、更新日志、IP 使用记录、捐助、支持

#### 社区页

源码位置：`software/ui/pages/more/community.py`

功能：

- 展示 QQ 群二维码 `assets/community_qr.jpg`。
- 打开联系开发者表单。
- 打开 GitHub 仓库。
- 展示 AGPL-3.0 许可入口。

#### 关于页

源码位置：`software/ui/pages/more/about.py`

功能：

- 展示 Logo、应用名称、免责声明。
- 展示当前版本，当前源码版本为 `3.1.0`。
- 检查更新。
- 打开 GitHub、官方文档 `https://surveydoc.hungrym0.top/`。
- 展示许可证、贡献者、服务条款/隐私相关对话框。

#### 更新日志页

源码位置：`software/ui/pages/more/changelog.py`

功能：

- 后台调用 `UpdateManager.get_all_releases()` 获取 GitHub Releases。
- 列表展示版本、日期、摘要。
- 点击进入详情，渲染发布说明文本。

#### IP 使用记录页

源码位置：`software/ui/pages/more/ip_usage.py`

功能：

- 展示随机 IP 剩余额度。
- 用 QtCharts 展示每日提取 IP 数。
- 后台读取 `software.io.reports.get_usage_summary()`。
- 首次打开且已认证随机 IP 时可能触发彩蛋额度领取和动效。

#### 捐助页

源码位置：`software/ui/pages/more/donate.py`

功能：

- 展示微信赞赏码 `assets/WeDonate.png`。
- 展示支付宝二维码 `assets/AliDonate.jpg`。
- 小屏下自适应纵向布局。

#### 支持/联系表单

源码位置：`software/ui/pages/support/*`

可设置项与功能：

| 表单项 | 说明 |
| --- | --- |
| 消息类型 | 报错反馈、额度申请、新功能建议、纯聊天 |
| 联系邮箱 | 额度申请时要求邮箱验证 |
| 邮箱验证码 | 额度申请时显示 6 位验证码输入 |
| 需求额度 | 0.5 步进，最高约 19999 |
| 支付金额 | 预设 8.88、11.45、20.26、50、78.91、114.51，也可编辑 |
| 紧急程度 | 低、中（本月内）、高（本周内）、紧急（两天内） |
| 支付方式 | 微信/支付宝 |
| 已完成支付确认 | 额度申请前需确认，且随机 IP 用户 ID 需有效 |
| 消息正文 | 文本描述 |
| 图片附件 | 非额度申请最多 3 张，每张最多 10MB，可文件添加或粘贴 |
| 自动附件 | 报错反馈可附当前配置、日志、崩溃日志 |

额度申请金额门槛：

| 需求额度 | 最低金额 |
| --- | --- |
| 1500 及以上 | 11.45 |
| 2000 及以上 | 20.26 |
| 3500 及以上 | 50 |
| 8000 及以上 | 78.91 |
| 13000 及以上 | 114.51 |

## 5. 配置模型与持久化

### 5.1 RuntimeConfig 字段

源码位置：`software/core/config/schema.py`

主要字段：

| 类别 | 字段 |
| --- | --- |
| 问卷元信息 | `url`、`survey_title`、`survey_provider` |
| 执行规模 | `target`、`threads`、`browser_preference` |
| 时间 | `submit_interval`、`answer_duration`、`timed_mode_enabled`、`timed_mode_interval` |
| 随机 IP | `random_ip_enabled`、`proxy_source`、`custom_proxy_api`、`proxy_area_code` |
| 随机 UA | `random_ua_enabled`、`random_ua_keys`、`random_ua_ratios` |
| 控制策略 | `fail_stop_enabled`、`pause_on_aliyun_captcha`、`reliability_mode_enabled`、`psycho_target_alpha`、`headless_mode` |
| AI | `ai_mode`、`ai_provider`、`ai_api_key`、`ai_base_url`、`ai_api_protocol`、`ai_model`、`ai_system_prompt` |
| 反填 | `reverse_fill_enabled`、`reverse_fill_source_path`、`reverse_fill_format`、`reverse_fill_start_row` |
| 题目与策略 | `answer_rules`、`dimension_groups`、`question_entries`、`questions_info` |

### 5.2 配置文件

源码位置：`software/io/config/store.py`、`software/core/config/codec.py`

行为：

- 保存格式为 JSON，UTF-8，缩进 2。
- 默认运行时配置路径为 runtime 目录下的 `config.json`。
- 用户配置列表位于 runtime `configs/`。
- 默认配置文件名由问卷标题生成。
- 加载时支持剥离 JSON 行注释和块注释。
- 当前配置 schema 版本为 5。
- 支持从 v3/v4 迁移。
- 对旧字段 `random_proxy_api`、`ai_enabled` 有拒绝或迁移约束。
- 随机 UA 权重会归一化为合计 100。
- 随机 IP 是否启用会结合会话/额度状态归一化。

### 5.3 QSettings 与安全存储

QSettings：

- 组织/应用：`SurveyController/Settings`
- 主要 key：
  - `navigation_selected_text_visible`
  - `window_topmost`
  - `ask_save_on_close`
  - `prevent_sleep_during_run`
  - `auto_check_update`
  - `download_source`
  - `random_ip_auth/*`
  - 社区提示、彩蛋播放等内部状态

安全存储：

- `software/system/secure_store.py` 在 Windows 上使用 DPAPI。
- 后端存储位置为 HKCU `Software\SurveyController\SecureStore`。
- 随机 IP 设备 ID 等敏感身份信息会走安全存储。

运行时目录：

- `logs/`：运行日志、`last_session.log`、`fatal_crash.log`
- `configs/`：用户配置与问卷解析缓存
- 下载更新的 EXE 也会放入 runtime 目录。

## 6. 自动作答执行功能

### 6.1 控制器层

源码位置：`software/ui/controller/run_controller.py` 与 `run_controller_parts/*`

`RunController` 负责 GUI 与执行引擎的边界：

- 异步解析问卷。
- 构建默认题目配置。
- 校验题目配置和 RuntimeConfig。
- 构建 `ExecutionConfig`。
- 初始化反填规格。
- 同步运行 UI 状态。
- 管理启动、停止、恢复、关闭。
- 汇报线程进度、状态文本、错误、暂停状态。
- 运行前做 headless 多并发浏览器探针。
- 运行时按设置阻止系统休眠。

关键运行信号：

| 信号 | 含义 |
| --- | --- |
| `surveyParsed` | 问卷结构解析成功 |
| `surveyParseFailed` | 问卷解析失败 |
| `runStateChanged` | 运行状态变化 |
| `runFailed` | 运行失败 |
| `statusUpdated` | 主状态文本更新 |
| `threadProgressUpdated` | 单线程进度更新 |
| `pauseStateChanged` | 暂停/恢复 |
| `cleanupFinished` | 浏览器清理完成 |
| `quickBugReportSuggested` | 建议快速报错反馈 |
| `freeAiUnstableSuggested` | 免费 AI 不稳定提示 |

### 6.2 执行引擎

源码位置：`software/core/engine/execution_loop.py`、`browser_session.py`、`submission.py`

核心执行流程：

1. 每个 worker 创建或复用浏览器会话。
2. 若启用随机 IP，为当前会话申请代理。
3. 若启用随机 UA，为当前会话选取 User-Agent。
4. 打开问卷或在定时模式下轮询等待问卷开放。
5. 检测设备限制、验证码、问卷结束等状态。
6. 获取心理测量联合样本或反填样本。
7. 调用提供商适配器填答。
8. 提交后检测成功页、验证页、完成标识。
9. 成功则计数、提交分布统计、释放样本。
10. 失败则按失败策略重试、换代理、停止或回退。

停止策略：

- 默认连续失败阈值为 5。
- `fail_stop_enabled` 在配置中存在，但运行参数页当前固定写入 true。
- 成功后重置连续失败计数。
- 用户停止时会设置停止事件并清理浏览器。
- 反填样本失败一次可回队列，二次失败丢弃。

浏览器：

- 使用 Playwright sync API。
- 优先 Edge，再 Chrome。
- `browser_preference` 在配置中存在，但运行参数页当前写入空列表，实际走默认优先级。
- 无头模式默认开启。
- 有头窗口默认约 550x650。
- 跨线程清理时会尽量关闭 Playwright，必要时 fallback 到 `taskkill /PID /T /F`。

## 7. 提供商能力

### 7.1 WJX 问卷星

源码位置：`software/providers/wjx/*`

解析：

- HTTP 直接抓取优先，失败时用 headless Playwright 临时渲染。
- 解析 `divQuestion`、分页、题号、题干、选项、行、显示条件、跳题规则。
- 支持题型码：
  - 3 单选
  - 4 多选
  - 5 量表/评分
  - 6/9 矩阵
  - 7 下拉
  - 8 滑块
  - 11 排序
  - 1/2 填空/多空
- 能识别可填写选项、嵌入下拉、定位题、描述题、滑块矩阵、多选限制、强制选项等。

作答：

- 按页面逐题处理。
- 支持单选、多选、矩阵、下拉、量表、评分、滑块、排序、文本、多空。
- 单选/多选支持其他项填写与规则约束。
- 文本支持固定答案、随机身份类值、随机整数、AI。
- 矩阵支持行级权重、倾向、信效度优化。
- headless WJX 支持拦截 `joinnew/processjq.ashx`，用 `httpx` 携带浏览器 cookie/header 直接提交。
- 检测 Aliyun captcha 和设备额度限制。

### 7.2 腾讯问卷

源码位置：`software/providers/qq/*`

解析：

- 从 `/s2/...` URL 提取 survey id/hash。
- 优先调用腾讯问卷 API 获取元信息与题目。
- 若 API 失败可用 headless Playwright 在页面上下文中拉取。
- 支持语言/locale：简中、繁中、英文等。
- 可检测登录限制并给出友好错误。

题型映射：

| 腾讯题型 | 内部题型 |
| --- | --- |
| radio | single |
| checkbox | multiple |
| select | dropdown |
| text/textarea | text |
| nps/star | scale/score |
| matrix_radio/matrix_star | matrix |

作答：

- 按页面分组。
- 支持单选、多选、下拉、文本、评分/星级、矩阵。
- 支持 persona boost、一致性规则、严格比例、分布修正、AI 文本。
- 检测腾讯安全验证和完成页。
- unsupported 题型会阻止启动。

### 7.3 Credamo

源码位置：`software/providers/credamo/*`

解析：

- 使用 headless Playwright 渲染页面。
- 解析 `.answer-page .question`。
- 通过预作答当前页发现动态显隐后的后续问题。
- 推断单选、多选、下拉、量表、排序、文本、多空等题型。
- 可识别强制选择说明、简单算术陷阱、强制文本提示。

作答：

- 等待题目根节点出现。
- 按题号配置映射作答。
- 支持单选、下拉、量表、多选、排序、文本。
- 可处理强制选择与简单算术题。
- 检测完成、验证码/验证、名额/次数限制等关键词。

限制：

- 高级能力覆盖比 WJX/QQ 少。
- AI、反填、严格比例、信效度等能力在 Credamo 运行侧不是完整主路径。

## 8. 随机 IP、随机 UA 与网络能力

### 8.1 随机 IP

源码位置：`software/network/proxy/*`、`software/core/proxy/*`、`software/ui/controller/run_controller_parts/runtime_random_ip.py`

能力：

- 支持默认官方代理、限时福利代理、自定义代理 API。
- 官方/福利源需要随机 IP 身份和额度。
- 自定义源不需要服务端授权。
- 支持地区筛选。
- 支持代理批量获取、去重、过期剔除、线程占用管理。
- 支持浏览器代理与 headless WJX HTTP 提交代理分离。
- 代理连接错误会丢弃当前代理并重试。

额度与成本：

- 作答时长换算代理分钟数。
- 代理分钟数换算额度成本。
- 界面会在概览页展示消耗预警和低额度提示。
- 额度耗尽时可自动关闭随机 IP 或提示申请。

自定义代理 API：

- 支持 JSON 中递归提取代理字符串。
- 支持 `ip:port`、带协议代理、对象字段。
- 如果 URL 包含 `{num}` 则替换，否则尝试追加 `num` 参数。
- 会识别白名单、余额不足、API key、授权、过期、账户禁用等致命错误。

### 8.2 随机 UA

源码位置：`software/app/config.py`、`software/core/engine/browser_session.py`

可选 UA 类别：

| UI 维度 | 内部 key | 说明 |
| --- | --- | --- |
| 微信访问占比 | `wechat` | 映射到微信 Android UA |
| 手机访问占比 | `mobile` | 映射到移动 Android UA |
| 链接访问占比 | `pc` | 映射到 PC Web UA |

运行时根据三类权重随机选择设备类别，再选择具体 UA。

## 9. AI 主观题能力

源码位置：`software/integrations/ai/client.py`、`software/core/ai/runtime.py`

能力：

- 填空题、多空填空、其他项填写可启用 AI。
- AI prompt 会包含题干、题型、persona、最近答题上下文。
- 多空题会要求返回多个答案并校验数量。
- 自定义服务商支持 Chat Completions 和 Responses 两种协议。
- 自定义 endpoint 可自动识别 `/chat/completions` 或 `/responses`。
- 对 429、502、503、504、超时等会有限重试。

免费模式：

- 使用 `AI_FREE_ENDPOINT`。
- 请求包含用户 ID、设备 ID、题型、题干、多空数量、系统提示词。
- 免费模式不稳定时有专门的连续失败提示。

## 10. 分布、规则、人格与信效度

源码位置：`software/core/questions/*`、`software/core/persona/*`、`software/core/psychometrics/*`

主要能力：

| 能力 | 说明 |
| --- | --- |
| 严格比例 | 自定义权重可转为目标分布，执行中根据已提交统计动态修正 |
| 条件规则 | 以前题答案约束后题选项，支持必须选择/必须不选 |
| Persona | 每份问卷生成性别、年龄、教育、职业、收入、婚育、满意度倾向 |
| 选项 boost | 与 persona 匹配的选项可提高权重 |
| 上下文记录 | 保存最近答案供规则、AI prompt、连贯性使用 |
| 维度倾向 | 维度内建立基础倾向，避免同维度答案完全随机 |
| 心理测量优化 | 对量表/矩阵类维度生成联合样本，逼近目标 Cronbach's α |
| 反向题推断 | 尝试识别反向题并调整一致性方向 |

信效度模式：

- 默认开启。
- 目标 alpha 默认 0.9。
- 若题目有维度分组，按维度建立计划。
- 若没有显式维度，运行侧可能为可支持题型归入全局信效度维度。
- 联合优化会在保留配置分布的前提下寻找更接近目标 alpha 的样本序列。

## 11. 问卷解析缓存、导入导出与关闭保存

### 11.1 问卷解析缓存

源码位置：`software/providers/survey_cache.py`

行为：

- 缓存在 runtime `configs/survey_cache`。
- 缓存版本为 1。
- TTL 约 2 小时。
- WJX/QQ 会使用远程指纹辅助判断缓存是否仍有效。
- Credamo 主要依赖 TTL。
- 设置页可清除缓存。

### 11.2 配置导入导出

行为：

- 保存当前 URL、运行参数、题目条目、规则、维度、反填、AI 等。
- 支持配置抽屉选择已有配置。
- 加载配置后会同步到概览、运行参数、题目策略、反填页面。
- 关闭窗口时若 `ask_save_on_close` 开启，会询问是否保存当前配置。

### 11.3 关闭与清理

关闭窗口时：

- 可提示保存当前配置到 runtime `configs/*.json`。
- 保存当前会话日志到 `logs/last_session.log`。
- 停止控制器、停止日志刷新、停止支持页轮询、停止更新下载线程相关引用。
- 清理浏览器进程。

## 12. 软件更新功能

源码位置：`software/update/updater.py`、`software/ui/shell/main_window_parts/update.py`

功能：

- 从 GitHub Releases latest 检查新版本。
- 当前版本来自 `software/app/version.py`，源码中为 `3.1.0`。
- 本地版本等于远端：状态 latest。
- 本地版本高于远端：状态 preview。
- 远端版本更高且 Release 中存在 `.exe`：状态 outdated。
- 支持获取全部 releases 用于更新日志页。
- 支持下载进度、速度、取消、下载完成后启动新版本。
- 下载源失败时可自动切换下一个下载源。
- 下载成功后清理 runtime 目录下旧版 SurveyController exe。
- frozen EXE 场景下可调度退出后删除旧正在运行文件。

## 13. 隐含配置与迁移注意点

以下配置在模型或运行链路中存在，但 UI 暴露程度有限或固定写入：

| 配置 | 当前情况 | 迁移建议 |
| --- | --- | --- |
| `fail_stop_enabled` | 配置存在，运行页更新时固定为 true | Go 版本如需可控，可考虑暴露“连续失败后停止” |
| `pause_on_aliyun_captcha` | 配置存在，运行页固定 true | 可做成高级开关 |
| `browser_preference` | 配置存在，运行页写空列表，执行侧走 Edge/Chrome 默认顺序 | 可在高级设置暴露浏览器优先级 |
| `timed_mode_interval` | 配置存在，默认 3 秒 | 定时模式页可补充刷新间隔输入 |
| `random_ua_keys` | 旧/兼容字段存在 | 新 UI 主要使用三类比例 |
| `reverse_fill_format` | 配置存在 | 当前反填页未明显暴露格式选择 |
| `reverse_fill_start_row` | 配置存在 | 当前反填页未明显暴露起始行输入 |
| `download_source` | 可设置 | 恢复默认设置未清除此 key |

其他观察：

- 概览页目标份数最大 99999，运行参数页最大 9999，存在 UI 上限不一致。
- WJX 能力最完整，Tencent 次之，Credamo 更偏基础题型自动化。
- 反填页面可校验多平台链接，但 `_context_ready()` 当前要求 WJX 上下文，实际 V1 应视为 WJX-only。
- 自定义 AI API Key 保存到配置文件时有敏感信息风险。
- 更新模块会写入 runtime 目录并删除旧 exe，Go 迁移时要重新设计权限、签名和原子替换策略。

## 14. 对 Go 版本复刻的功能拆分建议

建议按以下边界迁移或验收：

1. 基础壳层：主导航、主题、设置、日志、更新、关于、社区。
2. 配置模型：兼容 schema v5，先实现加载/保存/迁移，再接 UI。
3. 问卷解析：WJX 优先，其次 Tencent，Credamo 单独适配。
4. 题目配置：先覆盖题型权重、文本随机/AI、矩阵，再做嵌入下拉和其他项。
5. 执行引擎：线程/worker、浏览器会话、停止策略、进度信号。
6. WJX 提交：优先复刻 headless HTTP 提交和验证检测，因为这是性能与稳定性的关键。
7. 随机 IP/UA：先完成来源、额度、地区、代理池，再接成本提示。
8. 规则与分布：先实现条件规则和严格比例，再实现 persona 与信效度联合优化。
9. 反填：先实现 WJX Excel 预检和单选/量表/文本回填，再扩展矩阵、多空。
10. 支持与反馈：按现有表单字段复刻即可，但建议重新审视邮箱验证码和支付确认逻辑。

## 15. 源码核对清单

本报告主要依据以下源码区域：

- `SurveyController.py`
- `README.md`
- `requirements.txt`
- `software/app/*`
- `software/ui/shell/*`
- `software/ui/pages/workbench/*`
- `software/ui/pages/settings/settings.py`
- `software/ui/pages/more/*`
- `software/ui/pages/support/*`
- `software/ui/dialogs/question_wizard/*`
- `software/ui/controller/*`
- `software/core/config/*`
- `software/core/engine/*`
- `software/core/questions/*`
- `software/core/psychometrics/*`
- `software/core/reverse_fill/*`
- `software/providers/wjx/*`
- `software/providers/qq/*`
- `software/providers/credamo/*`
- `software/network/proxy/*`
- `software/integrations/ai/*`
- `software/update/updater.py`

## 16. 快速功能索引

| 功能 | 是否存在 | 入口 |
| --- | --- | --- |
| 问卷链接解析 | 是 | 概览、反填 |
| 二维码解析 | 是 | 概览、反填 |
| 配置导入导出 | 是 | 概览 |
| 配置列表 | 是 | 概览 |
| 手动题目配置 | 是 | 概览题目清单、题目向导 |
| 解析后题目向导 | 是 | 自动配置问卷 |
| 每题概率/权重 | 是 | 题目向导 |
| 文本随机生成 | 是 | 题目向导 |
| AI 填空 | 是 | 运行参数、题目向导 |
| 条件规则 | 是 | 题目策略 |
| 维度分组 | 是 | 题目策略 |
| 信效度提升 | 是 | 运行参数 |
| 随机 IP | 是 | 概览、运行参数 |
| 自定义代理 API | 是 | 运行参数 |
| 指定代理地区 | 是 | 运行参数 |
| 随机 UA | 是 | 运行参数 |
| 定时模式 | 是 | 运行参数 |
| 无头模式 | 是 | 运行参数 |
| 多线程执行 | 是 | 概览、运行参数 |
| 暂停/恢复 | 是 | 运行控制、验证码场景 |
| 日志查看/导出 | 是 | 日志 |
| 报错反馈 | 是 | 日志、支持 |
| 额度申请 | 是 | 概览、支持 |
| IP 使用记录 | 是 | 更多 |
| 自动更新 | 是 | 设置、关于、启动 |
| 更新日志 | 是 | 更多 |
| 问卷解析缓存清理 | 是 | 设置 |
| Excel 反填 | 是，预览 | 反填 |
| 社区/捐助/关于 | 是 | 底部导航/更多 |

