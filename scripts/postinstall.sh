#!/bin/sh
# Post-install script for dnsres packages

# Create config directory if it doesn't exist
if [ ! -d "/etc/dnsres" ]; then
  mkdir -p /etc/dnsres
fi

echo "dnsres installed successfully!"
echo "Configuration example: /etc/dnsres/config.json.example"
echo "User config location: ~/.config/dnsres/config.json"
echo "Documentation: https://github.com/mikesale/dnsres"
echo "Report issues: https://github.com/mikesale/dnsres/issues"
