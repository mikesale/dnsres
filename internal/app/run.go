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
	configFile := flag.String("config", "config.json", "Path to configuration file")
	reportMode := flag.Bool("report", false, "Generate statistics report")
	hostname := flag.String("host", "", "Override hostname from config file")
	flag.Parse()

	// Load configuration
	fmt.Printf("Loading configuration from %s\n", *configFile)
	config, err := dnsres.LoadConfig(*configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	fmt.Println("Configuration loaded")

	// Override hostname if specified
	if *hostname != "" {
		config.Hostnames = []string{*hostname}
		fmt.Printf("Hostname override enabled: %s\n", *hostname)
	}

	// Create resolver
	fmt.Println("Validating configuration")
	resolver, err := dnsres.NewDNSResolver(config)
	if err != nil {
		return fmt.Errorf("failed to create DNS resolver: %w", err)
	}
	fmt.Println("Resolver initialized")

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
