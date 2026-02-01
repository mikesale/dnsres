package app

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"dnsres/internal/dnsres"
)

func Run() error {
	// Parse command line flags
	configFile := flag.String("config", "", "Path to configuration file (default: auto-detect)")
	reportMode := flag.Bool("report", false, "Generate statistics report")
	hostname := flag.String("host", "", "Override hostname from config file")
	flag.Parse()

	args := flag.Args()
	var positionalHost string
	if len(args) > 0 {
		positionalHost = strings.TrimSpace(args[0])
	}

	// Resolve config path
	configPath, wasCreated, err := dnsres.ResolveConfigPath(*configFile)
	if err != nil {
		return fmt.Errorf("failed to resolve config path: %w", err)
	}

	var config *dnsres.Config
	if configPath == "" {
		fmt.Println("No configuration file found; using built-in defaults")
		config = dnsres.DefaultConfig()
	} else {
		fmt.Printf("Loading configuration from %s\n", configPath)
		if wasCreated {
			fmt.Printf("Created default configuration file at %s\n", configPath)
		}

		config, err = dnsres.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		fmt.Println("Configuration loaded")
	}

	// Override hostname if specified
	if positionalHost != "" {
		config.Hostnames = []string{positionalHost}
		fmt.Printf("Hostname set from CLI: %s\n", positionalHost)
	} else if *hostname != "" {
		config.Hostnames = []string{*hostname}
		fmt.Printf("Hostname override enabled: %s\n", *hostname)
	}

	if len(config.Hostnames) == 0 {
		return fmt.Errorf("hostname required: provide a domain as the first argument or use -host")
	}

	// Create resolver
	fmt.Println("Validating configuration")
	resolver, err := dnsres.NewDNSResolver(config)
	if err != nil {
		return fmt.Errorf("failed to create DNS resolver: %w", err)
	}
	fmt.Println("Resolver initialized")

	// Report log directory fallback
	if resolver.LogDirWasFallback() {
		fmt.Printf("\nNote: Using fallback log directory at %s\n", resolver.GetLogDir())
		fmt.Printf("(XDG state directory unavailable)\n\n")
	}

	// Handle report mode
	if *reportMode {
		fmt.Println("Report mode enabled; generating report")
		fmt.Println(resolver.GenerateReport())
		return nil
	}

	fmt.Printf("Monitoring %d hostnames across %d DNS servers every %s\n", len(config.Hostnames), len(config.DNSServers), config.QueryInterval.Duration)
	fmt.Println("Press q then Enter to quit")

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		fmt.Printf("Shutdown signal received (%s)\n", sig)
		cancel()
	}()

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input := strings.TrimSpace(scanner.Text())
			if strings.EqualFold(input, "q") {
				fmt.Println("Quit requested; shutting down")
				cancel()
				return
			}
		}
	}()

	// Start resolution
	if err := resolver.Start(ctx); err != nil {
		return fmt.Errorf("failed to start DNS resolver: %w", err)
	}

	return nil
}
