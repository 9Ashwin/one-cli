# one-cli

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.23-blue.svg)](https://go.dev/)

[中文版](./README.zh.md) | [English](./README.md)

一个 **Agent-Native CLI 生成框架** —— 从 OpenAPI 3.x 规范脚手架生成对 AI Agent 友好的 CLI。
抽取自 [`lark-cli`](https://github.com/larksuite/cli)（飞书/Lark 官方 CLI）的工程实践，泛化后
让任何拥有 REST API 的平台都能在几分钟内产出一个 Agent 友好的 CLI。

## 为什么选 one-cli？

- **为 Agent 原生设计** —— 结构化 stderr 错误信封、schema 自省、自动生成 AI Skills，Agent
  无需额外适配即可解析输出、从错误中恢复。
- **规范驱动** —— 喂一份 OpenAPI 3.x 规范（带 `x-onecli-*` 扩展），得到可运行的 CLI。
- **三层命令** —— 友好快捷命令（`+list`）→ 元数据驱动 API 命令 → 通用 `api METHOD /path`
  调用，按需选择粒度。
- **可插拔认证** —— OAuth2 device flow、OS 密钥链存储、多 profile、身份切换。
- **默认安全** —— 输入路径校验、文件系统抽象、副作用 dry-run、输出净化（CLI 参数来自 AI
  Agent，视为不可信）。
- **质量门禁** —— 只有经 Agent 实测的命令进入 CLI 的 stable 表面。

## 与 lark-cli 的关系

`one-cli` 泛化了 `lark-cli` 的平台无关核心。`lark-cli` 自身正在基于本框架重构（dogfooding）
——其飞书专属层（scope 目录、飞书 OpenAPI 元数据、keychain 命名空间）保留在 `lark-cli`，
所有可复用部分都在这里。

## 状态

🚧 **开发中**。框架按 issue 逐步构建，生成器命令（`init`、`skill gen`、`quality check`）
渐进落地。PRD 与技术 SPEC 见 [`tasks/`](./tasks)。

## 快速开始（`onecli init` 落地后）

```bash
go install github.com/9Ashwin/one-cli@latest

# 从 OpenAPI 规范生成 CLI
onecli init --spec examples/petstore/openapi.json --name petstore-cli

cd petstore-cli
go build -o petstore-cli .
./petstore-cli --help
./petstore-cli api GET /pets --dry-run
```

## 目录结构

完整源码布局与贡献规范见 [AGENTS.md](./AGENTS.md)。

## 许可证

MIT，见 [LICENSE](./LICENSE)。
