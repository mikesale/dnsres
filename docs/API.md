# API Documentation

## Health Check Endpoint

### GET /health

Returns the health status of the DNS resolver and its components.

#### Response Format
```json
{
  "status": "healthy",
  "timestamp": "2024-03-14T10:00:00Z",
  "details": {
    "8.8.8.8": "ok",
    "1.1.1.1": "ok"
  }
}
```

#### Status Codes
- 200: Service is healthy (at least one DNS server is responding)
- 503: Service is unhealthy (no DNS servers are responding)

The health check performs a TCP connection test to each configured DNS server every 30 seconds. A server is considered healthy if it accepts TCP connections within 5 seconds.

## Metrics Endpoint

### GET /metrics

Returns Prometheus metrics for the DNS resolver.

#### Available Metrics

##### DNS Resolution Metrics
- `dns_resolution_total`: Total number of DNS resolution attempts (Counter)
  - Labels: `server`, `hostname`
- `dns_resolution_success`: Number of successful DNS resolutions (Counter)
  - Labels: `server`, `hostname`
- `dns_resolution_failure`: Number of failed DNS resolutions (Counter)
  - Labels: `server`, `hostname`, `error_type`
- `dns_resolution_duration_seconds`: DNS resolution duration in seconds (Histogram)
  - Labels: `server`, `hostname`
- `dns_resolution_consistency`: Whether DNS responses are consistent across servers (Gauge)
  - Labels: `hostname`
- `dns_resolution_cycle_duration_seconds`: Duration of a full resolution cycle in seconds (Histogram)
- `dns_response_size_bytes`: Size of DNS responses in bytes (Histogram)
  - Labels: `server`, `hostname`

##### Circuit Breaker Metrics
- `circuit_breaker_state`: Current state of circuit breaker (Gauge)
  - Labels: `server`
  - Values: 0=Closed, 1=Open, 2=Half-Open
- `circuit_breaker_failures`: Number of consecutive failures for each server (Counter)
  - Labels: `server`

## Command Line Interface

### Basic Usage
```bash
./dnsres [options]
```

### Options
- `-config string`: Path to configuration file (default "config.json")
- `-host string`: Override hostname from config file
- `-report`: Generate statistics report

### Examples
```bash
# Run with default config
./dnsres

# Configless usage (defaults + CLI hostname)
./dnsres example.com

# Override hostname
./dnsres -host example.com

# Generate report
./dnsres -report

# Use custom config
./dnsres -config custom.json
```

## Configuration API

### Configuration Structure
```json
{
  "hostnames": ["example.com"],
  "dns_servers": ["8.8.8.8:53", "1.1.1.1:53"],
  "query_timeout": "5s",
  "query_interval": "30s",
  "health_port": 8880,
  "metrics_port": 9990,
  "log_dir": "logs",
  "instrumentation_level": "none",
  "circuit_breaker": {
    "threshold": 5,
    "timeout": "30s"
  }
}
```

### Field Descriptions

#### Required Fields
- `hostnames`: List of hostnames to monitor
- `dns_servers`: List of DNS server addresses (port :53 appended if missing)
- `query_timeout`: Timeout for each DNS query (e.g., "5s")
- `query_interval`: Interval between resolution checks (e.g., "30s")

#### Optional Fields
- `health_port`: Health check endpoint port (default: 8880)
- `metrics_port`: Prometheus metrics endpoint port (default: 9990)
- `log_dir`: Directory for log files (default: XDG state directory or $HOME/logs)
- `instrumentation_level`: Debug logging level: "none", "low", "medium", "high", "critical" (default: "none")
- `circuit_breaker`: Circuit breaker configuration
  - `threshold`: Number of consecutive failures before opening (default: 5)
  - `timeout`: Time to wait before attempting to close (default: "30s")

## Logging API

### Log Files
- `dnsres-success.log`: Successful resolution attempts
- `dnsres-error.log`: Failed resolution attempts

### Log Format
```
2024/03/14 10:00:00 Resolved example.com using 8.8.8.8 (state: normal)
2024/03/14 10:00:00 Failed to resolve example.com using 1.1.1.1: timeout
```

## Error Handling

### Common Error Types
- `ErrCircuitOpen`: Circuit breaker is open
- `ErrTimeout`: DNS query timeout
- `ErrInvalidConfig`: Invalid configuration
- `ErrDNSError`: DNS resolution error

### Error Response Format
```json
{
  "error": "error message",
  "code": "ERROR_CODE",
  "details": {
    "server": "8.8.8.8",
    "hostname": "example.com",
    "state": "circuit_open"
  }
}
```
