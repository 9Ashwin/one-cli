# PRD: Agent-Native CLI 生成框架

## Introduction

`lark-cli` 是一个成熟的 Agent-Native CLI 实现：它让人类和 AI Agent 都能在终端中操作飞书，覆盖三层命令架构、元数据驱动的命令生成、26 个 AI Skills、OAuth 认证与 keychain 凭证、类型化结构化错误、输入防注入/dry-run 安全层等。这些能力中，大量设计是**平台无关**的——任何"开放平台 + REST API + AI Agent 消费"的场景都需要它们。

本 PRD 描述将 `lark-cli` 的核心能力抽离为一个通用 **Agent-Native CLI 生成框架**（以下简称"框架"）。框架由两部分组成：

1. **`anicli-sdk`**：可嵌入的 Go 核心库，提供三层命令运行时、认证抽象、结构化错误、安全层、Schema 自省等能力。
2. **`anicli`**：脚手架生成器 CLI，读取 OpenAPI 3.x 规范（+ 自有扩展字段），交互式生成一个完整的新 CLI 项目骨架。

使用者只需提供 OpenAPI 规范与少量平台适配代码，即可产出一个具备 Agent 友好性（结构化输出、错误可恢复、Skills 开箱即用）的 CLI。`lark-cli` 自身将重构为基于该框架的上层实现（dogfooding），验证框架的通用性与迁移可行性。

> 读者假设：中级 Go 开发者或 AI Agent。下文涉及 `lark-cli` 现有概念时会注明对应源码路径。

## Goals

- 将 `lark-cli` 的平台无关核心抽离为独立可复用的 `anicli-sdk` + `anicli` 脚手架
- 支持仅凭一份 OpenAPI 3.x 规范（+ 扩展字段）生成可运行的 Agent-Native CLI 骨架
- 框架保留 lark-cli 的六项核心能力：三层命令架构、AI Skills 生成、认证抽象、结构化错误、安全层、Schema 自省 + Quality gates
- `lark-cli` 重构为框架上层实现，迁移后功能与回归测试不退化
- 框架可独立于飞书生态对外开源（MIT）

## User Stories

### US-001: 初始化框架仓库与模块边界
**Description:** As a framework maintainer, I want a clean repo/module split so that SDK and generator can evolve and be imported independently.

**Acceptance Criteria:**
- [ ] 新仓库 `one-cli` 含两个 Go module：`anicli-sdk`（核心库）、`anicli`（生成器），各自 `go.mod` 独立
- [ ] 顶层 Makefile 提供 `make build` / `make unit-test` / `make test`，覆盖两个模块
- [ ] 顶层 README 说明框架定位、与 lark-cli 的关系、目录结构
- [ ] `go vet ./...` 与 `gofmt -l .` 无输出

### US-002: 抽离三层命令运行时到 SDK
**Description:** As a CLI author, I want the three-tier command runtime (Shortcuts + API commands + generic call) in the SDK so I can assemble my platform's commands by metadata.

**Acceptance Criteria:**
- [ ] `anicli-sdk` 提供 `runtime` 包，封装快捷命令管道（对应 `shortcuts/common/runner.go`）、元数据自动注册的 API 命令（对应 `cmd/service/`）、通用调用命令
- [ ] SDK 接口不依赖任何飞书特定类型，命令注册通过平台无关的 `ServiceMeta` / `MethodMeta` 数据结构
- [ ] 三层命令均可由使用者通过 `runtime.New(...)` 装配，并提供 `--format`/`--dry-run`/`--page-*` 等通用 flag
- [ ] 单元测试覆盖三层命令的装配与执行（dry-run 请求结构断言）
- [ ] `go vet` / `gofmt` 无输出

### US-003: 定义元数据输入规范（OpenAPI 3.x + 扩展字段）
**Description:** As a CLI author, I want a documented metadata spec so I can describe my platform's services/methods/scopes/shortcuts in one place.

**Acceptance Criteria:**
- [ ] 框架文档定义输入规范：以 OpenAPI 3.x 为基础，扩展字段置于 `x-anicli-*`（含 scope、shortcut、身份映射、quality 等级）
- [ ] 提供 `examples/petstore.openapi.json` 一份最小可运行示例规范
- [ ] SDK 提供 `metadata` 包解析 OpenAPI + 扩展字段，解析失败返回类型化错误
- [ ] 单元测试覆盖合法/非法规范解析（缺失必填字段、非法 scope 字符串等）
- [ ] `go vet` / `gofmt` 无输出

