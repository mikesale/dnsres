package main

import (
	"log"

	"dnsres/internal/tui"
)

func main() {
	if err := tui.Run(); err != nil {
		log.Fatalf("dnsres-tui exited with error: %v", err)
	}
}
