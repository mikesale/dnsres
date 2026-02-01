package dnsres

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"dnsres/internal/xdg"
)

// setupLoggers initializes the loggers
// Returns: (successLog, errorLog, appLog, actualPath, wasFallback, error)
func setupLoggers(logDir string) (*log.Logger, *log.Logger, *log.Logger, string, bool, error) {
	wasFallback := false

	// If empty or default "logs", use XDG
	if logDir == "" || logDir == "logs" {
		dir, isFallback, err := xdg.EnsureStateDir()
		if err != nil {
			return nil, nil, nil, "", false, fmt.Errorf("failed to create log directory: %w", err)
		}
		logDir = dir
		wasFallback = isFallback
	} else {
		// User specified explicit path
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, nil, nil, "", false, fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	successLogFile, err := os.OpenFile(
		filepath.Join(logDir, "dnsres-success.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return nil, nil, nil, "", false, fmt.Errorf("failed to open success log file: %w", err)
	}

	errorLogFile, err := os.OpenFile(
		filepath.Join(logDir, "dnsres-error.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return nil, nil, nil, "", false, fmt.Errorf("failed to open error log file: %w", err)
	}

	appLogFile, err := os.OpenFile(
		filepath.Join(logDir, "dnsres-app.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return nil, nil, nil, "", false, fmt.Errorf("failed to open app log file: %w", err)
	}

	successLog := log.New(successLogFile, "", log.LstdFlags)
	errorLog := log.New(errorLogFile, "", log.LstdFlags)
	appLog := log.New(appLogFile, "", log.LstdFlags)

	return successLog, errorLog, appLog, logDir, wasFallback, nil
}