### US-004: 实现 `anicli init` 脚手架生成器
**Description:** As a CLI author, I want to run `anicli init` to scaffold a new CLI project from an OpenAPI spec so I can start within minutes.

**Acceptance Criteria:**
- [ ] `anicli init --spec <path> --name <cli-name>` 生成一个可编译的 Go CLI 项目骨架（main.go、cmd 注册、go.mod 引用 anicli-sdk）
- [ ] 生成的项目 `go build ./...` 成功
- [ ] 生成的 CLI 能执行 `--help`，列出从规范解析出的服务与命令
- [ ] 生成的 CLI 能执行任一 API 命令的 `--dry-run` 并输出请求结构（method/URL/params）
- [ ] 生成器对已存在目标目录采取非破坏策略（拒绝覆盖或要求 `--force`）
- [ ] 单元测试覆盖生成器输出（文件存在性、可编译性）
- [ ] `go vet` / `gofmt` 无输出

### US-005: 抽离认证抽象层到 SDK
**Description:** As a CLI author, I want a pluggable auth layer so my CLI can support OAuth, API keys, or custom credential flows without rewriting command code.

**Acceptance Criteria:**
- [ ] `anicli-sdk` 提供 `auth` 包定义 `CredentialProvider` 接口，含 OAuth2 device flow、keychain 存储默认实现（对应 `internal/credential/`、`internal/keychain/`）
- [ ] 提供 `auth login` / `auth logout` / `auth status` / `auth check` 通用命令实现，平台通过注入 scope 目录复用
- [ ] 支持多 profile（对应 `cmd/profile/`）与身份切换（`--as user|bot`）抽象
- [ ] 凭证默认存入 OS 原生 keychain，并提供 env/in-memory provider 作为可插拔备选
- [ ] 单元测试覆盖 provider chain 选择与 profile 切换
- [ ] `go vet` / `gofmt` 无输出

### US-006: 抽离类型化结构化错误体系到 SDK
**Description:** As a CLI author, I want the typed error envelope in the SDK so AI agents can parse `type`/`subtype`/`param`/`hint` to decide recovery.

**Acceptance Criteria:**
- [ ] `anicli-sdk` 提供 `errs` 包，迁移 lark-cli 的错误分类（ValidationError/InternalError/NetworkError 等）与构造器（对应 `errs/`）
- [ ] stderr envelope 结构与 lark-cli 一致，包含 `category`/`subtype`/`param`/`hint`/`cause` 字段
- [ ] 提供 lint 规则（移植 `lint/errscontract`）阻止裸 `fmt.Errorf` 终态与 legacy 错误助手
- [ ] 单元测试断言类型化元数据（`errs.ProblemOf`）而非仅消息子串
- [ ] `go vet` / `gofmt` 无输出

### US-007: 抽离安全层到 SDK
**Description:** As a CLI author, I want the safety layer (anti-injection, path validation, dry-run, output sanitization) in the SDK so Agent-supplied input is bounded by default.

**Acceptance Criteria:**
- [ ] `anicli-sdk` 提供 `validate` 包（`SafeInputPath`/`SafeOutputPath`，对应 `internal/validate/`）与输出净化（对应 `internal/security/`）
- [ ] 所有文件 I/O 通过 `vfs` 抽象（对应 `internal/vfs/`），便于测试
- [ ] 副作用命令默认支持 `--dry-run`，运行时可在平台层声明某命令为"只读/可写"
- [ ] 单元测试覆盖路径越权、注入载荷、dry-run 输出净化
- [ ] `go vet` / `gofmt` 无输出

### US-008: 实现 Schema 自省命令
**Description:** As an AI agent, I want a `schema` command to introspect available methods, params, and scopes so I can self-discover capabilities.

**Acceptance Criteria:**
- [ ] SDK 提供可注入元数据目录的 `schema` 命令实现（对应 `cmd/schema`、`internal/schema/`）
- [ ] `anicli` 生成项目默认注册 `schema` 命令，支持 `schema`（列出全部）与 `schema <method>`（单方法详情）
- [ ] 输出包含参数、请求体、响应结构、所需 scope、支持身份
- [ ] 单元测试覆盖 schema 列表与单方法详情输出结构
- [ ] `go vet` / `gofmt` 无输出

