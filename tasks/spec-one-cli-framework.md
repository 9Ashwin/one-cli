# SPEC: one-cli — Agent-Native CLI 生成框架

> Technical specification derived from: tasks/prd-agent-native-cli-framework.md
> Generated: 2026-07-02 | Repo: github.com/9Ashwin/one-cli (public) | Module: github.com/9Ashwin/one-cli

## 1. Summary

### 1.1 What This SPEC Covers
将 `lark-cli` 的平台无关核心能力抽离为通用 Agent-Native CLI 生成框架 `one-cli`。框架由一个主 Go module 构成：公开 SDK API 放在导出包（`runtime`/`auth`/`errs`/`validate`/`vfs`/`schema`/`metadata`/`qualitygate`），`extension/` 与 `lint/` 为独立 module（沿用 lark-cli 多 module 模式），生成器二进制 `onecli` 作为主 module 内 `cmd/onecli` 入口。使用者通过 `onecli init --spec <openapi>` 生成可运行 CLI 骨架。工程化（AGENTS.md、Makefile、多层 CI、自定义 lint、错误契约、quality-gate、goreleaser、gitleaks）全面复刻 lark-cli。

### 1.2 PRD Reference
- Source: tasks/prd-agent-native-cli-framework.md
- User Stories covered: US-001 ~ US-012（全部）
- Functional Requirements covered: FR-1 ~ FR-12（全部）

### 1.3 Design Decisions Summary
| Decision | Choice | Rationale |
|----------|--------|-----------|
| 模块组织 | 单主 module + `extension`/`lint` 独立 module | 沿用 lark-cli 模式（`extension/`、`lint/` 各自 go.mod）；SDK 公开 API 放主 module 导出包，避免双 module 循环依赖 |
| 生成器入口 | `cmd/onecli` 主 module 内，产出 `onecli` 二进制 | 与 lark-cli `main.go` 一致；不拆独立 module，简化依赖 |
| 元数据输入 | OpenAPI 3.x + `x-onecli-*` 扩展 | lark-cli 自有 meta 非通用；框架须吃业界标准。扩展字段命名从 `x-anicli-*` 改为 `x-onecli-*` 以对齐项目名 |
| OpenAPI 解析库 | `kin-openapi` | 社区主流、活跃；避免自研解析器维护成本（与"尽量按 lark-cli 来"不冲突——lark-cli 的 meta 是平台私有，框架必须吃标准） |
| 插件 SDK | 纳入，独立 module `extension/` | 直接复刻 lark-cli `extension/platform/`（capabilities/lifecycle/risk/rule/selector） |
| 错误体系 | 移植 `errs/` + `ERROR_CONTRACT.md` | RFC 7807 对齐，wire-stable，AI agent 可解析 |
| 工程化 | 全套复刻 lark-cli | AGENTS.md / Makefile 多 target / 多层 CI / 自定义 lint / quality-gate / goreleaser / gitleaks |
| 迁移节奏 | 串行：框架核心稳定（US-002~010）后再迁 lark-cli（US-011） | 风险低；与 lark-ci 分层 CI 的串行 gate 理念一致 |
| GitHub owner | `github.com/9Ashwin/one-cli` | 已确认 |

---

## 2. Architecture

### 2.1 System Context
```
                      ┌─────────────────────────────────────────┐
   CLI 作者 ───────▶  │ onecli init --spec openapi.json          │
                      │   (生成器: cmd/onecli, 主 module)         │
                      └───────────────┬─────────────────────────┘
                                      │ 生成
                                      ▼
                      ┌─────────────────────────────────────────┐
                      │ 生成的目标 CLI 项目 (go build)            │
                      │   import "github.com/9Ashwin/one-cli/..." │
                      │   + 平台适配层 (scope 目录/OAuth config)  │
                      └───────────────┬─────────────────────────┘
                                      │ 运行时调用
                                      ▼
   AI Agent / 人类 ──▶ 目标 CLI ──▶ anicli-sdk 运行时 ──▶ 平台 REST API
                                      │
                                      └─ errs envelope (stderr) / stdout JSON
```

### 2.2 Component Design
主 module `github.com/9Ashwin/one-cli` 下导出包：

