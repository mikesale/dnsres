#!/bin/bash
# dnsres installation script
# Usage: curl -sSL https://raw.githubusercontent.com/mikesale/dnsres/main/install.sh | bash

set -e

# Configuration
REPO="mikesale/dnsres"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.config/dnsres"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print colored output
print_info() {
    echo -e "${GREEN}==>${NC} $1"
}

print_error() {
    echo -e "${RED}Error:${NC} $1" >&2
}

print_warning() {
    echo -e "${YELLOW}Warning:${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    local os=""
    local arch=""
    
    # Detect OS
    case "$(uname -s)" in
        Darwin)
            os="Darwin"
            ;;
        Linux)
            os="Linux"
            ;;
        *)
            print_error "Unsupported operating system: $(uname -s)"
            print_error "This script supports macOS (Darwin) and Linux only."
            exit 1
            ;;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)
            arch="x86_64"
            ;;
        aarch64|arm64)
            arch="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $(uname -m)"
            print_error "This script supports x86_64 and arm64 only."
            exit 1
            ;;
    esac
    
    echo "${os}_${arch}"
}

# Get latest release version from GitHub
get_latest_version() {
    local version
    version=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$version" ]; then
        print_error "Failed to fetch latest version from GitHub"
        exit 1
    fi
    
    echo "$version"
}

# Download and install dnsres
install_dnsres() {
    local platform="$1"
    local version="$2"
    local archive_name="dnsres_${version#v}_${platform}.tar.gz"
    local download_url="https://github.com/${REPO}/releases/download/${version}/${archive_name}"
    local temp_dir
    temp_dir=$(mktemp -d)
    
    print_info "Downloading dnsres ${version} for ${platform}..."
    
    if ! curl -sL "$download_url" -o "${temp_dir}/${archive_name}"; then
        print_error "Failed to download dnsres from ${download_url}"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    print_info "Extracting archive..."
    tar -xzf "${temp_dir}/${archive_name}" -C "$temp_dir"
    
    # Check if we need sudo for installation
    if [ -w "$INSTALL_DIR" ]; then
        SUDO=""
    else
        SUDO="sudo"
        print_warning "Installation requires sudo privileges for ${INSTALL_DIR}"
    fi
    
    print_info "Installing binaries to ${INSTALL_DIR}..."
    $SUDO install -m 755 "${temp_dir}/dnsres" "${INSTALL_DIR}/dnsres"
    $SUDO install -m 755 "${temp_dir}/dnsres-tui" "${INSTALL_DIR}/dnsres-tui"
    
    # Create config directory if it doesn't exist
    if [ ! -d "$CONFIG_DIR" ]; then
        print_info "Creating config directory: ${CONFIG_DIR}"
        mkdir -p "$CONFIG_DIR"
    fi
    
    # Copy example config if it exists in the archive
    if [ -f "${temp_dir}/examples/config.json" ]; then
        if [ ! -f "${CONFIG_DIR}/config.json" ]; then
            print_info "Installing example configuration..."
            cp "${temp_dir}/examples/config.json" "${CONFIG_DIR}/config.json"
            print_info "Configuration file created at: ${CONFIG_DIR}/config.json"
        else
            print_warning "Configuration file already exists at ${CONFIG_DIR}/config.json"
            print_warning "Example config is available in the archive if you need it."
        fi
    fi
    
    # Install completions if available
    if [ -d "${temp_dir}/completions" ]; then
        install_completions "$temp_dir" "$SUDO"
    fi
    
    # Cleanup
    rm -rf "$temp_dir"
    
    print_info "Installation complete!"
}

# Install shell completions
install_completions() {
    local temp_dir="$1"
    local sudo_cmd="$2"
    
    # Bash completion
    if [ -d "/usr/local/etc/bash_completion.d" ]; then
        print_info "Installing bash completions..."
        $sudo_cmd cp "${temp_dir}/completions/dnsres.bash" "/usr/local/etc/bash_completion.d/dnsres"
    elif [ -d "/etc/bash_completion.d" ]; then
        print_info "Installing bash completions..."
        $sudo_cmd cp "${temp_dir}/completions/dnsres.bash" "/etc/bash_completion.d/dnsres"
    fi
    
    # Zsh completion
    if [ -d "/usr/local/share/zsh/site-functions" ]; then
        print_info "Installing zsh completions..."
        $sudo_cmd cp "${temp_dir}/completions/dnsres.zsh" "/usr/local/share/zsh/site-functions/_dnsres"
    fi
    
    # Fish completion
    if [ -d "$HOME/.config/fish/completions" ]; then
        print_info "Installing fish completions..."
        cp "${temp_dir}/completions/dnsres.fish" "$HOME/.config/fish/completions/dnsres.fish"
    fi
}

# Main installation flow
main() {
    echo ""
    print_info "dnsres installation script"
    echo ""
    
    # Check for required commands
    for cmd in curl tar; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            print_error "Required command not found: $cmd"
            print_error "Please install $cmd and try again."
            exit 1
        fi
    done
    
    # Detect platform
    platform=$(detect_platform)
    print_info "Detected platform: ${platform}"
    
    # Get latest version
    version=$(get_latest_version)
    print_info "Latest version: ${version}"
    
    # Install
    install_dnsres "$platform" "$version"
    
    # Print success message
    echo ""
    print_info "dnsres has been successfully installed!"
    echo ""
    echo "Two binaries are available:"
    echo "  • dnsres      - CLI for continuous monitoring"
    echo "  • dnsres-tui  - Interactive terminal UI"
    echo ""
    echo "Quick start:"
    echo "  dnsres example.com          # CLI mode"
    echo "  dnsres-tui example.com      # Interactive TUI mode"
    echo ""
    echo "Configuration: ${CONFIG_DIR}/config.json"
    echo ""
    echo "Documentation: https://github.com/${REPO}"
    echo "Report issues: https://github.com/${REPO}/issues"
    echo ""
}

main "$@"
