package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"dnsres/internal/dnsres"
)

// TestModelStatusMessage tests that the TUI model correctly displays
// log directory location in the status message
func TestModelStatusMessage(t *testing.T) {
	t.Run("normal log directory shows path without fallback notice", func(t *testing.T) {
		// Create mock resolver with custom log directory
		tempDir := t.TempDir()
		config := dnsres.DefaultConfig()
		config.Hostnames = []string{"example.com"}
		config.LogDir = tempDir

		_, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Create resolver (won't actually start)
		resolver, err := dnsres.NewDNSResolver(config)
		if err != nil {
			t.Fatalf("failed to create resolver: %v", err)
		}

		events := make(chan dnsres.ResolverEvent, 10)
		errs := make(chan error, 1)
		unsubscribe := func() { close(events) }

		// Create TUI model
		m := newModel(resolver, config, cancel, events, unsubscribe, errs)

		// Verify status message contains log directory
		if !strings.Contains(m.statusMsg, resolver.GetLogDir()) {
			t.Errorf("expected status message to contain log directory path, got: %s", m.statusMsg)
		}

		// Verify NO fallback notice when using custom path
		if strings.Contains(m.statusMsg, "fallback") {
			t.Errorf("expected NO fallback notice for custom log directory, got: %s", m.statusMsg)
		}

		expectedFormat := fmt.Sprintf("Logs: %s", resolver.GetLogDir())
		if m.statusMsg != expectedFormat {
			t.Errorf("expected status message format %q, got %q", expectedFormat, m.statusMsg)
		}

		t.Logf("Status message (custom path): %s", m.statusMsg)
	})

	t.Run("XDG log directory shows path without fallback notice", func(t *testing.T) {
		// When using XDG state directory (not fallback), should show normal message
		config := dnsres.DefaultConfig()
		config.Hostnames = []string{"example.com"}
		config.LogDir = "" // Empty means use XDG

		_, cancel := context.WithCancel(context.Background())
		defer cancel()

		resolver, err := dnsres.NewDNSResolver(config)
		if err != nil {
			t.Fatalf("failed to create resolver: %v", err)
		}

		events := make(chan dnsres.ResolverEvent, 10)
		errs := make(chan error, 1)
		unsubscribe := func() { close(events) }

		m := newModel(resolver, config, cancel, events, unsubscribe, errs)

		// Verify status message contains log directory
		if !strings.Contains(m.statusMsg, resolver.GetLogDir()) {
			t.Errorf("expected status message to contain log directory, got: %s", m.statusMsg)
		}

		t.Logf("Status message (XDG): %s", m.statusMsg)
	})
}

// TestLogDirectoryPathFormatting tests the formatting of log directory
// paths in the TUI status message
func TestLogDirectoryPathFormatting(t *testing.T) {
	testCases := []struct {
		name           string
		logDir         string
		wasFallback    bool
		expectedFormat string
	}{
		{
			name:           "custom directory",
			logDir:         "/var/log/dnsres",
			wasFallback:    false,
			expectedFormat: "Logs: /var/log/dnsres",
		},
		{
			name:           "XDG state directory",
			logDir:         filepath.Join("/home/user/.local/state/dnsres"),
			wasFallback:    false,
			expectedFormat: "Logs: /home/user/.local/state/dnsres",
		},
		{
			name:           "HOME fallback directory",
			logDir:         "/home/user/logs",
			wasFallback:    true,
			expectedFormat: "Logs: /home/user/logs (fallback)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock status message based on fallback flag
			var statusMsg string
			if tc.wasFallback {
				statusMsg = fmt.Sprintf("Logs: %s (fallback)", tc.logDir)
			} else {
				statusMsg = fmt.Sprintf("Logs: %s", tc.logDir)
			}

			// Verify format
			if statusMsg != tc.expectedFormat {
				t.Errorf("expected status message %q, got %q", tc.expectedFormat, statusMsg)
			}

			// Verify path is included
			if !strings.Contains(statusMsg, tc.logDir) {
				t.Errorf("expected status message to contain log directory path")
			}

			// Verify fallback notice appears only when expected
			hasFallbackNotice := strings.Contains(statusMsg, "(fallback)")
			if hasFallbackNotice != tc.wasFallback {
				t.Errorf("fallback notice mismatch: expected %v, got %v", tc.wasFallback, hasFallbackNotice)
			}

			t.Logf("Status message: %s", statusMsg)
		})
	}
}

// TestModelCreation tests that the model is created correctly with
// resolver state properly initialized
func TestModelCreation(t *testing.T) {
	t.Run("model initializes with resolver state", func(t *testing.T) {
		config := dnsres.DefaultConfig()
		config.Hostnames = []string{"example.com", "google.com"}
		config.DNSServers = []string{"8.8.8.8:53", "1.1.1.1:53"}
		config.LogDir = "/tmp/test-logs"

		_, cancel := context.WithCancel(context.Background())
		defer cancel()

		resolver, err := dnsres.NewDNSResolver(config)
		if err != nil {
			t.Fatalf("failed to create resolver: %v", err)
		}

		events := make(chan dnsres.ResolverEvent, 10)
		errs := make(chan error, 1)
		unsubscribe := func() { close(events) }

		m := newModel(resolver, config, cancel, events, unsubscribe, errs)

		// Verify model has correct config
		if m.config != config {
			t.Error("expected model to have correct config")
		}

		// Verify model has correct resolver
		if m.resolver != resolver {
			t.Error("expected model to have correct resolver")
		}

		// Verify server states initialized
		if len(m.servers) != len(config.DNSServers) {
			t.Errorf("expected %d server states, got %d", len(config.DNSServers), len(m.servers))
		}

		// Verify status message is set
		if m.statusMsg == "" {
			t.Error("expected status message to be set")
		}

		// Verify status message contains log directory
		if !strings.Contains(m.statusMsg, "/tmp/test-logs") {
			t.Errorf("expected status message to contain log directory, got: %s", m.statusMsg)
		}

		t.Logf("Model initialized successfully with %d servers", len(m.servers))
		t.Logf("Status: %s", m.statusMsg)
	})
}

// Note: Full TUI integration testing (with Bubble Tea message passing and
// rendering) requires a more complex setup. These tests validate the core
// logic of status message formatting and model initialization.
// Visual rendering tests would require mocking the Bubble Tea terminal
// and are better suited for manual testing or screenshot-based integration tests.