| 包 | 职责 | lark-cli 对应 |
|----|------|---------------|
| `runtime` | 三层命令运行时：Shortcut 管道、元数据驱动 API 命令、通用调用 | `shortcuts/common/runner.go`、`cmd/service/`、`cmd/api/` |
| `metadata` | OpenAPI 3.x + `x-onecli-*` 解析，产出 `ServiceMeta`/`MethodMeta` | `internal/cmdmeta/`、`internal/apicatalog/`、`internal/registry/` |
| `auth` | 认证抽象：`CredentialProvider` 接口、OAuth2 device flow、keychain、多 profile、身份切换 | `internal/credential/`、`internal/keychain/`、`cmd/profile/`、`internal/auth/` |
| `errs` | 类型化错误分类 + wire envelope（RFC 7807 对齐） | `errs/`（直接移植） |
| `validate` | 路径校验 `SafeInputPath`/`SafeOutputPath`、输入净化 | `internal/validate/` |
| `vfs` | 文件系统抽象（测试可 mock） | `internal/vfs/` |
| `security` | 输出净化、防注入 | `internal/security/` |
| `schema` | Schema 自省命令实现 | `cmd/schema/`、`internal/schema/` |
| `qualitygate` | 命令准入评测：manifest 导出 + check | `internal/qualitygate/` |
| `skillgen` | AI Skills 生成（master + domain + shared） | `skill-template/`、`cmd/skill/` |
| `output` | 输出格式（json/pretty/table/ndjson/csv）+ 分页 + envelope | `internal/output/` |
| `client` | `APIClient`：`DoSDKRequest`/`DoStream`，HTTP 传输 | `internal/client/`、`internal/transport/` |

独立 module：
- `extension/`（go.mod: `github.com/9Ashwin/one-cli/extension`）—— 公开插件 SDK，复刻 `extension/platform/`：capabilities/builder/handler/identity/invocation/lifecycle/plugin/register/risk/rule/selector/view
- `lint/`（go.mod: `github.com/9Ashwin/one-cli/lint`）—— 自定义 lint analyzer，含 `errscontract`（阻止裸 `fmt.Errorf` 终态、legacy 助手、未声明 subtype）与 `lintapi`

### 2.3 Module Interactions
```
cmd/onecli (生成器)  ──imports──▶  metadata, skillgen, runtime(模板示例)
   │
   └─ templates/ (go embed) ──▶ 生成的项目骨架引用主 module 导出包

生成的目标 CLI ──imports──▶  runtime, auth, errs, validate, vfs, schema, output, client
                              │
                              └─ extension (插件作者按需 import)
```
依赖方向约束：`extension` 与 `lint` 不反向依赖主 module；主 module 不依赖 `cmd/onecli`。`errs` 是最底层，无内部依赖。

### 2.4 File Structure
```
one-cli/
├── go.mod                          # module github.com/9Ashwin/one-cli
├── main.go                         # 入口（路由到 cmd/onecli 或目标 CLI）
├── Makefile                        # 复刻 lark-cli 多 target
├── AGENTS.md                       # 贡献规范 + 源码布局 + 错误契约
├── README.md / README.zh.md
├── LICENSE (MIT)
├── .golangci.yml                   # 复刻
├── .goreleaser.yml                 # 复刻，binary: onecli
├── .gitleaks.toml
├── .licenserc.yaml
├── .codecov.yml
├── .github/
│   ├── CODEOWNERS
│   ├── pull_request_template.md
│   └── workflows/                  # ci / release / semantic-review / arch-audit / pr-labels / skill-format-check
├── runtime/                        # [NEW] 三层命令运行时（导出包，从 internal 提升并去飞书化）
├── metadata/                       # [NEW] OpenAPI 解析
├── auth/                           # [NEW] 认证抽象
├── errs/                           # [移植] + ERROR_CONTRACT.md
├── validate/                       # [移植]
├── vfs/                            # [移植]
├── security/                       # [移植]
├── schema/                         # [NEW]
├── qualitygate/                    # [移植]
├── skillgen/                       # [NEW] + templates/
├── output/                         # [移植]
├── client/                         # [移植]
├── cmd/
│   └── onecli/                     # [NEW] 生成器入口
│       ├── init.go                 # anicli init
│       ├── skill.go                # onecli skill gen
│       └── quality.go              # onecli quality check
├── extension/                      # [独立 module, 移植] 插件 SDK
├── lint/                           # [独立 module, 移植] 自定义 lint
├── scripts/
│   ├── fetch_meta.py               # [保留概念] 框架无飞书 meta，改为示例规范拉取
│   ├── resolve-changed-from.sh
│   ├── check-doc-tokens.sh
│   └── *.test.sh / *.test.js
├── skill-template/                 # [NEW] 参数化模板（master/domain/shared）
├── examples/
│   └── petstore/                   # [NEW] openapi.json + 生成结果 + skills
├── tests/
│   └── cli_e2e/                    # dry-run / live E2E
└── docs/
    ├── input-spec.md               # OpenAPI + x-onecli-* 规范
    ├── auth-extension.md
    ├── error-contract.md           # (另见 errs/ERROR_CONTRACT.md)
    └── security-model.md
```

