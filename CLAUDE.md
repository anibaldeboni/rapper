# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is Rapper

Rapper is a Go TUI application for batch HTTP requests driven by CSV data. It uses the Bubbletea framework (Elm Architecture: Model → Update → View) with Lipgloss styling and Bubbles components.

## Build & Development Commands

```bash
make build          # Build to ./build/rapper (embeds VCS info)
make dev            # Build with race detector
make run            # Build and run
make test           # Run all tests: go test -v -cover -race -timeout 30s ./...
make test-coverage  # Generate HTML coverage report at ./build/coverage.html
make lint           # Run golangci-lint
make mocks          # Regenerate mocks via go generate ./...
make clean          # Remove build artifacts
make install_deps   # Download Go dependencies
```

Run a single test:
```bash
go test -v -run TestName ./internal/package/...
```

## Architecture

### Core Flow

CSV file → Processor (worker pool) → Web Gateway (templated HTTP requests) → Logs

### Key Packages (`internal/`)

- **ui/**: Bubbletea TUI app split into `app.go` (model), `app_update.go` (update logic), `app_view.go` (rendering). Four views under `ui/views/`: files, logs, settings, workers. Navigation via F1-F4 keys.
- **processor/**: Worker pool with dynamic sizing (1 to CPU count). Uses goroutines with `sync.WaitGroup`, context cancellation, and atomic counters for metrics.
- **web/**: `gateway.go` builds HTTP requests using Go `text/template` with CSV column values. `client.go` is the HTTP client abstraction.
- **config/**: YAML config with hot-reload (`manager.go` observer pattern), multi-profile support (`profile.go`), and legacy format auto-conversion.
- **logs/**: Logger interface for structured request/response logging.
- **updates/**: Checks GitHub API for new releases at startup.

### Patterns

- **Dependency injection** via interfaces — all major components have interfaces with generated mocks (`internal/*/mock/`).
- **Template-based requests**: URL, body, and headers use `text/template` syntax with CSV fields as variables (e.g., `{{.column_name}}`).
- **Config hot-reload**: `Manager.OnChange()` registers callbacks; config changes propagate without restart.

## Testing

- Uses `stretchr/testify` for assertions and `go.uber.org/mock` for mock generation.
- Mocks are generated from `//go:generate` directives in interface files. Run `make mocks` after changing interfaces.
- Tests exist for all packages under `internal/`.

## Linting

Configured in `.golangci.yml`. Notable enabled linters: `copyloopvar`, `misspell`, `nilerr`, `perfsprint`, `prealloc`, `unconvert`, `unparam`. Formatters: `gofmt`, `goimports`.

## Configuration Format

```yaml
request:
  method: POST
  url_template: "https://api.example.com/{{.id}}"
  body_template: '{"name": "{{.name}}"}'
  headers:
    Authorization: "Bearer {{.token}}"
    Content-Type: "application/json"

csv:
  fields: [id, name, token]
  separator: ","

workers: 4
```

Multi-profile: place multiple `.yml` files in the config directory and switch between them in the settings view.
