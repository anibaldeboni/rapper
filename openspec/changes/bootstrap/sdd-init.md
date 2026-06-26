# sdd-init — rapper

- project: rapper
- workspace: /Users/anibal.neto/Documents/anibal/rapper
- detected_at: 2026-06-26T15:20:00-03:00
- artifact_store: openspec
- strict_tdd: true
- test_command: go test -v -cover -race -timeout 30s ./...

## Stack

- **Language**: Go 1.25.0
- **Module**: `github.com/anibaldeboni/rapper`
- **Framework**: Bubbletea v2 (Elm Architecture: Model → Update → View)
- **UI styling**: Lipgloss v2, Bubbles v2
- **Dependencies**: `stretchr/testify` (testing), `go.uber.org/mock` (mocks), `gopkg.in/yaml.v3` (config), `charm.land/*` (TUI)
- **Build system**: Makefile with `make build`, `make test`, `make lint`, `make mocks`, `make run`
- **Release**: GoReleaser (`.goreleaser.yml`) + semantic-release (`.releaserc.json`)
- **Runtime**: TUI application, CLI flags for config/dir/output/workers, hot-reload config

## Conventions

- **Linting**: golangci-lint v2 (`.golangci.yml`) with `gofmt` + `goimports` formatters; linters: `copyloopvar`, `misspell`, `nilerr`, `perfsprint`, `prealloc`, `unconvert`, `unparam`
- **Mock generation**: `//go:generate mockgen` directives in interface files under `internal/`; mocks land in `internal/<pkg>/mock/`; run via `make mocks` (`go generate ./...`)
- **Commit style**: Conventional commits (`feat:`, `fix:`, `refactor:`, `chore:`, `docs:`, `chore(deps):` for Dependabot)
- **Test naming**: Descriptive `TestXxx_Yyy` or `When ...` pattern using `t.Run()` sub-tests
- **Test packages**: External test packages (`package <pkg>_test`) for black-box testing; internal (`package <pkg>`) for white-box where needed
- **Directories**: Single `main.go` root entry point, all library code under `internal/`

## Architecture patterns

- **Layout**: Standard Go single-module layout — `main.go` at root, all packages under `internal/`
- **Packages**: `internal/config/` (YAML config with hot-reload, multi-profile), `internal/web/` (HTTP gateway + client), `internal/processor/` (worker pool), `internal/ui/` (Bubbletea app with Model/Update/View split), `internal/logs/` (structured logging), `internal/updates/` (GitHub release checker), `internal/styles/` (Lipgloss themes), `internal/utils/` (helpers)
- **Dependency injection**: Constructor functions (`New*`) accepting interfaces as parameters; interfaces defined in the consumer package (processor ports, UI ports)
- **Port/adapter**: `internal/ui/ports/` defines interfaces that UI components depend on — `ConfigManager`, `ConfigProvider`, `ProcessorController`, `LogService`, etc.
- **Mock pattern**: Each interface has a corresponding `//go:generate mockgen` directive and a `mock/` subdirectory with generated mocks
- **UI**: Bubbletea Model-Update-View split across `app.go`, `app_update.go`, `app_view.go`; views under `internal/ui/views/`
- **Wiring**: Done in `main.go` — constructs config manager, HTTP gateway, processor, and UI app, then starts the Bubbletea program

## Testing capability

- **Test runner**: `go test` via `make test` (`go test -v -cover -race -timeout 30s ./...`)
- **Framework**: `stretchr/testify` (assertions) + `go.uber.org/mock/gomock` (mocking)
- **Coverage**: Available via `make test-coverage` (generates HTML report at `build/coverage.html`)
- **Race detector**: Enabled in all test commands (`-race`)
- **CI**: GitHub Actions — tests (`gotestsum`) and lint (`golangci-lint`) on push/PR to master; matrix: ubuntu-latest + macos-latest
- **Test files**: 8 test files across packages: `logs/`, `updates/`, `utils/`, `processor/`, `web/` (client + gateway), `ui/`, `ui/assets/`
- **Test layers**: Unit tests only — no integration, E2E, or snapshot/golden tests detected
- **Missing coverage**: `internal/config/` has no test files; `internal/styles/` has no tests
- **Bubbletea tests**: Using `tea.NewProgram` with mock I/O in `ui/ui_test.go` for integration-style TUI testing

| Layer       | Available | Tool |
| ----------- | --------- | ---- |
| Unit        | ✅        | `go test` + testify + gomock |
| Integration | ❌        | — |
| E2E         | ❌        | — |

| Tool         | Available | Command |
| ------------ | --------- | ------- |
| Linter       | ✅        | `golangci-lint run -v` |
| Type checker | partial   | `go vet` (no dedicated config, implicit via `go test`) |
| Formatter    | ✅        | `gofmt` + `goimports` (enforced via golangci-lint) |

## Strict TDD evidence

**Verdict: strict_tdd: true** — based on default fallback (test runner exists, no explicit TDD marker found).

The project demonstrates strong testing discipline:
- Tests exist alongside implementation in 6 of 8 internal packages
- `make test` (with race + coverage) is the default test command and runs as part of `make all`
- GitHub Actions CI runs `gotestsum` on every push/PR to master (two OS matrix)
- Mocks are pre-generated via `//go:generate` directives and committed alongside interfaces
- The Bubbletea UI tests use `gomock.NewController(t)` with mocked port interfaces and `tea.NewProgram` for integration-level TUI testing

Limitations:
- No explicit TDD marker/config found in project files
- No pre-commit/pre-push hooks installed (only git sample hooks present)
- `internal/config/` and `internal/styles/` have no test files
- The test suite has no integration or E2E coverage

## Detection commands run

- `read go.mod` — module identity, dependencies, Go version
- `read Makefile` — build, test, lint, mock generation commands
- `read .golangci.yml` — linter and formatter configuration
- `read internal/*/ports.go` — interface definitions and `//go:generate` directives
- `read internal/*/*_test.go` — test patterns, framework usage, coverage
- `read .github/workflows/ci.yml` — CI pipeline configuration
- `read .github/copilot-instructions.md` — project conventions document
- `ls .git/hooks/` — checked for pre-commit/push hooks (only samples)
- `go test -list '.*' ./internal/logs/...` — verified test infrastructure compiles and runs
- `go version` — confirmed Go runtime
- `git log --oneline -20` — commit convention analysis

## Next recommended action

`sdd-explore` — explore improvements to the project (e.g., add tests for `internal/config/` and `internal/styles/`, introduce integration/E2E testing, or plan a new feature) before committing to any change.
