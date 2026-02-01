package tui

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"dnsres/internal/dnsres"

	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the interactive TUI application.
func Run() error {
	configFile := flag.String("config", "", "Path to configuration file (default: auto-detect)")
	hostname := flag.String("host", "", "Override hostname from config file")
	flag.Parse()

	args := flag.Args()
	var positionalHost string
	if len(args) > 0 {
		positionalHost = strings.TrimSpace(args[0])
	}

	// Resolve config path
	configPath, _, err := dnsres.ResolveConfigPath(*configFile)
	if err != nil {
		return fmt.Errorf("failed to resolve config path: %w", err)
	}

	var config *dnsres.Config
	if configPath == "" {
		config = dnsres.DefaultConfig()
	} else {
		config, err = dnsres.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	if positionalHost != "" {
		config.Hostnames = []string{positionalHost}
	} else if *hostname != "" {
		config.Hostnames = []string{*hostname}
	}

	if len(config.Hostnames) == 0 {
		return fmt.Errorf("hostname required: provide a domain as the first argument or use -host")
	}

	resolver, err := dnsres.NewDNSResolver(config)
	if err != nil {
		return fmt.Errorf("failed to create DNS resolver: %w", err)
	}
	resolver.SetOutputWriter(io.Discard)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- resolver.Start(ctx)
		close(errCh)
	}()

	events, unsubscribe := resolver.SubscribeEvents(200)
	model := newModel(resolver, config, cancel, events, unsubscribe, errCh)

	program := tea.NewProgram(model, tea.WithAltScreen())
	if err := program.Start(); err != nil {
		cancel()
		return fmt.Errorf("failed to start TUI: %w", err)
	}

	if err := <-errCh; err != nil {
		return fmt.Errorf("resolver stopped with error: %w", err)
	}

	return nil
}
