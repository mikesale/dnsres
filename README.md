# DNS Resolution Monitor

A tool for monitoring DNS resolution across multiple servers with advanced features for reliability and performance.

## Why?
We ran into some issues with name resolution that were causing API calls to fail intermittently, but we couldn't independently identify if there were actual problems with the name resolution and if so, across what servers and for how long. This utility will help you monitor and report on DNS and name resolution issues so you can correlate them with your applications' calling APIs that are getting network errors.

## Getting started
What you need:
- The nameserver(s) your infrastructure uses
- The hostname you need to test resolution against


## Ideas
What would I do next just for fun? If I want to have fun with it, I'd like to use an edge LLM to analyze the logs and look for anomalies with some mini-ML. Then maybe set it up as an MCP server 😎

## Features

- Concurrent DNS resolution checks across multiple servers
- Sharded cache implementation for high-performance caching
- Circuit breaker pattern for fault tolerance
- Configurable query timeout and interval
- Separate logs for resolution success, failures, and application health
- Statistical reporting of resolution success rates
- Graceful shutdown handling
- Configurable via JSON configuration file
- Prometheus metrics for monitoring
- Sophisticated DNS error handling
- Health check endpoint for monitoring

## Installation

I'm working on providing precompiled binaries, I just don't have anywhere to host them as of yet. 

For now, make sure you have Go installed, clone the project, and then:

```bash
make build
```

**Note:** The build process automatically disables CGO (`CGO_ENABLED=0`) to ensure a stable, static binary that avoids kernel hangs on macOS systems.

## Configuration

The tool uses a `config.json` file for configuration. See the example at
`examples/config.json`:

```json
{
  "hostnames": ["example.com"],
  "dns_servers": ["8.8.8.8", "1.1.1.1"],
  "query_timeout": "5s",
  "query_interval": "1m",
  "health_port": 8880,
  "metrics_port": 9990,
  "log_dir": "logs",
  "instrumentation_level": "none",
  "circuit_breaker": {
    "threshold": 5,
    "timeout": "30s"
  },
  "cache": {
    "max_size": 1000
  }
}
```

### Configuration Options

- `hostnames`: List of hostnames to monitor
- `dns_servers`: List of DNS server IP addresses. If no port is specified, port 53 is automatically appended (e.g., `8.8.8.8` becomes `8.8.8.8:53`).
- `query_timeout`: Timeout for each DNS query (e.g., "5s", "10s")
- `query_interval`: Interval between resolution checks (e.g., "1m", "5m")
- `health_port`: Port for health check endpoint (default: 8880)
- `metrics_port`: Port for Prometheus metrics (default: 9990)
- `log_dir`: Directory for log files (default: "logs")
- `instrumentation_level`: Debug instrumentation level (`none`, `low`, `medium`, `high`, `critical`)
- `circuit_breaker`: Circuit breaker configuration
  - `threshold`: Number of failures before opening (default: 5)
  - `timeout`: Time to wait before resetting (default: "30s")
- `cache`: Cache configuration
  - `max_size`: Maximum number of cache entries (default: 1000)

## Architecture

The tool is built with a modular architecture:

- `dnspool`: Manages a pool of DNS clients for efficient resolution
- `cache`: Implements a sharded cache for high-performance caching
- `circuitbreaker`: Implements the circuit breaker pattern
- `health`: Provides health check functionality
- `metrics`: Exposes Prometheus metrics
- `dnsanalysis`: Analyzes DNS responses and compares results

## Circuit Breaker Pattern

I wanted to explicitly implement a circuit breaker in Go as a test because I deal a LOT with APIs that can hit rate limits. Helping customers use OpenAI on a low tier, rate limits are constantly a problem. So the tool implements a circuit breaker pattern to prevent cascading failures and provide fault tolerance. Each DNS server has its own circuit breaker with three states:

1. **Closed (Normal)**: The circuit is closed and requests are allowed through
2. **Open (Failing)**: The circuit is open and requests are blocked
3. **Half-Open (Testing)**: The circuit is testing if the service has recovered

The circuit breaker as configured in the core configuration json file will:
- Open after `failure_threshold` consecutive failures
- Wait `reset_timeout` before attempting to close
- Require `failure_threshold` successful attempts in half-open state to fully close
- Track failures independently for each DNS server

## Usage

After installation, you can use the DNS resolver tool:

```bash
# Basic usage
dnsres -config examples/config.json

# Override hostname
dnsres -config examples/config.json -host example.com

# Generate statistics report
dnsres -config examples/config.json -report
```

## Sample Output

### Monitor Output (Success Log)
```
2024/03/14 10:00:00 Resolved example.com using 8.8.8.8:53 (state: closed)
2024/03/14 10:00:00 Resolved example.com using 1.1.1.1:53 (state: closed)
```

### Statistics Report
```
Hour              | DNS Server     | Total    | Fails    | Fail %  
-----------------------------------------------------------------
2024-03-14 10:00 | 1.1.1.1        | 60       | 0        |   0.00%
2024-03-14 10:00 | 8.8.8.8        | 60       | 2        |   3.33%
2024-03-14 11:00 | 1.1.1.1        | 60       | 0        |   0.00%
2024-03-14 11:00 | 8.8.8.8        | 60       | 1        |   1.67%
```

## Metrics

The tool exposes Prometheus metrics on port 9990. Available metrics include:

- `dns_resolution_total`: Total number of DNS resolution attempts
- `dns_resolution_success`: Number of successful DNS resolutions
- `dns_resolution_failure`: Number of failed DNS resolutions
- `dns_resolution_duration_seconds`: DNS resolution duration in seconds
- `circuit_breaker_state`: Current state of each DNS server's circuit breaker (0=Closed, 1=Open, 2=Half-Open)
- `circuit_breaker_failures`: Number of consecutive failures for each DNS server

## Log Files

The tool maintains three separate log files in the configured `log_dir` to separate concerns and simplify monitoring.

### 1. `dnsres-success.log`
Contains a clean audit trail of successful DNS resolutions. This log is intended for long-term auditing and traffic analysis.

**Format:**
```
2024/03/14 10:00:00 Resolved example.com using 8.8.8.8:53 (state: closed)
```

### 2. `dnsres-error.log`
Contains details of failed DNS resolutions. This log is empty when the system and targets are healthy.

**Format:**
```
2024/03/14 10:01:00 Failed to resolve example.com using 8.8.8.8:53: dial udp 8.8.8.8:53: i/o timeout
```

### 3. `dnsres-app.log`
Contains internal application health events, such as startup sequences, HTTP server status (health/metrics ports), configuration errors, and shutdown events. Monitor this file to ensure the *binary itself* is healthy.

**Format:**
```
2024/03/14 10:00:00 Health server error: listen tcp :8880: bind: address already in use
```

## Building from Source

```bash
# Clone the repository
git clone <repository-url>
cd dnsres

# Build the project (creates static binary 'dnsres')
make build
```

## Prometheus
I admit I'm wandering in the dark with this, but with the Go integration and my previous time doing cloud monitoring at the big O I wanted to take a swing and this and get familiar with the package.

## Requirements

- Go 1.21 or later
- Network access to configured DNS servers
- Write permissions in the directory for log files
- Prometheus (optional, for metrics collection)

## License

This project is open source and available under the GNU General Public License v3.0. See the [LICENSE](LICENSE) file for details. 
