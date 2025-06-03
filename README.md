# DNS Resolution Monitor

A tool for monitoring DNS resolution across multiple servers with advanced features for reliability and performance.

## Why?
We ran into some issues with name resolution that were causing API calls to fail intermittently, but we couldn't independently identify if there were actual problems with the name resolution and if so, across what servers and for how long. This utility will help you monitor and report on DNS and name resolution issues so you can correlate them with your applications' calling APIs that are getting network errors.

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

The tool creates two log files:

1. `dnsres-success.log`: Contains successful resolution attempts
2. `dnsres-error.log`: Contains failed resolution attempts

Each log entry includes:
- Timestamp
- Hostname
- DNS server used
- Circuit breaker state
- Error message (for failed resolutions)

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