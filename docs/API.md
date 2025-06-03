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
- `dns_resolution_total`: Total resolution attempts
- `dns_resolution_success`: Successful resolutions
- `dns_resolution_failure`: Failed resolutions
- `dns_resolution_duration_seconds`: Resolution duration
- `dns_resolution_consistency`: Response consistency
- `dns_response_size_bytes`: Size of DNS responses
- `dns_record_count`: Number of records in responses
- `dns_resolution_latency_seconds`: Latency between servers
- `dns_resolution_ttl_seconds`: TTL values from responses
- `dns_resolution_retries_total`: Retry attempts
- `dns_resolution_timeout_total`: Timeout occurrences
- `dns_resolution_nxdomain_total`: NXDOMAIN responses
- `dns_resolution_servfail_total`: SERVFAIL responses
- `dns_resolution_refused_total`: REFUSED responses
- `dns_resolution_rate_limit_total`: Rate limit occurrences
- `dns_resolution_network_error_total`: Network-related errors
- `dns_resolution_dnssec_total`: DNSSEC validation results
- `dns_resolution_edns_support`: EDNS support status
- `dns_resolution_dnssec_support`: DNSSEC support status
- `dns_resolution_protocol_total`: Protocol usage

##### Circuit Breaker Metrics
- `circuit_breaker_state`: Current state (0=Closed, 1=Open, 2=Half-Open)
- `circuit_breaker_failures`: Consecutive failures
- `dns_circuit_breaker_trips_total`: Circuit breaker trips

##### Cache Metrics
- `dns_cache_size`: Current cache size
- `dns_cache_hits_total`: Cache hits
- `dns_cache_misses_total`: Cache misses
- `dns_cache_evictions_total`: Cache evictions

##### Health Check Metrics
- `dns_resolver_health_status`: Component health status
- `dns_resolver_health_check_duration_seconds`: Health check duration

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
  "dns_servers": ["8.8.8.8", "1.1.1.1"],
  "query_timeout": "5s",
  "query_interval": "1m",
  "failure_threshold": 5,
  "reset_timeout": "30s",
  "half_open_timeout": "5s",
  "max_cache_ttl": "1h",
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

### Field Descriptions

#### Required Fields
- `hostnames`: List of hostnames to monitor
- `dns_servers`: List of DNS server IP addresses
- `query_timeout`: Timeout for DNS queries
- `query_interval`: Interval between resolution checks

#### Optional Fields
- `circuit_breaker`: Circuit breaker configuration
  - `threshold`: Number of failures before opening (default: 5)
  - `timeout`: Time to wait before resetting (default: "30s")
- `cache`: Cache configuration
  - `max_size`: Maximum number of cache entries (default: 1000)
- `health_port`: Health check endpoint port (default: 8080)
- `metrics_port`: Metrics endpoint port (default: 9090)
- `log_dir`: Log directory (default: "logs")

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