# Agent Guide for dnsres

This file is a concise runbook for agentic coding tools working in this
repository. Keep changes consistent with existing patterns unless you have
explicit direction to refactor.

## Rules Sources
- Cursor rules: none found in `.cursor/rules/` or `.cursorrules`.
- Copilot rules: none found in `.github/copilot-instructions.md`.

## Quick Facts
- Language: Go 1.24.0 (see `go.mod`).
- Binary names: `dnsres` (main CLI), `dnsres-tui` (interactive TUI).
- Configuration: JSON file (default `config.json`).
- Logging: separate success, error, and app logs in `log_dir`.

## Build / Lint / Test
Use the Makefile targets first; they encode repo-specific behavior.

### Build
- `make build` (builds static binary; `CGO_ENABLED=0` enforced)
- `make build-tui` (builds interactive TUI binary)
- `make build-all` (cross-compile to multiple OS/arch)
- `make release` (build all + archive artifacts)

### Lint / Format / Vet
- `make lint` (requires `golangci-lint`)
- `make fmt` (`go fmt ./...`)
- `make vet` (`go vet ./...`)

### Tests
- `make test` (runs `go test -v ./...`)
- `make coverage` (generates coverage report)

### Running a Single Test
The repository contains 13 test files across multiple packages:
- Unit tests: `cache/`, `circuitbreaker/`, `dnsanalysis/`, `dnspool/`, `health/`, `metrics/`, `instrumentation/`
- Package tests: `internal/dnsres/` (resolver, cycle, loop tests)
- Integration tests: `internal/integration/` (E2E tests with build tag)

When running tests, use standard Go patterns:
- `go test ./... -run TestName`
- `go test ./path/to/pkg -run TestName`
- `go test ./path/to/pkg -run TestName -count=1` (avoid cached results)
- `go test ./path/to/pkg -run TestName/Subcase` (table-driven subtests)

### Running Integration Tests
Integration tests use the `//go:build integration` build tag and require the full binary to be built:
- `go test -tags=integration ./internal/integration -v`
- `go test -tags=integration -short ./internal/integration` (skip in short mode)
- Integration tests build `cmd/dnsres`, create temp configs, and validate end-to-end behavior

### Docker
- `make docker-build`
- `make docker-run`

### Dependencies / Mocks
- `make deps` (installs `golangci-lint`, `mockgen`)
- `make mocks` (note: target paths reference older `internal/...` layout)

## Repository Layout (Current)

### Command Binaries
- `cmd/dnsres/`: main CLI application entrypoint (thin wrapper)
- `cmd/dnsres-tui/`: interactive TUI application entrypoint

### Internal Packages (unexported, module-private)
- `internal/app/`: application runtime, flag parsing, orchestration (`run.go`)
- `internal/dnsres/`: core DNS resolver logic
  - `resolver.go`: DNSResolver implementation, Start loop
  - `config.go`: configuration loading and validation
  - `events.go`: event bus for TUI integration
  - `logging.go`: log file setup
  - `report.go`: statistics reporting
- `internal/tui/`: TUI application (Bubble Tea framework)
  - `model.go`: TUI state and update logic
  - `run.go`: TUI initialization and event subscription
  - `theme.go`: color schemes and styling
- `internal/integration/`: end-to-end integration tests (requires `integration` build tag)

### Public Packages (root-level, importable by external code)
- `cache/`: sharded cache implementation with expiration
- `circuitbreaker/`: circuit breaker pattern with metrics integration
- `dnspool/`: DNS client pooling for connection reuse
- `dnsanalysis/`: DNS response analysis and consistency checking
- `health/`: HTTP health check endpoint
- `metrics/`: Prometheus metric definitions and registration
- `instrumentation/`: debug instrumentation level parsing (`none`, `low`, `medium`, `high`, `critical`)

### Other
- `docs/`: development guide, API documentation
- `examples/`: example configuration files (`config.json`)
- `logs/`: default log output directory (created at runtime)

## Code Style Guidelines
Follow standard Go style plus the conventions below, inferred from the code.

### Formatting
- Always run `gofmt` (or `make fmt`) before committing.
- Use tabs for indentation and keep line length reasonable.

### Imports
- Group imports as: standard library, internal packages (`dnsres/internal/...`), public packages (`dnsres/cache`, `dnsres/metrics`, etc.), third-party.
- Keep groups separated by a blank line.
- Internal packages are only accessible within this module.
- Public packages at root level can be imported by external projects.

