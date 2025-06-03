# Changelog

All notable changes to the DNS Resolution Monitor project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Circuit breaker pattern implementation
- Prometheus metrics collection
- Health check endpoints
- DNS response caching
- Log file support
- Statistics reporting
- Command-line options for hostname override and reporting
- Configuration validation
- Graceful shutdown handling

### Changed
- Improved error handling
- Enhanced logging with circuit breaker states
- Updated configuration format
- Improved metrics collection

### Fixed
- DNS response validation
- Cache cleanup
- Health check implementation
- Metrics registration

## [0.1.0] - 2024-03-14

### Added
- Initial release
- Basic DNS resolution monitoring
- Configuration file support
- Simple logging
- Basic error handling 