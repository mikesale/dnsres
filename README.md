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

1. Ensure you have Go 1.21 or later installed
2. Clone the repository
3. Build the project:
   ```bash
   go build -o dnsres
   ```

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

### Running the Monitor

```bash
./dnsres
```

This will start the DNS resolution monitor using the configuration from `config.json`. The program will:
- Create two log files: `dnsres-success.log` and `dnsres-error.log`
- Perform DNS resolution checks at the configured interval
- Log successful and failed resolutions with circuit breaker states
- Handle graceful shutdown with Ctrl+C
- Expose Prometheus metrics on port 9090

### Viewing Statistics

```bash
./dnsres -report
```

This will generate a report showing resolution statistics grouped by hour and DNS server.

### Command Line Options

- `-host`: Override the hostname from config file
  ```bash
  ./dnsres -host example.com
  ```
- `-report`: Generate and display statistics report
  ```bash
  ./dnsres -report
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