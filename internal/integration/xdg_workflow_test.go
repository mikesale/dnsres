//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestXDGWorkflowEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Build the binary first
	tempDir := t.TempDir()
	binPath := filepath.Join(tempDir, "dnsres-test")

	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/dnsres")
	buildCmd.Env = os.Environ()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	// Build from repo root (two levels up from internal/integration)
	buildCmd.Dir = filepath.Dir(filepath.Dir(cwd))
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, string(output))
	}

	t.Run("fresh install creates XDG config automatically", func(t *testing.T) {
		testDir := t.TempDir()
		xdgConfig := filepath.Join(testDir, "config")
		xdgState := filepath.Join(testDir, "state")

		cmd := exec.Command(binPath, "-host", "example.com")
		cmd.Env = append(os.Environ(),
			"XDG_CONFIG_HOME="+xdgConfig,
			"XDG_STATE_HOME="+xdgState,
		)
		cmd.Dir = testDir // Change to temp dir so no local config.json exists

		// Run for 2 seconds then kill
		if err := cmd.Start(); err != nil {
			t.Fatalf("failed to start dnsres: %v", err)
		}

		time.Sleep(2 * time.Second)
		cmd.Process.Kill()
		cmd.Wait()

		// Verify XDG config was created
		configPath := filepath.Join(xdgConfig, "dnsres", "config.json")
		if _, err := os.Stat(configPath); err != nil {
			t.Errorf("expected XDG config to be created at %s: %v", configPath, err)
		}

		// Verify config has required fields
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read config: %v", err)
		}

		var cfg map[string]interface{}
		if err := json.Unmarshal(data, &cfg); err != nil {
			t.Fatalf("failed to parse config: %v", err)
		}

		requiredFields := []string{"hostnames", "dns_servers", "query_interval"}
		for _, field := range requiredFields {
			if _, ok := cfg[field]; !ok {
				t.Errorf("expected config to have field %q", field)
			}
		}

		// Verify XDG state directory for logs was created
		logDir := filepath.Join(xdgState, "dnsres")
		if _, err := os.Stat(logDir); err != nil {
			t.Errorf("expected XDG state log directory at %s: %v", logDir, err)
		}

		// Verify log files exist
		logFiles := []string{"dnsres-success.log", "dnsres-error.log", "dnsres-app.log"}
		for _, logFile := range logFiles {
			path := filepath.Join(logDir, logFile)
			if _, err := os.Stat(path); err != nil {
				t.Errorf("expected log file %s to exist: %v", path, err)
			}
		}
	})

	t.Run("existing local config.json takes precedence", func(t *testing.T) {
		testDir := t.TempDir()
		xdgConfig := filepath.Join(testDir, "config")
		xdgState := filepath.Join(testDir, "state")

		// Create local config.json with custom settings
		localConfig := map[string]interface{}{
			"hostnames":      []string{"local-test.com"},
			"dns_servers":    []string{"1.1.1.1:53"},
			"query_interval": "10s",
		}
		localConfigData, _ := json.MarshalIndent(localConfig, "", "  ")
		localConfigPath := filepath.Join(testDir, "config.json")
		if err := os.WriteFile(localConfigPath, localConfigData, 0644); err != nil {
			t.Fatalf("failed to write local config: %v", err)
		}

		cmd := exec.Command(binPath, "-host", "example.com")
		cmd.Env = append(os.Environ(),
			"XDG_CONFIG_HOME="+xdgConfig,
			"XDG_STATE_HOME="+xdgState,
		)
		cmd.Dir = testDir

		if err := cmd.Start(); err != nil {
			t.Fatalf("failed to start dnsres: %v", err)
		}

		time.Sleep(2 * time.Second)
		cmd.Process.Kill()
		cmd.Wait()

		// Verify local config was NOT modified
		data, _ := os.ReadFile(localConfigPath)
		var cfg map[string]interface{}
		json.Unmarshal(data, &cfg)

		hostnames := cfg["hostnames"].([]interface{})
		if len(hostnames) != 1 || hostnames[0] != "local-test.com" {
			t.Error("expected local config to remain unchanged")
		}

		// Verify XDG config was NOT created (local takes precedence)
		xdgConfigPath := filepath.Join(xdgConfig, "dnsres", "config.json")
		if _, err := os.Stat(xdgConfigPath); !os.IsNotExist(err) {
			t.Error("expected XDG config to NOT be created when local config exists")
		}
	})

	t.Run("XDG fallback to HOME/logs when state dir fails", func(t *testing.T) {
		testDir := t.TempDir()
		homeDir := filepath.Join(testDir, "home")
		os.MkdirAll(homeDir, 0755)

		readOnlyState := filepath.Join(testDir, "readonly-state")
		os.Mkdir(readOnlyState, 0000)
		defer os.Chmod(readOnlyState, 0755)

		// Create minimal config
		configData := []byte(`{
			"hostnames": ["example.com"],
			"dns_servers": ["8.8.8.8:53"],
			"query_timeout": "5s",
			"query_interval": "5s",
			"health_port": 18880,
			"metrics_port": 19990,
			"instrumentation_level": "none",
			"circuit_breaker": {
				"threshold": 5,
				"timeout": "30s"
			},
			"cache": {
				"max_size": 1000
			}
		}`)
		configPath := filepath.Join(testDir, "config.json")
		os.WriteFile(configPath, configData, 0644)

		cmd := exec.Command(binPath, "-config", configPath)
		cmd.Env = append(os.Environ(),
			"XDG_STATE_HOME="+readOnlyState,
			"HOME="+homeDir,
		)
		cmd.Dir = testDir

		// Capture stderr to check for fallback message
		stderr, _ := cmd.StderrPipe()
		if err := cmd.Start(); err != nil {
			t.Fatalf("failed to start dnsres: %v", err)
		}

		time.Sleep(2 * time.Second)
		cmd.Process.Kill()

		// Read stderr output
		output := make([]byte, 1024)
		n, _ := stderr.Read(output)
		stderrText := string(output[:n])

		cmd.Wait()

		// Verify fallback to HOME/logs
		fallbackLogDir := filepath.Join(homeDir, "logs")
		if _, err := os.Stat(fallbackLogDir); err != nil {
			t.Errorf("expected fallback log directory at %s: %v", fallbackLogDir, err)
		}

		// Verify warning message about fallback (if implementation prints one)
		if !strings.Contains(stderrText, "logs") && !strings.Contains(stderrText, "fallback") {
			// Some implementations might not print warnings - just check the directory exists
			t.Logf("Fallback directory created successfully at %s", fallbackLogDir)
		}
	})

	t.Run("explicit -config flag overrides XDG", func(t *testing.T) {
		testDir := t.TempDir()
		xdgConfig := filepath.Join(testDir, "config")

		// Create custom config
		customConfigPath := filepath.Join(testDir, "custom-config.json")
		customConfig := map[string]interface{}{
			"hostnames":      []string{"custom.example.com"},
			"dns_servers":    []string{"1.1.1.1:53"},
			"query_interval": "5s",
		}
		customConfigData, _ := json.MarshalIndent(customConfig, "", "  ")
		os.WriteFile(customConfigPath, customConfigData, 0644)

		cmd := exec.Command(binPath, "-config", customConfigPath)
		cmd.Env = append(os.Environ(),
			"XDG_CONFIG_HOME="+xdgConfig,
		)
		cmd.Dir = testDir

		if err := cmd.Start(); err != nil {
			t.Fatalf("failed to start dnsres: %v", err)
		}

		time.Sleep(2 * time.Second)
		cmd.Process.Kill()
		cmd.Wait()

		// Verify custom config was used (not modified)
		data, _ := os.ReadFile(customConfigPath)
		var cfg map[string]interface{}
		json.Unmarshal(data, &cfg)

		hostnames := cfg["hostnames"].([]interface{})
		if hostnames[0] != "custom.example.com" {
			t.Error("expected custom config to be used and unchanged")
		}

		// Verify XDG config was NOT created
		xdgConfigPath := filepath.Join(xdgConfig, "dnsres", "config.json")
		if _, err := os.Stat(xdgConfigPath); !os.IsNotExist(err) {
			t.Error("expected XDG config to NOT be created when -config flag used")
		}
	})
}