### US-009: 实现 AI Skills 生成与分发
**Description:** As a CLI author, I want the generator to produce AI Agent Skills (master skill + domain skills) from my metadata so agents can use my CLI out-of-the-box.

**Acceptance Criteria:**
- [ ] `anicli` 提供 `anicli skill gen` 从元数据生成 master skill + 各 domain skill Markdown（对应 `skill-template/`）
- [ ] 生成的 skill 文件结构兼容主流 AI 工具（Claude Code / Cursor / Gemini CLI 的 skill 加载约定）
- [ ] 提供 `shared` skill 模板（认证、安全规则），各 domain skill 自动引用
- [ ] 单元测试覆盖生成输出（文件存在性、master 引用 domain、shared 引用完整性）
- [ ] `go vet` / `gofmt` 无输出

### US-010: 实现 Quality gates（命令准入评测）
**Description:** As a framework maintainer, I want a quality gate harness so only Agent-tested commands are admitted into a CLI's stable surface.

**Acceptance Criteria:**
- [ ] SDK 提供 `qualitygate` 包，定义命令准入评测接口（dry-run 结构校验 + Agent 调用成功率指标，对应 `internal/qualitygate/`）
- [ ] 提供 CLI 子命令 `anicli quality check <spec>` 输出每方法通过/未通过报告
- [ ] 未通过命令默认不出现在生成的 CLI stable 表面（可 `--include-experimental` 放开）
- [ ] 单元测试覆盖通过/未通过/降级标记逻辑
- [ ] `go vet` / `gofmt` 无输出

### US-011: 迁移 lark-cli 为框架上层实现（dogfooding）
**Description:** As a framework maintainer, I want lark-cli rebuilt on anicli-sdk so the framework is proven on a real, large-scale platform.

**Acceptance Criteria:**
- [ ] lark-cli 改为 `import anicli-sdk`，移除已抽离的重复实现，保留飞书专属适配层（scope 目录、飞书 OpenAPI 元数据、keychain 命名空间）
- [ ] lark-cli 现有 dry-run E2E（`tests/cli_e2e/dryrun/`）全量通过
- [ ] lark-cli `make unit-test` 全量通过，无新增 `go vet` / lint 告警
- [ ] lark-cli `--help`、`schema`、`auth status`、`api GET/POST`、至少 1 个 shortcut 行为与迁移前一致（回归断言）
- [ ] 记录迁移过程中发现的框架缺口至 Open Questions

### US-012: 框架文档与开源就绪
**Description:** As an open-source user, I want docs and examples so I can build my own Agent-Native CLI without reading lark internals.

**Acceptance Criteria:**
- [ ] 顶层 README 含快速开始：`anicli init` → 生成 → `--help` → dry-run，三步内可运行
- [ ] `docs/` 含输入规范、认证扩展点、错误体系、安全模型四篇文档
- [ ] `examples/` 含 petstore 全流程示例（规范 + 生成结果 + skills）
- [ ] LICENSE 为 MIT，CI 配置（lint + test + build）就绪
- [ ] `go vet` / `gofmt` 无输出

## Functional Requirements

- FR-1: 系统必须提供两个独立 Go module：`anicli-sdk`（核心库）与 `anicli`（生成器）
- FR-2: 系统必须提供三层命令运行时，支持快捷命令、元数据驱动 API 命令、通用调用三种粒度装配
- FR-3: 系统必须以 OpenAPI 3.x 作为元数据输入基础，并通过 `x-anicli-*` 扩展字段承载 scope、shortcut、身份映射、quality 等级
- FR-4: `anicli init` 必须从规范生成可编译、可执行 `--help` 与 `--dry-run` 的 CLI 项目骨架
- FR-5: 系统必须提供可插拔认证层，含 OAuth2 device flow、keychain 存储、多 profile、身份切换默认实现
- FR-6: 系统必须提供类型化结构化错误体系，stderr envelope 含 `category`/`subtype`/`param`/`hint`/`cause`
- FR-7: 系统必须提供安全层：路径校验、vfs 抽象、输出净化、副作用命令默认 dry-run
- FR-8: 系统必须提供 `schema` 自省命令，输出方法参数、请求体、响应结构、scope、支持身份
- FR-9: `anicli skill gen` 必须从元数据生成 master + domain + shared AI Skills
- FR-10: 系统必须提供 quality gate 评测，未通过命令默认不进入生成的 CLI stable 表面
- FR-11: lark-cli 必须重构为 `anicli-sdk` 上层实现，且现有 dry-run E2E 与 unit-test 全量通过
- FR-12: 系统必须提供 MIT 许可证与 CI 配置（lint + test + build）

