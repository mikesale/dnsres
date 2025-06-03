# Development Guide

## Prerequisites

- Go 1.21 or later
- Make (optional, for using Makefile)
- Docker (optional, for containerized development)

## Development Setup

1. Clone the repository:
```bash
git clone https://github.com/yourusername/dnsres.git
cd dnsres
```

2. Install dependencies:
```bash
go mod download
```

3. Build the project:
```bash
go build -o dnsres
```

## Project Structure

```
dnsres/
├── cmd/
│   └── dnsres/
│       └── main.go
├── internal/
│   ├── cache/
│   │   └── cache.go
│   ├── circuitbreaker/
│   │   ├── circuitbreaker.go
│   │   └── errors.go
│   ├── health/
│   │   └── health.go
│   └── metrics/
│       └── metrics.go
├── pkg/
│   └── dnsanalysis/
│       └── dnsanalysis.go
├── config.json
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Development Workflow

1. Create a new branch for your feature:
```bash
git checkout -b feature/your-feature-name
```

2. Make your changes and run tests:
```bash
go test ./...
```

3. Run linters:
```bash
go vet ./...
golangci-lint run
```

4. Commit your changes:
```bash
git add .
git commit -m "feat: your feature description"
```

5. Push your branch and create a pull request.

## Testing

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/cache
```

### Writing Tests

1. Create test files with `_test.go` suffix
2. Use table-driven tests for multiple test cases
3. Mock external dependencies using interfaces
4. Test both success and failure scenarios

Example test:
```go
func TestDNSResolver_Resolve(t *testing.T) {
    tests := []struct {
        name     string
        hostname string
        wantErr  bool
    }{
        {
            name:     "valid hostname",
            hostname: "example.com",
            wantErr:  false,
        },
        {
            name:     "invalid hostname",
            hostname: "invalid",
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            r := NewDNSResolver(Config{})
            _, err := r.Resolve(tt.hostname)
            if (err != nil) != tt.wantErr {
                t.Errorf("DNSResolver.Resolve() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Code Style

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for code formatting
- Keep functions small and focused
- Use meaningful variable and function names
- Add comments for exported functions and types

## Error Handling

1. Use custom error types for specific error cases
2. Wrap errors with context using `fmt.Errorf`
3. Check error types using `errors.Is` and `errors.As`
4. Log errors with appropriate context

Example:
```go
if err != nil {
    return nil, fmt.Errorf("failed to resolve %s: %w", hostname, err)
}
```

## Logging

1. Use structured logging with appropriate log levels
2. Include relevant context in log messages
3. Use separate log files for different concerns
4. Implement log rotation

Example:
```go
log.Printf("Resolved %s using %s (state: %s)", hostname, server, state)
```

## Metrics

1. Use Prometheus metrics for monitoring
2. Add appropriate labels to metrics
3. Document metric names and labels
4. Use appropriate metric types (Counter, Gauge, Histogram)

Example:
```go
metrics.DNSResolutionTotal.WithLabelValues(hostname, server).Inc()
```

## Documentation

1. Keep README.md up to date
2. Document all exported functions and types
3. Include examples in documentation
4. Update API documentation when making changes

Example:
```go
// Resolve performs DNS resolution for the given hostname.
// It returns the resolved IP addresses and any error encountered.
func (r *DNSResolver) Resolve(hostname string) ([]string, error) {
    // ...
}
```

## Release Process

1. Update version in code and documentation
2. Update CHANGELOG.md
3. Create a new tag
4. Build and test release artifacts
5. Create GitHub release

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Update documentation
6. Submit a pull request

## Code Review Process

1. All changes require at least one reviewer
2. CI checks must pass
3. Code must be properly documented
4. Tests must be included
5. Changes must follow the project's style guide

## Support

For questions and support, please:
1. Check the documentation
2. Search existing issues
3. Create a new issue if needed
4. Contact maintainers for urgent issues 