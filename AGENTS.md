# AGENTS.md

## Goal (pick one per PR)

- Make the framework better: improve UX, error messages, help text, flags, and output clarity.
- Improve reliability: fix bugs, edge cases, and regressions with tests.
- Improve developer velocity: simplify code paths, reduce complexity, keep behavior explicit.
- Improve quality gates: strengthen tests/lint/checks without adding heavy process.

## Build & Test

```bash
make build          # Build the onecli generator binary
make unit-test      # Required before PR (runs with -race where supported)
make test           # Full: vet + fmt-check + unit + examples-build + integration
make examples-build # Keep extension/ and lint/ independent modules compilable
```

## Pre-PR Checks (match CI gates)

1. `make unit-test`
2. `go vet ./...`
3. `gofmt -l .` — must produce no output
4. `go mod tidy` — must not change `go.mod`/`go.sum`
5. `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 run --new-from-rev=origin/main`
6. If dependencies changed: `go run github.com/google/go-licenses/v2@v2.0.1 check ./... --disallowed_types=forbidden,restricted,reciprocal,unknown`

## Source Layout

| Path | What it does |
|------|-------------|
| `main.go` | Entry point for the `onecli` generator binary |
| `cmd/onecli/` | Generator subcommands: `init`, `skill gen`, `quality check` |
| `runtime/` | Three-tier command runtime (shortcut / API / generic call) |
| `metadata/` | OpenAPI 3.x + `x-onecli-*` parsing into `ServiceMeta`/`MethodMeta` |
| `auth/` | Pluggable credential providers, OAuth2 device flow, keychain, multi-profile |
| `errs/` | Typed, RFC 7807-aligned error taxonomy (see `errs/ERROR_CONTRACT.md`) |
| `validate/` | Path safety validation (`SafeInputPath`/`SafeOutputPath`) |
| `vfs/` | Filesystem abstraction (use `vfs.*` instead of `os.*`) |
| `security/` | Output sanitization, anti-injection |
| `schema/` | Schema introspection command |
| `qualitygate/` | Command admission harness (manifest export + check) |
| `skillgen/` | AI Agent Skills generation (master + domain + shared) |
| `output/` | Output formats (json/pretty/table/ndjson/csv) + envelope |
| `client/` | `APIClient`: HTTP transport, `DoSDKRequest`/`DoStream` |
| `internal/build/` | Version metadata injected via -ldflags |
| `extension/` | Independent module — public plugin SDK |
| `lint/` | Independent module — custom golangci-lint analyzers (`errscontract`) |
| `examples/` | Reference specs and generated CLI samples (e.g. petstore) |
| `tests/cli_e2e/` | Dry-run and live E2E tests |

## Who Uses This Framework

The generated CLIs' primary consumers include AI agents (Claude Code, Cursor, Gemini CLI). Code
in the exported packages is read by machines — error messages, output format, and flag design
all directly affect agent success rates.

The one rule to internalize: **every error message you write will be parsed by an AI to decide
its next action.** Make errors structured, actionable, and specific.

## Code Conventions

### Structured errors

Command-facing failures must be typed `errs.*` errors — never a final bare `fmt.Errorf`. AI
agents parse the stderr envelope's `type` / `subtype` / `param` / `hint` fields to decide their
next action; the full taxonomy lives in `errs/ERROR_CONTRACT.md`.

Picking a constructor (populated by Issue #2; the table below is the contract all later issues
follow):

| Failure | Constructor |
|---------|-------------|
| User flag/arg fails validation | `errs.NewValidationError(errs.SubtypeInvalidArgument, ...).WithParam("--flag")` |
| Valid request, wrong system state | `errs.NewValidationError(errs.SubtypeFailedPrecondition, ...).WithHint(...)` |
| Upstream API returned an error code | `errclass.BuildAPIError` (never hand-build) |
| Network / transport failure | `errs.NewNetworkError(errs.SubtypeNetworkTransport, ...)` |
| Local file I/O failure | `errs.NewInternalError(errs.SubtypeFileIO, ...)` — validate the path first (`validate.SafeInputPath` / `SafeOutputPath`) and use `vfs.*` |
| Unclassified lower-layer error as final | `errs.NewInternalError(errs.SubtypeUnknown, ...).WithCause(err)` |
| Lower layer already returned a typed error | pass it through unchanged — re-wrapping downgrades its classification |

### stdout is data, stderr is everything else

Program output (JSON envelopes) goes to stdout. Progress, warnings, hints go to stderr. Mixing
them corrupts pipe chains.

### Use `vfs.*` instead of `os.*`

All filesystem access goes through the `vfs` package. This enables test mocking.

### Validate paths before reading

CLI arguments are untrusted (they come from AI agents). Call `validate.SafeInputPath` before any
file I/O.

### Tests

- Every behavior change needs a test alongside the change.
- Isolate config state with `t.Setenv("ONECLI_CONFIG_DIR", t.TempDir())`.

### E2E Testing

**Dry-run E2E (required for every command change)**
- Validates request structure without calling real APIs
- Place in `tests/cli_e2e/dryrun/`
- Use `--dry-run`, assert method/URL/params
- No secrets needed — runs on fork PRs

**Live E2E (required for new flows or behavior changes)**
- Validates real API round-trips
- Place in `tests/cli_e2e/<domain>/`
- Must be self-contained: create -> use -> cleanup

## Commit & PR

- Conventional Commits in English: `feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`, `ci:`
- PR title in the same format. Fill `.github/pull_request_template.md` completely.
- Never commit secrets, tokens, or internal sensitive data.
