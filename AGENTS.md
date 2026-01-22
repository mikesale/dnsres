# Agent Guide for dnsres

This file is a concise runbook for agentic coding tools working in this
repository. Keep changes consistent with existing patterns unless you have
explicit direction to refactor.

## Rules Sources
- Cursor rules: none found in `.cursor/rules/` or `.cursorrules`.
- Copilot rules: none found in `.github/copilot-instructions.md`.

## Quick Facts
- Language: Go 1.21 (see `go.mod`).
- Binary name: `dnsres`.
- Configuration: JSON file (default `config.json`).
- Logging: separate success, error, and app logs in `log_dir`.

## Build / Lint / Test
Use the Makefile targets first; they encode repo-specific behavior.

### Build
- `make build` (builds static binary; `CGO_ENABLED=0` enforced)
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
There are currently no `*_test.go` files in the repo. When adding tests, use
standard Go patterns:
- `go test ./... -run TestName`
- `go test ./path/to/pkg -run TestName`
- `go test ./path/to/pkg -run TestName -count=1` (avoid cached results)
- `go test ./path/to/pkg -run TestName/Subcase` (table-driven subtests)

### Docker
- `make docker-build`
- `make docker-run`

### Dependencies / Mocks
- `make deps` (installs `golangci-lint`, `mockgen`)
- `make mocks` (note: target paths reference older `internal/...` layout)

## Repository Layout (Current)
- `dnsres.go`: main entrypoint and core orchestration
- `cache/`: sharded cache implementation
- `circuitbreaker/`: circuit breaker logic
- `dnspool/`: DNS client pooling
- `dnsanalysis/`: response analysis and consistency checks
- `health/`: health check HTTP handler
- `metrics/`: Prometheus metric definitions
- `docs/`: development and API notes

## Code Style Guidelines
Follow standard Go style plus the conventions below, inferred from the code.

### Formatting
- Always run `gofmt` (or `make fmt`) before committing.
- Use tabs for indentation and keep line length reasonable.

### Imports
- Group imports as: standard library, local module (`dnsres/...`), third-party.
- Keep groups separated by a blank line.

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
- CLI flags live in `dnsres.go` (`-config`, `-report`, `-host`).
- Main loop starts in `DNSResolver.Start`.
- Health endpoint is served by `health.HealthChecker`.
- Metrics are served via Prometheus in `dnsres.go`.

## Non-Goals for Agents
- Do not change log formats unless requested.
- Do not remove `CGO_ENABLED=0` from build steps.
- Do not introduce new config keys without updating docs.
