package main

import "log"

func main() {
	if err := run(); err != nil {
		log.Fatalf("dnsres exited with error: %v", err)
	}
}
