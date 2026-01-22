# Architecture Overview

This document describes how `dnsres` is structured, how control and data flow
through the system, and how the core components interact. It is intended for
new contributors who need a clear mental model before making changes.

## Purpose and Scope

`dnsres` is a DNS resolution monitor. It periodically resolves a configured set
of hostnames against a configured set of DNS servers, logs results, updates
metrics, and exposes health/metrics HTTP endpoints. It is configured by JSON
and runs as a long-lived process. The period is defined by the config.json parameter config.query_interval.

## High-Level Flow

At runtime the application follows this lifecycle:

1. Parse CLI flags and load configuration.
2. Validate configuration and normalize DNS server addresses (ensure `:53`).
3. Initialize core components (loggers, client pool, circuit breakers, cache,
   health checker, metrics).
4. Start HTTP servers for health and Prometheus metrics.
5. Start the resolution loop that continuously queries DNS servers.
6. On shutdown signals, gracefully stop HTTP servers and exit.

## Entry Point and Initialization

The main entrypoint is `dnsres.go`.

### CLI
- Flags: `-config`, `-report`, `-host`.
- `-report` switches to report-only mode and prints statistics.
- `-host` overrides the `hostnames` in config for ad-hoc checks.

### Config Loading
- `loadConfig` reads JSON and decodes into `Config`.
- `Duration` is a wrapper around `time.Duration` to support strings like
  "5s"/"1m" in JSON.
- DNS servers are normalized to include ports using `net.SplitHostPort` and
  `net.JoinHostPort` (default `:53`).
- Validation occurs via `Config.Validate` and `validateConfig`.

### Logging Setup
`setupLoggers` creates three independent `log.Logger` instances, each bound to
a separate file within `log_dir`:
- `dnsres-success.log`
- `dnsres-error.log`
- `dnsres-app.log`

These are used consistently across the system to separate concerns.

## Core Components

### DNSResolver (orchestrator)
`DNSResolver` owns and coordinates the system:
- `config` (validated configuration)
- `clientPool` (`dnspool.ClientPool`)
- `breakers` (`circuitbreaker.CircuitBreaker` per server)
- `cache` (`cache.ShardedCache`)
- `health` (`health.HealthChecker`)
- `successLog`, `errorLog`, `appLog`
- `stats` (in-memory counters used by report mode)

Creation: `NewDNSResolver` sets up all dependencies and seeds per-server stats.

### Client Pool (`dnspool`)
The client pool reuses `*dns.Client` instances keyed by server address:
- Limits pool size per server.
- Applies per-request timeout from configuration.
- Records protocol metrics for pooled/new/returned/dropped usage.

### Circuit Breaker (`circuitbreaker`)
Each DNS server has its own circuit breaker that tracks failures:
- States: Closed, Open, Half-Open.
- `Allow` guards requests and updates state metrics.
- `RecordSuccess`/`RecordFailure` update failure counts and metrics.

### Cache (`cache`)
The sharded cache stores `DNSResponse` values:
- Sharded map for concurrency (`CacheShard` uses `sync.RWMutex`).
- TTL-based expiration on read.
- Eviction is size-based per shard.
- Metrics track cache hits, misses, evictions, and size.

### Health Checker (`health`)
Health checks are a TCP connectivity probe to DNS servers:
- Runs on a timer and updates a per-server status map.
- Exposes `/` returning "healthy" or "unhealthy".
- Updates DNS metrics with health check outcomes.

### Metrics (`metrics`)
Prometheus metrics are defined in a dedicated package:
- Counters, gauges, histograms for resolution, cache, circuit breaker, health.
- Metrics are updated throughout the resolver workflow.

## Resolution Workflow (Data Flow)

The resolution loop runs in `DNSResolver.Start` and `resolveAll`.

1. **Start loop:**
   - `Start` launches health and metrics HTTP servers.
   - A ticker triggers periodic resolution, with an immediate initial run.

2. **Hostnames fan-out:**
   - `resolveAll` creates a goroutine per hostname.
   - A semaphore (`chan struct{}`) caps concurrent hostname resolution.

3. **Servers fan-out:**
   - For each hostname, it queries all DNS servers concurrently.

4. **Per-server resolution path (resolveWithServer):**
   - **Cache lookup:** check `cache.Get(hostname)`.
     - On hit: increment cache hit metrics, return cached response.
     - On miss: increment cache miss metrics, continue.
   - **Circuit breaker:** call `Allow` before issuing network requests.
   - **Client pool:** get a DNS client (reused or new).
   - **Query:** send DNS request with `ExchangeContext`.
   - **Metrics and stats:**
     - Record success/failure counts.
     - Record response size, duration, and status.
   - **Response handling:**
     - Extract records, derive minimum TTL, build `DNSResponse`.
   - **Cache store:** store with TTL-based expiration.

5. **Consistency check:**
   - After all servers return for a hostname, `dnsanalysis.CompareResponses`
     checks IP address consistency and records a gauge.

## Statistics and Reporting

`ResolutionStats` is maintained in memory for optional reporting mode:
- Tracks total/failed counts and last error per server.
- `GenerateReport` produces an hourly summary table.
- Report mode exits after printing the report.

## HTTP Surfaces

- **Health endpoint:** `health.HealthChecker` implements `http.Handler` and
  returns `healthy` or `unhealthy` based on server status.
- **Metrics endpoint:** Prometheus `promhttp.Handler` is served on the metrics
  port for scraping.

## Concurrency and Synchronization

Key synchronization points:
- Cache shards: `RWMutex` for per-shard entries.
- Circuit breaker: `Mutex` for per-server counters and timestamps.
- Client pool: `Mutex` protects shared map of clients.
- Health checker: `RWMutex` protects status map.

Care is taken to keep lock scopes small and avoid I/O while locked.

## Configuration and Runtime Defaults

- DNS servers missing a port are normalized to include `:53`.
- Query timeout and interval are configurable via `Duration` in JSON.
- Health and metrics servers are configured by port number.
- Instrumentation logging defaults to `none` and is controlled by
  `instrumentation_level` in `config.json`.

## Extending the System

When adding functionality:
- Keep new metrics in `metrics/metrics.go` and reuse existing label order.
- Prefer extending `DNSResolver` rather than creating parallel control flows.
- Add new config fields to `Config`, update validation, and update `README.md`.
- Consider tests for cache behavior, circuit breaker state changes, and
  config parsing when adding new logic.

## Component Map

- Orchestration: `dnsres.go`
- DNS queries: `dnspool/pool.go`, `dnsres.go` (`resolveWithServer`)
- Cache: `cache/sharded.go`
- Circuit breaker: `circuitbreaker/circuitbreaker.go`
- Response analysis: `dnsanalysis/dnsanalysis.go`
- Health checks: `health/health.go`
- Metrics: `metrics/metrics.go`