---

## 3. Data Model

### 3.1 元数据输入规范（OpenAPI 3.x + `x-onecli-*`）

无数据库。框架的核心"数据"是元数据。规范定义：

```yaml
openapi: 3.0.3
info:
  title: Petstore
  version: 1.0.0
x-onecli:
  cli-name: petstore-cli        # 生成的 CLI 名
  default-identity: user         # user | bot
  auth:
    oauth:
      token-url: https://example.com/token
      device-auth-url: https://example.com/device
      scopes:                    # 全局 scope 目录
        - id: pet:read
          desc: 读取宠物
paths:
  /pets:
    get:
      operationId: petsList
      x-onecli:
        service: pets            # 三层命令的 service 维度
        shortcut: "+list"        # 可选：注册为快捷命令
        identity: [user]         # 支持身份
        scopes: [pet:read]       # 所需 scope
        quality: stable          # stable | experimental（准入）
        dry-run-safe: false      # 是否只读（影响默认 dry-run）
      parameters: [...]
      responses: {...}
```

### 3.2 内部实体定义（Go）
```go
// metadata 包
type Spec struct {
    CLIName        string
    DefaultIdentity Identity
    Auth           AuthConfig
    Services       map[string]ServiceMeta
}

type ServiceMeta struct {
    Name    string
    Methods []MethodMeta
}

type MethodMeta struct {
    OperationID  string
    HTTPMethod   string
    Path         string
    Shortcut     string // "" 表示无快捷命令
    Identities   []Identity
    Scopes       []string
    Quality      Quality // Stable | Experimental
    DryRunSafe   bool
    Params       []ParamMeta
    RequestBody  *SchemaMeta
    Responses    map[int]SchemaMeta
}

type ParamMeta struct {
    Name     string
    In       string // path|query|header
    Required bool
    Type     string
}
```

### 3.3 错误实体（移植 `errs/`）
```go
type Problem struct {
    Category Category   // 9 类，闭集
    Subtype  Subtype    // 声明式，未声明 CI 失败
    Code     int        // 上游码，omitempty
    Message  string
    Hint     string
    Param    string     // 仅 ValidationError
    Cause    error      // .WithCause
}
```

### 3.4 Schema 迁移与兼容
无 DB 迁移。元数据规范版本化：`x-onecli.spec-version: "1"`。解析器按版本路由；未知字段忽略并 warning（stderr）。

---

## 4. API Design

框架非网络服务，"API"指：(a) SDK Go 公开接口；(b) `onecli` 生成器子命令；(c) 生成的 CLI 的命令表面。

### 4.1 生成器子命令
| 命令 | 说明 | 关键 flag |
|------|------|-----------|
| `onecli init` | 生成 CLI 骨架 | `--spec`, `--name`, `--out`, `--force` |
| `onecli skill gen` | 生成 AI Skills | `--spec`, `--out`, `--format`(claude/cursor) |
| `onecli quality check` | 命令准入评测 | `--spec`, `--changed-from`, `--facts-out` |
| `onecli schema` | 自省框架自身能力 | — |

### 4.2 SDK 公开装配接口
```go
// runtime 包
func New(deps Deps) (*Runtime, error)

type Deps struct {
    Meta        *metadata.Spec
    Auth        auth.CredentialProvider
    Client      client.APIClient
    Output      output.Formatter
    VFS         vfs.VFS
}

type Runtime struct{ ... }
func (r *Runtime) RootCmd() *cobra.Command        // 装配三层命令
func (r *Runtime) RegisterShortcut(...) error
```

