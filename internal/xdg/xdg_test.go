package xdg

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigHome(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		wantDir  string
	}{
		{
			name:     "uses XDG_CONFIG_HOME when set",
			envValue: "/custom/config",
			wantDir:  "/custom/config",
		},
		{
			name:     "falls back to ~/.config when not set",
			envValue: "",
			wantDir:  "", // will check it contains .config
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldValue := os.Getenv("XDG_CONFIG_HOME")
			defer os.Setenv("XDG_CONFIG_HOME", oldValue)

			if tt.envValue != "" {
				os.Setenv("XDG_CONFIG_HOME", tt.envValue)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}

			got := ConfigHome()
			if tt.wantDir != "" && got != tt.wantDir {
				t.Errorf("ConfigHome() = %v, want %v", got, tt.wantDir)
			}
			if tt.wantDir == "" && got == "" {
				t.Error("ConfigHome() should not return empty string")
			}
		})
	}
}

func TestStateHome(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		wantDir  string
	}{
		{
			name:     "uses XDG_STATE_HOME when set",
			envValue: "/custom/state",
			wantDir:  "/custom/state",
		},
		{
			name:     "falls back to ~/.local/state when not set",
			envValue: "",
			wantDir:  "", // will check it contains .local/state
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldValue := os.Getenv("XDG_STATE_HOME")
			defer os.Setenv("XDG_STATE_HOME", oldValue)

			if tt.envValue != "" {
				os.Setenv("XDG_STATE_HOME", tt.envValue)
			} else {
				os.Unsetenv("XDG_STATE_HOME")
			}

			got := StateHome()
			if tt.wantDir != "" && got != tt.wantDir {
				t.Errorf("StateHome() = %v, want %v", got, tt.wantDir)
			}
			if tt.wantDir == "" && got == "" {
				t.Error("StateHome() should not return empty string")
			}
		})
	}
}

func TestDataHome(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		wantDir  string
	}{
		{
			name:     "uses XDG_DATA_HOME when set",
			envValue: "/custom/data",
			wantDir:  "/custom/data",
		},
		{
			name:     "falls back to ~/.local/share when not set",
			envValue: "",
			wantDir:  "", // will check it contains .local/share
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldValue := os.Getenv("XDG_DATA_HOME")
			defer os.Setenv("XDG_DATA_HOME", oldValue)

			if tt.envValue != "" {
				os.Setenv("XDG_DATA_HOME", tt.envValue)
			} else {
				os.Unsetenv("XDG_DATA_HOME")
			}

			got := DataHome()
			if tt.wantDir != "" && got != tt.wantDir {
				t.Errorf("DataHome() = %v, want %v", got, tt.wantDir)
			}
			if tt.wantDir == "" && got == "" {
				t.Error("DataHome() should not return empty string")
			}
		})
	}
}

func TestConfigFile(t *testing.T) {
	tempDir := t.TempDir()

	oldValue := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldValue)
	os.Setenv("XDG_CONFIG_HOME", tempDir)

	t.Run("creates config file when it does not exist", func(t *testing.T) {
		configPath, wasCreated, err := ConfigFile()
		if err != nil {
			t.Fatalf("ConfigFile() error = %v", err)
		}
		if !wasCreated {
			t.Error("ConfigFile() should report file was created")
		}

		expectedPath := filepath.Join(tempDir, "dnsres", "config.json")
		if configPath != expectedPath {
			t.Errorf("ConfigFile() path = %v, want %v", configPath, expectedPath)
		}

		// Verify file exists
		if _, err := os.Stat(configPath); err != nil {
			t.Errorf("config file should exist: %v", err)
		}

		// Verify file permissions
		info, err := os.Stat(configPath)
		if err != nil {
			t.Fatalf("failed to stat config file: %v", err)
		}
		if info.Mode().Perm() != 0644 {
			t.Errorf("config file permissions = %v, want 0644", info.Mode().Perm())
		}

		// Verify it's valid JSON with expected fields
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read config file: %v", err)
		}

		var config map[string]interface{}
		if err := json.Unmarshal(data, &config); err != nil {
			t.Fatalf("config file is not valid JSON: %v", err)
		}

		// Check essential fields
		requiredFields := []string{"hostnames", "dns_servers", "query_timeout", "query_interval"}
		for _, field := range requiredFields {
			if _, ok := config[field]; !ok {
				t.Errorf("config missing required field: %s", field)
			}
		}
	})

	t.Run("returns existing config file without creating new one", func(t *testing.T) {
		// Call again
		configPath, wasCreated, err := ConfigFile()
		if err != nil {
			t.Fatalf("ConfigFile() error = %v", err)
		}
		if wasCreated {
			t.Error("ConfigFile() should report file was not created when it exists")
		}

		expectedPath := filepath.Join(tempDir, "dnsres", "config.json")
		if configPath != expectedPath {
			t.Errorf("ConfigFile() path = %v, want %v", configPath, expectedPath)
		}
	})
}

