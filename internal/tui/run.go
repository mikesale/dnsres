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
	configFile := flag.String("config", "config.json", "Path to configuration file")
	hostname := flag.String("host", "", "Override hostname from config file")
	flag.Parse()

	args := flag.Args()
	var positionalHost string
	if len(args) > 0 {
		positionalHost = strings.TrimSpace(args[0])
	}

	config, err := dnsres.LoadConfig(*configFile)
	if err != nil {
		fmt.Printf("Config load failed (%s); using built-in defaults\n", err)
		config = dnsres.DefaultConfig()
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