### Naming
- Exported identifiers: PascalCase (e.g., `DNSResolver`, `HealthChecker`).
- Local identifiers: camelCase.
- Acronyms follow Go conventions (`DNS`, `HTTP`, `TTL`, `EDNS`).
- Avoid overly generic names; prefer clarity over brevity.

### Types and Structs
- Prefer concrete types; avoid interface{} unless required.
- Use `time.Duration` for time values; the custom `Duration` wrapper handles
  JSON parsing for configuration.
- Keep config structs in sync with `config.json` and `README.md` examples.

### Error Handling
- Return errors early with context: `fmt.Errorf("...: %w", err)`.
- Use `errors.Is`/`errors.As` for sentinel or wrapped error checks.
- Prefer explicit error messages over silent fallbacks.

### Logging
- Use the three logger streams created in `setupLoggers`:
  - `dnsres-success.log` for successful resolutions.
  - `dnsres-error.log` for failures and inconsistencies.
  - `dnsres-app.log` for startup, health, and server lifecycle.
- Keep log lines short and actionable; avoid dumping large structs.

### Metrics
- Metrics live in `metrics/metrics.go`. Add new metrics there.
- Use label sets that match existing label usage and ordering.
- Avoid updating metrics while holding locks when possible.

### Concurrency
- Protect shared state with `sync.Mutex`/`sync.RWMutex`.
- Keep lock scopes minimal; avoid calling metrics or I/O while locked.
- Use wait groups and semaphores for concurrency limits.

### DNS and Networking Conventions
- DNS servers should include ports; if missing, append `:53`.
- Use `context.Context` for request lifetimes and cancellation.
- Prefer pooled clients from `dnspool` where available.

### Event System
- The resolver publishes events via an internal event bus (`internal/dnsres/events.go`).
- Event types: `cycle_start`, `cycle_complete`, `resolve_success`, `resolve_failure`, `inconsistent`.
- Subscribers receive `ResolverEvent` structs with metadata (time, hostname, server, duration, error, etc.).
- Used primarily by the TUI for real-time updates.
- Subscribe with `eventBus.subscribe(bufferSize)` and cleanup with returned unsubscribe function.
- Events are non-blocking; slow consumers won't block the resolver.

### Instrumentation Levels
- Configurable debug instrumentation via `instrumentation_level` config field.
- Levels: `none` (default), `low`, `medium`, `high`, `critical`.
- Parsed by `instrumentation.ParseLevel()` and returned as `instrumentation.Level` enum.
- Higher levels emit more detailed diagnostic information.
- Used for troubleshooting DNS resolution issues in production.

### Configuration
- Validate config values before use (`Config.Validate` / `validateConfig`).
- Use clear error messages when config is invalid.
- Preserve backward compatibility in config fields when possible.

### Testing Style (When Added)
- Prefer table-driven tests with `t.Run` subtests.
- Cover success and failure paths; assert error text sparingly.
- Use `-count=1` for nondeterministic tests or when debugging.

### Documentation
- Update `README.md` when behavior or config changes.
- Keep `docs/DEVELOPMENT.md` aligned with actual repo layout.

## Agent Workflow Notes
- Keep changes minimal, scoped, and consistent with current structure.
- Avoid refactoring path layout unless explicitly requested.
- If adding new dependencies, update `go.mod` and `go.sum`.
- Run `make fmt` and targeted tests when modifying logic.

## Common Entry Points
- CLI flags live in `internal/app/run.go` (`-config`, `-report`, `-host`).
- Main loop starts in `DNSResolver.Start` (`internal/dnsres/resolver.go`).
- Health endpoint is served by `health.HealthChecker`.
- Metrics are served via Prometheus in `internal/app/run.go`.

## Testing Guidelines

### Unit Tests
- Test files co-locate with implementation: `cache/sharded_test.go`, `health/health_test.go`, etc.
- Use table-driven tests with `t.Run()` subtests.
- Mock external dependencies where needed (DNS servers, time, etc.).

### Integration Tests
- Located in `internal/integration/` with `//go:build integration` tag.
- Run with: `go test -tags=integration ./internal/integration -v`.
- Tests build the actual binary, create temp configs, and validate E2E behavior.
- Skip in `-short` mode: `if testing.Short() { t.Skip(...) }`.
- Require longer timeouts and real DNS servers.

### Test Execution
- Unit tests only: `make test` or `go test ./...`.
- With integration: `go test -tags=integration ./...`.
- Avoid cached results: `go test -count=1 ./path/to/pkg`.
- Coverage report: `make coverage`.

## Non-Goals for Agents
- Do not change log formats unless requested.
- Do not remove `CGO_ENABLED=0` from build steps.
- Do not introduce new config keys without updating docs.
