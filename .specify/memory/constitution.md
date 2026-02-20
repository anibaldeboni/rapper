<!--
Sync Impact Report
- Version change: N/A → 1.0.0
- Modified principles:
	- Template Principle 1 → I. Layered Architecture Boundaries
	- Template Principle 2 → II. Config-Driven HTTP Contracts
	- Template Principle 3 → III. Verification Gates (NON-NEGOTIABLE)
	- Template Principle 4 → IV. TUI Consistency and Operator Flow
	- Template Principle 5 → V. Operational Safety and Observability
- Added sections:
	- Engineering Constraints
	- Delivery Workflow & Quality Gates
- Removed sections:
	- None
- Templates requiring updates:
	- ✅ updated: .specify/templates/plan-template.md
	- ✅ updated: .specify/templates/spec-template.md
	- ✅ updated: .specify/templates/tasks-template.md
	- ⚠ pending: .specify/templates/commands/*.md (directory not present in repository)
- Follow-up TODOs:
	- None
-->

# Rapper Constitution

## Core Principles

### I. Layered Architecture Boundaries
All changes MUST preserve package responsibilities: `internal/ui` handles presentation and
interaction flow, `internal/processor` orchestrates CSV-driven execution, `internal/web`
builds and sends requests, `internal/config` owns profile/config lifecycle, and
`internal/logs` records outcomes. Cross-layer shortcuts are prohibited unless introduced via
explicit interfaces in the owning package. Rationale: strict boundaries keep the TUI,
processing pipeline, and HTTP transport independently testable and evolvable.

### II. Config-Driven HTTP Contracts
Request behavior MUST be expressed via YAML config and Go `text/template` inputs, not
hard-coded per dataset. URL/body/header templates MUST remain compatible with CSV field
mapping semantics, and profile switching MUST preserve deterministic outcomes across
environments (`dev`, `staging`, `production`). Rationale: the product promise is
configuration-first batch execution.

### III. Verification Gates (NON-NEGOTIABLE)
Every behavior change MUST include or update automated tests in the affected package(s), and
the repository MUST pass `make test` and `make lint` before merge. Interface changes MUST
regenerate mocks via `make mocks` and compile cleanly. Rationale: concurrency, templating,
and TUI state transitions are regression-prone without enforced verification.

### IV. TUI Consistency and Operator Flow
User-facing behavior MUST follow the Bubble Tea Model-Update-View split and existing view
contracts (`files`, `logs`, `settings`, `workers`). Keybindings, status messaging, and
navigation semantics MUST remain consistent unless a deliberate UX change is specified.
Rationale: operators rely on predictable keyboard-driven workflows during high-volume runs.

### V. Operational Safety and Observability
Long-running work MUST remain cancelable via context propagation, and request outcomes MUST
be observable through metrics/log outputs. New network behavior MUST define timeout/cancel
expectations and failure handling paths. Rationale: safe interruption and traceable outcomes
are critical for batch HTTP tooling.

### VI. Code duplication
Code duplication is prohibited. You must search for existing functions before creating a 
new one. If no reusable functions exist, a new one may be introduced.

## Engineering Constraints

- Language and tooling baseline MUST remain Go + existing Makefile workflows (`make build`,
	`make dev`, `make test`, `make lint`, `make mocks`) unless explicitly amended.
- Dependency injection through package-local interfaces is REQUIRED for components that cross
	package boundaries.
- Configuration files MUST stay YAML-compatible with profile management in `internal/config`.
- Security-sensitive changes MUST document implications for HTTP client behavior, credential
	handling in headers, and log output permissions.

## Delivery Workflow & Quality Gates

Work proceeds as specification → plan → tasks → implementation. Plans MUST include a
Constitution Check that verifies architecture boundaries, test/lint strategy, and config
compatibility. Task lists MUST map work to user stories and include explicit verification
tasks for behavior changes. Pull requests MUST describe impacted layers and how rollback or
safe cancellation behaves if execution fails.

## Governance
This constitution takes precedence over local conventions when conflicts arise. Amendments
require: (1) a documented proposal, (2) updates to impacted templates under `.specify`,
and (3) review approval by maintainers.

Versioning policy:
- MAJOR: incompatible redefinition/removal of principles or governance guarantees.
- MINOR: new principle/section or materially expanded mandatory guidance.
- PATCH: clarifications, wording improvements, typo fixes, non-semantic edits.

Compliance review is REQUIRED in planning and code review: each plan MUST pass Constitution
Check gates, and each PR MUST state how the change satisfies applicable principles.

Operational guidance sources are `README.md` and
`.github/copilot-instructions.md`; these MUST remain aligned with this constitution.

**Version**: 1.0.0 | **Ratified**: 2026-02-20 | **Last Amended**: 2026-02-20
