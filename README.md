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
- Detailed logging of successful and failed resolutions
- Statistical reporting of resolution success rates
- Graceful shutdown handling
- Configurable via JSON configuration file
- Prometheus metrics for monitoring
- Sophisticated DNS error handling
- Health check endpoint for monitoring
- DNSSEC and EDNS support detection

## Installation

### macOS
```bash
# Intel Mac
curl -L https://github.com/mikesale/dnsres/releases/download/v1.0.0/dnsres-darwin-amd64-v1.0.0.tar.gz | tar xz
sudo mv dnsres-darwin-amd64 /usr/local/bin/dnsres

# Apple Silicon
curl -L https://github.com/mikesale/dnsres/releases/download/v1.0.0/dnsres-darwin-arm64-v1.0.0.tar.gz | tar xz
sudo mv dnsres-darwin-arm64 /usr/local/bin/dnsres
```

### Linux
```bash
# AMD64
curl -L https://github.com/mikesale/dnsres/releases/download/v1.0.0/dnsres-linux-amd64-v1.0.0.tar.gz | tar xz
sudo mv dnsres-linux-amd64 /usr/local/bin/dnsres

# ARM64
curl -L https://github.com/mikesale/dnsres/releases/download/v1.0.0/dnsres-linux-arm64-v1.0.0.tar.gz | tar xz
sudo mv dnsres-linux-arm64 /usr/local/bin/dnsres
```

### Windows
1. Download `dnsres-windows-amd64-v1.0.0.zip`
2. Extract the zip file
3. Move `dnsres-windows-amd64.exe` to a directory in your PATH

## Configuration

The tool uses a `config.json` file for configuration. Here's an example:

```json
{
  "hostnames": ["example.com"],
  "dns_servers": ["8.8.8.8", "1.1.1.1"],
  "query_timeout": "5s",
  "query_interval": "1m",
  "health_port": 8080,
  "metrics_port": 9090,
  "log_dir": "logs",
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
- `dns_servers`: List of DNS server IP addresses
- `query_timeout`: Timeout for each DNS query (e.g., "5s", "10s")
- `query_interval`: Interval between resolution checks (e.g., "1m", "5m")
- `health_port`: Port for health check endpoint (default: 8080)
- `metrics_port`: Port for Prometheus metrics (default: 9090)
- `log_dir`: Directory for log files (default: "logs")
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

The tool implements a circuit breaker pattern to prevent cascading failures and provide fault tolerance. Each DNS server has its own circuit breaker with three states:

1. **Closed (Normal)**: The circuit is closed and requests are allowed through
2. **Open (Failing)**: The circuit is open and requests are blocked
3. **Half-Open (Testing)**: The circuit is testing if the service has recovered

The circuit breaker will:
- Open after `failure_threshold` consecutive failures
- Wait `reset_timeout` before attempting to close
- Require `failure_threshold` successful attempts in half-open state to fully close
- Track failures independently for each DNS server

## Usage

After installation, you can use the DNS resolver tool:

```bash
# Basic usage
dnsres -config config.json

# Override hostname
dnsres -config config.json -host example.com

# Generate statistics report
dnsres -config config.json -report
```

## Sample Output

### Monitor Output
```
2024/03/14 10:00:00 Resolved example.com using 8.8.8.8 (state: normal)
2024/03/14 10:00:00 Resolved example.com using 1.1.1.1 (state: normal)
2024/03/14 10:01:00 Resolution error for example.com using 8.8.8.8 (state: circuit open): lookup example.com: no such host
2024/03/14 10:01:00 Resolved example.com using 1.1.1.1 (state: circuit half-open)
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

The tool exposes Prometheus metrics on port 9090. Available metrics include:

- `dns_resolution_total`: Total number of DNS resolution attempts
- `dns_resolution_success`: Number of successful DNS resolutions
- `dns_resolution_failure`: Number of failed DNS resolutions
- `dns_resolution_duration_seconds`: DNS resolution duration in seconds
- `circuit_breaker_state`: Current state of each DNS server's circuit breaker (0=Closed, 1=Open, 2=Half-Open)
- `circuit_breaker_failures`: Number of consecutive failures for each DNS server

## Log Files

The tool creates a single log file `dnsres.log` that contains structured JSON logs for all events. Each log entry includes:

### Basic Information
- `timestamp`: When the event occurred
- `level`: Log level (INFO/ERROR)
- `hostname`: The domain being resolved
- `server`: The DNS server used
- `correlation_id`: Unique ID to track related events

### System Context
- `version`: The version of the DNS resolver
- `environment`: Development/Staging/Production
- `instance_id`: Unique identifier for the running instance

### DNS Query Details
- `query_type`: The type of DNS query (A, AAAA, MX, etc.)
- `edns_enabled`: Whether EDNS was used
- `dnssec_enabled`: Whether DNSSEC was enabled
- `recursion_desired`: Whether recursion was requested

### Performance Metrics
- `duration_ms`: Total time taken for the resolution
- `queue_time_ms`: Time spent waiting for a client from the pool
- `network_latency_ms`: Raw network latency (excluding processing time)
- `processing_time_ms`: Time spent processing the response
- `cache_ttl_seconds`: Time-to-live of cached entries

### Response Analysis
- `response_code`: The DNS response code (NOERROR, NXDOMAIN, etc.)
- `response_size`: Size of the DNS response in bytes
- `record_count`: Number of records in the response
- `authoritative`: Whether the response was authoritative
- `truncated`: Whether the response was truncated
- `response_flags`: Additional DNS response flags (AA, TC, RD, RA, AD, CD)

### Circuit Breaker and Cache
- `circuit_state`: Current state of the circuit breaker
- `cache_hit`: Whether the response came from cache

### Error Information
- `error`: Error message (for failed queries)
- `error_type`: Type of error (circuit_breaker, client_pool, query_error, dns_error)

Example log entries:

```json
{
  "timestamp": "2024-03-14T10:00:00Z",
  "level": "INFO",
  "hostname": "example.com",
  "server": "8.8.8.8",
  "correlation_id": "8.8.8.8-example.com-1710417600000000000",
  "version": "1.0.0",
  "environment": "production",
  "instance_id": "dnsres-1",
  "query_type": "A",
  "edns_enabled": true,
  "dnssec_enabled": true,
  "recursion_desired": true,
  "duration_ms": 45.2,
  "queue_time_ms": 0.5,
  "network_latency_ms": 30.1,
  "processing_time_ms": 14.6,
  "cache_ttl_seconds": 300,
  "response_code": "NOERROR",
  "response_size": 123,
  "record_count": 2,
  "authoritative": false,
  "truncated": false,
  "response_flags": ["RD", "RA"],
  "circuit_state": "closed",
  "cache_hit": false
}
```

The structured logging format makes it easy to:
- Parse logs using standard JSON tools
- Filter and search logs based on specific fields
- Track related events using correlation IDs
- Analyze performance metrics
- Monitor system health
- Debug DNS resolution issues
- Track cache effectiveness
- Monitor circuit breaker behavior

## Building from Source

```bash
# Clone the repository
git clone <repository-url>
cd dnsres

# Build the project
go build -o dnsres
```

## Requirements

- Go 1.21 or later
- Network access to configured DNS servers
- Write permissions in the directory for log files
- Prometheus (optional, for metrics collection)

## License

This project is open source and available under the GNU General Public License v3.0. See the [LICENSE](LICENSE) file for details. 