### 4.3 生成的 CLI 命令表面（三层）
| 层 | 形态 | 示例 |
|----|------|------|
| 快捷命令 | `<cli> <service> +<shortcut>` | `petstore-cli pets +list` |
| API 命令 | `<cli> <service> <resource> <method>` | `petstore-cli pets list` |
| 通用调用 | `<cli> api <METHOD> <path>` | `petstore-cli api GET /pets` |
| 自省 | `<cli> schema`, `<cli> schema pets.list` | — |
| 认证 | `<cli> auth login/logout/status/check/scopes/list` | — |

通用 flag（SDK 注入）：`--format`, `--dry-run`, `--page-all`, `--page-limit`, `--page-delay`, `--as`(user|bot)。

### 4.4 错误响应（stderr envelope）
```json
{
  "ok": false,
  "identity": "user",
  "error": {
    "type": "validation",
    "subtype": "invalid_argument",
    "param": "--chat-id",
    "message": "...",
    "hint": "run petstore-cli schema pets.list"
  }
}
```
字段稳定性与 lark-cli `errs/ERROR_CONTRACT.md` 完全一致。

### 4.5 Breaking Changes
对 lark-cli 的 wire envelope 字段（`ok`/`identity`/`error.type`/`error.subtype`）保持 wire-stable，重命名视为破坏性变更。

---

## 5. Business Logic

### 5.1 三层命令解析算法
```
输入 argv:
  if argv[0] == "api":         → 通用调用层（直接 METHOD + path + params/data）
  elif argv[1] 前缀为 "+":     → 快捷命令层（查 MethodMeta.Shortcut，注入智能默认值）
  else:                        → API 命令层（service.resource.method 映射 OperationID）

执行管道（shortcut）:
  1. 解析 flag → Flag.Input(@file/stdin) 解析
  2. validate.SafeInputPath 校验所有路径参数
  3. 若 DryRunSafe==false 且无 --dry-run → 警告（stderr）
  4. auth.CredentialProvider 取凭证，校验 scope（auth.check）
  5. client.DoSDKRequest，分类 code!=0 为 typed error
  6. output.Formatter 渲染 stdout；进度/hint → stderr
  7. 分页：--page-all 循环合并
```

### 5.2 `onecli init` 生成算法
```
1. metadata.Parse(spec) → *Spec；失败返回 errs typed error
2. 校验目标目录：存在且非空且无 --force → error
3. 渲染 templates/（go:embed）替换占位（cli-name, module-path, services...）
4. 生成：main.go, go.mod(import 主 module), cmd/, skills/, README
5. go build ./... 验证可编译（生成器内 exec）
6. 输出生成清单（stdout JSON）+ 下一步提示（stderr）
```

### 5.3 Quality gate 准入逻辑
```
对 Spec.Services 全部 MethodMeta:
  - stable: 须通过 dry-run 结构校验（method/URL/params 与 OpenAPI 一致）
  - experimental: 允许未通过，但不出现在生成 CLI 的默认表面
  - --include-experimental 放开
输出 facts.json：每方法 {passed, reason}
```

### 5.4 Edge Cases
- OpenAPI 缺 `operationId` → 用 `METHOD_path` 生成，并 stderr warning
- `x-onecli.service` 缺失 → 归入 `default` service
- 同一 shortcut 重复 → init 报 typed error（`duplicate_shortcut`）
- 凭证过期 → `auth` 返回 `authorization/missing_scope` 或 `token_expired`，envelope 带 `hint`
- `--as bot` 但平台未配 bot → `authorization/identity_unavailable`

---

## 6. Error Handling

### 6.1 Error Taxonomy（移植 9 类 Category）
| Category | 触发场景 | Exit Code |
|----------|----------|-----------|
| validation | flag/参数校验失败、错误系统状态 | 2 |
| authorization | 缺 scope、token 过期、身份不可用 | 3 |
| not_found | 资源/方法不存在 | 4 |
| network | 传输失败 | 5 |
| policy | 安全策略拦截（SecurityPolicyError） | 6 |
| internal | 文件 I/O、未分类 | 1 |
| (其余移植自 errs/category.go) | — | — |

