//go:build integration

package integration

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDNSResEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	runEndToEnd(t, "30s", 5*time.Minute, "30s", 3)
}

func TestDNSResEndToEndShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	runEndToEnd(t, "10s", time.Minute, "10s", 3)
}

func runEndToEnd(t *testing.T, interval string, runDuration time.Duration, expectedInterval string, minCycles int) {
	t.Helper()

	tempDir := t.TempDir()
	binPath := filepath.Join(tempDir, "dnsres")
	logDir := filepath.Join(tempDir, "logs")
	configPath := filepath.Join(tempDir, "config.json")

	config := map[string]interface{}{
		"hostnames":             []string{"google.com"},
		"dns_servers":           []string{"8.8.8.8", "1.1.1.1"},
		"query_timeout":         "5s",
		"query_interval":        interval,
		"health_port":           18880,
		"metrics_port":          19990,
		"log_dir":               logDir,
		"instrumentation_level": "low",
		"circuit_breaker": map[string]interface{}{
			"threshold": 5,
			"timeout":   "30s",
		},
		"cache": map[string]interface{}{
			"max_size": 1000,
		},
	}

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	if err := os.WriteFile(configPath, configJSON, 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	buildCmd := exec.Command("go", "build", "-o", binPath, "./")
	buildCmd.Env = os.Environ()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	buildCmd.Dir = filepath.Dir(filepath.Dir(cwd))
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, string(output))
	}

	ctx, cancel := context.WithTimeout(context.Background(), runDuration+30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binPath, "-config", configPath)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to get stdout: %v", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("failed to get stderr: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start dnsres: %v", err)
	}

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	var stdoutMu sync.Mutex
	var cycleStarts int
	var cycleCompletes int

	readStream := func(pipe *bufio.Reader, buf *bytes.Buffer, isStdout bool) {
		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			line := scanner.Text()
			buf.WriteString(line)
			buf.WriteByte('\n')
			if isStdout {
				stdoutMu.Lock()
				if strings.Contains(line, "Resolution cycle starting") {
					cycleStarts++
				}
				if strings.Contains(line, "Resolution cycle complete") {
					cycleCompletes++
				}
				stdoutMu.Unlock()
			}
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		readStream(bufio.NewReader(stdoutPipe), &stdoutBuf, true)
	}()
	go func() {
		defer wg.Done()
		readStream(bufio.NewReader(stderrPipe), &stderrBuf, false)
	}()

	select {
	case <-time.After(runDuration):
		_ = cmd.Process.Signal(os.Interrupt)
	case <-ctx.Done():
		if ctx.Err() != context.Canceled {
			t.Fatalf("dnsres exceeded timeout: %v", ctx.Err())
		}
	}

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			t.Fatalf("dnsres timed out: %v", ctx.Err())
		}
		t.Fatalf("dnsres exited with error: %v\nstdout:\n%s\nstderr:\n%s", err, stdoutBuf.String(), stderrBuf.String())
	}

	wg.Wait()

	stdout := stdoutBuf.String()
	if !strings.Contains(stdout, "Loading configuration") {
		t.Fatalf("expected config load output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Resolver initialized") {
		t.Fatalf("expected resolver initialized output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Resolution loop started (interval "+expectedInterval+")") {
		t.Fatalf("expected loop start output, got:\n%s", stdout)
	}
	if cycleStarts < minCycles || cycleCompletes < minCycles {
		t.Fatalf("expected at least %d cycles, got starts=%d completes=%d", minCycles, cycleStarts, cycleCompletes)
	}

	appLogPath := filepath.Join(logDir, "dnsres-app.log")
	successLogPath := filepath.Join(logDir, "dnsres-success.log")
	errorLogPath := filepath.Join(logDir, "dnsres-error.log")

	for _, path := range []string{appLogPath, successLogPath, errorLogPath} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected log file %s, got error: %v", path, err)
		}
	}

	appLog, err := os.ReadFile(appLogPath)
	if err != nil {
		t.Fatalf("failed to read app log: %v", err)
	}
	appLogText := string(appLog)
	if !strings.Contains(appLogText, "resolution cycle start") {
		t.Fatalf("expected resolution cycle start in app log")
	}
	if !strings.Contains(appLogText, "resolution cycle complete") {
		t.Fatalf("expected resolution cycle complete in app log")
	}
	if !strings.Contains(appLogText, "resolution tick fired interval="+expectedInterval) {
		t.Fatalf("expected resolution tick in app log")
	}
}
