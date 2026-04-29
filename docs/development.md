# 开发指南

## 默认约定

- GitHub 相关工作优先使用 GitHub CLI（`gh`）。
- 引导提交之后，开发流程为：议题 -> 分支 -> 拉取请求。
- 保持 `main` 随时可发布。
- 优先提交小而清晰、便于审查的改动。

## 本地命令

```powershell
go run ./cmd/surveyctl version
go test ./...
go test -race ./...
go vet ./...
gofmt -w (git ls-files '*.go')
```

在 Windows 上，`go test -race` 需要 CGO 和 `gcc` 之类的 C 编译器。如果本地竞态检查失败并提示 `C compiler "gcc" not found`，可以先运行普通测试，并依赖 Ubuntu CI 中的竞态检查，直到本地安装好 C 工具链。

## 先有议题

修改行为前先创建或认领议题：

```powershell
gh issue create
gh issue view 123
```

议题应写清用户目标、范围、验收标准和安全约束。

## 分支

使用简短、关联议题的分支名：

```text
codex/issue-123-config-schema
codex/issue-124-provider-contract
```

## 提交

提交信息保持简洁：

```text
chore: bootstrap go project
feat: add runtime engine mode parser
test: cover provider capability checks
```

## 拉取请求

使用 GitHub CLI 打开草稿拉取请求：

```powershell
gh pr create --draft --fill
```

每个拉取请求必须包含：

- 关联议题。
- 改了什么。
- 为什么改。
- 运行过哪些测试。
- 涉及时说明性能影响。
- 风险与回滚说明。

## 风格规则

- 参考 Google Go 风格指南和 Go 代码审查建议。
- 函数保持单一职责。
- 接口尽量靠近消费侧。
- 包名要清晰。
- 可取消的工作使用 `context.Context`。
- 遇到平台验证、登录或反滥用页面时停止并报告。

## 引导提交例外

由于远程仓库最初为空，`v0.1` 允许一次直接提交到 `main` 的引导提交。这个例外不适用于后续正常功能开发。