## Non-Goals (Out of Scope)

- 不支持非 REST/OpenAPI 的协议（gRPC、GraphQL）——首版仅 REST
- 不内置飞书特定业务逻辑（scope 目录、飞书 OpenAPI 元数据保留在 lark-cli 适配层，不进框架）
- 不提供 GUI / TUI 配置界面（仅 CLI + 交互式 prompt）
- 不实现 Skills 的自动版本升级分发服务（仅生成，分发沿用各 AI 工具现有机制）
- 不做多语言 SDK（首版仅 Go）
- 不重构 lark-cli 的飞书业务快捷命令实现，仅替换其底层框架依赖

## Design Considerations

- 目录结构建议：`anicli-sdk/{runtime,auth,errs,validate,vfs,schema,metadata,qualitygate,...}`、`anicli/{cmd,generator,templates}`、`examples/`、`docs/`
- skill 模板沿用 lark-cli 的 `master-skill-template.md` + `skill-template.md` + `domains/*.md` 结构，参数化平台名与命令前缀
- 输出格式 flag（json/pretty/table/ndjson/csv）与分页 flag（page-all/page-limit/page-delay）作为 SDK 通用 flag 注入
- 交互式 prompt 使用 `charmbracelet/huh`（与 lark-cli 一致），保持生成器体验一致

## Technical Considerations

- Go 版本与 lark-cli 对齐：`go 1.23+`
- 关键依赖复用：`spf13/cobra`（命令框架）、`tidwall/gjson`（JSON 处理）、`charmbracelet/huh`+`lipgloss`（交互/样式）、`larksuite/oapi-sdk-go` 仅 lark-cli 适配层保留
- OpenAPI 解析：优先 `kin-openapi` 或等价库；扩展字段通过 `x-anicli-*` 读取
- keychain 跨平台：复用 lark-cli 现有 keychain 抽象（macOS Keychain / Windows Credential Manager / Linux libsecret）
- 模块拆分需保证 `anicli-sdk` 不反向依赖 `anicli`，避免循环
- 迁移 lark-cli 时分阶段：先抽 SDK 接口 → lark-cli 切换实现 → 跑回归 → 删除旧代码

## Success Metrics

- 从一份 OpenAPI 规范到可运行 `--dry-run` 的 CLI，`anicli init` 耗时 < 30 秒
- lark-cli 迁移至框架后，dry-run E2E 与 unit-test 通过率 100%，无行为回归
- 框架抽离后，lark-cli 代码体积（去除已抽离实现）减少 ≥ 30%
- 新平台接入：按 examples/petstore 流程，用户 10 分钟内产出可执行 CLI 骨架
- 框架覆盖 lark-cli 六项核心能力，每项均有独立 US 与测试验证

## Open Questions

- 框架仓库名与 Go module 路径最终命名（当前假设 `anicli-sdk` / `anicli`，待确认）
- 是否需要支持多 OpenAPI 文件合并（多 service 拼接为一套 CLI）？
- quality gate 的"Agent 调用成功率"指标如何在没有真实 Agent 的 CI 中度量？是否以 dry-run 结构校验 + 规则集替代？
- Skills 生成是否需要支持多 AI 工具格式差异（Claude Code skill vs Cursor rule）？首版是否统一一套？
- lark-cli 迁移是否与框架开发并行，还是框架稳定后再迁？建议先稳定核心 SDK 再迁（见 US-011 依赖 US-002~US-010）
- 是否保留 lark-cli 的 `_notice.update` / `_notice.skills` 通知机制作为 SDK 通用能力？

[Assumption] 本 PRD 默认采用推荐方案：1C（SDK + 脚手架）、2ABCDEF（全保留）、3C（OpenAPI + 扩展）、4A（dogfooding 迁移 lark-cli）。以上 Open Questions 待用户在 review 阶段确认。
