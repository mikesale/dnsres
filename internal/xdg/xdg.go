package xdg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ConfigHome returns the XDG config home directory.
// Falls back to ~/.config if XDG_CONFIG_HOME is not set.
func ConfigHome() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config")
}

// StateHome returns the XDG state home directory.
// Falls back to ~/.local/state if XDG_STATE_HOME is not set.
func StateHome() string {
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return xdg
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".local", "state")
}

// DataHome returns the XDG data home directory.
// Falls back to ~/.local/share if XDG_DATA_HOME is not set.
func DataHome() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return xdg
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".local", "share")
}

// ConfigFile returns the path to the dnsres config file.
// If the file doesn't exist, it creates it with minimal defaults.
// Returns: (path, wasCreated, error)
func ConfigFile() (string, bool, error) {
	configDir := filepath.Join(ConfigHome(), "dnsres")
	configPath := filepath.Join(configDir, "config.json")

	// If config exists, return it
	if fileExists(configPath) {
		return configPath, false, nil
	}

	// Create directory
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", false, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create minimal default config file
	if err := createMinimalConfig(configPath); err != nil {
		return "", false, fmt.Errorf("failed to create default config: %w", err)
	}

	return configPath, true, nil
}

// EnsureStateDir ensures the dnsres state directory exists.
// Falls back to $HOME/logs if XDG state directory cannot be created.
// Returns: (path, wasFallback, error)
func EnsureStateDir() (string, bool, error) {
	// Try XDG state directory first
	stateDir := filepath.Join(StateHome(), "dnsres")
	if err := os.MkdirAll(stateDir, 0755); err == nil {
		return stateDir, false, nil
	}

	// Fall back to $HOME/logs
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", false, fmt.Errorf("failed to get home directory: %w", err)
	}

	fallbackDir := filepath.Join(homeDir, "logs")
	if err := os.MkdirAll(fallbackDir, 0755); err != nil {
		return "", false, fmt.Errorf("failed to create log directory: %w", err)
	}

	return fallbackDir, true, nil
}

// fileExists checks if a file exists and is not a directory.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// createMinimalConfig creates a minimal config file with essential fields only.
func createMinimalConfig(path string) error {
	// Minimal config with all required fields and reasonable defaults
	// Users can override hostnames via command line or edit this file
	config := map[string]interface{}{
		"hostnames":             []string{"example.com"},
		"dns_servers":           []string{"8.8.8.8:53", "1.1.1.1:53", "9.9.9.9:53"},
		"query_timeout":         "5s",
		"query_interval":        "30s",
		"health_port":           8880,
		"metrics_port":          9990,
		"log_dir":               "",
		"instrumentation_level": "none",
		"circuit_breaker": map[string]interface{}{
			"threshold": 5,
			"timeout":   "30s",
		},
		"cache": map[string]interface{}{
			"max_size": 1000,
		},
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}
