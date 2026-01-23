# Integration Testing

This repository includes a long-running end-to-end integration test that
executes the `dnsres` binary against real DNS servers and validates CLI output
and log files. The test is opt-in and intended for one-off runs, not CI.

## Running the End-to-End Test

Run the test with the integration build tag:

```bash
go test -tags=integration ./tests/integration -run TestDNSResEndToEnd -count=1
```

Notes:
- The test runs for about 5 minutes and uses a 30-second interval.
- It builds a temporary `dnsres` binary and writes a temporary config.
- It targets `google.com` and the DNS servers `8.8.8.8` and `1.1.1.1`.
- Logs are written into a temporary directory and validated by the test.

## Skipping

The test is skipped automatically when run with `-short`.
