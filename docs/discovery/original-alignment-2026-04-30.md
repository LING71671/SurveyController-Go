# 原 Python 项目对齐复盘

日期：2026-04-30

来源：本地 `B:\SurveyController\SurveyController-main` 工作副本。

## 本次对齐目标

本次对齐发生在 Go 版完成 v0.7 浏览器基础、v0.8 Phase 1 腾讯 API parser 原型之后。目标不是重新做一次完整盘点，而是校准下一轮开发：

- 当前 Go 版哪些能力已经和原项目方向一致。
- 哪些地方只是原型，不能误认为已迁移完成。
- v0.8、v0.9、v1.0 的后续 issue 应优先补哪些缺口。

## 已对齐的 Go 能力

| 原项目能力 | Go 版当前状态 | 结论 |
| --- | --- | --- |
| provider registry 和 URL 分发 | 已有 `internal/provider` registry、capability、URL matcher | 方向正确，但 completion/submission detector 仍待拆进 provider 契约 |
| 强类型问卷模型 | 已有 `internal/domain.SurveyDefinition`、题型、provider id | 已优于原项目动态 dict，后续应继续保留 `ProviderRaw` 承接平台字段 |
| HTTP client 和缓存基础 | 已有 `internal/httpclient`、`internal/parsecache` | 与原项目 `httpx` + survey cache 思路一致 |
| 问卷星 HTML parser 原型 | 已有 `internal/provider/wjx.ParseHTML` | 仍缺页码、跳题、显示条件、多选上下限、附加填空等完整字段 |
| 浏览器抽象 | 已有 `internal/browser` 接口、fake、`FetchHTML` | 与原项目“不要泄漏 Playwright 细节到 provider 外”一致 |
| doctor browser | 已有 `surveyctl doctor browser` 预检入口 | 当前只做静态预检，尚未等价原项目子进程真实启动 probe |
| 腾讯 API parser 原型 | 已有 `internal/provider/tencent.ParseAPI` | 只完成 JSON 映射原型，未完成真实 session/meta/questions API 流程 |

## 关键差距

### 腾讯问卷

原项目 `tencent/provider/parser.py` 的真实解析链路比当前 Go 原型更完整：

- 从 `/s2/{survey_id}/{hash}/` URL 提取 survey id 和 hash。
- 先调用 `session`，再按 locale 尝试 `meta` 与 `questions`。
- 请求头需要 `Origin`、`Referer`、JSON accept。
- 登录要求不只来自状态码，还来自跳转 URL、location header、响应 body 和 payload 内 token。
- 题型字段包含 `type`、`options`、`sub_titles`、`star_num`、`star_begin_num`、`min_length`、`max_length`。
- 选项文本需要去掉 `{fillblank-*}` token，并识别带附加填空的选项。
- API 失败后可回退浏览器，在页面内 `fetch` API。

Go 后续不能把当前 `ParseAPI` 视为完整腾讯 parser。下一步应新增：

- URL id/hash extractor。
- 可注入 HTTP client 的 `FetchAPI` 或 `ParseFromClient`。
- `session/meta/questions` 三段 API fixture。
- locale fallback table tests。
- 登录要求递归检测函数。
- 腾讯 provider raw 字段：`provider_page_id`、`provider_type`、`multi_min_limit`、`multi_max_limit`、`fillable_options`。

### Credamo

原项目 `credamo/provider/parser.py` 的复杂度主要来自动态页面：

- 通过 JS 从 `.answer-page .question` 抽取当前可见题目。
- 预填当前题目以触发动态显隐题，再收集新增题。
- 多页解析最多 20 页，有 next/submit 导航判断。
- DOM id 可能复用，去重 key 必须包含 page、id、num、title。
- 支持 forced select、算术陷阱题、forced text。
- parser 会复用 runtime 的答题器做“预填探索”。

Go 版 Credamo 应分两步，不要直接上完整 browser parser：

1. 先做 DOM snapshot parser：输入由浏览器 JS 抽出的 JSON，输出 `SurveyDefinition`。
2. 再做 browser explorer：负责打开页面、执行 JS、预填触发显隐、翻页和去重。

这样能先迁移原项目 `CI/unit_tests/providers/test_credamo_parser.py` 中的纯函数测试，降低 Playwright 集成风险。

### 浏览器 probe

原项目 `software/app/browser_probe.py` 用子进程启动真实 Playwright，避免主进程和 UI 线程被底层环境错误拖垮。Go 当前 `doctor browser` 只检查 OS、PATH、环境变量和代理占位。

Go 后续应保留两级 doctor：

- 静态预检：当前已有，快速、无副作用。
- 真实启动 probe：独立超时、可取消、返回结构化 `ok/browser/error_kind/message/elapsed_ms`。

### 提交判定与停止策略

原项目提交后由 provider 分发：

- 完成页识别。
- 提交后验证/风控识别。
- 校验文案提取。
- 设备填写次数上限识别。
- headless HTTP 成功信号短路。

Go 当前还没有提交判定契约。v0.9 进入运行时预览前，需要先补 provider detector 接口，否则 runner 容易重新长成大分支。

建议新增接口方向：

```go
type CompletionDetector interface {
    IsCompletion(ctx context.Context, page browser.Page) (bool, error)
}

type SubmissionDetector interface {
    DetectSubmissionState(ctx context.Context, page browser.Page) (SubmissionState, error)
}
```

## 对后续路线的调整

1. v0.8 剩余工作优先做 Credamo DOM snapshot parser，而不是直接真实浏览器探索。
2. 腾讯 parser 下一轮应补真实 API client skeleton 和 URL id/hash 提取，不急着接 runtime。
3. v0.9 开始前必须补 `SubmissionResult`、completion/verification/device quota detector 契约。
4. 真实 Playwright Go 集成放在 doctor probe 和 browser pool 实现中，以 mock/fake 测试为主，集成测试按环境跳过。
5. 继续保持合规边界：登录、验证、风控、设备次数上限都必须停止并报告，不做绕过。

## 下一批建议 issue

| 顺序 | 建议 issue | 理由 |
| --- | --- | --- |
| 1 | Credamo DOM snapshot parser 原型 | 迁移纯解析和陷阱题识别，不依赖真实浏览器 |
| 2 | 腾讯 API client skeleton | 把 URL id/hash、session/meta/questions、locale fallback 补齐 |
| 3 | Provider submission detector 契约 | v0.9 运行预览前先定义提交判定边界 |
| 4 | Browser real probe skeleton | 将 `doctor browser` 从静态检查扩展到真实启动检查 |
| 5 | 三平台 parser fixture 对齐 | 为 v1.0 的三平台解析闭环准备统一 fixture |