### 6.2 Retry Strategy
- 网络错误（`network/transport`）：SDK 不自动重试；分页请求遇 5xx 跳过该页并 stderr warning（与 lark-cli 一致）
- OAuth token 过期：`auth` 自动刷新一次，仍失败返回 typed error

### 6.3 Failure Modes
- keychain 不可用（无 GUI 的 CI）→ 降级 env provider，stderr warning
- 元数据远程拉取失败 → 用嵌入默认元数据（`loader_embedded.go` 模式）

---

## 7. Security

### 7.1 Authentication & Authorization
- 默认 OAuth2 device flow；`auth.CredentialProvider` 接口允许平台注入 API key / 自定义流
- scope 校验在命令执行前（`auth.check`），缺失即 `authorization/missing_scope`，envelope 带 `missing_scopes` 与 `hint`
- 多 profile 隔离凭证；身份切换 `--as` 受 `MethodMeta.Identities` 约束

### 7.2 Input Validation
- 所有文件路径经 `validate.SafeInputPath`/`SafeOutputPath`（禁止 `..` 越权、绝对路径逃逸）
- 所有文件 I/O 经 `vfs` 抽象
- 命令参数来自 AI agent，视为不可信：参数值经 `security` 净化后渲染

### 7.3 Data Protection
- 凭证存 OS 原生 keychain（macOS Keychain / Windows Credential Manager / Linux libsecret），不明文落盘
- `gitleaks` + `check-doc-tokens.sh` 阻止示例 token 泄漏（文档用 `_EXAMPLE_TOKEN` 占位）
- stdout 仅数据；敏感字段不进 envelope

---

## 8. Performance

### 8.1 Expected Load
- 单进程 CLI，无并发服务。关注点：启动延迟（解析元数据 < 100ms）、大分页吞吐
- 元数据嵌入（go:embed）避免运行时远程拉取

### 8.2 Optimization Strategy
- 元数据解析结果缓存于 `Runtime` 构造期
- 分页：流式 ndjson 输出，避免全量内存
- `--dry-run` 不发起真实请求

### 8.3 Database Considerations
无数据库。

---

## 9. Testing Strategy

### 9.1 Unit Tests
- 每个导出包独立测试，边界：合法/非法 OpenAPI、provider chain、profile 切换、路径越权、dry-run 净化
- 用 `cmdutil.TestFactory` 等价物构造测试 `Runtime`；`t.Setenv("ONECLI_CONFIG_DIR", t.TempDir())` 隔离
- 错误路径断言 `errs.ProblemOf` 的 `category`/`subtype`/`param` + cause 保留，而非仅消息子串

### 9.2 Integration Tests
- `onecli init` 生成后 `go build ./...` 必须成功
- 生成的 CLI 跑 `--help` / `schema` / `--dry-run`，断言输出结构

### 9.3 Edge Case Tests
- OpenAPI 缺 operationId / 重复 shortcut / 凭证过期 / `--as bot` 未配 / keychain 不可用降级

### 9.4 E2E（复刻 lark-cli 双层）
| 类型 | 位置 | 用途 |
|------|------|------|
| Dry-run E2E | `tests/cli_e2e/dryrun/` | 校验请求结构，无需 secret，fork PR 可跑 |
| Live E2E | `tests/cli_e2e/<domain>/` | 真实 API 往返，自包含 create→use→cleanup |

### 9.5 Acceptance Criteria Mapping
| US/FR | 测试 | 类型 |
|-------|------|------|
| US-003 / FR-3 | `metadata` 解析合法/非法规范 | unit |
| US-004 / FR-4 | `onecli init` 生成可编译 CLI | integration |
| US-006 / FR-6 | `errs` typed envelope 字段稳定性 | unit |
| US-007 / FR-7 | 路径越权/dry-run 净化 | unit |
| US-010 / FR-10 | quality gate stable/experimental 标记 | unit |
| US-011 / FR-11 | lark-cli 迁移后 dry-run E2E 全通过 | e2e |

---

## 10. Implementation Plan

