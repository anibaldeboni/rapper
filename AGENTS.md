# AGENTS.md — `rapper`

Operational guide for agents working in this repo. Optimize for signal density: every line should answer "would an agent miss this without help?". Generic Go advice is omitted.

## What this is

Go 1.25 CLI + Bubble Tea **v2** TUI that fires HTTP requests driven by CSV files (URL/body/header templates use Go `text/template` with CSV fields as the data map). Single binary at `./build/rapper`. **There is no Node.js code** — `package.json` and `package-lock.json` exist only to feed `semantic-release` in the release pipeline.

## Quick commands

| Task | Command |
| --- | --- |
| Build (VCS info embedded) | `make build` |
| Dev build with race detector | `make dev` |
| Release-optimized build | `make release` |
| Cross-platform binaries | `make build-all` |
| Run all tests | `make test` |
| Single test | `go test -v -run TestName ./internal/<pkg>/...` |
| Coverage report (opens HTML) | `make test-coverage` |
| Lint | `make lint` (requires `golangci-lint` v2 installed) |
| Regenerate mocks | `make mocks` then `make install_deps` |
| Full sweep | `make all` (lint → test → build) |
| Clean | `make clean` |

CI uses `gotest.tools/gotestsum` for output formatting; otherwise the same `go test` flags.

## Architecture (layered — do not cross without an interface in the owning package)

- `main.go` — wires: `config.Manager` → `web.HttpGateway` → `processor.Processor` → `ui.AppModel`. Reads `-config`, `-dir`, `-output`, `-workers` flags.
- `internal/ui` — Bubble Tea Model-Update-View. Files: `app.go` (model), `app_update.go` (Msg handling), `app_view.go` (rendering). Views live in `internal/ui/views/`; keymaps in `internal/ui/keymaps.go` and `internal/ui/kbind/`; components in `internal/ui/components/`. **Note: imports `charm.land/bubbletea/v2` (not v1) — v1 patterns do not apply.**
- `internal/processor` — CSV → templated HTTP request pipeline. Concurrent worker pool. `MaxWorkers = runtime.NumCPU()` (clamped at runtime via `utils.Clamp`).
- `internal/web` — builds + executes HTTP requests. Applies `text/template` to URL/body/headers. **No explicit `http.Client` timeout — relies on `context` cancellation.**
- `internal/config` — config manager + profiles. Hot-reload via `Manager.OnChange(func(*Config))` rewires the gateway live (profile switch in TUI does not require restart).
- `internal/logs` — in-memory + on-disk logger. File created with `O_WRONLY|O_APPEND|O_CREATE`, perms `0660`. The output filename comes from `-output`.
- `internal/updates` — checks GitHub Releases API on startup; message is shown in usage and exit screens.
- `internal/styles` / `internal/utils` — shared helpers.

## Conventions

- **Hexagonal ports:** define interfaces in the **consuming** package, not the implementing one. Examples: `internal/processor/ports.go` (`HttpGateway`, `RequestLogger`), `internal/ui/ports/ports.go`. Mark with `//go:generate mockgen -destination mock/<name>_mock.go -package mock_<pkg> <iface-path> <IfaceName>`.
- **Mock output dir:** `internal/<pkg>/mock/` (one per package). Always run `make mocks` after editing an interface, then `make install_deps` so `go mod tidy` sees any new dependencies.
- **Templates:** URL/body/header values are Go `text/template` strings. Data map = CSV row keys → cell values. Use `{{.field_name}}`.
- **Configuration schema:** see below — the loader accepts two formats, and which one is in `config.yml` is not what the README example shows.
- **Style:** golangci-lint v2 with `copyloopvar`, `misspell`, `nilerr`, `perfsprint`, `prealloc`, `unconvert`, `unparam` enabled (see `.golangci.yml`). Formatters: `gofmt` + `goimports`.
- **Tests:** colocated in the package they test (no central `tests/` Go code — `tests/example.csv` is a fixture only). Use `testify` and the gomock-generated mocks.

## Config profiles (gotcha)

- `config.yml` at the repo root uses the **legacy** schema (`token:` / `path:` / `payload:`). The loader accepts it for backwards compatibility, but it is the **old** shape.
- The **new** schema is `request:` (with `method`, `url_template`, `body_template`, `headers`) + `csv:` + `workers`. This is what the README and `production.yml` / `staging.yml` / `api1.yml` use.
- New profiles in the config dir become switchable in the TUI via `Ctrl+P`. `config.yml` is the default profile unless the user changes it.

## Verification gates (must pass before merge)

- `make lint` and `make test` are both required.
- Touched an interface → `make mocks` and confirm it compiles.
- Project-level gates (Constitution Check, layer boundaries, cancelable context, etc.) are in `.specify/memory/constitution.md` — that file takes precedence over local conventions.

## Versioning & release

- **Semantic-release** on `master` only (`.releaserc.json`). Conventional Commits drive version bumps.
- **GoReleaser** (`.goreleaser.yml`) builds: linux/amd64, darwin/amd64, darwin/arm64. **linux/arm64 is intentionally ignored.**
- Version is injected at build time via ldflags into `github.com/anibaldeboni/rapper/internal/ui.version`.
- `CHANGELOG.md` is generated and gitignored — do not hand-edit.
- Dependabot updates `go.mod` weekly (`.github/dependabot.yml`).

## Things agents typically miss

- `go 1.25.0` in `go.mod` — match local toolchain; CI uses `setup-go` from `go.mod`.
- HTTP client has **no explicit timeout**; cancellation is via `context`. Any new blocking call must accept and respect `ctx`.
- `internal/logs/logs.go` is **currently in flight** per `git status` (`M internal/logs/logs.go`). Don't treat it as baseline; expect a partially-applied refactor.
- `CLAUDE.md` is staged for deletion (`D CLAUDE.md`); `.github/copilot-instructions.md` is the surviving instruction file for GitHub Copilot and is roughly complementary to this one — keep both aligned.
- `package.json` is **not** a Node.js project manifest — it only declares `semantic-release` for the release workflow. Do not run `npm` locally; let `release.yml` install it.
- `MaxWorkers = runtime.NumCPU()`. The flag help text reflects this dynamically; the README's hardcoded "max: 5" is stale.
- `.specify/`, `openspec/`, `.atl/` are local tooling state (Specify framework, OpenSpec/SDD artifacts, Gentle AI skill-registry cache) — leave them alone unless the user asks.
- `internal/ui/` uses `charm.land/bubbletea/v2`, not v1. v1 idioms (`tea.KeyMsg.String()`, package layout, etc.) differ; do not paste v1 snippets.

## Reference

- `README.md` — user-facing setup, TUI shortcuts, build/test recipes (note: the `max: 5` worker line is stale — see `MaxWorkers` above).
- `.specify/memory/constitution.md` — governing principles. **Takes precedence over local conventions.**
- `.github/copilot-instructions.md` — supplementary style/architecture notes (GitHub Copilot-facing).
- `Makefile` — single source of truth for build/test/lint/install targets.
- `.golangci.yml` — lint configuration.
- `.goreleaser.yml` + `.releaserc.json` — release pipeline.
- `.atl/skill-registry.md` — local Go skill catalog. **Delegator-only** — read paths from it and pass to sub-agents, do not inject summaries.