func TestEnsureStateDir(t *testing.T) {
	tempDir := t.TempDir()

	oldValue := os.Getenv("XDG_STATE_HOME")
	defer os.Setenv("XDG_STATE_HOME", oldValue)
	os.Setenv("XDG_STATE_HOME", tempDir)

	t.Run("creates state directory when it does not exist", func(t *testing.T) {
		stateDir, wasFallback, err := EnsureStateDir()
		if err != nil {
			t.Fatalf("EnsureStateDir() error = %v", err)
		}
		if wasFallback {
			t.Error("EnsureStateDir() should not use fallback when XDG dir can be created")
		}

		expectedPath := filepath.Join(tempDir, "dnsres")
		if stateDir != expectedPath {
			t.Errorf("EnsureStateDir() path = %v, want %v", stateDir, expectedPath)
		}

		// Verify directory exists
		info, err := os.Stat(stateDir)
		if err != nil {
			t.Errorf("state directory should exist: %v", err)
		}
		if !info.IsDir() {
			t.Error("state path should be a directory")
		}

		// Verify directory permissions
		if info.Mode().Perm() != 0755 {
			t.Errorf("state directory permissions = %v, want 0755", info.Mode().Perm())
		}
	})

	t.Run("returns existing state directory", func(t *testing.T) {
		stateDir, wasFallback, err := EnsureStateDir()
		if err != nil {
			t.Fatalf("EnsureStateDir() error = %v", err)
		}
		if wasFallback {
			t.Error("EnsureStateDir() should not use fallback")
		}

		expectedPath := filepath.Join(tempDir, "dnsres")
		if stateDir != expectedPath {
			t.Errorf("EnsureStateDir() path = %v, want %v", stateDir, expectedPath)
		}
	})
}

func TestEnsureStateDirFallback(t *testing.T) {
	// Create a read-only directory to force fallback
	tempDir := t.TempDir()
	readOnlyDir := filepath.Join(tempDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0000); err != nil {
		t.Fatalf("failed to create readonly dir: %v", err)
	}
	defer func() {
		if err := os.Chmod(readOnlyDir, 0755); err != nil {
			t.Logf("warning: failed to restore permissions on readonly dir: %v", err)
		}
	}()

	oldValue := os.Getenv("XDG_STATE_HOME")
	defer os.Setenv("XDG_STATE_HOME", oldValue)
	os.Setenv("XDG_STATE_HOME", readOnlyDir)

	t.Run("falls back to HOME/logs when XDG state dir cannot be created", func(t *testing.T) {
		stateDir, wasFallback, err := EnsureStateDir()
		if err != nil {
			t.Fatalf("EnsureStateDir() error = %v", err)
		}
		if !wasFallback {
			t.Error("EnsureStateDir() should use fallback when XDG dir cannot be created")
		}

		homeDir, _ := os.UserHomeDir()
		expectedPath := filepath.Join(homeDir, "logs")
		if stateDir != expectedPath {
			t.Errorf("EnsureStateDir() fallback path = %v, want %v", stateDir, expectedPath)
		}
	})
}

func TestFileExists(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("returns false for non-existent file", func(t *testing.T) {
		path := filepath.Join(tempDir, "nonexistent.txt")
		if fileExists(path) {
			t.Error("fileExists() should return false for non-existent file")
		}
	})

	t.Run("returns true for existing file", func(t *testing.T) {
		path := filepath.Join(tempDir, "exists.txt")
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		if !fileExists(path) {
			t.Error("fileExists() should return true for existing file")
		}
	})

	t.Run("returns false for directory", func(t *testing.T) {
		dirPath := filepath.Join(tempDir, "testdir")
		if err := os.Mkdir(dirPath, 0755); err != nil {
			t.Fatalf("failed to create test directory: %v", err)
		}

		if fileExists(dirPath) {
			t.Error("fileExists() should return false for directory")
		}
	})
}

func TestCreateMinimalConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	if err := createMinimalConfig(configPath); err != nil {
		t.Fatalf("createMinimalConfig() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("config file should exist: %v", err)
	}

	// Verify file permissions
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("failed to stat config file: %v", err)
	}
	if info.Mode().Perm() != 0644 {
		t.Errorf("config file permissions = %v, want 0644", info.Mode().Perm())
	}

	// Verify JSON is valid and contains required fields
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("config file is not valid JSON: %v", err)
	}

	// Check all required fields exist
	requiredFields := map[string]bool{
		"hostnames":             true,
		"dns_servers":           true,
		"query_timeout":         true,
		"query_interval":        true,
		"health_port":           true,
		"metrics_port":          true,
		"log_dir":               true,
		"instrumentation_level": true,
		"circuit_breaker":       true,
		"cache":                 true,
	}

	for field := range requiredFields {
		if _, ok := config[field]; !ok {
			t.Errorf("config missing required field: %s", field)
		}
	}

	// Verify specific values
	if hostnames, ok := config["hostnames"].([]interface{}); !ok || len(hostnames) == 0 {
		t.Error("hostnames should be a non-empty array")
	}

	if servers, ok := config["dns_servers"].([]interface{}); !ok || len(servers) == 0 {
		t.Error("dns_servers should be a non-empty array")
	}

	if timeout, ok := config["query_timeout"].(string); !ok || timeout != "5s" {
		t.Errorf("query_timeout = %v, want 5s", config["query_timeout"])
	}

	if interval, ok := config["query_interval"].(string); !ok || interval != "30s" {
		t.Errorf("query_interval = %v, want 30s", config["query_interval"])
	}
}