### 10.1 Phases
1. **P1 仓库与工程化骨架**（US-001, US-012）：建仓、双独立 module、Makefile、CI、AGENTS.md、goreleaser、gitleaks
2. **P2 核心移植**（US-002, US-005, US-006, US-007）：runtime/auth/errs/validate/vfs/security 从 lark-cli 提升并去飞书化
3. **P3 元数据与生成器**（US-003, US-004, US-008）：metadata 包 + `onecli init` + schema 命令
4. **P4 Skills 与 Quality gate**（US-009, US-010）：skillgen + qualitygate
5. **P5 Dogfooding 迁移**（US-011）：lark-cli 切换为框架上层，跑回归

### 10.2 Issue Mapping
| Issue | SPEC Sections | Priority | Depends On |
|-------|--------------|----------|------------|
| #1 仓库+工程化骨架 | 2.4, 10.1-P1 | high | — |
| #2 errs 移植 | 3.3, 6.1 | high | #1 |
| #3 runtime 三层 | 2.2, 5.1 | high | #2 |
| #4 metadata 规范+解析 | 3.1, 3.2, 5.4 | high | #1 |
| #5 auth 抽象 | 2.2, 7.1 | high | #2 |
| #6 validate/vfs/security | 2.2, 7.2 | high | #2 |
| #7 onecli init 生成器 | 5.2, 4.1 | high | #3,#4 |
| #8 schema 命令 | 4.1, 4.3 | medium | #3,#4 |
| #9 skillgen | 2.2, 4.1 | medium | #4 |
| #10 qualitygate | 5.3, 4.1 | medium | #3,#4 |
| #11 extension/lint 独立 module | 2.2 | medium | #2 |
| #12 lark-cli 迁移 dogfooding | 10.1-P5 | high | #2~#10 |
| #13 文档+开源就绪 | 1.1, 2.4 docs/ | medium | #7~#10 |

### 10.3 Incremental Delivery
- 每阶段独立可发版：P1 后 `onecli` 可 build；P3 后可生成 petstore 示例
- lark-cli 迁移分块切换（先 auth→errs→runtime→service），每块跑 dry-run E2E 后再切下一块
- feature flag：`x-onecli.quality=experimental` 控制命令可见性，渐进放开

---

## 11. Open Questions & Risks

### 11.1 Unresolved Questions
- GitHub 仓库 owner 已定为 `9Ashwin`，仓库名 `one-cli`，但 **公开仓库由谁/何时用 `gh repo create` 创建**？SPEC 阶段不执行，留待实现 Issue #1
- quality gate 的"Agent 调用成功率"在无真实 Agent 的 CI 中如何度量？SPEC 倾向：以 dry-run 结构校验 + 规则集替代（见 5.3），待确认
- Skills 是否需多 AI 工具格式（Claude skill vs Cursor rule）？SPEC 默认首版统一一套（Claude Code skill 格式），`--format` 预留
- `fetch_meta.py` 在框架中的角色：lark-cli 用它拉飞书 meta；框架无平台 meta，是否保留为"示例规范拉取脚本"或删除？

### 11.2 Technical Risks
| Risk | Impact | Mitigation |
|------|--------|-----------|
| 从 `internal` 提升到导出包后 API 难收敛 | 中 | P2 先定义最小公开接口，内部细节留 unexported |
| lark-cli 迁移期间行为回归 | 高 | 串行迁移 + 每块 dry-run E2E gate（10.3） |
| `kin-openapi` 与 lark-cli 风格不一致 | 低 | 仅用其解析，输出转自有 `MethodMeta` |
| 公开 SDK API 变更破坏已生成 CLI | 高 | 主 module 走 semver；导出包加 `// Deprecated` 策略 |

### 11.3 Assumptions
- GitHub owner `9Ashwin`，公开仓库 `one-cli`，MIT 许可
- Go 1.23+（与 lark-cli 对齐）
- 关键依赖复用：`spf13/cobra`、`tidwall/gjson`、`charmbracelet/huh`+`lipgloss`、`kin-openapi`（新增）
- 工程化全面复刻 lark-cli：AGENTS.md / Makefile / 多层 CI / 自定义 lint / quality-gate / goreleaser / gitleaks
- 迁移节奏串行（P5 在 P2~P4 稳定后）
- `errs/ERROR_CONTRACT.md` wire 字段 wire-stable，直接移植不改名
