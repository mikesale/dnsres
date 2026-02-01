# Installation Guide for dnsres

This guide provides comprehensive installation instructions for dnsres across different platforms and package managers.

## Table of Contents

- [Quick Install](#quick-install)
- [Homebrew (macOS/Linux)](#homebrew-macoslinux)
- [Debian/Ubuntu (APT)](#debianubuntu-apt)
- [RHEL/Fedora/CentOS (RPM)](#rhelfedoracentos-rpm)
- [Snap (Universal Linux)](#snap-universal-linux)
- [Arch Linux (AUR)](#arch-linux-aur)
- [Direct Download Script](#direct-download-script)
- [Manual Installation](#manual-installation)
- [Windows Installation](#windows-installation)
- [Building from Source](#building-from-source)
- [Shell Completions](#shell-completions)
- [Configuration](#configuration)
- [Upgrading](#upgrading)
- [Troubleshooting](#troubleshooting)

---

## Quick Install

**macOS/Linux (using install script):**
```bash
curl -sSL https://raw.githubusercontent.com/mikesale/dnsres/main/install.sh | bash
```

**macOS/Linux (using Homebrew):**
```bash
brew tap mikesale/dnsres
brew install dnsres
```

---

## Homebrew (macOS/Linux)

Homebrew is the recommended installation method for macOS and Linux users.

### Installation

1. **Add the tap:**
   ```bash
   brew tap mikesale/dnsres
   ```

2. **Install dnsres:**
   ```bash
   brew install dnsres
   ```

### What Gets Installed

- **Binaries:** `/usr/local/bin/dnsres` and `/usr/local/bin/dnsres-tui`
- **Example Config:** `/usr/local/etc/dnsres/config.json.example`
- **Shell Completions:** Automatically installed for bash, zsh, and fish

### Usage

```bash
# CLI mode
dnsres example.com

# Interactive TUI mode
dnsres-tui example.com

# With custom config
dnsres -config ~/.config/dnsres/config.json
```

### Upgrading

```bash
brew update
brew upgrade dnsres
```

### Uninstalling

```bash
brew uninstall dnsres
brew untap mikesale/dnsres
```

---

## Debian/Ubuntu (APT)

For Debian-based distributions (Debian, Ubuntu, Linux Mint, Pop!_OS, etc.).

### Installation

1. **Download the .deb package:**
   ```bash
   VERSION=1.1.8  # Replace with latest version
   wget https://github.com/mikesale/dnsres/releases/download/v${VERSION}/dnsres_${VERSION}_Linux_x86_64.deb
   ```

2. **Install the package:**
   ```bash
   sudo dpkg -i dnsres_${VERSION}_Linux_x86_64.deb
   ```

### What Gets Installed

- **Binaries:** `/usr/local/bin/dnsres` and `/usr/local/bin/dnsres-tui`
- **Example Config:** `/etc/dnsres/config.json.example`
- **Documentation:** `/usr/share/doc/dnsres/`
- **Shell Completions:** `/usr/share/bash-completion/`, `/usr/share/zsh/`, `/usr/share/fish/`

### Upgrading

Download and install the newer .deb package:
```bash
VERSION=1.2.0  # New version
wget https://github.com/mikesale/dnsres/releases/download/v${VERSION}/dnsres_${VERSION}_Linux_x86_64.deb
sudo dpkg -i dnsres_${VERSION}_Linux_x86_64.deb
```

### Uninstalling

```bash
sudo dpkg -r dnsres
```

---

## RHEL/Fedora/CentOS (RPM)

For Red Hat-based distributions (RHEL, Fedora, CentOS, Rocky Linux, AlmaLinux, etc.).

### Installation

1. **Download the .rpm package:**
   ```bash
   VERSION=1.1.8  # Replace with latest version
   wget https://github.com/mikesale/dnsres/releases/download/v${VERSION}/dnsres_${VERSION}_Linux_x86_64.rpm
   ```

2. **Install the package:**
   ```bash
   sudo rpm -i dnsres_${VERSION}_Linux_x86_64.rpm
   ```
   
   Or using dnf/yum:
   ```bash
   sudo dnf install dnsres_${VERSION}_Linux_x86_64.rpm
   ```

### What Gets Installed

- **Binaries:** `/usr/local/bin/dnsres` and `/usr/local/bin/dnsres-tui`
- **Example Config:** `/etc/dnsres/config.json.example`
- **Documentation:** `/usr/share/doc/dnsres/`
- **Shell Completions:** `/usr/share/bash-completion/`, `/usr/share/zsh/`, `/usr/share/fish/`

### Upgrading

```bash
VERSION=1.2.0  # New version
wget https://github.com/mikesale/dnsres/releases/download/v${VERSION}/dnsres_${VERSION}_Linux_x86_64.rpm
sudo rpm -U dnsres_${VERSION}_Linux_x86_64.rpm
```

### Uninstalling

```bash
sudo rpm -e dnsres
```

---

## Snap (Universal Linux)

Snap packages work across most Linux distributions.

### Installation

```bash
sudo snap install dnsres --classic
```

**Note:** The `--classic` flag is required for network access and file system permissions.

### Usage

```bash
dnsres example.com
dnsres-tui example.com
```

### Upgrading

Snaps update automatically, or manually:
```bash
sudo snap refresh dnsres
```

### Uninstalling

```bash
sudo snap remove dnsres
```

---

## Arch Linux (AUR)

For Arch Linux and derivatives (Manjaro, EndeavourOS, etc.).

### Installation

Using an AUR helper like `yay`:

```bash
yay -S dnsres-bin
```

Or manually:

```bash
git clone https://aur.archlinux.org/dnsres-bin.git
cd dnsres-bin
makepkg -si
```

### Upgrading

```bash
yay -Syu dnsres-bin
```

### Uninstalling

```bash
yay -R dnsres-bin
```

---

## Direct Download Script

The install script automatically detects your platform and installs the latest version.

### Installation

```bash
curl -sSL https://raw.githubusercontent.com/mikesale/dnsres/main/install.sh | bash
```

### What It Does

1. Detects your OS (macOS/Linux) and architecture (x86_64/arm64)
2. Downloads the latest release from GitHub
3. Installs binaries to `/usr/local/bin` (may require sudo)
4. Creates config directory at `~/.config/dnsres/`
5. Installs example configuration
6. Installs shell completions (if directories exist)

### Manual Script Download

If you prefer to inspect the script first:

```bash
curl -sSL https://raw.githubusercontent.com/mikesale/dnsres/main/install.sh -o install.sh
chmod +x install.sh
./install.sh
```

---

## Manual Installation

For manual installation or custom setups.

### Steps

1. **Download the archive for your platform:**
   
   Visit the [releases page](https://github.com/mikesale/dnsres/releases) and download:
   - macOS (Intel): `dnsres_*_Darwin_x86_64.tar.gz`
   - macOS (Apple Silicon): `dnsres_*_Darwin_arm64.tar.gz`
   - Linux (x86_64): `dnsres_*_Linux_x86_64.tar.gz`
   - Linux (ARM64): `dnsres_*_Linux_arm64.tar.gz`

2. **Extract the archive:**
   ```bash
   tar -xzf dnsres_*_*.tar.gz
   ```

3. **Move binaries to your PATH:**
   ```bash
   sudo mv dnsres dnsres-tui /usr/local/bin/
   sudo chmod +x /usr/local/bin/dnsres /usr/local/bin/dnsres-tui
   ```

4. **Create config directory:**
   ```bash
   mkdir -p ~/.config/dnsres
   ```

5. **Copy example config (optional):**
   ```bash
   cp examples/config.json ~/.config/dnsres/config.json
   ```

6. **Install completions (optional):**
   ```bash
   # Bash
   sudo cp completions/dnsres.bash /usr/local/etc/bash_completion.d/dnsres
   
   # Zsh
   sudo cp completions/dnsres.zsh /usr/local/share/zsh/site-functions/_dnsres
   
   # Fish
   cp completions/dnsres.fish ~/.config/fish/completions/dnsres.fish
   ```

---

## Windows Installation

Windows support is currently manual installation only.

### Installation

1. **Download the Windows archive:**
   
   Visit the [releases page](https://github.com/mikesale/dnsres/releases) and download:
   - `dnsres_*_Windows_x86_64.zip`

2. **Extract the archive:**
   
   Right-click and "Extract All" or use PowerShell:
   ```powershell
   Expand-Archive -Path dnsres_*_Windows_x86_64.zip -DestinationPath C:\dnsres
   ```

3. **Add to PATH (optional):**
   
   Add `C:\dnsres` to your system PATH environment variable.

4. **Create config directory:**
   ```powershell
   mkdir $env:USERPROFILE\.config\dnsres
   ```

5. **Run the binaries:**
   ```powershell
   .\dnsres.exe example.com
   .\dnsres-tui.exe example.com
   ```

### Future Windows Support

We plan to add Winget and Chocolatey support in future releases.

---

## Building from Source

For developers or users who want to build from source.

### Prerequisites

- Go 1.24 or later
- Git

### Steps

```bash
# Clone the repository
git clone https://github.com/mikesale/dnsres.git
cd dnsres

# Build CLI
make build

# Build TUI
make build-tui

# Build both for all platforms
make build-all

# Install locally
sudo cp dnsres dnsres-tui /usr/local/bin/
```

### Development Build

```bash
# Run without installing
go run ./cmd/dnsres example.com
go run ./cmd/dnsres-tui example.com

# Run tests
make test

# Run integration tests
go test -tags=integration ./internal/integration -v
```

---

## Shell Completions

Shell completions enable tab-completion for flags and arguments.

### Bash

Completions are automatically installed with package managers. To load manually:

```bash
source /usr/local/etc/bash_completion.d/dnsres
```

Add to `~/.bashrc` for persistence:
```bash
echo 'source /usr/local/etc/bash_completion.d/dnsres' >> ~/.bashrc
```

### Zsh

Completions are automatically installed. If not working, ensure the completion directory is in your `fpath`:

```zsh
# Add to ~/.zshrc
fpath=(/usr/local/share/zsh/site-functions $fpath)
autoload -U compinit && compinit
```

### Fish

Completions are automatically loaded from `~/.config/fish/completions/`.

### Testing Completions

```bash
dnsres -<TAB>        # Should show: -config, -host, -report, -help, -version
dnsres -config <TAB> # Should complete JSON files
```

---

## Configuration

### Default Config Locations

dnsres follows the XDG Base Directory Specification:

1. Explicit `-config` flag (if provided)
2. `./config.json` (current directory)
3. `~/.config/dnsres/config.json` (primary location)
4. Built-in defaults

### Creating Your First Config

1. **Copy the example:**
   ```bash
   mkdir -p ~/.config/dnsres
   cp /usr/local/etc/dnsres/config.json.example ~/.config/dnsres/config.json
   ```
   
   Or from package installations:
   ```bash
   cp /etc/dnsres/config.json.example ~/.config/dnsres/config.json
   ```

2. **Edit the configuration:**
   ```bash
   nano ~/.config/dnsres/config.json
   ```

3. **Key settings to configure:**
   - `hostnames`: List of domains to monitor
   - `dns_servers`: DNS servers to query (e.g., `["8.8.8.8:53", "1.1.1.1:53"]`)
   - `query_interval`: How often to check (e.g., `"30s"`, `"1m"`)
   - `log_dir`: Where to store logs (default: `~/.local/state/dnsres/`)

### Example Minimal Config

```json
{
  "hostnames": ["example.com"],
  "dns_servers": ["8.8.8.8:53", "1.1.1.1:53"],
  "query_timeout": "5s",
  "query_interval": "1m"
}
```

For full configuration options, see the [README.md](README.md#configuration).

---

## Upgrading

### Homebrew

```bash
brew update
brew upgrade dnsres
```

### APT (Debian/Ubuntu)

Download and install the new .deb package:
```bash
VERSION=1.2.0
wget https://github.com/mikesale/dnsres/releases/download/v${VERSION}/dnsres_${VERSION}_Linux_x86_64.deb
sudo dpkg -i dnsres_${VERSION}_Linux_x86_64.deb
```

### RPM (RHEL/Fedora)

```bash
VERSION=1.2.0
wget https://github.com/mikesale/dnsres/releases/download/v${VERSION}/dnsres_${VERSION}_Linux_x86_64.rpm
sudo rpm -U dnsres_${VERSION}_Linux_x86_64.rpm
```

### Snap

```bash
sudo snap refresh dnsres
```

### Install Script

Re-run the install script:
```bash
curl -sSL https://raw.githubusercontent.com/mikesale/dnsres/main/install.sh | bash
```

### Manual

Download the new release and replace the binaries in `/usr/local/bin/`.

---

## Troubleshooting

### Command Not Found

**Problem:** `dnsres: command not found`

**Solutions:**
1. Verify installation: `which dnsres`
2. Check PATH: `echo $PATH` (should include `/usr/local/bin`)
3. Reload shell: `source ~/.bashrc` or `source ~/.zshrc`
4. Reinstall: Follow installation steps again

### Permission Denied

**Problem:** Permission errors when running dnsres

**Solutions:**
1. Check execute permissions: `ls -l $(which dnsres)`
2. Fix permissions: `sudo chmod +x /usr/local/bin/dnsres /usr/local/bin/dnsres-tui`

### Port Already in Use

**Problem:** `bind: address already in use` for health or metrics ports

**Solutions:**
1. Check what's using the port: `lsof -i :8880` or `lsof -i :9990`
2. Change ports in config:
   ```json
   {
     "health_port": 8881,
     "metrics_port": 9991
   }
   ```

### Config File Not Found

**Problem:** dnsres can't find configuration

**Solutions:**
1. Verify config location: `ls ~/.config/dnsres/config.json`
2. Create config directory: `mkdir -p ~/.config/dnsres`
3. Copy example: See [Configuration](#configuration) section
4. Use explicit flag: `dnsres -config /path/to/config.json`

### DNS Resolution Failures

**Problem:** All DNS queries failing

**Solutions:**
1. Check network connectivity: `ping 8.8.8.8`
2. Verify DNS servers are accessible: `dig @8.8.8.8 example.com`
3. Check firewall rules (UDP port 53)
4. Try different DNS servers in config

### Completions Not Working

**Problem:** Tab completion not working

**Solutions:**

**Bash:**
```bash
# Check if completion file exists
ls /usr/local/etc/bash_completion.d/dnsres

# Source it manually
source /usr/local/etc/bash_completion.d/dnsres

# Add to ~/.bashrc
echo 'source /usr/local/etc/bash_completion.d/dnsres' >> ~/.bashrc
```

**Zsh:**
```zsh
# Check fpath
echo $fpath

# Rebuild completion cache
rm ~/.zcompdump
autoload -U compinit && compinit
```

**Fish:**
```fish
# Check completions directory
ls ~/.config/fish/completions/dnsres.fish

# Reload completions
fish_update_completions
```

### macOS Gatekeeper Warning

**Problem:** "dnsres cannot be opened because it is from an unidentified developer"

**Solutions:**
1. Use Homebrew installation (recommended)
2. Or bypass Gatekeeper: `xattr -d com.apple.quarantine $(which dnsres)`

### Snap Confinement Issues

**Problem:** Snap version can't access network or files

**Solutions:**
1. Ensure classic confinement: `sudo snap install dnsres --classic`
2. Check snap connections: `snap connections dnsres`

---

## Getting Help

If you encounter issues not covered here:

- **Documentation:** [README.md](README.md)
- **GitHub Issues:** [https://github.com/mikesale/dnsres/issues](https://github.com/mikesale/dnsres/issues)
- **Email:** mike.sale@gmail.com

When reporting issues, please include:
- Operating system and version
- Installation method
- dnsres version (`dnsres -version`)
- Error messages or logs
- Configuration file (redact sensitive data)

---

## Next Steps

After installation:

1. **Create a configuration file** (see [Configuration](#configuration))
2. **Run your first monitoring session:**
   ```bash
   dnsres example.com
   ```
3. **Try the interactive TUI:**
   ```bash
   dnsres-tui example.com
   ```
4. **Explore advanced features** in the [README.md](README.md)

---

**Installation Guide Version:** 1.0  
**Last Updated:** 2026-02-01  
**For dnsres version:** 1.2.0+
