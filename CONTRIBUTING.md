# 贡献指南

感谢你愿意一起建设 SurveyController-go。

## 必须遵守的协作流程

- GitHub 相关操作默认优先使用 GitHub CLI（`gh`）。
- `v0.1` 引导提交之后，所有非琐碎改动都必须先创建 GitHub 议题。
- 在独立主题分支上开发，然后打开拉取请求。
- 初始化例外结束后，不再直接向 `main` 推送。

建议流程：

```powershell
gh issue create
git checkout -b codex/issue-123-short-description
go test ./...
gh pr create --draft --fill
```

## Go 代码风格

- 参考 Google Go 风格指南和 Go 代码审查建议。
- 提交前运行 `gofmt`。
- 包要小而清晰，职责边界明确。
- 函数保持单一职责。
- 接口尽量定义在消费侧。
- 错误要显式返回，不要只写日志却吞掉失败。
- 可取消、可超时的工作使用 `context.Context`。
- 避免 `util`、`common`、`misc` 这类万能包名。

## 测试要求

打开拉取请求前请运行：

```powershell
gofmt -w (git ls-files '*.go')
go test ./...
go test -race ./...
go vet ./...
```

如果某项检查无法在本地运行，请在拉取请求中说明原因。

## 项目范围

`v0.1` 只做项目初始化。平台解析和真实运行能力会在后续版本实现。`v1.0` 目标是正式支持问卷星、腾讯问卷和 Credamo 见数平台。

## 安全边界

本项目仅供获得授权的学习与测试使用。任何绕过平台验证、自动化滥用，或帮助生成虚假答卷的贡献都不会被接受。
