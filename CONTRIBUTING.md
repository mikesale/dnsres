# Contributing to DNS Resolution Monitor

Thank you for your interest in contributing to the DNS Resolution Monitor project! This document provides guidelines and instructions for contributing.

## Development Setup

1. Ensure you have Go 1.21 or later installed
2. Fork and clone the repository
3. Install development dependencies:
   ```bash
   go mod download
   ```
4. Install required tools:
   ```bash
   go install golang.org/x/tools/cmd/goimports@latest
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

## Code Style

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `goimports` to format your code
- Run `golangci-lint` before submitting changes
- Write tests for new functionality
- Update documentation for any changes

## Pull Request Process

1. Create a new branch for your changes
2. Make your changes
3. Run tests and linters:
   ```bash
   go test ./...
   golangci-lint run
   ```
4. Update documentation if needed
5. Submit a pull request with a clear description of changes

## Testing

- Write unit tests for new functionality
- Run all tests before submitting:
  ```bash
  go test -v ./...
  ```
- Ensure test coverage is maintained or improved

## Documentation

- Update README.md for user-facing changes
- Add comments for complex code
- Update API documentation if needed
- Add examples for new features

## Project Structure

```
dnsres/
├── cache/           # DNS response caching
├── circuitbreaker/  # Circuit breaker implementation
├── dnsanalysis/     # DNS response analysis
├── health/          # Health check functionality
├── metrics/         # Prometheus metrics
├── dnsres.go        # Main application
├── examples         # Example configuration
│   └── config.json  # Sample configuration file
└── README.md        # Project documentation
```

## Adding New Features

1. Create a new branch
2. Implement the feature
3. Add tests
4. Update documentation
5. Submit a pull request

## Reporting Issues

- Use the GitHub issue tracker
- Include steps to reproduce
- Provide relevant logs
- Specify your environment

## Code of Conduct

- Be respectful and inclusive
- Focus on the technical aspects
- Help others learn and improve
- Give credit where due

## License

By contributing, you agree that your contributions will be licensed under the project's GNU General Public License v3.0. 
