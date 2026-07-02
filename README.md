# one-cli

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.23-blue.svg)](https://go.dev/)

[中文版](./README.zh.md) | [English](./README.md)

An **Agent-Native CLI generation framework** — scaffold an AI-agent-friendly CLI from an OpenAPI 3.x
spec. Extracted from the engineering practices of
[`lark-cli`](https://github.com/larksuite/cli) (Feishu/Lark's official CLI), generalized so any
platform with a REST API can ship an agent-friendly CLI in minutes.

## Why one-cli?

- **Agent-native by design** — structured stderr error envelopes, schema introspection, and
  auto-generated AI Skills, so agents parse output and recover from errors without extra adapters.
- **Spec-driven** — feed an OpenAPI 3.x spec (with `x-onecli-*` extensions), get a runnable CLI.
- **Three-tier commands** — friendly shortcuts (`+list`) → metadata-driven API commands → generic
  `api METHOD /path` calls, choose the right granularity per task.
- **Pluggable auth** — OAuth2 device flow, OS keychain storage, multi-profile, identity switching.
- **Safe by default** — input path validation, filesystem abstraction, dry-run for side effects,
  output sanitization — because CLI args come from AI agents.
- **Quality gates** — only agent-tested commands reach a CLI's stable surface.

## Relationship to lark-cli

`one-cli` generalizes `lark-cli`'s platform-agnostic core. `lark-cli` itself is being rebuilt on
top of this framework (dogfooding) — its Feishu-specific layer (scope catalog, Feishu OpenAPI
metadata, keychain namespace) stays in `lark-cli`; everything reusable lives here.

## Status

🚧 **Under active development.** The framework is being built issue-by-issue; the generator
commands (`init`, `skill gen`, `quality check`) land progressively. See
[`tasks/`](./tasks) for the PRD and technical SPEC.

## Quick start (once `onecli init` lands)

```bash
go install github.com/9Ashwin/one-cli@latest

# Scaffold a CLI from an OpenAPI spec
onecli init --spec examples/petstore/openapi.json --name petstore-cli

cd petstore-cli
go build -o petstore-cli .
./petstore-cli --help
./petstore-cli api GET /pets --dry-run
```

## Layout

See [AGENTS.md](./AGENTS.md) for the full source layout and contribution conventions.

## License

MIT. See [LICENSE](./LICENSE).
