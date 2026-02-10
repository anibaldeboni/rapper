# Project Guidelines

## Code Style
- Go formatting and imports are enforced by golangci-lint; follow patterns in [internal/ui/app.go](internal/ui/app.go), [internal/processor/processor.go](internal/processor/processor.go), and [internal/web/gateway.go](internal/web/gateway.go).
- Prefer client-defined interfaces with `//go:generate mockgen` in interface files; examples in [internal/processor/ports.go](internal/processor/ports.go) and [internal/ui/ports/ports.go](internal/ui/ports/ports.go).
- UI follows Bubble Tea Model-Update-View split across [internal/ui/app.go](internal/ui/app.go), [internal/ui/app_update.go](internal/ui/app_update.go), and [internal/ui/app_view.go](internal/ui/app_view.go).

## Architecture
- Main wiring happens in [main.go](main.go): config manager, HTTP gateway, processor, and TUI.
- Core pipeline: CSV processing in [internal/processor/processor.go](internal/processor/processor.go) drives requests via templates in [internal/web/gateway.go](internal/web/gateway.go) and logs via [internal/logs/logs.go](internal/logs/logs.go).
- Config profiles and hot-reload are managed in [internal/config/manager.go](internal/config/manager.go) and [internal/config/profile.go](internal/config/profile.go).

## Build and Test
- Build: `make build` (see [Makefile](Makefile)).
- Dev build with race: `make dev`.
- Tests: `make test` (race + coverage), or single test via `go test -v -run TestName ./internal/package/...` (see [CLAUDE.md](CLAUDE.md)).
- Generate mocks: `make mocks`.
- Lint: `make lint` (config in [.golangci.yml](.golangci.yml)).

## Project Conventions
- UI views live in [internal/ui/views/](internal/ui/views) and are instantiated from [internal/ui/app.go](internal/ui/app.go).
- Templates for URL/body/headers use `text/template` syntax with CSV fields (see [internal/web/gateway.go](internal/web/gateway.go)).
- Config files are YAML; loader and validation behavior in [internal/config/loader.go](internal/config/loader.go).

## Integration Points
- Update checks call GitHub Releases API on startup in [internal/updates/updates.go](internal/updates/updates.go).
- HTTP requests are executed via [internal/web/client.go](internal/web/client.go) and templated in [internal/web/gateway.go](internal/web/gateway.go).

## Security
- HTTP client uses default `http.Client` without explicit timeouts; cancellations rely on context in [internal/web/client.go](internal/web/client.go) and [internal/processor/processor.go](internal/processor/processor.go).
- Output logs may be written to a file with permissions `0660` in [internal/logs/logs.go](internal/logs/logs.go).
