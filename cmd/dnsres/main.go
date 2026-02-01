package main

import (
	"log"

	"dnsres/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("dnsres exited with error: %v", err)
	}
}
