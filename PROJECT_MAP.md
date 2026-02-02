# PROJECT_MAP.md - DNS Resolution Monitor Architecture

> **Long-Term Memory** for AI agents and developers working on this codebase.
> 
> Last Updated: 2026-02-02 | Go 1.24.0 | ~3,700 LOC

---

## Table of Contents
1. [High-Level Architecture](#section-1-high-level-architecture)
2. [Data Models](#section-2-data-models)
3. [Core Critical Flows](#section-3-core-critical-flows)
4. [Directory Map](#section-4-directory-map)
5. [Quick Reference](#quick-reference)

---

## Section 1: High-Level Architecture

### System Overview (10,000ft View)

The DNS Resolution Monitor is a long-running process that periodically resolves configured hostnames against multiple DNS servers, tracks consistency, exposes metrics, and provides both CLI and TUI interfaces.

```mermaid
graph TD
    subgraph "User Interfaces"
        CLI["dnsres CLI<br/>(cmd/dnsres)"]
        TUI["dnsres-tui<br/>(cmd/dnsres-tui)"]
    end

    subgraph "Application Layer"
        APP["internal/app<br/>Run() Orchestration"]
        TUIRUN["internal/tui<br/>Bubble Tea TUI"]
    end

    subgraph "Core Engine"
        RESOLVER["DNSResolver<br/>(internal/dnsres)"]
        CONFIG["Config Loader<br/>JSON Parsing"]
        EVENTS["Event Bus<br/>Pub/Sub System"]
    end

    subgraph "Resilience Layer"
        CB["Circuit Breaker<br/>(circuitbreaker/)"]
        CACHE["Sharded Cache<br/>(cache/)"]
        POOL["Client Pool<br/>(dnspool/)"]
    end

    subgraph "Observability"
        METRICS["Prometheus Metrics<br/>(metrics/)"]
        HEALTH["Health Checker<br/>(health/)"]
        LOGS["Three-Stream Logs<br/>success/error/app"]
        INSTR["Instrumentation<br/>Levels"]
    end

    subgraph "Analysis"
        ANALYSIS["DNS Analysis<br/>(dnsanalysis/)"]
    end

    subgraph "External Systems"
        DNS1["DNS Server 1<br/>(8.8.8.8:53)"]
        DNS2["DNS Server 2<br/>(1.1.1.1:53)"]
        DNSN["DNS Server N"]
        PROM["Prometheus<br/>Scraper"]
        MONITOR["Monitoring<br/>System"]
    end

    CLI -->|"flags: -config, -host, -report"| APP
    TUI -->|"flags: -config, -host"| TUIRUN
    
    APP -->|"creates"| RESOLVER
    TUIRUN -->|"creates"| RESOLVER
    TUIRUN -->|"subscribes to"| EVENTS
    
    RESOLVER -->|"loads"| CONFIG
    RESOLVER -->|"publishes"| EVENTS
    RESOLVER -->|"uses"| CB
    RESOLVER -->|"uses"| CACHE
    RESOLVER -->|"uses"| POOL
    RESOLVER -->|"uses"| ANALYSIS
    RESOLVER -->|"updates"| METRICS
    RESOLVER -->|"writes"| LOGS
    RESOLVER -->|"creates"| HEALTH
    
    POOL -->|"DNS queries"| DNS1
    POOL -->|"DNS queries"| DNS2
    POOL -->|"DNS queries"| DNSN
    
    HEALTH -->|"TCP probes"| DNS1
    HEALTH -->|"TCP probes"| DNS2
    
    HEALTH -->|"HTTP :8880"| MONITOR
    METRICS -->|"HTTP :9990"| PROM
    
    LOGS -->|"log files"| INSTR
```

### Component Interaction Matrix

| From \ To | Resolver | Cache | CircuitBreaker | Pool | Metrics | Events |
|-----------|----------|-------|----------------|------|---------|--------|
| **Resolver** | - | Get/Set | Allow/Record | Get/Put | Update | Publish |
| **Cache** | - | - | - | - | Update | - |
| **CircuitBreaker** | - | - | - | - | Update | - |
| **Pool** | - | - | - | - | Update | - |
| **Health** | - | - | - | - | Update | - |
| **TUI** | HealthSnapshot | - | - | - | - | Subscribe |

---

## Section 2: Data Models

### Entity-Relationship Diagram

```mermaid
erDiagram
    Config ||--o{ Hostname : contains
    Config ||--o{ DNSServer : contains
    Config ||--|| CircuitBreakerConfig : has
    Config ||--|| CacheConfig : has
    
    DNSResolver ||--|| Config : uses
    DNSResolver ||--o{ CircuitBreaker : manages
    DNSResolver ||--|| ShardedCache : owns
    DNSResolver ||--|| ClientPool : owns
    DNSResolver ||--|| HealthChecker : owns
    DNSResolver ||--|| EventBus : owns
    DNSResolver ||--|| ResolutionStats : tracks
    
    ShardedCache ||--o{ CacheShard : contains
    CacheShard ||--o{ CacheEntry : stores
    CacheEntry ||--|| DNSResponse : wraps
    
    EventBus ||--o{ Subscriber : notifies
    Subscriber }o--o{ ResolverEvent : receives
    
    ResolutionStats ||--o{ ServerStats : aggregates
    
    Config {
        string[] hostnames
        string[] dns_servers
        Duration query_timeout
        Duration query_interval
        int health_port
        int metrics_port
        string log_dir
        string instrumentation_level
    }
    
    CircuitBreakerConfig {
        int threshold
        Duration timeout
    }
    
    CacheConfig {
        int64 max_size
    }
    
    DNSResolver {
        Config config
        ClientPool clientPool
        map_CircuitBreaker breakers
        ShardedCache cache
        HealthChecker health
        Logger successLog
        Logger errorLog
        Logger appLog
        ResolutionStats stats
        Level instrumentationLevel
        EventBus events
    }
    
    CircuitBreaker {
        int threshold
        Duration timeout
        int failures
        Time lastError
        string server
        State state
    }
    
    ShardedCache {
        CacheShard[] shards
        int numShards
        int64 maxSize
    }
    
    CacheEntry {
        DNSResponse Response
        Time Expires
        int64 Size
    }
    
    DNSResponse {
        string Server
        string Hostname
        string[] Addresses
        dns_Msg Response
        map_int RecordCount
        uint32 TTL
        int Size
        bool DNSSEC
        bool EDNS
        string Protocol
        Duration Duration
    }
    
    ResolverEvent {
        EventType Type
        Time Time
        string Hostname
        string Server
        Duration Duration
        string Error
        string[] Addresses
        bool Consistent
        int HostnameCount
        int ServerCount
        string Source
    }
    
    ServerStats {
        int Total
        int Failures
        string LastError
    }
    
    ClientPool {
        map_clients clients
        int MaxSize
        Duration Timeout
    }
    
    HealthChecker {
        string[] servers
        map_bool status
        Logger appLog
        Level level
    }
```

### Event Types

| EventType | Trigger | Key Fields |
|-----------|---------|------------|
| `cycle_start` | Resolution cycle begins | HostnameCount, ServerCount |
| `cycle_complete` | Resolution cycle ends | Duration, HostnameCount, ServerCount |
| `resolve_success` | Single resolution succeeds | Hostname, Server, Duration, Addresses, Source |
| `resolve_failure` | Single resolution fails | Hostname, Server, Duration, Error, Source |
| `inconsistent` | DNS responses don't match | Hostname, Consistent=false |

### Circuit Breaker States

| State | Value | Description |
|-------|-------|-------------|
| `Closed` | 0 | Normal operation, requests allowed |
| `Open` | 1 | Too many failures, requests blocked |
| `HalfOpen` | 2 | Testing recovery, limited requests |

### Instrumentation Levels

| Level | Value | Description |
|-------|-------|-------------|
| `None` | 0 | No debug logging |
| `Low` | 1 | Basic lifecycle events |
| `Medium` | 2 | Failures and warnings |
| `High` | 3 | Detailed resolution info |
| `Critical` | 4 | Maximum verbosity |

---

## Section 3: Core Critical Flows

### Flow 1: DNS Resolution Cycle (Main Loop)

This is the core heartbeat of the application - the periodic resolution of all hostnames against all DNS servers.

```mermaid
sequenceDiagram
    autonumber
    participant Ticker as Ticker (QueryInterval)
    participant Resolver as DNSResolver
    participant Events as EventBus
    participant Sem as Semaphore (10)
    participant CB as CircuitBreaker
    participant Cache as ShardedCache
    participant Pool as ClientPool
    participant DNS as DNS Server
    participant Analysis as dnsanalysis
    participant Metrics as Prometheus

    Ticker->>Resolver: tick fires
    Resolver->>Events: publish(EventCycleStart)
    
    loop For each hostname (concurrent, capped by semaphore)
        Resolver->>Sem: acquire slot
        
        loop For each DNS server (concurrent)
            Resolver->>Resolver: resolveWithServer(ctx, server, hostname)
            
            Resolver->>Cache: Get(hostname)
            alt Cache Hit
                Cache-->>Resolver: DNSResponse, true
                Resolver->>Metrics: CacheHit.Inc()
                Resolver->>Events: publish(EventResolveSuccess, source="cache")
            else Cache Miss
                Resolver->>Metrics: CacheMiss.Inc()
                Resolver->>CB: Allow()
                
                alt Circuit Open
                    CB-->>Resolver: false
                    Resolver->>Metrics: Failure.Inc("circuit_breaker")
                    Resolver->>Events: publish(EventResolveFailure, source="circuit_breaker")
                else Circuit Allows
                    CB-->>Resolver: true
                    Resolver->>Pool: Get(server)
                    Pool-->>Resolver: *dns.Client
                    
                    Resolver->>DNS: ExchangeContext(msg, server)
                    
                    alt Query Success
                        DNS-->>Resolver: *dns.Msg, duration
                        Resolver->>CB: RecordSuccess()
                        Resolver->>Metrics: Duration.Observe()
                        Resolver->>Metrics: Success.Inc()
                        Resolver->>Cache: Set(hostname, response, TTL)
                        Resolver->>Events: publish(EventResolveSuccess, source="query")
                    else Query Failure
                        DNS-->>Resolver: error
                        Resolver->>CB: RecordFailure()
                        Resolver->>Metrics: Failure.Inc("query_error")
                        Resolver->>Events: publish(EventResolveFailure)
                    end
                    
                    Resolver->>Pool: Put(server, client)
                end
            end
        end
        
        Note over Resolver,Analysis: After all servers respond for hostname
        Resolver->>Analysis: CompareResponses(responses)
        Analysis-->>Resolver: consistent bool
        Resolver->>Metrics: Consistency.Set(consistent)
        
        alt Inconsistent
            Resolver->>Events: publish(EventInconsistent)
        end
        
        Resolver->>Sem: release slot
    end
    
    Resolver->>Metrics: CycleDuration.Observe()
    Resolver->>Events: publish(EventCycleComplete)
```

### Flow 2: TUI Event-Driven Rendering

The TUI uses Bubble Tea's Elm architecture with the resolver's event bus for real-time updates.

```mermaid
sequenceDiagram
    autonumber
    participant User as User Input
    participant Tea as Bubble Tea
    participant Model as TUI Model
    participant VP as Viewport
    participant Table as Server Table
    participant Resolver as DNSResolver
    participant Events as EventBus
    participant Health as HealthChecker

    Note over Tea,Model: Initialization
    Tea->>Model: Init()
    Model->>Resolver: SubscribeEvents(200)
    Resolver-->>Model: events chan, unsubscribe func
    Model->>Tea: Batch(waitForEvent, tickHealth, spinner.Tick)
    
    Note over Tea,Model: Event Loop
    
    loop Main Loop
        alt Resolver Event
            Events-->>Model: ResolverEvent via channel
            Model->>Model: applyEvent(event)
            
            alt EventCycleStart
                Model->>Model: cycleRunning = true
                Model->>VP: appendActivity("cycle start")
            else EventCycleComplete
                Model->>Model: cycleRunning = false, lastCycleDur = duration
                Model->>VP: appendActivity("cycle complete")
            else EventResolveSuccess
                Model->>Model: update serverState (latency, lastSuccess)
                Model->>VP: appendActivity("resolved hostname via server")
            else EventResolveFailure
                Model->>Model: update serverState (error, lastFailure)
                Model->>VP: appendActivity("failed hostname via server")
            else EventInconsistent
                Model->>VP: appendActivity("inconsistent responses")
            end
            
            Model->>Table: updateTableRows()
            Model->>Tea: waitForEvent(events)
            
        else Health Tick (every 2s)
            Model->>Resolver: HealthSnapshot()
            Resolver->>Health: StatusSnapshot()
            Health-->>Model: map[server]bool
            Model->>Table: updateTableRows()
            Model->>Tea: tickHealth()
            
        else Window Resize
            Tea->>Model: WindowSizeMsg{Width, Height}
            Model->>Model: resize()
            Model->>Table: SetHeight(), SetColumns()
            Model->>VP: Width, Height
            
        else Key Press
            User->>Tea: key event
            Tea->>Model: KeyMsg
            
            alt "q" or "ctrl+c"
                Model->>Resolver: cancel()
                Model->>Events: unsubscribe()
                Model->>Tea: Quit
            end
        end
    end
    
    Note over Tea,Model: View Rendering
    Tea->>Model: View()
    Model->>Model: summaryView() - status, counts, timing
    Model->>Table: View() - server table
    Model->>VP: View() - activity log
    Model-->>Tea: lipgloss.JoinVertical(summary+table, activity)
```

### Flow 3: Circuit Breaker State Machine

The circuit breaker protects against cascading failures when DNS servers become unavailable.

```mermaid
sequenceDiagram
    autonumber
    participant Resolver as DNSResolver
    participant CB as CircuitBreaker
    participant Metrics as Prometheus
    participant DNS as DNS Server

    Note over CB: Initial State: CLOSED (failures=0)
    
    rect rgb(200, 255, 200)
        Note over Resolver,DNS: Normal Operation (CLOSED)
        Resolver->>CB: Allow()
        CB->>Metrics: State.Set(0) [Closed]
        CB-->>Resolver: true
        Resolver->>DNS: query
        DNS-->>Resolver: success
        Resolver->>CB: RecordSuccess()
        CB->>CB: failures = 0
    end
    
    rect rgb(255, 255, 200)
        Note over Resolver,DNS: Failures Accumulating
        loop failures < threshold
            Resolver->>CB: Allow()
            CB-->>Resolver: true
            Resolver->>DNS: query
            DNS-->>Resolver: error
            Resolver->>CB: RecordFailure()
            CB->>CB: failures++, lastError = now
        end
    end
    
    rect rgb(255, 200, 200)
        Note over CB: State: OPEN (failures >= threshold)
        Resolver->>CB: Allow()
        CB->>CB: failures >= threshold?
        CB->>CB: time.Since(lastError) < timeout?
        CB->>Metrics: State.Set(1) [Open]
        CB-->>Resolver: false
        Note over Resolver: Request blocked, no DNS query
    end
    
    rect rgb(200, 200, 255)
        Note over CB: State: HALF-OPEN (timeout elapsed)
        Note over CB: Waiting for timeout to elapse...
        Resolver->>CB: Allow()
        CB->>CB: failures >= threshold?
        CB->>CB: time.Since(lastError) >= timeout?
        CB->>Metrics: State.Set(2) [HalfOpen]
        CB-->>Resolver: true
        Note over Resolver: Test request allowed
        
        alt Test Success
            Resolver->>DNS: query
            DNS-->>Resolver: success
            Resolver->>CB: RecordSuccess()
            CB->>CB: failures = 0
            CB->>Metrics: State.Set(0) [Closed]
            Note over CB: Back to CLOSED
        else Test Failure
            Resolver->>DNS: query
            DNS-->>Resolver: error
            Resolver->>CB: RecordFailure()
            CB->>CB: failures++, lastError = now
            CB->>Metrics: State.Set(1) [Open]
            Note over CB: Back to OPEN
        end
    end
```

---

## Section 4: Directory Map

```
dnsres/
├── cmd/                          # Command entrypoints (thin wrappers)
│   ├── dnsres/                   # CLI binary - headless resolver
│   │   └── main.go              # calls internal/app.Run()
│   └── dnsres-tui/              # TUI binary - interactive interface
│       └── main.go              # calls internal/tui.Run()
│
├── internal/                     # Private packages (not importable externally)
│   ├── app/                      # CLI application orchestration
│   │   └── run.go               # Flag parsing, signal handling, resolver startup
│   │
│   ├── dnsres/                   # Core DNS resolver implementation
│   │   ├── resolver.go          # DNSResolver struct, Start(), resolveAll(), resolveWithServer()
│   │   ├── config.go            # Config struct, LoadConfig(), Validate(), Duration wrapper
│   │   ├── events.go            # EventBus, ResolverEvent, EventType constants
│   │   ├── logging.go           # setupLoggers() - three-stream log setup
│   │   ├── report.go            # ResolutionStats, ServerStats, GenerateReport()
│   │   └── *_test.go            # Unit tests for resolver logic
│   │
│   ├── tui/                      # Terminal UI (Bubble Tea framework)
│   │   ├── run.go               # Run() - TUI entry point, event subscription
│   │   ├── model.go             # Bubble Tea model, Update(), View(), state management
│   │   └── theme.go             # Lipgloss styles, colors, borders
│   │
│   └── integration/              # End-to-end integration tests
│       └── dnsres_e2e_test.go   # Build tag: //go:build integration
│
├── cache/                        # Public: High-performance caching
│   ├── sharded.go               # ShardedCache, CacheShard, CacheEntry
│   └── sharded_test.go          # Cache unit tests
│
├── circuitbreaker/               # Public: Fault tolerance pattern
│   ├── circuitbreaker.go        # CircuitBreaker, State enum, Allow(), Record*()
│   ├── errors.go                # ErrCircuitOpen sentinel error
│   └── *_test.go                # Circuit breaker tests
│
├── dnspool/                      # Public: DNS client connection pooling
│   ├── pool.go                  # ClientPool, Get(), Put()
│   └── pool_test.go             # Pool tests
│
├── dnsanalysis/                  # Public: DNS response analysis
│   ├── dnsanalysis.go           # DNSResponse, AnalyzeResponse(), CompareResponses()
│   └── dnsanalysis_test.go      # Analysis tests
│
├── health/                       # Public: HTTP health endpoint
│   ├── health.go                # HealthChecker, ServeHTTP(), checkLoop()
│   └── health_test.go           # Health tests
│
├── metrics/                      # Public: Prometheus metrics definitions
│   ├── metrics.go               # All metric vars (counters, gauges, histograms)
│   └── metrics_test.go          # Metrics tests
│
├── instrumentation/              # Public: Debug logging levels
│   ├── level.go                 # Level enum, ParseLevel()
│   └── level_test.go            # Level tests
│
├── docs/                         # Documentation
│   ├── ARCHITECTURE.md          # System design overview
│   ├── DEVELOPMENT.md           # Developer guide
│   ├── API.md                   # API documentation
│   └── INTEGRATION_TESTING.md   # Testing guide
│
├── examples/                     # Example configurations
│   └── config.json              # Sample config.json
│
├── logs/                         # Runtime log output (created at runtime)
│   ├── dnsres-success.log       # Successful resolutions
│   ├── dnsres-error.log         # Failed resolutions
│   └── dnsres-app.log           # Application lifecycle events
│
├── AGENTS.md                     # AI agent instructions
├── README.md                     # Project documentation
├── CONTRIBUTING.md               # Contribution guidelines
├── SECURITY.md                   # Security policy
├── CHANGELOG.md                  # Version history
├── Makefile                      # Build commands
├── Dockerfile                    # Container build
├── go.mod                        # Go module definition
└── go.sum                        # Dependency checksums
```

### Package Dependency Graph

```mermaid
graph BT
    subgraph "External Dependencies"
        DNS["github.com/miekg/dns"]
        PROM["prometheus/client_golang"]
        CHARM["charmbracelet/bubbletea"]
        LIPGLOSS["charmbracelet/lipgloss"]
    end

    subgraph "Public Packages (importable)"
        METRICS["metrics"]
        INSTR["instrumentation"]
        CACHE["cache"]
        CB["circuitbreaker"]
        POOL["dnspool"]
        ANALYSIS["dnsanalysis"]
        HEALTH["health"]
    end

    subgraph "Internal Packages (private)"
        DNSRES["internal/dnsres"]
        APP["internal/app"]
        TUI["internal/tui"]
    end

    subgraph "Commands"
        CMD["cmd/dnsres"]
        CMDTUI["cmd/dnsres-tui"]
    end

    %% Dependencies
    METRICS --> PROM
    CACHE --> METRICS
    CACHE --> ANALYSIS
    CB --> METRICS
    POOL --> DNS
    POOL --> METRICS
    ANALYSIS --> DNS
    ANALYSIS --> METRICS
    HEALTH --> METRICS
    HEALTH --> INSTR

    DNSRES --> CACHE
    DNSRES --> CB
    DNSRES --> POOL
    DNSRES --> ANALYSIS
    DNSRES --> HEALTH
    DNSRES --> METRICS
    DNSRES --> INSTR
    DNSRES --> DNS
    DNSRES --> PROM

    APP --> DNSRES

    TUI --> DNSRES
    TUI --> CHARM
    TUI --> LIPGLOSS

    CMD --> APP
    CMDTUI --> TUI
```

---

## Quick Reference

### Build Commands

```bash
make build          # Build dnsres CLI (CGO_ENABLED=0)
make build-tui      # Build dnsres-tui
make build-all      # Cross-compile for multiple platforms
make test           # Run unit tests
make lint           # Run golangci-lint
make coverage       # Generate coverage report
```

### Run Commands

```bash
# CLI Mode
./dnsres -config config.json              # With config file
./dnsres example.com                       # Quick check with defaults
./dnsres -config config.json -report       # Generate stats report

# TUI Mode
./dnsres-tui -config config.json          # Interactive terminal UI
./dnsres-tui example.com                   # Quick TUI with defaults
```

### HTTP Endpoints

| Port | Path | Purpose |
|------|------|---------|
| 8880 | `/` | Health check (returns "healthy" or "unhealthy") |
| 9990 | `/metrics` | Prometheus metrics scraping |

### Key Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `dns_resolution_total` | Counter | server, hostname | Total resolution attempts |
| `dns_resolution_success` | Counter | server, hostname | Successful resolutions |
| `dns_resolution_failure` | Counter | server, hostname, error_type | Failed resolutions |
| `dns_resolution_duration_seconds` | Histogram | server, hostname | Query latency |
| `dns_resolution_consistency` | Gauge | hostname | Cross-server consistency (1=match) |
| `circuit_breaker_state` | Gauge | server | 0=Closed, 1=Open, 2=HalfOpen |
| `dns_resolver_cache_size` | Gauge | - | Current cache entries |
| `dns_resolver_cache_hits_total` | Counter | - | Cache hit count |

### Log Files

| File | Purpose |
|------|---------|
| `dnsres-success.log` | Audit trail of successful resolutions |
| `dnsres-error.log` | Failed resolutions and inconsistencies |
| `dnsres-app.log` | Application lifecycle (startup, shutdown, errors) |

### Configuration Schema

```json
{
  "hostnames": ["example.com"],           // Required: domains to monitor
  "dns_servers": ["8.8.8.8:53"],          // Required: DNS servers (port auto-added)
  "query_timeout": "5s",                   // Per-query timeout
  "query_interval": "30s",                 // Resolution cycle interval
  "health_port": 8880,                     // Health endpoint port
  "metrics_port": 9990,                    // Prometheus metrics port
  "log_dir": "logs",                       // Log file directory
  "instrumentation_level": "none",         // none|low|medium|high|critical
  "circuit_breaker": {
    "threshold": 5,                        // Failures before open
    "timeout": "30s"                       // Reset timeout
  },
  "cache": {
    "max_size": 1000                       // Max cache entries
  }
}
```

---

## Notes for AI Agents

1. **Entry Points**: Start at `internal/app/run.go` (CLI) or `internal/tui/run.go` (TUI)
2. **Core Logic**: `internal/dnsres/resolver.go` contains the main resolution loop
3. **No Database**: All state is in-memory; logs are append-only files
4. **Concurrency**: Semaphore limits concurrent hostname resolution to 10
5. **Event Bus**: Non-blocking pub/sub; slow TUI won't block resolver
6. **Metrics**: All metrics in `metrics/metrics.go`; follow existing label patterns
7. **Testing**: Unit tests co-located; integration tests require `-tags=integration`
